package replay

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/gc"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/tracked"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/vision"
	"google.golang.org/protobuf/proto"
)

const checkpointInterval = 500

type event struct {
	offset    int64
	timestamp int64
	msgType   persistence.MessageId
}

type snapshot struct {
	detections map[int]*vision.SSL_DetectionFrame
	geometry   *vision.SSL_GeometryData
	tracked    map[string]*tracked.TrackerWrapperPacket
	referee    *gc.Referee
}

type checkpoint struct {
	eventIndex int
	snapshot   *snapshot
}

type Engine struct {
	mu sync.Mutex

	reader            *persistence.Reader
	events            []event
	displayEventIdx   []int
	displayTimestamps []int64
	annotations       []Annotation
	cachedAnnotations []Annotation
	checkpoints       []checkpoint

	fileName string
	errorMsg string

	tmpDir      string
	workingPath string
	uploadPath  string

	loaded          bool
	playing         bool
	rate            float64
	positionNs      int64
	durationNs      int64
	baseTimestampNs int64

	currentEventIndex int
	currentSnapshot   *snapshot

	stopPlayback chan struct{}
}

func NewEngine() *Engine {
	e := &Engine{
		tmpDir:       filepath.Join(os.TempDir(), "ssl-vision-client", "replay"),
		rate:         1.0,
		stopPlayback: make(chan struct{}),
	}
	e.currentSnapshot = emptySnapshot()
	e.currentEventIndex = -1
	go e.runPlaybackLoop()
	return e
}

func (e *Engine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	close(e.stopPlayback)
	e.cleanupLocked()
}

func (e *Engine) LoadFromReader(reader io.Reader, fileName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cleanupLocked()
	e.errorMsg = ""
	e.fileName = fileName

	if err := os.MkdirAll(e.tmpDir, 0o755); err != nil {
		e.errorMsg = err.Error()
		return err
	}

	tmpBase := filepath.Join(e.tmpDir, fmt.Sprintf("%d", time.Now().UnixNano()))
	uploadPath := tmpBase + "_upload"
	f, err := os.Create(uploadPath)
	if err != nil {
		e.errorMsg = err.Error()
		return err
	}
	if _, err := io.Copy(f, reader); err != nil {
		_ = f.Close()
		e.errorMsg = err.Error()
		return err
	}
	if err := f.Close(); err != nil {
		e.errorMsg = err.Error()
		return err
	}
	e.uploadPath = uploadPath

	logPath := uploadPath
	if filepath.Ext(fileName) == ".gz" {
		logPath = tmpBase + ".log"
		if err := gunzipFile(uploadPath, logPath); err != nil {
			e.errorMsg = err.Error()
			return err
		}
	}
	e.workingPath = logPath

	return e.finalizeLocked(logPath)
}

func (e *Engine) LoadFromPath(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.cleanupLocked()
	e.errorMsg = ""
	e.fileName = filepath.Base(path)

	logPath := path
	if filepath.Ext(path) == ".gz" {
		if err := os.MkdirAll(e.tmpDir, 0o755); err != nil {
			e.errorMsg = err.Error()
			return err
		}
		tmpLog := filepath.Join(e.tmpDir, fmt.Sprintf("%d.log", time.Now().UnixNano()))
		if err := gunzipFile(path, tmpLog); err != nil {
			e.errorMsg = err.Error()
			return err
		}
		logPath = tmpLog
		e.workingPath = tmpLog
	}

	return e.finalizeLocked(logPath)
}

func (e *Engine) finalizeLocked(logPath string) error {
	pr, err := persistence.NewReader(logPath)
	if err != nil {
		e.errorMsg = err.Error()
		return err
	}
	e.reader = pr

	if err := e.buildIndexLocked(); err != nil {
		e.errorMsg = err.Error()
		return err
	}

	e.loaded = len(e.displayEventIdx) > 0
	if e.loaded {
		e.baseTimestampNs = e.displayTimestamps[0]
		e.durationNs = e.displayTimestamps[len(e.displayTimestamps)-1] - e.baseTimestampNs
		e.cachedAnnotations = normalizeAnnotations(e.annotations, e.baseTimestampNs)
		if err := e.setDisplayIndexLocked(0); err != nil {
			e.errorMsg = err.Error()
			return err
		}
	} else {
		e.baseTimestampNs = 0
		e.durationNs = 0
		e.cachedAnnotations = nil
	}

	return nil
}

func (e *Engine) State(mode Mode) State {
	e.mu.Lock()
	defer e.mu.Unlock()
	annotations := e.cachedAnnotations
	if annotations == nil {
		annotations = []Annotation{}
	}
	return State{
		Mode:        mode,
		Loaded:      e.loaded,
		Playing:     e.playing,
		Rate:        e.rate,
		PositionNs:  e.positionNs,
		DurationNs:  e.durationNs,
		FileName:    e.fileName,
		Error:       e.errorMsg,
		Annotations: annotations,
	}
}

func (e *Engine) Control(req ControlRequest) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch req.Action {
	case ActionPlay:
		if e.loaded {
			e.playing = true
		}
	case ActionPause:
		e.playing = false
	case ActionSeek:
		if !e.loaded {
			return nil
		}
		return e.seekPositionLocked(req.PositionNs)
	case ActionStep:
		if !e.loaded {
			return nil
		}
		return e.stepLocked(req.Delta)
	case ActionSetRate:
		if req.Rate <= 0 {
			return nil
		}
		e.rate = req.Rate
	default:
		return fmt.Errorf("unknown action: %s", req.Action)
	}

	return nil
}

func (e *Engine) CurrentDetection() *vision.SSL_DetectionFrame {
	e.mu.Lock()
	defer e.mu.Unlock()
	return combineDetections(e.currentSnapshot.detections)
}

func (e *Engine) CurrentTracked() map[string]*tracked.TrackerWrapperPacket {
	e.mu.Lock()
	defer e.mu.Unlock()
	frames := map[string]*tracked.TrackerWrapperPacket{}
	for k, v := range e.currentSnapshot.tracked {
		frames[k] = cloneTrackerWrapper(v)
	}
	return frames
}

func (e *Engine) CurrentGeometry() *vision.SSL_GeometryData {
	e.mu.Lock()
	defer e.mu.Unlock()
	return cloneGeometry(e.currentSnapshot.geometry)
}

func (e *Engine) CurrentReferee() *gc.Referee {
	e.mu.Lock()
	defer e.mu.Unlock()
	return cloneReferee(e.currentSnapshot.referee)
}

func (e *Engine) buildIndexLocked() error {
	working := emptySnapshot()
	lastCommand := gc.Referee_Command(-1)
	hasLastCommand := false

	offset := int64(persistence.HeaderSize)
	for e.reader.HasMessage() {
		msg, err := e.reader.ReadMessage()
		if err != nil {
			return err
		}
		e.events = append(e.events, event{offset: offset, timestamp: msg.Timestamp, msgType: msg.MessageType.Id})
		eventIndex := len(e.events) - 1

		if isDisplayMessageType(msg.MessageType.Id) {
			e.displayEventIdx = append(e.displayEventIdx, eventIndex)
			e.displayTimestamps = append(e.displayTimestamps, msg.Timestamp)
		}

		if msg.MessageType.Id == persistence.MessageSslRefbox2013 {
			ref := new(gc.Referee)
			if err := proto.Unmarshal(msg.Message, ref); err == nil {
				cmd := ref.GetCommand()
				if !hasLastCommand || cmd != lastCommand {
					e.annotations = append(e.annotations, Annotation{TimestampNs: msg.Timestamp, Label: cmd.String()})
					lastCommand = cmd
					hasLastCommand = true
				}
			}
		}

		if err := applyMessageToSnapshot(working, msg); err != nil {
			return err
		}
		if eventIndex%checkpointInterval == 0 {
			e.checkpoints = append(e.checkpoints, checkpoint{eventIndex: eventIndex, snapshot: cloneSnapshot(working)})
		}

		offset += int64(16 + len(msg.Message))
	}

	if len(e.events) > 0 {
		e.checkpoints = append(e.checkpoints, checkpoint{eventIndex: len(e.events) - 1, snapshot: cloneSnapshot(working)})
	}

	return nil
}

func (e *Engine) runPlaybackLoop() {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	lastTick := time.Now()
	for {
		select {
		case <-e.stopPlayback:
			return
		case <-ticker.C:
			now := time.Now()
			delta := now.Sub(lastTick)
			lastTick = now

			e.mu.Lock()
			if e.playing && e.loaded {
				advance := int64(float64(delta.Nanoseconds()) * e.rate)
				nextPos := e.positionNs + advance
				if nextPos >= e.durationNs {
					nextPos = e.durationNs
					e.playing = false
				}
				_ = e.seekPositionLocked(nextPos)
			}
			e.mu.Unlock()
		}
	}
}

func (e *Engine) seekPositionLocked(positionNs int64) error {
	if positionNs < 0 {
		positionNs = 0
	}
	if positionNs > e.durationNs {
		positionNs = e.durationNs
	}
	if len(e.displayTimestamps) == 0 {
		e.positionNs = 0
		return nil
	}

	targetTs := e.baseTimestampNs + positionNs
	idx := sort.Search(len(e.displayTimestamps), func(i int) bool {
		return e.displayTimestamps[i] > targetTs
	}) - 1
	if idx < 0 {
		idx = 0
	}
	if err := e.setDisplayIndexLocked(idx); err != nil {
		return err
	}
	e.positionNs = e.displayTimestamps[idx] - e.baseTimestampNs
	return nil
}

func (e *Engine) stepLocked(delta int) error {
	if len(e.displayEventIdx) == 0 {
		return nil
	}
	if delta == 0 {
		delta = 1
	}

	current := 0
	if e.currentEventIndex >= 0 {
		current = e.displayIndexFromEventIndexLocked(e.currentEventIndex)
	}
	next := current + delta
	if next < 0 {
		next = 0
	}
	if next >= len(e.displayEventIdx) {
		next = len(e.displayEventIdx) - 1
	}
	if err := e.setDisplayIndexLocked(next); err != nil {
		return err
	}
	e.positionNs = e.displayTimestamps[next] - e.baseTimestampNs
	return nil
}

func (e *Engine) displayIndexFromEventIndexLocked(eventIdx int) int {
	pos := sort.SearchInts(e.displayEventIdx, eventIdx)
	if pos < len(e.displayEventIdx) && e.displayEventIdx[pos] == eventIdx {
		return pos
	}
	if pos <= 0 {
		return 0
	}
	return pos - 1
}

func (e *Engine) setDisplayIndexLocked(displayIdx int) error {
	if displayIdx < 0 || displayIdx >= len(e.displayEventIdx) {
		return nil
	}
	targetEvent := e.displayEventIdx[displayIdx]
	if targetEvent == e.currentEventIndex {
		return nil
	}

	if e.currentEventIndex >= 0 && targetEvent == e.currentEventIndex+1 {
		msg, err := e.reader.ReadMessageAt(e.events[targetEvent].offset)
		if err != nil {
			return err
		}
		if err := applyMessageToSnapshot(e.currentSnapshot, msg); err != nil {
			return err
		}
		e.currentEventIndex = targetEvent
		return nil
	}

	cpIndex := sort.Search(len(e.checkpoints), func(i int) bool {
		return e.checkpoints[i].eventIndex > targetEvent
	}) - 1

	startEvent := 0
	if cpIndex >= 0 {
		e.currentSnapshot = cloneSnapshot(e.checkpoints[cpIndex].snapshot)
		e.currentEventIndex = e.checkpoints[cpIndex].eventIndex
		startEvent = e.currentEventIndex + 1
	} else {
		e.currentSnapshot = emptySnapshot()
		e.currentEventIndex = -1
		startEvent = 0
	}

	for i := startEvent; i <= targetEvent; i++ {
		msg, err := e.reader.ReadMessageAt(e.events[i].offset)
		if err != nil {
			return err
		}
		if err := applyMessageToSnapshot(e.currentSnapshot, msg); err != nil {
			return err
		}
		e.currentEventIndex = i
	}

	return nil
}

func (e *Engine) cleanupLocked() {
	e.loaded = false
	e.playing = false
	e.events = nil
	e.displayEventIdx = nil
	e.displayTimestamps = nil
	e.annotations = nil
	e.cachedAnnotations = nil
	e.checkpoints = nil
	e.currentEventIndex = -1
	e.positionNs = 0
	e.durationNs = 0
	e.baseTimestampNs = 0
	e.currentSnapshot = emptySnapshot()

	if e.reader != nil {
		_ = e.reader.Close()
		e.reader = nil
	}
	if e.workingPath != "" {
		_ = os.Remove(e.workingPath)
		e.workingPath = ""
	}
	if e.uploadPath != "" {
		_ = os.Remove(e.uploadPath)
		e.uploadPath = ""
	}
}

func gunzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gz.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, gz)
	return err
}

func isDisplayMessageType(id persistence.MessageId) bool {
	switch id {
	case persistence.MessageSslVision2014:
		return true
	case persistence.MessageSslVisionTracker2020:
		return true
	case persistence.MessageSslRefbox2013:
		return true
	default:
		return false
	}
}

func applyMessageToSnapshot(s *snapshot, msg *persistence.Message) error {
	switch msg.MessageType.Id {
	case persistence.MessageSslVision2014:
		wrapper := new(vision.SSL_WrapperPacket)
		if err := proto.Unmarshal(msg.Message, wrapper); err != nil {
			return nil
		}
		if wrapper.Detection != nil {
			camID := int(wrapper.Detection.GetCameraId())
			s.detections[camID] = cloneDetection(wrapper.Detection)
		}
		if wrapper.Geometry != nil {
			s.geometry = cloneGeometry(wrapper.Geometry)
		}
	case persistence.MessageSslVisionTracker2020:
		packet := new(tracked.TrackerWrapperPacket)
		if err := proto.Unmarshal(msg.Message, packet); err != nil {
			return nil
		}
		uuid := packet.GetUuid()
		if uuid != "" && packet.TrackedFrame != nil {
			s.tracked[uuid] = cloneTrackerWrapper(packet)
		}
	case persistence.MessageSslRefbox2013:
		ref := new(gc.Referee)
		if err := proto.Unmarshal(msg.Message, ref); err != nil {
			return nil
		}
		s.referee = cloneReferee(ref)
	}
	return nil
}

func emptySnapshot() *snapshot {
	return &snapshot{
		detections: map[int]*vision.SSL_DetectionFrame{},
		tracked:    map[string]*tracked.TrackerWrapperPacket{},
	}
}

func cloneSnapshot(s *snapshot) *snapshot {
	out := emptySnapshot()
	for k, v := range s.detections {
		out.detections[k] = cloneDetection(v)
	}
	for k, v := range s.tracked {
		out.tracked[k] = cloneTrackerWrapper(v)
	}
	out.geometry = cloneGeometry(s.geometry)
	out.referee = cloneReferee(s.referee)
	return out
}

func cloneDetection(in *vision.SSL_DetectionFrame) *vision.SSL_DetectionFrame {
	if in == nil {
		return nil
	}
	out, ok := proto.Clone(in).(*vision.SSL_DetectionFrame)
	if !ok {
		return nil
	}
	return out
}

func cloneGeometry(in *vision.SSL_GeometryData) *vision.SSL_GeometryData {
	if in == nil {
		return nil
	}
	out, ok := proto.Clone(in).(*vision.SSL_GeometryData)
	if !ok {
		return nil
	}
	return out
}

func cloneTrackerWrapper(in *tracked.TrackerWrapperPacket) *tracked.TrackerWrapperPacket {
	if in == nil {
		return nil
	}
	out, ok := proto.Clone(in).(*tracked.TrackerWrapperPacket)
	if !ok {
		return nil
	}
	return out
}

func cloneReferee(in *gc.Referee) *gc.Referee {
	if in == nil {
		return nil
	}
	out, ok := proto.Clone(in).(*gc.Referee)
	if !ok {
		return nil
	}
	return out
}

func normalizeAnnotations(annotations []Annotation, baseTimestampNs int64) []Annotation {
	out := make([]Annotation, len(annotations))
	for i, a := range annotations {
		out[i] = Annotation{TimestampNs: a.TimestampNs - baseTimestampNs, Label: a.Label}
	}
	return out
}

func combineDetections(detections map[int]*vision.SSL_DetectionFrame) *vision.SSL_DetectionFrame {
	f := &vision.SSL_DetectionFrame{
		FrameNumber:  new(uint32),
		CameraId:     new(uint32),
		TCapture:     new(float64),
		TSent:        new(float64),
		Balls:        make([]*vision.SSL_DetectionBall, 0),
		RobotsYellow: make([]*vision.SSL_DetectionRobot, 0),
		RobotsBlue:   make([]*vision.SSL_DetectionRobot, 0),
	}
	for _, b := range detections {
		if b == nil {
			continue
		}
		if b.FrameNumber != nil && *b.FrameNumber > *f.FrameNumber {
			*f.FrameNumber = *b.FrameNumber
		}
		if b.TCapture != nil && *b.TCapture > *f.TCapture {
			*f.TCapture = *b.TCapture
		}
		if b.TSent != nil && *b.TSent > *f.TSent {
			*f.TSent = *b.TSent
		}
		f.Balls = append(f.Balls, b.Balls...)
		f.RobotsYellow = append(f.RobotsYellow, b.RobotsYellow...)
		f.RobotsBlue = append(f.RobotsBlue, b.RobotsBlue...)
	}
	return f
}

package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/config"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/gc"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/grsim"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/replay"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/tracked"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/vision"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

var address = flag.String("address", ":8082", "The address on which the UI and API is served, default: :8082")
var skipInterfaces = flag.String("skipInterfaces", "", "Comma separated list of interface names to ignore when receiving multicast packets")
var verbose = flag.Bool("verbose", false, "Verbose output")

// ReceiverManager レシーバーの管理
type ReceiverManager struct {
	visionReceiver  *vision.Receiver
	trackedReceiver *tracked.Receiver
	refereeReceiver *gc.Receiver
	replayEngine    *replay.Engine
	mode            replay.Mode
	grSimSender     *grsim.Sender
	skipIfis        []string
	verbose         bool
	visionAddr      string
	trackedAddr     string
	refereeAddr     string
	mu              sync.Mutex
}

func main() {
	flag.Parse()

	// 設定を読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("UI is available at %v", formattedAddress())

	httpServer := setupServer(cfg)

	err = httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func setupServer(cfg *config.Config) *http.Server {
	skipIfis := parseSkipInterfaces()

	// grSim Senderを初期化
	grSimSender := grsim.NewSender(cfg.GetGrSimAddress())
	if err := grSimSender.Connect(); err != nil {
		log.Printf("Warning: Failed to connect to grSim: %v", err)
	}

	// レシーバーマネージャーを初期化
	manager := &ReceiverManager{
		skipIfis:     skipIfis,
		verbose:      *verbose,
		grSimSender:  grSimSender,
		replayEngine: replay.NewEngine(),
		mode:         replay.ModeLive,
	}

	// 初回起動
	visionAddr, trackedAddr, refereeAddr := cfg.GetAddresses()
	manager.startReceivers(visionAddr, trackedAddr, refereeAddr)

	srv := NewServer(
		manager.GetDetectionFrames,
		manager.GetTrackedFrames,
		manager.GetGeometry,
		manager.GetRefereeMsg,
		cfg,
		manager,
		manager,
		grSimSender,
	)
	return &http.Server{
		Addr:    *address,
		Handler: srv,
	}
}

// startReceivers レシーバーを起動
func (rm *ReceiverManager) startReceivers(visionAddr, trackedAddr, refereeAddr string) {
	rm.visionAddr = visionAddr
	rm.trackedAddr = trackedAddr
	rm.refereeAddr = refereeAddr

	rm.visionReceiver = vision.NewReceiver(visionAddr)
	rm.trackedReceiver = tracked.NewReceiver(trackedAddr)
	rm.refereeReceiver = gc.NewReceiver(refereeAddr)

	rm.visionReceiver.MulticastServer.SkipInterfaces = rm.skipIfis
	rm.visionReceiver.MulticastServer.Verbose = rm.verbose
	rm.trackedReceiver.MulticastServer.SkipInterfaces = rm.skipIfis
	rm.trackedReceiver.MulticastServer.Verbose = rm.verbose
	rm.refereeReceiver.MulticastServer.SkipInterfaces = rm.skipIfis
	rm.refereeReceiver.MulticastServer.Verbose = rm.verbose

	rm.visionReceiver.Start()
	rm.trackedReceiver.Start()
	rm.refereeReceiver.Start()

	log.Printf("Receivers started: vision=%s, tracked=%s, referee=%s", visionAddr, trackedAddr, refereeAddr)
}

// GetDetectionFrames 現在のビジョンレシーバーから検出フレームを取得
func (rm *ReceiverManager) GetDetectionFrames() *vision.SSL_DetectionFrame {
	rm.mu.Lock()
	mode := rm.mode
	receiver := rm.visionReceiver
	replayEngine := rm.replayEngine
	rm.mu.Unlock()

	if mode == replay.ModeReplay {
		return replayEngine.CurrentDetection()
	}
	if receiver != nil {
		return receiver.CombinedDetectionFrames()
	}
	return nil
}

// GetTrackedFrames 現在のトラッキングレシーバーからフレームを取得
func (rm *ReceiverManager) GetTrackedFrames() map[string]*tracked.TrackerWrapperPacket {
	rm.mu.Lock()
	mode := rm.mode
	receiver := rm.trackedReceiver
	replayEngine := rm.replayEngine
	rm.mu.Unlock()

	if mode == replay.ModeReplay {
		return replayEngine.CurrentTracked()
	}
	if receiver != nil {
		return receiver.TrackedFrames()
	}
	return nil
}

// GetGeometry 現在のビジョンレシーバーからジオメトリを取得
func (rm *ReceiverManager) GetGeometry() *vision.SSL_GeometryData {
	rm.mu.Lock()
	mode := rm.mode
	receiver := rm.visionReceiver
	replayEngine := rm.replayEngine
	rm.mu.Unlock()

	if mode == replay.ModeReplay {
		return replayEngine.CurrentGeometry()
	}
	if receiver != nil {
		return receiver.CurrentGeometry()
	}
	return nil
}

// GetRefereeMsg 現在のレフェリーレシーバーからメッセージを取得
func (rm *ReceiverManager) GetRefereeMsg() *gc.Referee {
	rm.mu.Lock()
	mode := rm.mode
	receiver := rm.refereeReceiver
	replayEngine := rm.replayEngine
	rm.mu.Unlock()

	if mode == replay.ModeReplay {
		return replayEngine.CurrentReferee()
	}
	if receiver != nil {
		return receiver.RefereeMsg()
	}
	return nil
}

// stopReceiversLocked 既存レシーバーを停止（mu保持状態で呼ぶこと）
func (rm *ReceiverManager) stopReceiversLocked() {
	if rm.visionReceiver != nil && rm.visionReceiver.MulticastServer != nil {
		rm.visionReceiver.MulticastServer.Stop()
	}
	if rm.trackedReceiver != nil && rm.trackedReceiver.MulticastServer != nil {
		rm.trackedReceiver.MulticastServer.Stop()
	}
	if rm.refereeReceiver != nil && rm.refereeReceiver.MulticastServer != nil {
		rm.refereeReceiver.MulticastServer.Stop()
	}
}

// RestartReceivers レシーバーを再起動（config.ReceiverRestarterインターフェースの実装）
func (rm *ReceiverManager) RestartReceivers(visionAddr, trackedAddr, refereeAddr string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.stopReceiversLocked()
	rm.startReceivers(visionAddr, trackedAddr, refereeAddr)
	rm.mode = replay.ModeLive

	log.Printf("Receivers restarted: vision=%s, tracked=%s, referee=%s", visionAddr, trackedAddr, refereeAddr)
	return nil
}

func (rm *ReceiverManager) SetMode(mode replay.Mode) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if mode == rm.mode {
		return nil
	}
	if mode != replay.ModeLive && mode != replay.ModeReplay {
		return fmt.Errorf("invalid mode: %s", mode)
	}

	if mode == replay.ModeReplay {
		rm.stopReceiversLocked()
		rm.mode = replay.ModeReplay
		log.Println("Data mode switched to replay")
		return nil
	}

	rm.startReceivers(rm.visionAddr, rm.trackedAddr, rm.refereeAddr)
	rm.mode = replay.ModeLive
	log.Println("Data mode switched to live")
	return nil
}

func (rm *ReceiverManager) State() replay.State {
	rm.mu.Lock()
	mode := rm.mode
	replayEngine := rm.replayEngine
	rm.mu.Unlock()
	return replayEngine.State(mode)
}

func (rm *ReceiverManager) Control(req replay.ControlRequest) error {
	return rm.replayEngine.Control(req)
}

func (rm *ReceiverManager) LoadReplay(file io.Reader, filename string) error {
	return rm.replayEngine.LoadFromReader(file, filename)
}

func (rm *ReceiverManager) LoadReplayFromPath(path string) error {
	return rm.replayEngine.LoadFromPath(path)
}

func parseSkipInterfaces() []string {
	return strings.Split(*skipInterfaces, ",")
}

func formattedAddress() string {
	if strings.HasPrefix(*address, ":") {
		return "http://localhost" + *address
	}
	return "http://" + *address
}

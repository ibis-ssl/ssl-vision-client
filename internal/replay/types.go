package replay

type Mode string

const (
	ModeLive   Mode = "live"
	ModeReplay Mode = "replay"
)

type Annotation struct {
	TimestampNs int64  `json:"timestampNs"`
	Label       string `json:"label"`
}

type State struct {
	Mode        Mode         `json:"mode"`
	Loaded      bool         `json:"loaded"`
	Playing     bool         `json:"playing"`
	Rate        float64      `json:"rate"`
	PositionNs  int64        `json:"positionNs"`
	DurationNs  int64        `json:"durationNs"`
	FileName    string       `json:"fileName"`
	Error       string       `json:"error"`
	Annotations []Annotation `json:"annotations"`
}

const (
	ActionPlay    = "play"
	ActionPause   = "pause"
	ActionSeek    = "seek"
	ActionStep    = "step"
	ActionSetRate = "set_rate"
	ActionSetMode = "set_mode"
)

type ControlRequest struct {
	Action     string  `json:"action"`
	Mode       string  `json:"mode"`
	PositionNs int64   `json:"positionNs"`
	Delta      int     `json:"delta"`
	Rate       float64 `json:"rate"`
}

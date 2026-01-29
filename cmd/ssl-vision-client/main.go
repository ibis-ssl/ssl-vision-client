package main

import (
	"flag"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/config"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/gc"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/grsim"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/tracked"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/vision"
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
	grSimSender     *grsim.Sender
	skipIfis        []string
	verbose         bool
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
		skipIfis:    skipIfis,
		verbose:     *verbose,
		grSimSender: grSimSender,
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
		grSimSender,
	)
	return &http.Server{
		Addr:    *address,
		Handler: srv,
	}
}

// startReceivers レシーバーを起動
func (rm *ReceiverManager) startReceivers(visionAddr, trackedAddr, refereeAddr string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

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
	// 読み取り専用ロックを使用してパフォーマンスを向上
	rm.mu.Lock()
	receiver := rm.visionReceiver
	rm.mu.Unlock()

	if receiver != nil {
		return receiver.CombinedDetectionFrames()
	}
	return nil
}

// GetTrackedFrames 現在のトラッキングレシーバーからフレームを取得
func (rm *ReceiverManager) GetTrackedFrames() map[string]*tracked.TrackerWrapperPacket {
	rm.mu.Lock()
	receiver := rm.trackedReceiver
	rm.mu.Unlock()

	if receiver != nil {
		return receiver.TrackedFrames()
	}
	return nil
}

// GetGeometry 現在のビジョンレシーバーからジオメトリを取得
func (rm *ReceiverManager) GetGeometry() *vision.SSL_GeometryData {
	rm.mu.Lock()
	receiver := rm.visionReceiver
	rm.mu.Unlock()

	if receiver != nil {
		return receiver.CurrentGeometry()
	}
	return nil
}

// GetRefereeMsg 現在のレフェリーレシーバーからメッセージを取得
func (rm *ReceiverManager) GetRefereeMsg() *gc.Referee {
	rm.mu.Lock()
	receiver := rm.refereeReceiver
	rm.mu.Unlock()

	if receiver != nil {
		return receiver.RefereeMsg()
	}
	return nil
}

// RestartReceivers レシーバーを再起動（config.ReceiverRestarterインターフェースの実装）
func (rm *ReceiverManager) RestartReceivers(visionAddr, trackedAddr, refereeAddr string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 既存のレシーバーを停止
	if rm.visionReceiver != nil && rm.visionReceiver.MulticastServer != nil {
		rm.visionReceiver.MulticastServer.Stop()
	}
	if rm.trackedReceiver != nil && rm.trackedReceiver.MulticastServer != nil {
		rm.trackedReceiver.MulticastServer.Stop()
	}
	if rm.refereeReceiver != nil && rm.refereeReceiver.MulticastServer != nil {
		rm.refereeReceiver.MulticastServer.Stop()
	}

	// 新しいレシーバーを起動
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

	log.Printf("Receivers restarted: vision=%s, tracked=%s, referee=%s", visionAddr, trackedAddr, refereeAddr)
	return nil
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

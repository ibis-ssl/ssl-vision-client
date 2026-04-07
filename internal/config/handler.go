package config

import (
	"encoding/json"
	"log"
	"net/http"
)

// ConfigUpdateRequest 設定更新リクエスト
type ConfigUpdateRequest struct {
	VisionPort                 int    `json:"visionPort"`
	TrackedPort                int    `json:"trackedPort"`
	RefereePort                int    `json:"refereePort"`
	SimAddress                 string `json:"simAddress"`
	SimPort                    int    `json:"simPort"`
	AutoBallPlacementEnabled   bool   `json:"autoBallPlacementEnabled"`
	AutoCenterAfterGoalEnabled bool   `json:"autoCenterAfterGoalEnabled"`
}

// ReceiverRestarter レシーバーを再起動するインターフェース
type ReceiverRestarter interface {
	RestartReceivers(visionAddr, trackedAddr, refereeAddr string) error
}

// SimSenderReconnector シミュレータ送信者を再接続するインターフェース
type SimSenderReconnector interface {
	ReconnectSimSender(addr string) error
}

// HandleConfig 設定の取得・更新を処理するハンドラー
func HandleConfig(cfg *Config, restarter ReceiverRestarter, reconnector SimSenderReconnector) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method == http.MethodGet {
			handleGetConfig(w, cfg)
			return
		}

		if r.Method == http.MethodPost {
			handleUpdateConfig(w, r, cfg, restarter, reconnector)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})
}

func handleGetConfig(w http.ResponseWriter, cfg *Config) {
	w.Header().Set("Content-Type", "application/json")

	cfg.mu.RLock()
	visionPort := cfg.VisionPort
	trackedPort := cfg.TrackedPort
	refereePort := cfg.RefereePort
	simAddress := cfg.SimAddress
	simPort := cfg.SimPort
	autoBallPlacement := cfg.AutoBallPlacementEnabled
	autoCenterAfterGoal := cfg.AutoCenterAfterGoalEnabled
	cfg.mu.RUnlock()

	response := map[string]interface{}{
		"visionPort":                 visionPort,
		"trackedPort":                trackedPort,
		"refereePort":                refereePort,
		"visionIP":                   VisionIP,
		"trackedIP":                  TrackedIP,
		"refereeIP":                  RefereeIP,
		"simAddress":                 simAddress,
		"simPort":                    simPort,
		"autoBallPlacementEnabled":   autoBallPlacement,
		"autoCenterAfterGoalEnabled": autoCenterAfterGoal,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding config response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleUpdateConfig(w http.ResponseWriter, r *http.Request, cfg *Config, restarter ReceiverRestarter, reconnector SimSenderReconnector) {
	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// ポート番号の検証
	if req.VisionPort <= 0 || req.VisionPort > 65535 ||
		req.TrackedPort <= 0 || req.TrackedPort > 65535 ||
		req.RefereePort <= 0 || req.RefereePort > 65535 {
		http.Error(w, "Invalid port number", http.StatusBadRequest)
		return
	}

	// シミュレータポートの検証（0 は「変更なし」として許可）
	if req.SimPort < 0 || req.SimPort > 65535 {
		http.Error(w, "Invalid simulator port number", http.StatusBadRequest)
		return
	}

	// 設定を更新
	cfg.Update(
		req.VisionPort,
		req.TrackedPort,
		req.RefereePort,
		req.SimAddress,
		req.SimPort,
		req.AutoBallPlacementEnabled,
		req.AutoCenterAfterGoalEnabled,
	)

	// ファイルに保存
	if err := cfg.Save(); err != nil {
		log.Printf("Error saving config: %v", err)
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// アドレスを構築してレシーバーを再起動
	vision, tracked, referee := cfg.GetAddresses()
	if err := restarter.RestartReceivers(vision, tracked, referee); err != nil {
		log.Printf("Error restarting receivers: %v", err)
		http.Error(w, "Failed to restart receivers", http.StatusInternalServerError)
		return
	}

	if err := reconnector.ReconnectSimSender(cfg.GetSimAddress()); err != nil {
		log.Printf("Warning: Failed to reconnect sim sender: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

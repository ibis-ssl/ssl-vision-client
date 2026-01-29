package grsim

import (
	"encoding/json"
	"log"
	"net/http"
)

// BallReplacementRequest ボール配置リクエスト
type BallReplacementRequest struct {
	X float64 `json:"x"` // メートル単位
	Y float64 `json:"y"` // メートル単位
}

// RobotReplacementRequest ロボット配置リクエスト
type RobotReplacementRequest struct {
	X          float64 `json:"x"`          // メートル単位
	Y          float64 `json:"y"`          // メートル単位
	Dir        float64 `json:"dir"`        // 度単位
	ID         uint32  `json:"id"`         // ロボットID
	YellowTeam bool    `json:"yellowTeam"` // true=黄色チーム、false=青色チーム
}

// HandleReplaceBall ボール配置APIハンドラー
func HandleReplaceBall(sender *Sender) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req BallReplacementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := sender.ReplaceBall(req.X, req.Y); err != nil {
			log.Printf("Error replacing ball: %v", err)
			http.Error(w, "Failed to replace ball", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
}

// HandleReplaceRobot ロボット配置APIハンドラー
func HandleReplaceRobot(sender *Sender) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RobotReplacementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := sender.ReplaceRobot(req.X, req.Y, req.Dir, req.ID, req.YellowTeam); err != nil {
			log.Printf("Error replacing robot: %v", err)
			http.Error(w, "Failed to replace robot", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})
}

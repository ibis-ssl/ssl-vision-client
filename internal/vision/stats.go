package vision

import (
	"encoding/json"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/common"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type ServiceStats struct {
	VisionHz      float64 `json:"visionHz"`
	CameraCount   int     `json:"cameraCount"`
	TrackerHz     float64 `json:"trackerHz"`
	TrackerActive bool    `json:"trackerActive"`
	RefereeHz     float64 `json:"refereeHz"`
}

func HandleServiceStats(statsProvider func() ServiceStats) http.Handler {
	return common.UpgradeToWebsocket(
		func(_ *http.Request, conn *websocket.Conn) {
			log.Println("Client for service stats connected")
			defer log.Println("Client for service stats disconnected")

			var lastJSON []byte
			for {
				stats := statsProvider()
				payload, err := json.Marshal(stats)
				if err != nil {
					log.Println(err)
					return
				}
				if string(payload) != string(lastJSON) {
					if err := common.SendJSONMessage(conn, stats); err != nil {
						log.Println(err)
						return
					}
					lastJSON = payload
				}
				time.Sleep(500 * time.Millisecond)
			}
		},
	)
}

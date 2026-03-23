package portmonitor

import (
	"encoding/json"
	"github.com/RoboCup-SSL/ssl-vision-client/internal/common"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const publishInterval = 500 * time.Millisecond

func HandlePortStatus(monitor *Monitor) http.Handler {
	return common.UpgradeToWebsocket(
		func(_ *http.Request, conn *websocket.Conn) {
			log.Println("Client for port status connected")
			defer log.Println("Client for port status disconnected")

			var lastJSON []byte
			for {
				status := monitor.Status()
				payload, err := json.Marshal(status)
				if err != nil {
					log.Println(err)
					return
				}
				if string(payload) != string(lastJSON) {
					if err := common.SendJSONMessage(conn, status); err != nil {
						log.Println(err)
						return
					}
					lastJSON = payload
				}
				time.Sleep(publishInterval)
			}
		},
	)
}

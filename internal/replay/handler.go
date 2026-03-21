package replay

import (
	"encoding/json"
	"io"
	"net/http"
)

type Service interface {
	SetMode(mode Mode) error
	State() State
	Control(req ControlRequest) error
	LoadReplay(file io.Reader, filename string) error
	LoadReplayFromPath(path string) error
}

func HandleUpload(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(256 << 20); err != nil {
			http.Error(w, "Invalid multipart form", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Missing file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if err := service.LoadReplay(file, header.Filename); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.SetMode(ModeReplay); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, service.State())
	})
}

func HandleLoadPath(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
			http.Error(w, "Invalid request body: missing path", http.StatusBadRequest)
			return
		}

		if err := service.LoadReplayFromPath(req.Path); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := service.SetMode(ModeReplay); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, service.State())
	})
}

func HandleState(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, service.State())
	})
}

func HandleControl(service Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ControlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Action == ActionSetMode {
			if err := service.SetMode(Mode(req.Mode)); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			if err := service.Control(req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		writeJSON(w, service.State())
	})
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

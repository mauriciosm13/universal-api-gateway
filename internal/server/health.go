package server

import (
	"encoding/json"
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func writeHealth(w http.ResponseWriter, status string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	_ = json.NewEncoder(w).Encode(healthResponse{
		Status:    status,
		Service:   "universal-api-gateway",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func handleLive(w http.ResponseWriter, _ *http.Request) {
	writeHealth(w, "alive", http.StatusOK)
}

func handleReady(w http.ResponseWriter, _ *http.Request) {
	writeHealth(w, "ready", http.StatusOK)
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeHealth(w, "ok", http.StatusOK)
}

package httpapi

import (
	"encoding/json"
	"net/http"
)

type health struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Handler serves a local-only health endpoint; no account or platform API exists yet.
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Browser origins used by the local Vite server and Tauri on Windows/macOS.
		origin := r.Header.Get("Origin")
		switch origin {
		case "http://127.0.0.1:1420", "http://localhost:1420", "http://tauri.localhost", "tauri://localhost", "https://tauri.localhost":
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(health{Status: "ok", Service: "una-seller-api"})
	})
	return mux
}

package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status: %d", w.Code)
	}
	var v struct{ Status, Service string }
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v.Status != "ok" || v.Service != "una-seller-api" {
		t.Fatalf("unexpected health: %+v", v)
	}
}

func TestMethodAndCORS(t *testing.T) {
	for _, tc := range []struct{ origin, want string }{
		{"http://127.0.0.1:1420", "http://127.0.0.1:1420"},
		{"tauri://localhost", "tauri://localhost"},
		{"https://evil.example", ""},
	} {
		r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		r.Header.Set("Origin", tc.origin)
		w := httptest.NewRecorder()
		Handler().ServeHTTP(w, r)
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != tc.want {
			t.Errorf("origin %q: got %q, want %q", tc.origin, got, tc.want)
		}
	}
	r := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	w := httptest.NewRecorder()
	Handler().ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST status: %d", w.Code)
	}
}

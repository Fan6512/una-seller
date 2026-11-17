package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
)

func TestExtensionOriginPreflightAndAuth(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{err: ErrInvalidSession})
	origin := "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	r := httptest.NewRequest(http.MethodOptions, "/api/v1/collections", nil)
	r.Header.Set("Origin", origin)
	r.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") != origin {
		t.Fatalf("extension preflight status=%d origin=%q", w.Code, w.Header().Get("Access-Control-Allow-Origin"))
	}
	r = httptest.NewRequest(http.MethodPost, "/api/v1/collections", nil)
	r.Header.Set("Origin", origin)
	r.Header.Set("Authorization", "Bearer invalid")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || w.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("extension origin must not bypass auth: %d", w.Code)
	}
}

func TestExtensionOriginRejectsMalformedAndWebOrigins(t *testing.T) {
	h := HandlerWithCollections(nil, nil)
	for _, origin := range []string{
		"chrome-extension://not-valid", "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.evil.com",
		"chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/path",
		"https://amazon.com", "https://evil.example", "null",
	} {
		r := httptest.NewRequest(http.MethodOptions, "/api/v1/collections", nil)
		r.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 403 {
			t.Errorf("untrusted origin %q status=%d, want 403", origin, w.Code)
		}
	}
}

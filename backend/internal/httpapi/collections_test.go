package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
)

type stubVerifier struct {
	id  string
	err error
}

func (v stubVerifier) Verify(_ context.Context, _ string) (string, error) {
	return v.id, v.err
}

func TestUnconfiguredCollectionIsNotFakeSuccess(t *testing.T) {
	h := HandlerWithCollections(nil, nil)
	for _, method := range []string{"GET", "POST"} {
		r := httptest.NewRequest(method, "/api/v1/collections", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d", method, w.Code)
		}
	}
}

func TestCollectionRequiresValidSessionBeforeReadingData(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb"})
	for _, path := range []string{"/api/v1/collections", "/api/v1/collections/6cbe41df-69cb-4594-9bad-a5473aa1a5fb"} {
		for _, header := range []string{"", "Basic foo", "Bearer"} {
			r := httptest.NewRequest(http.MethodGet, path, nil)
			r.Header.Set("Authorization", header)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			if w.Code != http.StatusUnauthorized || w.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("path=%s header=%q status=%d challenge=%q", path, header, w.Code, w.Header().Get("WWW-Authenticate"))
			}
		}
	}
}

func TestCollectionAuthFailureAndCORS(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{err: ErrInvalidSession})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	r.Header.Set("Authorization", "Bearer invalid-token")
	r.Header.Set("Origin", "http://tauri.localhost")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 401 || w.Header().Get("Access-Control-Allow-Origin") != "http://tauri.localhost" || w.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("invalid token response=%d cors=%q", w.Code, w.Header().Get("Access-Control-Allow-Origin"))
	}

	r = httptest.NewRequest(http.MethodOptions, "/api/v1/collections", nil)
	r.Header.Set("Origin", "http://tauri.localhost")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
		t.Fatalf("preflight response=%d headers=%q", w.Code, w.Header().Get("Access-Control-Allow-Headers"))
	}

	r = httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	r.Header.Set("Origin", "http://evil.example")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatalf("untrusted origin status=%d", w.Code)
	}

	unavailable := HandlerWithCollections(&collections.Store{}, stubVerifier{err: ErrAuthUnavailable})
	r = httptest.NewRequest(http.MethodGet, "/api/v1/collections", nil)
	r.Header.Set("Authorization", "Bearer some-token")
	w = httptest.NewRecorder()
	unavailable.ServeHTTP(w, r)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("auth outage must not pass as authorized, status=%d", w.Code)
	}
}

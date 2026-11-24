package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
)

func TestCBTC0ProbeRequiresPostAndSession(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{})
	const path = "/api/v1/stores/6cbe41df-69cb-4594-9bad-a5473aa1a5fb/publication-mode/probe"
	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, path, nil))
	if get.Code != http.StatusMethodNotAllowed || get.Header().Get("Allow") != "POST" {
		t.Fatalf("GET status=%d allow=%q", get.Code, get.Header().Get("Allow"))
	}
	post := httptest.NewRecorder()
	h.ServeHTTP(post, httptest.NewRequest(http.MethodPost, path, nil))
	if post.Code != http.StatusUnauthorized || post.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("unauthenticated POST status=%d challenge=%q", post.Code, post.Header().Get("WWW-Authenticate"))
	}
}

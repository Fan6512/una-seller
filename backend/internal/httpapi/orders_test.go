package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func TestOrdersRequireAuthenticatedDesktopSession(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		r := httptest.NewRequest(method,
			"/api/v1/stores/00000000-0000-4000-8000-000000000001/orders?site=MLB&logistic_type=remote", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s unauthenticated order access: %d", method, w.Code)
		}
	}
}

func TestOrdersRejectMissingScopeAndInvalidFiltersBeforeDatabase(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders/sync?site=MLB"},
		{http.MethodGet, "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders?site=MLB&logistic_type=remote&offset=-1"},
		{http.MethodGet, "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders?site=MLB&logistic_type=remote&status=not%3Bvalid"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		req.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s %s: status=%d body=%s", tt.method, tt.path, w.Code, w.Body.String())
		}
	}
}

package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnconfiguredCollectionIsNotFakeSuccess(t *testing.T) {
	h := HandlerWithCollections(nil, "")
	for _, method := range []string{"GET", "POST"} {
		r := httptest.NewRequest(method, "/api/v1/collections", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d", method, w.Code)
		}
	}
}

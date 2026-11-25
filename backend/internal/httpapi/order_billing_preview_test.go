package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func TestBillingPreviewRejectsInvalidRequestsBeforeDatabase(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	const path = "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders/billing-preview"
	const scope = "?site=MLB&logistic_type=remote"
	for _, tc := range []struct {
		method string
		query  string
		body   string
		auth   bool
		want   int
	}{
		{http.MethodGet, scope, "", true, http.StatusMethodNotAllowed},
		{http.MethodPost, scope, `{"order_ids":["11"]}`, false, http.StatusUnauthorized},
		{http.MethodPost, "?site=MLB", `{"order_ids":["11"]}`, true, http.StatusBadRequest},
		{http.MethodPost, scope + "&site=MLM", `{"order_ids":["11"]}`, true, http.StatusBadRequest},
		{http.MethodPost, scope, `{"order_ids":["11"],"unknown":true}`, true, http.StatusBadRequest},
		{http.MethodPost, scope, `{"order_ids":["11"]} {}`, true, http.StatusBadRequest},
		{http.MethodPost, scope, `{"order_ids":["11"]}` + strings.Repeat(" ", 4096), true, http.StatusBadRequest},
	} {
		r := httptest.NewRequest(tc.method, path+tc.query, strings.NewReader(tc.body))
		if tc.auth {
			r.Header.Set("Authorization", "Bearer test-only")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("billing %s: status=%d want=%d body=%s", tc.method, w.Code, tc.want, w.Body.String())
		}
		if tc.want == http.StatusMethodNotAllowed && w.Header().Get("Allow") != http.MethodPost {
			t.Fatal("missing POST Allow header")
		}
	}
}

func TestBillingPreviewErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
		code string
	}{
		{meli.ErrOrderBillingInvalid, http.StatusBadRequest, "order_billing_invalid"},
		{meli.ErrOrderNotFound, http.StatusNotFound, "order_not_found"},
		{meli.ErrSiteNotAuthorized, http.StatusForbidden, "order_site_not_authorized"},
		{meli.ErrOrderAuthorizationChanged, http.StatusConflict, "order_auth_changed"},
		{meli.ErrOrderBusy, http.StatusConflict, "order_busy"},
		{meli.ListingPlatformError(http.StatusTooManyRequests), http.StatusTooManyRequests, "order_rate_limited"},
		{meli.ListingPlatformError(http.StatusNotFound), http.StatusBadGateway, "order_platform_unavailable"},
	} {
		w := httptest.NewRecorder()
		writeOrderBillingError(w, tc.err)
		if tc.want == http.StatusTooManyRequests && w.Header().Get("Access-Control-Expose-Headers") != "Retry-After" {
			t.Fatal("desktop cannot read the retry header over CORS")
		}
		if w.Code != tc.want || !strings.Contains(w.Body.String(), tc.code) {
			t.Fatalf("billing error %v: status=%d body=%s", tc.err, w.Code, w.Body.String())
		}
	}
}

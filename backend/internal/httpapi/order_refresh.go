package httpapi

import (
	"errors"
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// The user explicitly requests a fresh official detail for one existing order.
// Payment/claim reconciliation is separate; a cancelled order isn't necessarily refunded.
func (a storeAPI) orderRefresh(w http.ResponseWriter, r *http.Request, store, id string) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "order_method_not_allowed"})
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	if q.Get("site") == "" || q.Get("logistic_type") == "" ||
		len(q["site"]) != 1 || len(q["logistic_type"]) != 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_scope_required"})
		return
	}
	scope := meli.OrderScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}
	record, err := a.service.RefreshOneOrder(r.Context(), owner, store, scope, id)
	if err == nil {
		writeJSON(w, http.StatusOK, record)
		return
	}
	code := meli.OrderErrorCode(err)
	status := http.StatusServiceUnavailable
	switch code {
	case "order_not_found":
		status = http.StatusNotFound
	case "order_site_not_authorized", "order_reauth_required":
		status = http.StatusForbidden
	case "order_auth_changed", "order_refresh_unknown", "order_busy":
		status = http.StatusConflict
	case "order_rate_limited":
		status = http.StatusTooManyRequests
		var limit *meli.OrderRateLimitError
		if errors.As(err, &limit) {
			w.Header().Set("Retry-After", limit.RetryAfter())
		}
	case "order_timeout":
		status = http.StatusGatewayTimeout
	case "order_canceled":
		status = http.StatusRequestTimeout
	case "order_invalid_response", "order_platform_unavailable":
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]string{"error": code})
}

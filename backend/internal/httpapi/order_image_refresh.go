package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// Manual image-only refresh; does not edit an official listing, a shipment,
// payment or any third-party order. OAuth token stays on the Go service.
func (a storeAPI) orderImageRefresh(w http.ResponseWriter, r *http.Request, store, id string) {
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
		len(q["site"]) != 1 || len(q["logistic_type"]) != 1 ||
		strings.Contains(store, "/") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_scope_required"})
		return
	}
	scope := meli.OrderScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}
	order, err := a.service.RefreshOrderPrimaryImage(r.Context(), owner, store, scope, id)
	if err == nil {
		writeJSON(w, http.StatusOK, order)
		return
	}
	code := meli.OrderErrorCode(err)
	status := http.StatusBadGateway
	switch {
	case errors.Is(err, meli.ErrOrderImageUnavailable):
		code, status = "order_image_unavailable", http.StatusNotFound
	case code == "order_not_found":
		status = http.StatusNotFound
	case code == "order_site_not_authorized" || code == "order_reauth_required":
		status = http.StatusForbidden
	case code == "order_rate_limited":
		status = http.StatusTooManyRequests
		var limit *meli.OrderRateLimitError
		if errors.As(err, &limit) {
			w.Header().Set("Retry-After", limit.RetryAfter())
		}
	case code == "order_busy" || code == "order_auth_changed" || code == "order_refresh_unknown":
		status = http.StatusConflict
	case code == "order_timeout":
		status = http.StatusGatewayTimeout
	case code == "order_canceled":
		status = http.StatusRequestTimeout
	}
	writeJSON(w, status, map[string]string{"error": code})
}

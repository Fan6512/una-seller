package httpapi

import (
	"errors"
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func (a storeAPI) orderDispatchSLA(w http.ResponseWriter, r *http.Request, store, orderID string) {
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
	result, err := a.service.RefreshOrderDispatchSLA(r.Context(), owner, store, scope, orderID)
	if err == nil {
		writeJSON(w, http.StatusOK, result)
		return
	}
	code := meli.OrderErrorCode(err)
	status := http.StatusServiceUnavailable
	switch {
	case errors.Is(err, meli.ErrOrderSLANoShipment):
		code, status = "order_sla_no_shipment", http.StatusNotFound
	case errors.Is(err, meli.ErrOrderSLAInvalid):
		code, status = "order_sla_invalid", http.StatusBadGateway
	case errors.Is(err, meli.ErrOrderSLANotApplicable):
		code, status = "order_sla_not_applicable", http.StatusConflict
	case code == "order_not_found":
		status = http.StatusNotFound
	case code == "order_site_not_authorized" || code == "order_reauth_required":
		status = http.StatusForbidden
	case code == "order_auth_changed" || code == "order_refresh_unknown" || code == "order_busy":
		status = http.StatusConflict
	case code == "order_rate_limited":
		status = http.StatusTooManyRequests
		var limit *meli.OrderRateLimitError
		if errors.As(err, &limit) {
			w.Header().Set("Retry-After", limit.RetryAfter())
		}
	case code == "order_timeout":
		status = http.StatusGatewayTimeout
	case code == "order_canceled":
		status = http.StatusRequestTimeout
	case code == "order_invalid_response" || code == "order_platform_unavailable":
		status = http.StatusBadGateway
	}
	writeJSON(w, status, map[string]string{"error": code})
}

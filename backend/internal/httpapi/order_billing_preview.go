package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// Official CBT billing is fetched only on manual read. A complete single-order
// result can be stored locally; the cache endpoint never contacts the platform.
// Do not expose raw JSON (payer PII), global merchant credentials, or an inferred net.
func (a storeAPI) orderBillingPreview(w http.ResponseWriter, r *http.Request, store string) {
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
	var request struct {
		OrderIDs []string `json:"order_ids"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_billing_invalid"})
		return
	}
	page, err := a.service.PreviewOrderBilling(r.Context(), owner, store,
		meli.OrderScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}, request.OrderIDs)
	if err == nil {
		writeJSON(w, http.StatusOK, page)
		return
	}
	writeOrderBillingError(w, err)
}

// orderBillingCache is a read-only local snapshot, not an official API call.
func (a storeAPI) orderBillingCache(w http.ResponseWriter, r *http.Request, store, orderID string) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
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
	page, err := a.service.GetCachedOrderBilling(r.Context(), owner, store,
		meli.OrderScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}, orderID)
	if err != nil {
		writeOrderBillingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func writeOrderBillingError(w http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	code := meli.OrderErrorCode(err)
	switch {
	case errors.Is(err, meli.ErrOrderBillingInvalid):
		status = http.StatusBadRequest
		code = "order_billing_invalid"
	case code == "order_not_found":
		status = http.StatusNotFound
	case code == "order_site_not_authorized" || code == "order_reauth_required":
		status = http.StatusForbidden
	case code == "order_rate_limited":
		status = http.StatusTooManyRequests
		// Retry-After is not a CORS-safelisted response header. Expose only
		// this non-sensitive value; the existing origin and owner gates remain.
		w.Header().Set("Access-Control-Expose-Headers", "Retry-After")
		var cooldown *meli.OrderRateLimitError
		if errors.As(err, &cooldown) {
			w.Header().Set("Retry-After", cooldown.RetryAfter())
		}
	case code == "order_busy" || code == "order_refresh_unknown" || code == "order_auth_changed":
		status = http.StatusConflict
	case code == "order_timeout":
		status = http.StatusGatewayTimeout
	case code == "order_canceled":
		status = http.StatusRequestTimeout
	}
	writeJSON(w, status, map[string]string{"error": code})
}

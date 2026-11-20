package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// All order routes use the existing store-scoped owner verifier.
// Nothing exposes Mercado Libre access tokens or buyer PII to the desktop.
func (a storeAPI) orders(w http.ResponseWriter, r *http.Request, store, tail string) {
	w.Header().Set("Cache-Control", "no-store")
	if store == "" || strings.Contains(store, "/") {
		writeError(w, http.StatusNotFound, "store not found")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	site, logistic := q.Get("site"), q.Get("logistic_type")
	if site == "" || logistic == "" || len(q["site"]) != 1 || len(q["logistic_type"]) != 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_scope_required"})
		return
	}
	scope := meli.OrderScope{Site: site, Logistic: logistic}
	var err error
	switch {
	case tail == "sync" && r.Method == http.MethodPost:
		var result meli.OrderSyncResult
		result, err = a.service.SyncOrders(r.Context(), owner, store, scope)
		if err == nil {
			writeJSON(w, http.StatusOK, result)
			return
		}
	case tail == "" && r.Method == http.MethodGet:
		offset := 0
		if q.Has("offset") {
			var parseErr error
			offset, parseErr = strconv.Atoi(q.Get("offset"))
			if parseErr != nil || offset < 0 || offset > 100000 || len(q["offset"]) != 1 {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_offset_invalid"})
				return
			}
		}
		status := q.Get("status")
		if !meli.ValidOrderStatus(status) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_status_invalid"})
			return
		}
		var page meli.OrderPage
		page, err = a.service.ListOrders(r.Context(), owner, store, scope, status, offset)
		if err == nil {
			writeJSON(w, http.StatusOK, page)
			return
		}
	case tail != "" && tail != "sync" && r.Method == http.MethodGet:
		var order meli.OrderRecord
		order, err = a.service.GetOrder(r.Context(), owner, store, scope, tail)
		if err == nil {
			writeJSON(w, http.StatusOK, order)
			return
		}
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order_route_not_found"})
		return
	}
	code := meli.OrderErrorCode(err)
	status := http.StatusServiceUnavailable
	switch code {
	case "order_not_found":
		status = http.StatusNotFound
	case "order_site_not_authorized", "order_reauth_required":
		status = http.StatusForbidden
	case "order_busy", "order_refresh_unknown", "order_auth_changed":
		status = http.StatusConflict
	case "order_invalid_response", "order_platform_unavailable":
		status = http.StatusBadGateway
	case "order_rate_limited":
		status = http.StatusTooManyRequests
	}
	writeJSON(w, status, map[string]string{"error": code})
}

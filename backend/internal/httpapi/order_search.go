package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// GET /api/v1/stores/{store_uuid|all}/orders/search?site=*&logistic_type=*&q=... .
//
// The path is authenticated by storeAPI.orders before reaching this handler.
func (a storeAPI) searchOrders(w http.ResponseWriter, r *http.Request, owner, store string) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "order_method_not_allowed"})
		return
	}
	q := r.URL.Query()
	for _, key := range []string{"site", "logistic_type", "status", "purchase", "q", "offset"} {
		if len(q[key]) > 1 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_search_invalid"})
			return
		}
	}
	offset := 0
	if q.Has("offset") {
		value, err := strconv.Atoi(q.Get("offset"))
		if err != nil || value < 0 || value > 100000 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_search_invalid"})
			return
		}
		offset = value
	}
	filter := meli.OrderSearchFilter{
		Store: store, Site: q.Get("site"), Logistic: q.Get("logistic_type"),
		Status: q.Get("status"), Purchase: q.Get("purchase"),
		Query: strings.TrimSpace(q.Get("q")), Offset: offset,
	}
	if !filter.Valid() || (q.Has("q") && q.Get("q") != filter.Query) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_search_invalid"})
		return
	}
	page, err := a.service.SearchOrders(r.Context(), owner, filter)
	if err == nil {
		writeJSON(w, http.StatusOK, page)
		return
	}
	if errors.Is(err, meli.ErrOrderInvalidResponse) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_search_invalid"})
		return
	}
	writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "order_search_unavailable"})
}

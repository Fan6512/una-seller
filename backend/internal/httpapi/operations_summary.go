package httpapi

import (
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func (a storeAPI) operationsSummary(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	store, site := r.URL.Query().Get("store_id"), r.URL.Query().Get("site")
	if !meli.ValidOperationsFilter(store, site) {
		writeError(w, http.StatusBadRequest, "invalid operations filter")
		return
	}
	summary, err := a.service.OperationsOverview(r.Context(), owner, store, site)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "operations summary unavailable")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

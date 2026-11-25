package httpapi

import "net/http"

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
	summary, err := a.service.OperationsSummary(r.Context(), owner)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "operations summary unavailable")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

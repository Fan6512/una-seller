package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/collections"
)

func (a collectionAPI) workingDraft(w http.ResponseWriter, r *http.Request, owner, id string) {
	switch r.Method {
	case http.MethodGet:
		draft, err := a.store.GetWorkingDraft(r.Context(), owner, id)
		if errors.Is(err, collections.ErrNotFound) {
			writeError(w, http.StatusNotFound, "collection not found")
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "working draft unavailable")
		} else {
			writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
		}
	case http.MethodPut:
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			writeError(w, http.StatusUnsupportedMediaType, "expected application/json")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var in collections.WorkingDraftInput
		if err := dec.Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid working draft JSON")
			return
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			writeError(w, http.StatusBadRequest, "trailing working draft data")
			return
		}
		if err := in.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, "invalid working draft fields")
			return
		}
		draft, err := a.store.SaveWorkingDraft(r.Context(), owner, id, in)
		if errors.Is(err, collections.ErrNotFound) {
			writeError(w, http.StatusNotFound, "collection not found")
		} else if err != nil {
			writeError(w, http.StatusInternalServerError, "working draft not saved")
		} else {
			writeJSON(w, http.StatusOK, map[string]any{"draft": draft})
		}
	default:
		w.Header().Set("Allow", "GET, PUT")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

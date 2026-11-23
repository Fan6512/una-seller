package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// Desktop-only, authenticated and read-only global CBT category suggestions.
// Site-specific category rules are still validated at publication.
func (a storeAPI) predictCBTCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && !allowedAppOrigin(origin) {
		writeError(w, http.StatusForbidden, "origin not allowed")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	if a.service == nil {
		writeError(w, http.StatusServiceUnavailable, "Mercado Libre category predictor is not configured")
		return
	}
	title := strings.TrimSpace(r.URL.Query().Get("q"))
	items, err := a.service.PredictCBTCategories(r.Context(), owner, title)
	if errors.Is(err, meli.ErrCBTPredictorInput) {
		writeError(w, http.StatusBadRequest, "edit the English title before requesting category recommendations")
		return
	}
	if errors.Is(err, meli.ErrCBTPredictorStore) {
		writeJSON(w, http.StatusPreconditionFailed, map[string]string{
			"error": "connect a Mercado Libre store to use the official CBT category predictor",
			"code":  "CBT_STORE_REQUIRED",
		})
		return
	}
	if err != nil {
		writeError(w, http.StatusBadGateway, "official CBT category recommendations unavailable; choose category manually")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	var texts []string
	for _, item := range items {
		texts = append(texts, item.CategoryName)
		for _, node := range item.Path {
			texts = append(texts, node.Name)
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "source": "mercado_libre_cbt", "confirmed": false,
		"translations_zh": a.translator.translate(r.Context(), "en", texts)})
}

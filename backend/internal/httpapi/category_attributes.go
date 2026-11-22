package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func (a storeAPI) writeCategoryRules(w http.ResponseWriter, r *http.Request, rules meli.CategoryRules) {
	if a.localizeCategory != nil {
		rules.Translations = a.localizeCategory(r.Context(), rules)
	}
	writeJSON(w, http.StatusOK, rules)
}

// Read-only, store-scoped category rules for the dynamic listing editor.
// The caller must already pass requireOwner, and the service verifies
// the CBT parent / marketplace child association before calling the platform.
func (a storeAPI) categoryRules(w http.ResponseWriter, r *http.Request, owner, store string) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	scope := meli.ListingScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}
	category := strings.TrimSpace(q.Get("category_id"))
	if len(q["site"]) != 1 || len(q["logistic_type"]) != 1 || len(q["category_id"]) != 1 {
		writeListingError(w, http.StatusBadRequest, "category_invalid", "missing category scope", nil, meli.ListingSync{})
		return
	}
	result, err := a.service.GetCategoryRules(r.Context(), owner, store, scope, category)
	if err == nil {
		a.writeCategoryRules(w, r, result)
		return
	}
	var platform meli.ListingPlatformError
	status, code := http.StatusBadGateway, "category_unavailable"
	switch {
	case errors.Is(err, meli.ErrCategoryInvalid):
		status, code = http.StatusBadRequest, "category_invalid"
	case errors.Is(err, meli.ErrSiteNotAuthorized):
		status, code = http.StatusForbidden, "listing_site_unverified"
	case errors.Is(err, meli.ErrStoreNotFound):
		status, code = http.StatusNotFound, "listing_store_missing"
	case errors.Is(err, meli.ErrReauthorizationRequired):
		status, code = http.StatusForbidden, "listing_reauthorization"
	case errors.Is(err, meli.ErrListingTokenExpired):
		status, code = http.StatusConflict, "listing_token_expired"
	case errors.Is(err, meli.ErrRefreshUnknown):
		status, code = http.StatusConflict, "listing_refresh_unknown"
	case errors.Is(err, context.Canceled):
		status, code = http.StatusRequestTimeout, "listing_canceled"
	case errors.Is(err, context.DeadlineExceeded):
		status, code = http.StatusGatewayTimeout, "listing_timeout"
	case errors.As(err, &platform):
		if int(platform) == http.StatusTooManyRequests {
			status, code = http.StatusTooManyRequests, "listing_rate_limited"
		} else if int(platform) == http.StatusUnauthorized || int(platform) == http.StatusForbidden {
			status, code = http.StatusForbidden, "listing_platform_auth"
		} else if int(platform) == http.StatusNotFound {
			status, code = http.StatusNotFound, "category_unavailable"
		}
	}
	writeListingError(w, status, code, "category rules could not be retrieved", nil, meli.ListingSync{})
}

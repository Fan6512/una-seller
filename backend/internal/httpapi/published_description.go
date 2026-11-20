package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// One on-demand, read-only description; never changes listings or the sync cursor.
func (a storeAPI) publishedDescription(w http.ResponseWriter, r *http.Request, owner, store, listing string) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	if len(q["site"]) != 1 || len(q["logistic_type"]) != 1 {
		writeListingError(w, http.StatusBadRequest, "listing_invalid_response", "invalid description scope", nil, meli.ListingSync{})
		return
	}
	desc, err := a.service.GetPublishedDescription(r.Context(), owner, store,
		meli.ListingScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}, listing)
	if err == nil {
		writeJSON(w, http.StatusOK, map[string]any{"description": desc})
		return
	}
	status, code := http.StatusBadGateway, "description_unavailable"
	var platform meli.ListingPlatformError
	switch {
	case errors.Is(err, meli.ErrListingResponse):
		status, code = http.StatusBadRequest, "listing_invalid_response"
	case errors.Is(err, meli.ErrListingDescriptionMissing), errors.Is(err, meli.ErrStoreNotFound):
		status, code = http.StatusNotFound, "listing_store_missing"
	case errors.Is(err, meli.ErrListingDescriptionLinkMissing):
		status, code = http.StatusNotFound, "description_link_missing"
	case errors.Is(err, meli.ErrSiteNotAuthorized):
		status, code = http.StatusForbidden, "listing_site_unverified"
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
		switch int(platform) {
		case http.StatusNotFound:
			status, code = http.StatusNotFound, "description_unavailable"
		case http.StatusTooManyRequests:
			status, code = http.StatusTooManyRequests, "listing_rate_limited"
		case http.StatusUnauthorized, http.StatusForbidden:
			status, code = http.StatusForbidden, "listing_platform_auth"
		}
	}
	writeListingError(w, status, code, "listing description could not be retrieved", nil, meli.ListingSync{})
}

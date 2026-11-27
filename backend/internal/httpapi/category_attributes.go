package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func (a storeAPI) attachPublishedCategoryBreadcrumbs(page *meli.ListingPage) {
	if a.categoryDictionary == nil {
		return
	}
	dictionary := a.categoryDictionary.Load()
	for index, id := range page.CBTCategoryIDs {
		breadcrumb := dictionary.PublishedBreadcrumb(id)
		if breadcrumb == nil {
			continue
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(page.Items[index], &fields) != nil || fields == nil {
			continue
		}
		path, err := json.Marshal(breadcrumb)
		if err != nil {
			continue
		}
		fields["cbt_category_path"] = path
		if item, err := json.Marshal(fields); err == nil {
			page.Items[index] = item
		}
	}
}

func (a storeAPI) writeCategoryRules(w http.ResponseWriter, r *http.Request, owner string, rules meli.CategoryRules) {
	rules.PathZH = a.categoryDictionary.Load().Localize(rules.Path)
	if a.accounts != nil {
		translator, err := a.accounts.Translation(r.Context(), owner)
		if errors.Is(err, errAccountCredentialRebind) {
			log.Printf("account translation unavailable: credentials need rebind, owner=%s", owner)
		} else if err != nil {
			log.Printf("account translation unavailable: provider config read error, owner=%s", owner)
		} else if translator != nil {
			rules.Translations = translator.translateCategoryRules(r.Context(), rules)
		}
	} else if a.localizeCategoryAttributes != nil {
		rules.Translations = a.localizeCategoryAttributes(r.Context(), rules)
	}
	writeJSON(w, http.StatusOK, rules)
}

func (a storeAPI) cbtCategoryRules(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	ids := r.URL.Query()["category_id"]
	if len(ids) != 1 {
		writeError(w, http.StatusBadRequest, "provide a single CBT category_id")
		return
	}
	rules, err := a.service.GetCBTCategoryRules(r.Context(), owner, ids[0])
	var platform meli.ListingPlatformError
	switch {
	case err == nil:
		a.writeCategoryRules(w, r, owner, rules)
	case errors.Is(err, meli.ErrCategoryInvalid):
		writeError(w, http.StatusBadRequest, "invalid CBT category ID")
	case errors.Is(err, meli.ErrCBTPredictorStore), errors.Is(err, meli.ErrReauthorizationRequired):
		writeError(w, http.StatusPreconditionFailed, "connect a Mercado Libre CBT store first")
	case errors.As(err, &platform) && int(platform) == http.StatusTooManyRequests:
		writeError(w, http.StatusTooManyRequests, "Mercado Libre category request rate limited")
	case errors.Is(err, context.DeadlineExceeded):
		writeError(w, http.StatusGatewayTimeout, "official category rule request timed out")
	default:
		writeError(w, http.StatusBadGateway, "official CBT category rules unavailable")
	}
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
		a.writeCategoryRules(w, r, owner, result)
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

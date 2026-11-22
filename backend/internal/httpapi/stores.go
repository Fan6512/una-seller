package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

type storeAPI struct {
	auth       collectionAPI
	service    *meli.Service
	translator *TranslationService
}

func (a storeAPI) requireOwner(w http.ResponseWriter, r *http.Request) (string, bool) {
	owner, ok := a.auth.owner(w, r)
	if !ok {
		return "", false
	}
	if a.service == nil {
		writeError(w, http.StatusServiceUnavailable, "Mercado Libre OAuth is not configured")
		return "", false
	}
	return owner, true
}

func (a storeAPI) stores(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	authStart := time.Now()
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	readStart := time.Now()
	items, err := a.service.List(r.Context(), owner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "stores unavailable")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	setStoreReadTiming(w, authStart, readStart)
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func setStoreReadTiming(w http.ResponseWriter, authStart, readStart time.Time) {
	w.Header().Set("Server-Timing", fmt.Sprintf("auth;dur=%.1f, read;dur=%.1f",
		float64(readStart.Sub(authStart))/float64(time.Millisecond),
		float64(time.Since(readStart))/float64(time.Millisecond)))
}

func (a storeAPI) start(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	link, err := a.service.Start(r.Context(), owner)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "could not create authorization")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"authorization_url": link})
}

// storeAction is a reusable ownership/status gate for store-scoped operations.
// The platform Access Token never leaves Go, including on a successful check.
func (a storeAPI) storeAction(w http.ResponseWriter, r *http.Request) {
	// Orders stay isolated from listing sync and desktop diagnostics.
	orderPath := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/stores/"), "/")
	if len(orderPath) == 4 && orderPath[0] != "" && orderPath[1] == "orders" && orderPath[2] != "" && orderPath[3] == "sla" {
		a.orderDispatchSLA(w, r, orderPath[0], orderPath[2])
		return
	}
	if len(orderPath) == 4 && orderPath[0] != "" && orderPath[1] == "orders" && orderPath[2] != "" && orderPath[3] == "refresh" {
		a.orderRefresh(w, r, orderPath[0], orderPath[2])
		return
	}
	if len(orderPath) == 4 && orderPath[0] != "" && orderPath[1] == "orders" && orderPath[2] != "" && orderPath[3] == "image-refresh" {
		a.orderImageRefresh(w, r, orderPath[0], orderPath[2])
		return
	}
	if len(orderPath) == 4 && orderPath[0] != "" && orderPath[1] == "orders" && orderPath[2] != "" && orderPath[3] == "purchase" {
		a.orderPurchase(w, r, orderPath[0], orderPath[2])
		return
	}
	if len(orderPath) >= 2 && len(orderPath) <= 3 && orderPath[0] != "" && orderPath[1] == "orders" {
		tail := ""
		if len(orderPath) == 3 {
			tail = orderPath[2]
		}
		a.orders(w, r, orderPath[0], tail)
		return
	}
	descriptionParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/stores/"), "/")
	if len(descriptionParts) == 4 && descriptionParts[0] != "" && descriptionParts[1] == "published-listings" &&
		descriptionParts[2] != "" && descriptionParts[3] == "description" {
		owner, ok := a.requireOwner(w, r)
		if ok {
			a.publishedDescription(w, r, owner, descriptionParts[0], descriptionParts[2])
		}
		return
	}
	if store, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/stores/"), "/published-listings"); ok && !strings.Contains(store, "/") {
		authStart := time.Now()
		owner, ok := a.requireOwner(w, r)
		if ok {
			a.publishedListings(w, r, owner, store, authStart)
		}
		return
	}
	if store, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/stores/"), "/category-attributes"); ok && store != "" && !strings.Contains(store, "/") {
		owner, allowed := a.requireOwner(w, r)
		if allowed {
			a.categoryRules(w, r, owner, store)
		}
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	part := strings.TrimPrefix(r.URL.Path, "/api/v1/stores/")
	pieces := strings.Split(part, "/")
	if len(pieces) != 2 || (pieces[1] != "disconnect" && pieces[1] != "connection-check") {
		writeError(w, http.StatusNotFound, "store not found")
		return
	}
	id := pieces[0]
	if strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "store not found")
		return
	}
	if pieces[1] == "connection-check" {
		authorized, err := a.service.RequireStoreConnection(r.Context(), owner, id)
		switch {
		case errors.Is(err, meli.ErrStoreNotFound):
			writeError(w, http.StatusNotFound, "store not found")
		case errors.Is(err, meli.ErrReauthorizationRequired):
			writeError(w, http.StatusForbidden, "store needs reauthorization")
		case errors.Is(err, meli.ErrRefreshUnknown):
			writeError(w, http.StatusConflict, "refresh outcome unknown; reauthorize or check on Mercado Libre")
		case errors.Is(err, meli.ErrRefreshBusy):
			writeError(w, http.StatusServiceUnavailable, "store token refresh in progress")
		case err != nil:
			writeError(w, http.StatusServiceUnavailable, "store authorization unavailable")
		default:
			w.Header().Set("Cache-Control", "no-store")
			writeJSON(w, http.StatusOK, map[string]any{
				"status": "connected", "seller_id": authorized.SellerID,
				"marketplaces_verified": authorized.Verified,
			})
		}
		return
	}
	err := a.service.Disconnect(r.Context(), owner, id)
	if errors.Is(err, meli.ErrStoreNotFound) {
		writeError(w, http.StatusNotFound, "store not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not disconnect store")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

// The callback is intentionally public. Its high-entropy, single-use state
// binds it to the authenticated owner who initiated the flow.
func (a storeAPI) callback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if a.service == nil {
		http.Error(w, "OAuth not configured", http.StatusServiceUnavailable)
		return
	}
	q := r.URL.Query()
	if len(q["state"]) != 1 || len(q["code"]) > 1 || len(q["error"]) > 1 || len(q.Get("code")) > 4096 {
		http.Error(w, "invalid authorization callback", http.StatusBadRequest)
		return
	}
	_, err := a.service.Complete(r.Context(), q.Get("state"), q.Get("code"), q.Get("error"))
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, meli.ErrSellerOwned) {
			status = http.StatusConflict
		}
		// Never put the provider's error text, access token, code or state in HTML.
		w.WriteHeader(status)
		_, _ = w.Write([]byte("<!doctype html><html lang='zh'><meta charset='utf-8'><title>授权未完成</title><h2>授权未完成</h2><p>请关闭此页面，在 Una Seller 店铺管理中重新授权。</p></html>"))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("<!doctype html><html lang='zh'><meta charset='utf-8'><title>授权成功</title><h2>美客多店铺授权成功</h2><p>可以关闭此页面，返回 Una Seller 点击刷新店铺。</p></html>"))
}

func (a storeAPI) publishedListings(w http.ResponseWriter, r *http.Request, owner, store string, authStart time.Time) {
	readStart := time.Now()
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		writeListingError(w, http.StatusMethodNotAllowed, "listing_method", "method not allowed", nil, meli.ListingSync{})
		return
	}
	query := r.URL.Query()
	scope := meli.ListingScope{Site: query.Get("site"), Logistic: query.Get("logistic_type")}
	if scope.Site == "" {
		scope.Site = "CBT"
	}
	offset := 0
	var expected *int
	restartValue := query.Get("restart")
	restart := r.Method == http.MethodPost && restartValue == "1"
	if query.Has("offset") {
		value, err := strconv.Atoi(query.Get("offset"))
		if err != nil || value < 0 || value > 100000 || (r.Method == http.MethodPost && value >= 1000) {
			writeListingError(w, http.StatusBadRequest, "listing_offset", "invalid listing offset", nil, meli.ListingSync{})
			return
		}
		offset, expected = value, &value
	}
	if restartValue != "" && restartValue != "0" && (!restart || (expected != nil && offset != 0)) {
		writeListingError(w, http.StatusBadRequest, "listing_offset", "invalid listing restart", nil, meli.ListingSync{})
		return
	}
	var skip struct {
		ListingID string `json:"listing_id"`
	}
	skipRequested := query.Has("skip_invalid")
	if skipRequested {
		if r.Method != http.MethodPost || query.Get("skip_invalid") != "1" || restart || expected == nil {
			writeListingError(w, http.StatusBadRequest, "listing_review_invalid", "invalid listing review confirmation", nil, meli.ListingSync{})
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
		if err != nil || json.Unmarshal(body, &skip) != nil || skip.ListingID == "" {
			writeListingError(w, http.StatusBadRequest, "listing_review_invalid", "invalid listing review confirmation", nil, meli.ListingSync{})
			return
		}
	}
	var err error
	var result meli.ListingSync
	if r.Method == http.MethodGet {
		var page meli.ListingPage
		page, err = a.service.ListPublished(r.Context(), owner, store, scope, offset)
		if err == nil {
			setStoreReadTiming(w, authStart, readStart)
			writeJSON(w, http.StatusOK, page)
			return
		}
	} else {
		if skipRequested {
			result, err = a.service.SkipPublishedReview(r.Context(), owner, store, scope, offset, skip.ListingID)
		} else {
			result, err = a.service.SyncPublished(r.Context(), owner, store, scope, expected, restart)
		}
		if err == nil {
			writeJSON(w, http.StatusOK, result)
			return
		}
	}
	if contextErr := r.Context().Err(); contextErr != nil {
		err = contextErr
	}
	status, code, message := http.StatusBadGateway, "listing_unavailable", "published listings unavailable"
	var interruption *meli.ListingInterruption
	switch {
	case errors.Is(err, context.Canceled):
		status, code, message = http.StatusRequestTimeout, "listing_canceled", "listing request canceled"
	case errors.Is(err, context.DeadlineExceeded):
		status, code, message = http.StatusGatewayTimeout, "listing_timeout", "listing request timed out"
	case errors.Is(err, meli.ErrListingReviewChanged):
		status, code, message = http.StatusConflict, "listing_review_changed", "listing review changed or cannot be skipped"
	case errors.Is(err, meli.ErrStoreNotFound):
		status, code, message = http.StatusNotFound, "listing_store_missing", "store not found"
	case errors.Is(err, meli.ErrSiteNotAuthorized):
		status, code, message = http.StatusForbidden, "listing_site_unverified", "marketplace site is not verified"
	case errors.Is(err, meli.ErrReauthorizationRequired):
		status, code, message = http.StatusForbidden, "listing_reauthorization", "store needs reauthorization"
	case errors.Is(err, meli.ErrListingTokenExpired):
		status, code, message = http.StatusConflict, "listing_token_expired", "store token expired; check connection in store management"
	case errors.Is(err, meli.ErrRefreshUnknown):
		status, code, message = http.StatusConflict, "listing_refresh_unknown", "store token refresh outcome unknown"
	case errors.Is(err, meli.ErrListingAuthorizationChanged):
		status, code, message = http.StatusConflict, "listing_auth_changed", "store authorization changed during sync"
	case errors.Is(err, meli.ErrListingCursor):
		status, code, message = http.StatusConflict, "listing_cursor_changed", "listing cursor changed or completed"
	case errors.Is(err, meli.ErrListingBusy):
		status, code, message = http.StatusConflict, "listing_busy", "listing sync already running"
	case errors.As(err, &interruption):
		code = interruption.Code
		message = "Mercado Libre listing sync interrupted"
		switch code {
		case "listing_platform_auth":
			status = http.StatusForbidden
		case "listing_rate_limited":
			status = http.StatusTooManyRequests
		case "listing_timeout":
			status = http.StatusGatewayTimeout
		}
	}
	writeListingError(w, status, code, message, interruption, result)
}

func writeListingError(w http.ResponseWriter, status int, code, message string, interruption *meli.ListingInterruption, result meli.ListingSync) {
	response := map[string]any{"error": message, "code": code}
	if interruption != nil {
		response["platform_status"] = interruption.PlatformStatus
		response["retryable"] = interruption.Retryable
		response["processed"] = result.Processed
		response["skipped"] = result.Skipped
		response["needs_review"] = result.NeedsReview
		if result.Review != nil {
			response["review"] = result.Review
		}
	}
	writeJSON(w, status, response)
}

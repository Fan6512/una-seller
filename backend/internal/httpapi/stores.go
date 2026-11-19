package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

type storeAPI struct {
	auth    collectionAPI
	service *meli.Service
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
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	items, err := a.service.List(r.Context(), owner)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "stores unavailable")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
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
	if store, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/stores/"), "/published-listings"); ok && !strings.Contains(store, "/") {
		owner, ok := a.requireOwner(w, r)
		if ok {
			a.publishedListings(w, r, owner, store)
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

func (a storeAPI) publishedListings(w http.ResponseWriter, r *http.Request, owner, store string) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	query := r.URL.Query()
	scope := meli.ListingScope{Site: query.Get("site"), Logistic: query.Get("logistic_type")}
	if scope.Site == "" {
		scope.Site = "CBT"
	}
	offset := 0
	if query.Has("offset") {
		var err error
		offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil || offset < 0 || offset > 100000 || (r.Method == http.MethodPost && offset >= 1000) {
			writeError(w, http.StatusBadRequest, "invalid listing offset")
			return
		}
	}
	var data any
	var err error
	if r.Method == http.MethodGet {
		data, err = a.service.ListPublished(r.Context(), owner, store, scope, offset)
	} else {
		data, err = a.service.SyncPublished(r.Context(), owner, store, scope, offset)
	}
	if err == nil {
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, http.StatusOK, data)
		return
	}
	var remote meli.ListingPlatformError
	switch {
	case errors.Is(err, meli.ErrStoreNotFound):
		writeError(w, http.StatusNotFound, "store not found")
	case errors.Is(err, meli.ErrSiteNotAuthorized):
		writeError(w, http.StatusForbidden, "marketplace site is not verified")
	case errors.Is(err, meli.ErrReauthorizationRequired):
		writeError(w, http.StatusForbidden, "store needs reauthorization")
	case errors.Is(err, meli.ErrListingTokenExpired):
		writeError(w, http.StatusConflict, "store token expired; check connection in store management")
	case errors.Is(err, meli.ErrRefreshUnknown):
		writeError(w, http.StatusConflict, "store token refresh outcome unknown")
	case errors.Is(err, meli.ErrListingBusy):
		writeError(w, http.StatusConflict, "listing sync already running")
	case errors.As(err, &remote):
		if remote == http.StatusTooManyRequests {
			writeError(w, http.StatusTooManyRequests, "Mercado Libre rate limited listing sync")
		} else {
			writeError(w, http.StatusBadGateway, remote.Error())
		}
	default:
		writeError(w, http.StatusBadGateway, "published listings unavailable")
	}
}

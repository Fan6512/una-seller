package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/cbtcategory"
	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
)

type collectionAPI struct {
	store    *collections.Store
	verifier IdentityVerifier
	accounts *AccountServices
}

func HandlerWithCollections(store *collections.Store, verifier IdentityVerifier, stores ...*meli.Service) http.Handler {
	var service *meli.Service
	if len(stores) != 0 {
		service = stores[0]
	}
	return HandlerWithServices(store, verifier, service, nil, nil)
}

func HandlerWithServices(store *collections.Store, verifier IdentityVerifier, storeService *meli.Service, translator *TranslationService, aiCopy *AICopyService) http.Handler {
	return HandlerWithCategoryDictionary(store, verifier, storeService, translator, nil, aiCopy)
}

func HandlerWithCategoryDictionary(store *collections.Store, verifier IdentityVerifier, storeService *meli.Service, translator *TranslationService, dictionary *cbtcategory.Reader, aiCopy *AICopyService) http.Handler {
	return HandlerWithAccountServices(store, verifier, storeService, translator, dictionary, aiCopy, nil)
}

func HandlerWithAccountServices(store *collections.Store, verifier IdentityVerifier, storeService *meli.Service, translator *TranslationService, dictionary *cbtcategory.Reader, aiCopy *AICopyService, accounts *AccountServices) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", Handler())
	// Public, read-only signed updater manifest. No account or store tokens involved.
	mux.HandleFunc("/api/v1/desktop/updates", desktopUpdaterHandler(nil, desktopReleasesURL))
	api := collectionAPI{store: store, verifier: verifier, accounts: accounts}
	mux.HandleFunc("/api/v1/collections", api.collections)
	mux.HandleFunc("/api/v1/collections/", api.detail)
	mux.HandleFunc("/api/v1/products", api.ownProducts)
	mux.HandleFunc("/api/v1/products/", api.ownProductDetail)
	mux.HandleFunc("/api/v1/products/ai-copy", api.aiCopyHandler(aiCopy))
	mux.HandleFunc("/api/v1/account/service-configs", api.accountServiceConfigs(accounts))
	mux.HandleFunc("/api/v1/account/service-configs/", api.accountServiceConfigAction(accounts))
	s := storeAPI{auth: api, service: storeService, translator: translator, categoryDictionary: dictionary, accounts: accounts}
	if translator != nil {
		s.localizeCategoryAttributes = func(ctx context.Context, rules meli.CategoryRules) map[string]string {
			return translator.translate(ctx, categoryTranslationSource(rules.CategoryID), categoryAttributeTranslationTexts(rules))
		}
	}
	mux.HandleFunc("/api/v1/stores/order-title-translations", s.orderTitleTranslations)
	mux.HandleFunc("/api/v1/categories/predict", s.predictCBTCategory)
	mux.HandleFunc("/api/v1/stores/cbt-category-attributes", s.cbtCategoryRules)
	mux.HandleFunc("/api/v1/stores", s.stores)
	mux.HandleFunc("/api/v1/stores/oauth/start", s.start)
	mux.HandleFunc("/api/v1/stores/", s.storeAction)
	mux.HandleFunc(meli.CallbackPath, s.callback)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || strings.HasPrefix(r.URL.Path, "/api/v1/account/") || strings.HasPrefix(r.URL.Path, "/api/v1/collections") || strings.HasPrefix(r.URL.Path, "/api/v1/stores") || strings.HasPrefix(r.URL.Path, "/api/v1/products") || r.URL.Path == "/api/v1/categories/predict" {
			// Chrome MV3 IDs are per-install; only accept syntactically valid
			// extension origins, and only for the collection API. Store and OAuth
			// routes stay Tauri-only. API access still requires a Bearer token.
			allowOrigin := allowedAppOrigin
			if strings.HasPrefix(r.URL.Path, "/api/v1/collections") {
				allowOrigin = allowedCollectionOrigin
			}
			origin := r.Header.Get("Origin")
			if origin != "" {
				if !allowOrigin(origin) {
					writeError(w, http.StatusForbidden, "origin not allowed")
					return
				}
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				if (strings.HasPrefix(r.URL.Path, "/api/v1/collections") && allowedAppOrigin(origin)) ||
					strings.HasPrefix(r.URL.Path, "/api/v1/products") || strings.HasPrefix(r.URL.Path, "/api/v1/account/") {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				} else {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				}
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				if allowedAppOrigin(origin) && r.Method != http.MethodOptions && timedStoreRoute(r.Method, r.URL.Path) {
					w.Header().Set("Access-Control-Expose-Headers", "Server-Timing")
					w = storeTimingResponse{w, time.Now()}
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func timedStoreRoute(method, path string) bool {
	if path == "/api/v1/stores" {
		return method == http.MethodGet
	}
	part, ok := strings.CutPrefix(path, "/api/v1/stores/")
	if !ok {
		return false
	}
	parts := strings.Split(part, "/")
	return len(parts) == 2 && parts[0] != "" &&
		(method == http.MethodPost && parts[1] == "connection-check" ||
			(method == http.MethodGet || method == http.MethodPost) && parts[1] == "published-listings") ||
		len(parts) == 4 && parts[0] != "" && parts[1] == "published-listings" &&
			parts[2] != "" && parts[3] == "description" && method == http.MethodGet
}

type storeTimingResponse struct {
	http.ResponseWriter
	started time.Time
}

func (w storeTimingResponse) WriteHeader(status int) {
	w.Header().Add("Server-Timing", "api;dur="+strconv.FormatFloat(float64(time.Since(w.started).Microseconds())/1000, 'f', 1, 64))
	w.ResponseWriter.WriteHeader(status)
}

// Forward ResponseController operations to the original writer when supported.
func (w storeTimingResponse) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// allowedAppOrigin only grants CORS transport to the desktop shell, not API access.
// Requests never rely on browser cookies for authentication.
func allowedAppOrigin(origin string) bool {
	switch origin {
	case "http://127.0.0.1:1420", "http://localhost:1420", "http://tauri.localhost", "tauri://localhost", "https://tauri.localhost":
		return true
	}
	return false
}

// allowedCollectionOrigin additionally accepts any syntactically valid Chrome MV3
// extension origin; the extension posts collections without a cookie session.
func allowedCollectionOrigin(origin string) bool {
	if allowedAppOrigin(origin) {
		return true
	}
	const prefix = "chrome-extension://"
	if !strings.HasPrefix(origin, prefix) {
		return false
	}
	id := strings.TrimPrefix(origin, prefix)
	if len(id) != 32 {
		return false
	}
	for _, b := range id {
		if b < 'a' || b > 'p' {
			return false
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	writeError(w, http.StatusUnauthorized, msg)
}
func (a collectionAPI) ready(w http.ResponseWriter) bool {
	if a.store == nil || a.verifier == nil {
		writeError(w, http.StatusServiceUnavailable, "collections require configured database and Supabase Auth")
		return false
	}
	return true
}

// authFastReadRoute explicitly names read-only product endpoints. New routes
// default to remote verification: GET alone does not imply safe caching.
func authFastReadRoute(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	switch r.URL.Path {
	case "/api/v1/collections", "/api/v1/products":
		return true
	}
	if id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/collections/"); ok {
		return collections.ValidUserID(id)
	}
	if id, ok := strings.CutPrefix(r.URL.Path, "/api/v1/products/"); ok {
		return collections.ValidUserID(id)
	}
	return false
}

// Owner identity comes exclusively from a verified Supabase access token.
func (a collectionAPI) owner(w http.ResponseWriter, r *http.Request) (string, bool) {
	if !a.ready(w) {
		return "", false
	}
	fields := strings.Fields(r.Header.Get("Authorization"))
	if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
		writeUnauthorized(w, "authentication required")
		return "", false
	}
	// Fail closed to remote verification on every route not in the allowlist.
	// IdentityVerifier mocks and custom verifiers remain compatible.
	var id string
	var err error
	if authFastReadRoute(r) {
		if fast, ok := a.verifier.(interface {
			VerifyFast(context.Context, string) (string, error)
		}); ok {
			id, err = fast.VerifyFast(r.Context(), fields[1])
		} else {
			id, err = a.verifier.Verify(r.Context(), fields[1])
		}
	} else {
		id, err = a.verifier.Verify(r.Context(), fields[1])
	}
	if err != nil {
		if errors.Is(err, ErrAuthUnavailable) {
			writeError(w, http.StatusServiceUnavailable, "authentication temporarily unavailable")
		} else {
			writeUnauthorized(w, "invalid session")
		}
		return "", false
	}
	if !collections.ValidUserID(id) {
		writeUnauthorized(w, "invalid session")
		return "", false
	}
	return id, true
}
func (a collectionAPI) collections(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := a.owner(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodPost:
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			writeError(w, http.StatusUnsupportedMediaType, "expected application/json")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var in collections.Input
		if err := dec.Decode(&in); err != nil {
			writeError(w, http.StatusBadRequest, "invalid collection JSON")
			return
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			writeError(w, http.StatusBadRequest, "trailing JSON data")
			return
		}
		if _, err := in.Validate(); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		p, created, err := a.store.Save(r.Context(), ownerID, in)
		if errors.Is(err, collections.ErrReadback) {
			// The transaction committed: do not lie to the caller that nothing was saved.
			writeJSON(w, http.StatusAccepted, map[string]any{"result": "saved_unverified", "id": p.ID, "collection_status": p.Status})
			return
		}
		if err != nil {
			if strings.Contains(err.Error(), "title is required") {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeError(w, http.StatusInternalServerError, "collection not saved")
			return
		}
		status := http.StatusOK
		action := "updated"
		if created {
			status = http.StatusCreated
			action = "created"
		}
		writeJSON(w, status, map[string]any{"result": action, "capture_status": in.Status, "collection": p})
	case http.MethodGet:
		limit, offset := 30, 0
		var err error
		if raw := r.URL.Query().Get("limit"); raw != "" {
			limit, err = strconv.Atoi(raw)
			if err != nil {
				writeError(w, 400, "invalid limit")
				return
			}
		}
		if raw := r.URL.Query().Get("offset"); raw != "" {
			offset, err = strconv.Atoi(raw)
			if err != nil {
				writeError(w, 400, "invalid offset")
				return
			}
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if len(q) > 200 || limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
			writeError(w, 400, "invalid query")
			return
		}
		inboxOnly := r.URL.Query().Get("inbox") == "1"
		items, err := a.store.ListWithInboxFilter(r.Context(), ownerID, q, limit, offset, inboxOnly)
		if err != nil {
			writeError(w, 500, "collections unavailable")
			return
		}
		writeJSON(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
func (a collectionAPI) detail(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := a.owner(w, r)
	if !ok {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/approve") {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/collections/"), "/approve")
		if !collections.ValidUserID(id) {
			writeError(w, 404, "collection not found")
			return
		}
		a.approveCollection(w, r, ownerID, id)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/working-draft") {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v1/collections/"), "/working-draft")
		if !collections.ValidUserID(id) {
			writeError(w, http.StatusNotFound, "collection not found")
			return
		}
		a.workingDraft(w, r, ownerID, id)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		w.Header().Set("Allow", "GET, DELETE")
		writeError(w, 405, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/collections/")
	if !collections.ValidUserID(id) {
		writeError(w, 404, "collection not found")
		return
	}
	if r.Method == http.MethodDelete {
		err := a.store.Delete(r.Context(), ownerID, id)
		if errors.Is(err, collections.ErrNotFound) {
			writeError(w, http.StatusNotFound, "collection not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "collection could not be deleted")
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	p, err := a.store.Get(r.Context(), ownerID, id)
	if errors.Is(err, collections.ErrNotFound) {
		writeError(w, 404, "collection not found")
		return
	}
	if err != nil {
		writeError(w, 500, "collection unavailable")
		return
	}
	writeJSON(w, 200, p)
}

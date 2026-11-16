package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/collections"
)

type collectionAPI struct {
	store   *collections.Store
	ownerID string
}

func HandlerWithCollections(store *collections.Store, ownerID string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/healthz", Handler())
	api := collectionAPI{store: store, ownerID: ownerID}
	mux.HandleFunc("/api/v1/collections", api.collections)
	mux.HandleFunc("/api/v1/collections/", api.detail)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/collections") {
			// This is local-only development code; PR-C will add extension pairing.
			origin := r.Header.Get("Origin")
			switch origin {
			case "http://127.0.0.1:1420", "http://localhost:1420", "http://tauri.localhost", "tauri://localhost", "https://tauri.localhost":
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			case "":
			default:
				writeError(w, http.StatusForbidden, "origin not allowed")
				return
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
func (a collectionAPI) ready(w http.ResponseWriter) bool {
	if a.store == nil || !collections.ValidUserID(a.ownerID) {
		writeError(w, http.StatusServiceUnavailable, "collections require configured local database and owner")
		return false
	}
	return true
}
func (a collectionAPI) collections(w http.ResponseWriter, r *http.Request) {
	if !a.ready(w) {
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
		p, created, err := a.store.Save(r.Context(), a.ownerID, in)
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
		items, err := a.store.List(r.Context(), a.ownerID, q, limit, offset)
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
	if !a.ready(w) {
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, 405, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/collections/")
	if !collections.ValidUserID(id) {
		writeError(w, 404, "collection not found")
		return
	}
	p, err := a.store.Get(r.Context(), a.ownerID, id)
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

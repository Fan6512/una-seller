package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/jackc/pgx/v5/pgconn"
)

func (a collectionAPI) ownProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	owner, ok := a.owner(w, r)
	if !ok {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
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
	if len(q) > 200 || limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		writeError(w, 400, "invalid query")
		return
	}
	items, err := a.store.ListOwnProducts(r.Context(), owner, q, limit, offset)
	if err != nil {
		writeError(w, 500, "product box unavailable")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, 200, map[string]any{"items": items, "limit": limit, "offset": offset})
}

func (a collectionAPI) ownProductDetail(w http.ResponseWriter, r *http.Request) {
	owner, ok := a.owner(w, r)
	if !ok {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/products/")
	if !collections.ValidUserID(id) {
		writeError(w, 404, "product not found")
		return
	}
	switch r.Method {
	case http.MethodGet:
		p, err := a.store.GetOwnProduct(r.Context(), owner, id)
		if errors.Is(err, collections.ErrNotFound) {
			writeError(w, 404, "product not found")
			return
		}
		if err != nil {
			writeError(w, 500, "product unavailable")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, p)
	case http.MethodPut:
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			writeError(w, 415, "expected application/json")
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 512<<10)
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		var input collections.OwnProductEdit
		if err := dec.Decode(&input); err != nil {
			writeError(w, 400, "invalid product edit JSON")
			return
		}
		if err := dec.Decode(new(any)); err != io.EOF {
			writeError(w, 400, "trailing product edit JSON")
			return
		}
		if err := input.Validate(); err != nil {
			writeError(w, 400, err.Error())
			return
		}
		p, err := a.store.UpdateOwnProduct(r.Context(), owner, id, input)
		var constraint *pgconn.PgError
		if errors.As(err, &constraint) && constraint.ConstraintName == "cbt_publication_variant_identity" {
			writeError(w, 409, "已关联或待核实的刊登规格必须保留")
			return
		}
		if errors.Is(err, collections.ErrNotFound) {
			writeError(w, 404, "product not found")
			return
		}
		if errors.Is(err, collections.ErrVariantUpgradeRequired) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error(), "code": "VARIANT_CLIENT_UPGRADE_REQUIRED"})
			return
		}
		if errors.Is(err, collections.ErrDraftConflict) {
			writeError(w, 409, "product changed; reload before saving")
			return
		}
		if errors.Is(err, collections.ErrLocalizedTitleTooLong) || errors.Is(err, collections.ErrInvalidTranslationSource) {
			writeError(w, 400, err.Error())
			return
		}
		if errors.Is(err, collections.ErrReadback) {
			writeJSON(w, http.StatusAccepted, map[string]any{"result": "saved_unverified", "id": p.ID, "revision": p.Revision})
			return
		}
		if err != nil {
			writeError(w, 500, "product not saved")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		writeJSON(w, 200, map[string]any{"product": p})
	default:
		w.Header().Set("Allow", "GET, PUT")
		writeError(w, 405, "method not allowed")
	}
}

// Explicit desktop-only promotion. A missing/incorrect draft revision
// receives 409 rather than copying data with unknown saved state.
func (a collectionAPI) approveCollection(w http.ResponseWriter, r *http.Request, owner, id string) {
	if origin := r.Header.Get("Origin"); origin != "" && !allowedAppOrigin(origin) {
		writeError(w, 403, "origin not allowed")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		writeError(w, 405, "method not allowed")
		return
	}
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		writeError(w, 415, "expected application/json")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var input struct {
		ExpectedRevision *int64 `json:"expected_revision"`
	}
	if err := dec.Decode(&input); err != nil {
		writeError(w, 400, "invalid approval JSON")
		return
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "trailing approval JSON")
		return
	}
	if input.ExpectedRevision == nil || *input.ExpectedRevision < 0 {
		writeError(w, 400, "expected_revision required")
		return
	}
	product, created, err := a.store.Promote(r.Context(), owner, id, *input.ExpectedRevision)
	if errors.Is(err, collections.ErrNotFound) {
		writeError(w, 404, "collection not found")
		return
	}
	if errors.Is(err, collections.ErrDraftConflict) {
		writeError(w, 409, "working draft has changed; reload before approval")
		return
	}
	if errors.Is(err, collections.ErrReadback) {
		writeJSON(w, http.StatusAccepted, map[string]any{"result": "saved_unverified", "id": product.ID})
		return
	}
	if err != nil {
		writeError(w, 500, "approval state unknown; reread product box")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if created {
		writeJSON(w, 201, map[string]any{"result": "created", "product": product})
	} else {
		writeJSON(w, 200, map[string]any{"result": "already_approved", "product": product})
	}
}

func (a collectionAPI) productPublication(service *meli.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner, ok := a.owner(w, r)
		if !ok {
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			w.Header().Set("Allow", "GET, POST")
			writeError(w, 405, "method not allowed")
			return
		}
		if service == nil {
			if r.Method == http.MethodGet {
				writeJSON(w, 200, map[string]any{"targets": []meli.PublicationTarget{}})
			} else {
				writeError(w, 503, "刊登服务不可用")
			}
			return
		}
		ctx := r.Context()
		p, err := a.store.GetOwnProduct(ctx, owner, r.PathValue("id"))
		if errors.Is(err, collections.ErrNotFound) {
			writeError(w, 404, "product not found")
			return
		}
		if err != nil {
			writeError(w, 500, "product unavailable")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if r.Method == http.MethodGet {
			targets, err := service.PublicationTargets(ctx, owner, p)
			if err != nil {
				writeError(w, 503, "刊登状态暂时不可用")
				return
			}
			writeJSON(w, 200, map[string]any{"targets": targets})
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && !allowedAppOrigin(origin) {
			writeError(w, 403, "origin not allowed")
			return
		}
		if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
			writeError(w, 415, "expected application/json")
			return
		}
		var input struct {
			meli.PublicationConfirmation
			Verify bool `json:"verify"`
		}
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		dec.DisallowUnknownFields()
		if dec.Decode(&input) != nil || dec.Decode(new(any)) != io.EOF {
			writeError(w, 400, "invalid publication JSON")
			return
		}
		var result meli.PublicationResult
		if input.Verify {
			if input.Revision != 0 || input.TitleMode != "" {
				writeError(w, 400, "invalid verification request")
				return
			}
			result, err = service.VerifyPublication(ctx, owner, p.ID, input.StoreID, input.Site)
		} else {
			result, err = service.SubmitPublication(ctx, owner, p, input.PublicationConfirmation)
		}
		var invalid meli.PublicationInputError
		switch {
		case err == nil:
			writeJSON(w, 200, result)
		case errors.As(err, &invalid):
			writeError(w, 422, invalid.Error())
		case errors.Is(err, meli.ErrPublicationConflict):
			writeError(w, 409, "产品修订或目标已变化，请重新加载核对")
		case errors.Is(err, meli.ErrPublicationNotReady):
			writeError(w, 422, "目标授权、刊登额度、官方图片或类目规则尚未核实")
		case errors.Is(err, meli.ErrStoreNotFound), errors.Is(err, meli.ErrSiteNotAuthorized), errors.Is(err, meli.ErrReauthorizationRequired):
			writeError(w, 422, "目标店铺授权不可用")
		default:
			writeError(w, 503, "刊登状态待核实，请重新读取已有尝试")
		}
	}
}

package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

// orderPurchase stores only a manually verified purchase number and CNY cost.
// It neither accesses 1688 nor submits a shipment to Mercado Libre.
func (a storeAPI) orderPurchase(w http.ResponseWriter, r *http.Request, store, orderID string) {
	w.Header().Set("Cache-Control", "no-store")
	owner, ok := a.requireOwner(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "order_method_not_allowed"})
		return
	}
	q := r.URL.Query()
	if q.Get("site") == "" || q.Get("logistic_type") == "" ||
		len(q["site"]) != 1 || len(q["logistic_type"]) != 1 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_scope_required"})
		return
	}
	scope := meli.OrderScope{Site: q.Get("site"), Logistic: q.Get("logistic_type")}
	var record *meli.OrderPurchase
	var err error
	if r.Method == http.MethodGet {
		record, err = a.service.GetOrderPurchase(r.Context(), owner, store, scope, orderID)
	} else {
		var request struct {
			PurchaseOrderID string `json:"purchase_order_id"`
			CostCNY         string `json:"cost_cny"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF ||
			request.PurchaseOrderID == "" || strings.TrimSpace(request.PurchaseOrderID) != request.PurchaseOrderID {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_purchase_invalid"})
			return
		}
		record, err = a.service.SaveOrderPurchase(r.Context(), owner, store, scope, orderID,
			request.PurchaseOrderID, request.CostCNY)
	}
	switch {
	case errors.Is(err, meli.ErrOrderPurchaseInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_purchase_invalid"})
	case errors.Is(err, meli.ErrStoreNotFound), errors.Is(err, meli.ErrOrderNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "order_not_found"})
	case errors.Is(err, meli.ErrSiteNotAuthorized):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "order_site_not_authorized"})
	case err != nil:
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "order_unavailable"})
	default:
		writeJSON(w, http.StatusOK, map[string]any{"purchase": record})
	}
}

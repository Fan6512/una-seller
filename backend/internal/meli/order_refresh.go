package meli

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// RefreshOneOrder re-reads only an existing CBT order from the official
// /marketplace/orders/{id} endpoint. It does not advance the bulk sync cursor,
// overwrite procurement, nor infer a financial settlement from sale amount.
func (s *Service) RefreshOneOrder(ctx context.Context, owner, store string, scope OrderScope, id string) (OrderRecord, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !orderNumber(id) {
		return OrderRecord{}, ErrOrderNotFound
	}
	if _, err := s.GetOrder(ctx, owner, store, scope, id); err != nil {
		return OrderRecord{}, err
	}
	auth, childID, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderRecord{}, err
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var raw json.RawMessage
	status, err := s.orderGET(fetchCtx, auth.AccessToken, "/marketplace/orders/"+id, &raw)
	if status == http.StatusNotFound {
		return OrderRecord{}, ErrOrderNotFound
	}
	if err != nil {
		return OrderRecord{}, err
	}
	record, err := parseOrderDetail(raw, id, childID, scope.Site)
	if err != nil {
		return OrderRecord{}, err
	}
	items, err := json.Marshal(record.Items)
	if err != nil {
		return OrderRecord{}, err
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrderRecord{}, err
	}
	defer tx.Rollback(ctx)
	var seller, connectionStatus, mappingStatus string
	var rawMarkets []byte
	err = tx.QueryRow(ctx, `SELECT seller_id,status,marketplaces_status,marketplaces
 FROM app.meli_store_connections
 WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR SHARE`,
		store, owner).Scan(&seller, &connectionStatus, &mappingStatus, &rawMarkets)
	if err != nil {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	var markets []Marketplace
	if json.Unmarshal(rawMarkets, &markets) != nil ||
		seller != auth.SellerID || connectionStatus != "connected" || mappingStatus != "verified" {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	allowed := false
	for _, market := range markets {
		if market.SiteID == scope.Site && market.LogisticType == scope.Logistic &&
			market.UserID.String() == childID {
			allowed = true
			break
		}
	}
	if !allowed {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	_, err = tx.Exec(ctx, `UPDATE app.meli_orders SET
 pack_id=NULLIF($6,''),status=$7,currency_id=$8,total_amount=$9::numeric,
 last_updated=$10,shipment_id=NULLIF($11,''),order_items=$12::jsonb,updated_at=now()
 WHERE store_id=$1::uuid AND site_id=$3 AND logistic_type=$4
 AND seller_id=$5 AND order_id=$2 AND last_updated <= $10`,
		store, id, scope.Site, scope.Logistic, childID,
		record.PackID, record.Status, record.CurrencyID, record.TotalAmount,
		record.LastUpdated, record.ShipmentID, string(items))
	if err != nil {
		return OrderRecord{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OrderRecord{}, err
	}
	// Return the committed latest local snapshot, even when the platform
	// response was older than a concurrent update.
	return s.GetOrder(ctx, owner, store, scope, id)
}

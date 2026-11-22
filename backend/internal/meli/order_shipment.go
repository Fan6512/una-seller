package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode"
)

var ErrOrderShipmentMissing = errors.New("shipment missing or unavailable")
var ErrOrderShipmentDenied = errors.New("official shipment access denied")

type OrderShipment struct {
	ID             string     `json:"shipment_id"`
	Status         string     `json:"status"`
	Substatus      string     `json:"substatus,omitempty"`
	TrackingNumber string     `json:"tracking_number,omitempty"`
	LogisticType   string     `json:"logistic_type,omitempty"`
	LastUpdated    *time.Time `json:"last_updated,omitempty"`
	CheckedAt      time.Time  `json:"checked_at"`
}

func (s *Service) RefreshOrderShipment(ctx context.Context, owner, store string, scope OrderScope, orderID string) (OrderRecord, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !orderNumber(orderID) {
		return OrderRecord{}, ErrOrderNotFound
	}
	order, err := s.GetOrder(ctx, owner, store, scope, orderID)
	if err != nil {
		return OrderRecord{}, err
	}
	if !orderSLAShipmentID.MatchString(order.ShipmentID) {
		return OrderRecord{}, ErrOrderShipmentMissing
	}
	auth, childID, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderRecord{}, err
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var payload struct {
		ID             json.Number `json:"id"`
		Status         string      `json:"status"`
		Substatus      string      `json:"substatus"`
		TrackingNumber string      `json:"tracking_number"`
		LogisticType   string      `json:"logistic_type"`
		LastUpdated    *time.Time  `json:"last_updated"`
		Logistic       struct {
			Type string `json:"type"`
		} `json:"logistic"`
	}
	code, err := s.orderGET(fetchCtx, auth.AccessToken, "/marketplace/shipments/"+order.ShipmentID, &payload)
	switch code {
	case http.StatusNotFound:
		return OrderRecord{}, ErrOrderShipmentMissing
	case http.StatusUnauthorized, http.StatusForbidden:
		return OrderRecord{}, ErrOrderShipmentDenied
	}
	if err != nil {
		return OrderRecord{}, err
	}
	if payload.LogisticType == "" {
		payload.LogisticType = payload.Logistic.Type
	}
	if code != http.StatusOK || payload.ID.String() != order.ShipmentID || payload.Status == "" ||
		!ValidOrderStatus(payload.Status) || !ValidOrderStatus(payload.Substatus) ||
		!ValidOrderStatus(payload.LogisticType) || len(payload.TrackingNumber) > 200 ||
		strings.IndexFunc(payload.TrackingNumber, unicode.IsControl) >= 0 {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	snapshot := OrderShipment{
		ID: order.ShipmentID, Status: payload.Status, Substatus: payload.Substatus,
		TrackingNumber: payload.TrackingNumber, LogisticType: payload.LogisticType,
		LastUpdated: payload.LastUpdated, CheckedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		return OrderRecord{}, err
	}
	// Lock the store mapping while checking the order's current shipment. A
	// late response cannot attach to a replaced shipment or a revoked mapping.
	tag, err := s.db.Exec(ctx, `WITH authorized AS (
 SELECT c.id FROM app.meli_store_connections c
 WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid
   AND c.status='connected' AND c.marketplaces_status='verified' AND c.seller_id=$9
   AND EXISTS (
     SELECT 1 FROM jsonb_array_elements(c.marketplaces) AS m(value)
     WHERE m.value->>'site_id'=$3 AND m.value->>'logistic_type'=$4 AND m.value->>'user_id'=$5
   )
 FOR SHARE
)
UPDATE app.meli_orders o SET shipment_snapshot=CASE
 WHEN o.shipment_snapshot->>'shipment_id'=$7
   AND (o.shipment_snapshot->>'last_updated')::timestamptz IS NOT NULL
   AND ($10::timestamptz IS NULL OR (o.shipment_snapshot->>'last_updated')::timestamptz > $10)
 THEN o.shipment_snapshot
 ELSE $8::jsonb END
FROM authorized c
WHERE o.store_id=c.id AND o.site_id=$3 AND o.logistic_type=$4 AND o.seller_id=$5
 AND o.order_id=$6 AND o.shipment_id=$7`,
		store, owner, scope.Site, scope.Logistic, childID, orderID, order.ShipmentID, data,
		auth.SellerID, payload.LastUpdated)
	if err != nil {
		return OrderRecord{}, err
	}
	if tag.RowsAffected() != 1 {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	return s.GetOrder(ctx, owner, store, scope, orderID)
}

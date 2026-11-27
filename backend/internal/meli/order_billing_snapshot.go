package meli

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// storedOrderBilling retains only the already-sanitized preview, with its
// shipment identity and refresh start timestamp for late-response protection.
type storedOrderBilling struct {
	ShipmentID string              `json:"shipment_id"`
	StartedAt  time.Time           `json:"started_at"`
	Preview    BillingOrderPreview `json:"preview"`
}

func (s *Service) saveOrderBillingSnapshot(ctx context.Context, owner, store string, scope OrderScope,
	child, seller, shipment string, startedAt time.Time, item BillingOrderPreview) (BillingOrderPreview, error) {
	checkedAt := time.Now().UTC()
	item.CheckedAt = &checkedAt
	data, err := json.Marshal(storedOrderBilling{ShipmentID: shipment, StartedAt: startedAt, Preview: item})
	if err != nil {
		return BillingOrderPreview{}, err
	}
	// Match the current verified store mapping and shipment at write time.
	// Older in-flight requests must not overwrite newer successful snapshots.
	var saved []byte
	err = s.db.QueryRow(ctx, `WITH authorized AS (
 SELECT c.id FROM app.meli_store_connections c
 WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid
   AND c.status='connected' AND c.marketplaces_status='verified' AND c.seller_id=$9
   AND EXISTS (
     SELECT 1 FROM jsonb_array_elements(c.marketplaces) AS m(value)
     WHERE m.value->>'site_id'=$3 AND m.value->>'logistic_type'=$4 AND m.value->>'user_id'=$5
   )
 FOR SHARE
)
UPDATE app.meli_orders o SET billing_snapshot=CASE
 WHEN o.billing_snapshot->>'shipment_id'=$7
   AND (o.billing_snapshot->>'started_at')::timestamptz >= $10::timestamptz
 THEN o.billing_snapshot ELSE $8::jsonb END
FROM authorized c
WHERE o.store_id=c.id AND o.site_id=$3 AND o.logistic_type=$4 AND o.seller_id=$5
 AND o.order_id=$6 AND COALESCE(o.shipment_id,'')=$7
RETURNING o.billing_snapshot`,
		store, owner, scope.Site, scope.Logistic, child, item.OrderID, shipment, data, seller, startedAt).Scan(&saved)
	if errors.Is(err, pgx.ErrNoRows) {
		return BillingOrderPreview{}, ErrOrderAuthorizationChanged
	}
	if err != nil {
		return BillingOrderPreview{}, err
	}
	var snapshot storedOrderBilling
	if err := json.Unmarshal(saved, &snapshot); err != nil || snapshot.Preview.OrderID != item.OrderID ||
		!snapshot.Preview.Available || snapshot.Preview.CheckedAt == nil || snapshot.ShipmentID != shipment {
		return BillingOrderPreview{}, ErrOrderInvalidResponse
	}
	return snapshot.Preview, nil
}

// GetCachedOrderBilling performs only a scoped PostgreSQL read. No OAuth
// refresh, official request, or calculation of unverified net proceeds.
func (s *Service) GetCachedOrderBilling(ctx context.Context, owner, store string, scope OrderScope,
	orderID string) (BillingPreviewPage, error) {
	if !validUUID(store) || !validUUID(owner) || !orderScopeValid(scope) || !billingIDsValid([]string{orderID}) {
		return BillingPreviewPage{}, ErrOrderBillingInvalid
	}
	var saved []byte
	err := s.db.QueryRow(ctx, `WITH connection AS (`+orderConnection+`)
 SELECT CASE WHEN o.billing_snapshot->>'shipment_id'=COALESCE(o.shipment_id,'')
 THEN o.billing_snapshot ELSE NULL END
 FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 WHERE c.status='connected' AND o.site_id=$3 AND o.logistic_type=$4 AND o.order_id=$5`,
		store, owner, scope.Site, scope.Logistic, orderID).Scan(&saved)
	if errors.Is(err, pgx.ErrNoRows) {
		return BillingPreviewPage{}, ErrOrderNotFound
	}
	if err != nil {
		return BillingPreviewPage{}, err
	}
	if len(saved) == 0 {
		return BillingPreviewPage{Items: []BillingOrderPreview{{
			OrderID: orderID, Reconciliation: "not_cached",
			Payments: []BillingPayment{}, Taxes: []BillingTax{}, Charges: []BillingCharge{},
		}}}, nil
	}
	var snapshot storedOrderBilling
	if err := json.Unmarshal(saved, &snapshot); err != nil || snapshot.Preview.OrderID != orderID ||
		!snapshot.Preview.Available || snapshot.Preview.CheckedAt == nil {
		return BillingPreviewPage{}, ErrOrderInvalidResponse
	}
	return BillingPreviewPage{Items: []BillingOrderPreview{snapshot.Preview}}, nil
}

package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrOrderSLANoShipment = errors.New("order has no shipment")
var ErrOrderSLAInvalid = errors.New("invalid dispatch SLA response")
var ErrOrderSLANotApplicable = errors.New("dispatch SLA not applicable")
var orderSLAShipmentID = regexp.MustCompile(`^[1-9][0-9]{0,29}$`)
var orderSLAStatus = regexp.MustCompile(`^[a-z_]{1,40}$`)

// RefreshOrderDispatchSLA fetches the actual official CBT shipping cutoff.
// This is an explicit operation, separate from bulk order sync and the buyer
// delivery estimate. It never marks a shipment dispatched.
func (s *Service) RefreshOrderDispatchSLA(ctx context.Context, owner, store string, scope OrderScope, orderID string) (OrderRecord, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !orderNumber(orderID) {
		return OrderRecord{}, ErrOrderNotFound
	}
	order, err := s.getOrder(ctx, owner, store, scope, orderID, false)
	if err != nil {
		return OrderRecord{}, err
	}
	if order.Status == "cancelled" || order.Status == "cancellation_requested" || order.Status == "refunded" ||
		strings.EqualFold(scope.Logistic, "fulfillment") || strings.EqualFold(scope.Logistic, "fulfilment") {
		return OrderRecord{}, ErrOrderSLANotApplicable
	}
	if !orderSLAShipmentID.MatchString(order.ShipmentID) {
		return OrderRecord{}, ErrOrderSLANoShipment
	}
	auth, childID, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderRecord{}, err
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var raw json.RawMessage
	statusCode, err := s.orderGET(fetchCtx, auth.AccessToken,
		"/marketplace/shipments/"+order.ShipmentID+"/sla", &raw)
	if statusCode == http.StatusNotFound {
		return OrderRecord{}, ErrOrderSLANoShipment
	}
	if err != nil {
		return OrderRecord{}, err
	}
	var payload struct {
		Status       string `json:"status"`
		ExpectedDate string `json:"expected_date"`
	}
	if json.Unmarshal(raw, &payload) != nil || !orderSLAStatus.MatchString(payload.Status) ||
		(payload.Status != "on_time" && payload.Status != "delayed" &&
			payload.Status != "early" && payload.Status != "insuficient_info") {
		return OrderRecord{}, ErrOrderSLAInvalid
	}
	var cutoff *time.Time
	if payload.ExpectedDate != "" {
		when, parseErr := time.Parse(time.RFC3339Nano, payload.ExpectedDate)
		if parseErr != nil {
			return OrderRecord{}, ErrOrderSLAInvalid
		}
		utc := when.UTC()
		cutoff = &utc
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrderRecord{}, err
	}
	defer tx.Rollback(ctx)
	var currentShipment string
	err = tx.QueryRow(ctx, `SELECT o.shipment_id
 FROM app.meli_orders o
 JOIN app.meli_store_connections c ON c.id=o.store_id
   AND c.owner_user_id=$2::uuid AND c.status='connected'
   AND c.marketplaces_status='verified'
 WHERE o.store_id=$1::uuid AND o.site_id=$3 AND o.logistic_type=$4
   AND o.seller_id=$5 AND o.order_id=$6
   AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) AS m(value)
     WHERE m.value->>'site_id'=$3 AND m.value->>'logistic_type'=$4
       AND m.value->>'user_id'=$5)
 FOR SHARE OF c FOR NO KEY UPDATE OF o`,
		store, owner, scope.Site, scope.Logistic, childID, orderID).Scan(&currentShipment)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	if err != nil {
		return OrderRecord{}, err
	}
	if currentShipment != order.ShipmentID {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.meli_order_dispatch_sla
 (store_id,site_id,logistic_type,seller_id,order_id,shipment_id,sla_status,expected_date)
 VALUES ($1::uuid,$2,$3,$4,$5,$6,$7,$8)
 ON CONFLICT (store_id,site_id,logistic_type,seller_id,order_id)
 DO UPDATE SET shipment_id=EXCLUDED.shipment_id,sla_status=EXCLUDED.sla_status,
 expected_date=EXCLUDED.expected_date,checked_at=now()`,
		store, scope.Site, scope.Logistic, childID, orderID,
		order.ShipmentID, payload.Status, cutoff)
	if err != nil {
		return OrderRecord{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OrderRecord{}, err
	}
	return s.getOrder(ctx, owner, store, scope, orderID, false)
}

package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
	"unicode"
)

var ErrOrderShipmentMissing = errors.New("shipment missing or unavailable")
var ErrOrderShipmentDenied = errors.New("official shipment access denied")
var ErrOrderShipmentChanged = errors.New("shipment changed during refresh")

type OrderShipment struct {
	ID               string                `json:"shipment_id"`
	Status           string                `json:"status"`
	Substatus        string                `json:"substatus,omitempty"`
	TrackingNumber   string                `json:"tracking_number,omitempty"`
	LogisticType     string                `json:"logistic_type,omitempty"`
	FirstMileCarrier string                `json:"first_mile_carrier,omitempty"`
	ShippedAt        *time.Time            `json:"shipped_at,omitempty"`
	DeliveredAt      *time.Time            `json:"delivered_at,omitempty"`
	LastUpdated      *time.Time            `json:"last_updated,omitempty"`
	CheckedAt        time.Time             `json:"checked_at"`
	Details          *OrderShipmentDetails `json:"details,omitempty"`
}

type OrderShipmentDetails struct {
	CarrierName string               `json:"carrier_name,omitempty"`
	TrackingURL string               `json:"tracking_url,omitempty"`
	History     []OrderShipmentEvent `json:"history"`
}

type OrderShipmentEvent struct {
	Status    string    `json:"status"`
	Substatus string    `json:"substatus,omitempty"`
	Date      time.Time `json:"date"`
}

func (s *Service) RefreshOrderShipment(ctx context.Context, owner, store string, scope OrderScope, orderID string, withDetails bool) (OrderRecord, error) {
	checkedAt := time.Now().UTC().Truncate(time.Microsecond)
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
		ID               json.Number `json:"id"`
		Status           string      `json:"status"`
		Substatus        string      `json:"substatus"`
		TrackingNumber   string      `json:"tracking_number"`
		LogisticType     string      `json:"logistic_type"`
		FirstMileCarrier string      `json:"fm_carrier"`
		LastUpdated      *time.Time  `json:"last_updated"`
		Logistic         struct {
			Type string `json:"type"`
		} `json:"logistic"`
		StatusHistory struct {
			Shipped   string `json:"date_shipped"`
			Delivered string `json:"date_delivered"`
		} `json:"status_history"`
	}
	path := "/marketplace/shipments/" + order.ShipmentID
	code, err := s.orderGET(fetchCtx, auth.AccessToken, path, &payload)
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
		FirstMileCarrier: shipmentCarrierName(payload.FirstMileCarrier),
		ShippedAt:        shipmentEventTime(payload.StatusHistory.Shipped),
		DeliveredAt:      shipmentEventTime(payload.StatusHistory.Delivered),
		LastUpdated:      payload.LastUpdated, CheckedAt: checkedAt,
	}
	var details *OrderShipmentDetails
	if withDetails {
		details, err = s.fetchOrderShipmentDetails(fetchCtx, auth.AccessToken, path)
		if err != nil {
			return OrderRecord{}, err
		}
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
   AND (
     (o.shipment_snapshot->>'last_updated')::timestamptz > $10
     OR ($10::timestamptz IS NULL AND (o.shipment_snapshot->>'last_updated') IS NOT NULL)
     OR ((o.shipment_snapshot->>'last_updated')::timestamptz IS NOT DISTINCT FROM $10::timestamptz
         AND (o.shipment_snapshot->>'checked_at')::timestamptz > $11)
   )
 THEN o.shipment_snapshot
 ELSE $8::jsonb END
FROM authorized c
WHERE o.store_id=c.id AND o.site_id=$3 AND o.logistic_type=$4 AND o.seller_id=$5
 AND o.order_id=$6 AND o.shipment_id=$7`,
		store, owner, scope.Site, scope.Logistic, childID, orderID, order.ShipmentID, data,
		auth.SellerID, payload.LastUpdated, checkedAt)
	if err != nil {
		return OrderRecord{}, err
	}
	if tag.RowsAffected() != 1 {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	result, err := s.GetOrder(ctx, owner, store, scope, orderID)
	if err != nil {
		return OrderRecord{}, err
	}
	if result.ShipmentID != order.ShipmentID || result.Shipment == nil ||
		(details != nil && !result.Shipment.CheckedAt.Equal(checkedAt)) {
		return OrderRecord{}, ErrOrderShipmentChanged
	}
	if details != nil {
		result.Shipment.Details = details
		for _, event := range details.History {
			// A timestamp for a substatus update cannot date the status transition.
			if event.Substatus != "" {
				continue
			}
			switch {
			case event.Status == "shipped" && result.Shipment.ShippedAt == nil:
				result.Shipment.ShippedAt = &event.Date
			case event.Status == "delivered" && result.Shipment.DeliveredAt == nil:
				result.Shipment.DeliveredAt = &event.Date
			}
		}
	}
	return result, nil
}

func (s *Service) fetchOrderShipmentDetails(ctx context.Context, token, path string) (*OrderShipmentDetails, error) {
	var carrier struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	var history []struct {
		Status    string `json:"status"`
		Substatus string `json:"substatus"`
		Date      string `json:"date"`
	}
	for _, resource := range []struct {
		suffix string
		result any
	}{{"/carrier", &carrier}, {"/history", &history}} {
		code, err := s.orderGET(ctx, token, path+resource.suffix, resource.result)
		switch code {
		case http.StatusNotFound:
			continue
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, ErrOrderShipmentDenied
		}
		if err != nil {
			return nil, err
		}
		if code != http.StatusOK {
			return nil, ErrOrderInvalidResponse
		}
	}
	parsed, err := url.Parse(carrier.URL)
	trackingURL := ""
	if err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") &&
		parsed.Hostname() != "" && parsed.User == nil && !strings.Contains(carrier.URL, "\\") {
		trackingURL = parsed.String()
	}
	var valid []OrderShipmentEvent
	if history != nil {
		valid = make([]OrderShipmentEvent, 0, len(history))
	}
	for _, event := range history {
		at := shipmentEventTime(event.Date)
		if at == nil || event.Status == "" || !ValidOrderStatus(event.Status) || !ValidOrderStatus(event.Substatus) {
			continue
		}
		valid = append(valid, OrderShipmentEvent{Status: event.Status, Substatus: event.Substatus, Date: *at})
	}
	slices.SortStableFunc(valid, func(a, b OrderShipmentEvent) int {
		return a.Date.Compare(b.Date)
	})
	return &OrderShipmentDetails{CarrierName: shipmentCarrierName(carrier.Name), TrackingURL: trackingURL, History: valid}, nil
}

func shipmentCarrierName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) > 200 || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return ""
	}
	return name
}

func shipmentEventTime(value string) *time.Time {
	at, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || at.IsZero() {
		return nil
	}
	return &at
}

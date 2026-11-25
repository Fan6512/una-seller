package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
)

const maxNotificationShipmentOrders = 10

// NotificationHandlers are read-only upstream adapters. The worker owns the
// lease, final scope checks, local transaction and retry policy.
func (s *Service) NotificationHandlers() map[string]notifications.Prepare {
	return map[string]notifications.Prepare{
		"marketplace_orders":    s.prepareOrderNotification,
		"marketplace_shipments": s.prepareShipmentNotification,
	}
}

type notifiedOrder struct {
	scope  OrderScope
	seller string
	record OrderRecord
}

func notificationAuthError(err error) error {
	var limited *OrderRateLimitError
	if errors.As(err, &limited) {
		return &notifications.RateLimited{RetryAfter: time.Until(limited.until)}
	}
	return err
}

func notificationStoreMatches(auth AuthorizedStore, job notifications.Job) bool {
	if !auth.Verified || auth.SellerID != job.ParentSellerID {
		return false
	}
	if job.NotificationUserID == auth.SellerID {
		return true
	}
	for _, market := range auth.Markets {
		if market.UserID.String() == job.NotificationUserID {
			return true
		}
	}
	return false
}

func notificationChildMatches(parent, user, child string) bool {
	return user == parent || user == child
}

// Match the official order's exact site and seller against the current
// marketplace mapping. Multiple logistic modes for one child/site are
// ambiguous: do not guess from shipment logistic.type or a default mode.
func resolveNotifiedOrder(raw json.RawMessage, id, user string, auth AuthorizedStore) (notifiedOrder, error) {
	var result notifiedOrder
	matches := 0
	for _, market := range auth.Markets {
		child := market.UserID.String()
		scope := OrderScope{Site: market.SiteID, Logistic: market.LogisticType}
		if !orderScopeValid(scope) || !orderNumber(child) || !notificationChildMatches(auth.SellerID, user, child) {
			continue
		}
		record, err := parseOrderDetail(raw, id, child, scope.Site)
		if err != nil {
			continue
		}
		result = notifiedOrder{scope: scope, seller: child, record: record}
		matches++
	}
	if matches != 1 {
		return notifiedOrder{}, notifications.ErrScope
	}
	return result, nil
}

func (s *Service) readNotifiedOrder(ctx context.Context, auth AuthorizedStore, user, id string) (notifiedOrder, error) {
	if !orderNumber(id) {
		return notifiedOrder{}, ErrOrderInvalidResponse
	}
	var raw json.RawMessage
	status, err := s.orderGET(ctx, auth.AccessToken, "/marketplace/orders/"+id, &raw)
	if err != nil {
		return notifiedOrder{}, notificationAuthError(err)
	}
	if status != http.StatusOK {
		return notifiedOrder{}, ErrOrderInvalidResponse
	}
	return resolveNotifiedOrder(raw, id, user, auth)
}

// Recheck exact site/logistics/child AND uniqueness in the worker transaction.
// The worker holds the parent mapping lock and checks the notification identity;
// this check protects against mapping changes between official read and commit.
func checkNotifiedOrderScope(ctx context.Context, tx notifications.LocalWriter, job notifications.Job, order notifiedOrder) error {
	var raw []byte
	err := tx.QueryRow(ctx, `SELECT marketplaces FROM app.meli_store_connections
 WHERE id=$1::uuid AND owner_user_id=$2::uuid AND seller_id=$3
 AND status='connected' AND marketplaces_status='verified' FOR SHARE`,
		job.StoreID, job.OwnerID, job.ParentSellerID).Scan(&raw)
	if err != nil {
		return err
	}
	var markets []Marketplace
	if json.Unmarshal(raw, &markets) != nil {
		return notifications.ErrScope
	}
	matches := 0
	exact := false
	for _, market := range markets {
		if market.SiteID == order.scope.Site && market.UserID.String() == order.seller {
			matches++
			exact = exact || market.LogisticType == order.scope.Logistic
		}
	}
	if matches != 1 || !exact || !notificationChildMatches(job.ParentSellerID, job.NotificationUserID, order.seller) {
		return notifications.ErrScope
	}
	return nil
}

func (s *Service) prepareOrderNotification(ctx context.Context, job notifications.Job) (notifications.VerifiedResource, error) {
	if job.Topic != "marketplace_orders" || !orderNumber(job.ResourceID) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	auth, err := s.RequireStoreConnection(ctx, job.OwnerID, job.StoreID)
	if err != nil {
		return notifications.VerifiedResource{}, err
	}
	if !notificationStoreMatches(auth, job) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	order, err := s.readNotifiedOrder(ctx, auth, job.NotificationUserID, job.ResourceID)
	if err != nil {
		return notifications.VerifiedResource{}, err
	}
	return notifications.VerifiedResource{
		Topic: job.Topic, ResourceID: order.record.ID, SellerID: order.seller,
		Apply: func(ctx context.Context, tx notifications.LocalWriter) error {
			if err := checkNotifiedOrderScope(ctx, tx, job, order); err != nil {
				return err
			}
			return upsertOrderRecord(ctx, tx, job.StoreID, order.scope, order.seller, order.record)
		},
	}, nil
}

type shipmentOrderLink struct {
	OrderID  json.Number `json:"order_id"`
	PackID   json.Number `json:"pack_id"`
	SellerID json.Number `json:"seller_id"`
}

// A shipment can precede its orders. Use the documented hosted association
// endpoint, never the removed shipment.order_id or a guessed pack identifier.
func (s *Service) readShipmentOrderLinks(ctx context.Context, token, shipmentID string) ([]shipmentOrderLink, error) {
	var rows []shipmentOrderLink
	status, err := s.orderGET(ctx, token, "/shipments/"+shipmentID+"/orders", &rows)
	if err != nil {
		return nil, notificationAuthError(err)
	}
	if status != http.StatusOK || len(rows) == 0 || len(rows) > 1000 {
		return nil, ErrOrderInvalidResponse
	}
	seen := make(map[string]shipmentOrderLink)
	for _, row := range rows {
		id := row.OrderID.String()
		if !orderNumber(id) || !orderNumber(row.SellerID.String()) || (row.PackID != "" && !orderNumber(row.PackID.String())) {
			return nil, ErrOrderInvalidResponse
		}
		if old, ok := seen[id]; ok && old != row {
			return nil, ErrOrderInvalidResponse
		}
		seen[id] = row
	}
	if len(seen) > maxNotificationShipmentOrders {
		return nil, ErrOrderInvalidResponse
	}
	result := make([]shipmentOrderLink, 0, len(seen))
	for _, row := range seen {
		result = append(result, row)
	}
	// Stable order also keeps concurrent row locks deterministic.
	slices.SortFunc(result, func(a, b shipmentOrderLink) int {
		if a.OrderID.String() < b.OrderID.String() {
			return -1
		}
		if a.OrderID.String() > b.OrderID.String() {
			return 1
		}
		return 0
	})
	return result, nil
}

func (s *Service) prepareShipmentNotification(ctx context.Context, job notifications.Job) (notifications.VerifiedResource, error) {
	if job.Topic != "marketplace_shipments" || !orderSLAShipmentID.MatchString(job.ResourceID) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	auth, err := s.RequireStoreConnection(ctx, job.OwnerID, job.StoreID)
	if err != nil {
		return notifications.VerifiedResource{}, err
	}
	if !notificationStoreMatches(auth, job) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	checkedAt := time.Now().UTC().Truncate(time.Microsecond)
	snapshot, sender, site, err := s.fetchShipmentSnapshot(ctx, auth.AccessToken, job.ResourceID, checkedAt, false)
	if err != nil {
		return notifications.VerifiedResource{}, notificationAuthError(err)
	}
	links, err := s.readShipmentOrderLinks(ctx, auth.AccessToken, job.ResourceID)
	if err != nil {
		return notifications.VerifiedResource{}, err
	}
	orders := make([]notifiedOrder, 0, len(links))
	for _, link := range links {
		order, err := s.readNotifiedOrder(ctx, auth, job.NotificationUserID, link.OrderID.String())
		if err != nil {
			return notifications.VerifiedResource{}, err
		}
		// All orders must establish the same forward shipment, seller, site
		// and exact logistics scope. Mixed/ambiguous shipments fail closed.
		if link.SellerID.String() != order.seller && link.SellerID.String() != auth.SellerID {
			return notifications.VerifiedResource{}, notifications.ErrScope
		}
		if sender != auth.SellerID && sender != order.seller {
			return notifications.VerifiedResource{}, notifications.ErrScope
		}
		if site != "" && site != order.scope.Site {
			return notifications.VerifiedResource{}, notifications.ErrScope
		}
		if order.record.ShipmentID != snapshot.ID || (link.PackID != "" && link.PackID.String() != order.record.PackID) {
			return notifications.VerifiedResource{}, ErrOrderShipmentChanged
		}
		if len(orders) > 0 && (orders[0].seller != order.seller || orders[0].scope != order.scope) {
			return notifications.VerifiedResource{}, notifications.ErrScope
		}
		orders = append(orders, order)
	}
	return notifications.VerifiedResource{
		Topic: job.Topic, ResourceID: snapshot.ID, SellerID: orders[0].seller,
		Apply: func(ctx context.Context, tx notifications.LocalWriter) error {
			for _, order := range orders {
				if err := checkNotifiedOrderScope(ctx, tx, job, order); err != nil {
					return err
				}
				if err := upsertOrderRecord(ctx, tx, job.StoreID, order.scope, order.seller, order.record); err != nil {
					return err
				}
				tag, err := writeOrderShipmentSnapshot(ctx, tx, job.OwnerID, job.StoreID, order.scope,
					order.seller, order.record.ID, job.ParentSellerID, snapshot)
				if err != nil {
					return err
				}
				if tag.RowsAffected() != 1 {
					return ErrOrderShipmentChanged
				}
			}
			return nil
		},
	}, nil
}

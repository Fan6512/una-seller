package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

const maxNotificationShipmentOrders = 30

// NotificationHandlers are read-only upstream adapters. The worker owns the
// lease, final scope checks, local transaction and retry policy.
func (s *Service) NotificationHandlers() map[string]notifications.Prepare {
	return map[string]notifications.Prepare{
		"marketplace_orders":    s.prepareOrderNotification,
		"marketplace_shipments": s.prepareShipmentNotification,
		"marketplace_questions": s.prepareCustomerNotification,
		"marketplace_messages":  s.prepareCustomerNotification,
	}
}

type notifiedOrder struct {
	scope  OrderScope
	seller string
	record OrderRecord
}

func notificationAuthError(err error) error {
	if errors.Is(err, ErrOrderBusy) {
		return &notifications.Deferred{Cause: err, Reason: notifications.DeferBusy, Delay: 2 * time.Second}
	}
	if errors.Is(err, ErrReauthorizationRequired) || errors.Is(err, ErrRefreshUnknown) || errors.Is(err, ErrRefreshBusy) {
		return &notifications.Deferred{Cause: err, Reason: notifications.DeferAuthorization, Delay: time.Minute}
	}
	var limited *OrderRateLimitError
	if errors.As(err, &limited) {
		return &notifications.RateLimited{RetryAfter: time.Until(limited.until), Shared: limited.shared}
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
	if errors.Is(err, pgx.ErrNoRows) {
		// The mapping was revoked or replaced between the official read and the
		// commit. Report it as a scope rejection, not a generic handler failure.
		return notifications.ErrScope
	}
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
	ctx = backgroundOrderRead(ctx)
	if job.Topic != "marketplace_orders" || !orderNumber(job.ResourceID) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	auth, err := s.RequireStoreConnection(ctx, job.OwnerID, job.StoreID)
	if err != nil {
		return notifications.VerifiedResource{}, notificationAuthError(err)
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
			if err := upsertOrderRecord(ctx, tx, job.StoreID, order.scope, order.seller, order.record); err != nil {
				return err
			}
			return inheritKnownShipmentSnapshot(ctx, tx, job, order)
		},
	}, nil
}

type shipmentOrderLink struct {
	OrderID  json.Number `json:"order_id"`
	PackID   json.Number `json:"pack_id"`
	SellerID json.Number `json:"seller_id"`
}

// A shipment can precede its orders. Prefer the documented hosted association
// endpoint. Real CBT shipments still expose shipment.order_id, so a single
// linked order is used when the association endpoint is not permitted (401/403)
// or has no rows yet; a multi-order pack is never guessed.
func (s *Service) readShipmentOrderLinks(ctx context.Context, token, shipmentID string) ([]shipmentOrderLink, error) {
	var rows []shipmentOrderLink
	status, err := s.orderGET(ctx, token, "/shipments/"+shipmentID+"/orders", &rows)
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return nil, &notifications.Deferred{Cause: ErrOrderShipmentDenied, Reason: notifications.DeferAuthorization, Delay: time.Minute}
	}
	if status == http.StatusNoContent || (status == http.StatusOK && err == nil && len(rows) == 0) {
		return nil, &notifications.Deferred{Cause: ErrOrderInvalidResponse, Reason: notifications.DeferAssociation, Delay: time.Minute}
	}
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
	ctx = backgroundOrderRead(ctx)
	if job.Topic != "marketplace_shipments" || !orderSLAShipmentID.MatchString(job.ResourceID) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	auth, err := s.RequireStoreConnection(ctx, job.OwnerID, job.StoreID)
	if err != nil {
		return notifications.VerifiedResource{}, notificationAuthError(err)
	}
	if !notificationStoreMatches(auth, job) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	checkedAt := time.Now().UTC().Truncate(time.Microsecond)
	snapshot, sender, site, shipmentOrderID, err := s.fetchShipmentSnapshot(ctx, auth.AccessToken, job.ResourceID, checkedAt, false)
	if err != nil {
		return notifications.VerifiedResource{}, notificationAuthError(err)
	}
	links, err := s.readShipmentOrderLinks(ctx, auth.AccessToken, job.ResourceID)
	fallback := false
	if err != nil {
		// Real CBT shipments still return one order_id. Use it only when the
		// association is unavailable or empty, never to fill a partial list.
		if shipmentOrderID == "" || !shipmentAssociationFallbackAllowed(err) {
			return notifications.VerifiedResource{}, err
		}
		fallback = true
		links = []shipmentOrderLink{{OrderID: json.Number(shipmentOrderID)}}
	}
	orders := make([]notifiedOrder, 0, len(links))
	for _, link := range links {
		order, err := s.readNotifiedOrder(ctx, auth, job.NotificationUserID, link.OrderID.String())
		if err != nil {
			return notifications.VerifiedResource{}, err
		}
		// All orders must establish the same forward shipment, seller, site
		// and exact logistics scope. Mixed/ambiguous shipments fail closed.
		// A fallback link, a zero sender and an absent site are unknown, not
		// contradictions: real CBT shipments return sender_id 0 and no site.
		if link.SellerID != "" && link.SellerID.String() != order.seller && link.SellerID.String() != auth.SellerID {
			return notifications.VerifiedResource{}, notifications.ErrPermanentScope
		}
		if sender != "" && sender != "0" && sender != auth.SellerID && sender != order.seller {
			return notifications.VerifiedResource{}, notifications.ErrPermanentScope
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
			if fallback {
				return updateKnownShipmentOrders(ctx, tx, job, orders[0], snapshot)
			}
			return nil
		},
	}, nil
}

// A denied/empty association response does not prove that the returned order_id
// is the only order. Update already verified local siblings in the exact same
// scope and current shipment, never invent or infer an unknown order ID.
// Keep the same hard bound; overflow rolls back the entire fenced transaction.
func updateKnownShipmentOrders(ctx context.Context, tx notifications.LocalWriter, job notifications.Job, order notifiedOrder, snapshot OrderShipment) error {
	rows, err := tx.Query(ctx, `SELECT order_id FROM app.meli_orders
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND seller_id=$4
AND shipment_id=$5 AND order_id<>$6 ORDER BY order_id LIMIT $7 FOR UPDATE`,
		job.StoreID, order.scope.Site, order.scope.Logistic, order.seller, snapshot.ID, order.record.ID, maxNotificationShipmentOrders)
	if err != nil {
		return err
	}
	ids := make([]string, 0, maxNotificationShipmentOrders)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(ids) >= maxNotificationShipmentOrders {
		return ErrOrderInvalidResponse
	}
	for _, id := range ids {
		tag, err := writeOrderShipmentSnapshot(ctx, tx, job.OwnerID, job.StoreID, order.scope,
			order.seller, id, job.ParentSellerID, snapshot)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrOrderShipmentChanged
		}
	}
	return nil
}

// A later order notification can establish another order's relationship to an
// already-read shipment, even when the hosted association endpoint is denied.
// Reuse only a current same-scope snapshot, and use the existing stale guard.
// This does not establish completeness of the official association list.
func inheritKnownShipmentSnapshot(ctx context.Context, tx notifications.LocalWriter, job notifications.Job, order notifiedOrder) error {
	if !orderSLAShipmentID.MatchString(order.record.ShipmentID) {
		return nil
	}
	var raw []byte
	err := tx.QueryRow(ctx, `SELECT shipment_snapshot FROM app.meli_orders
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND seller_id=$4
AND shipment_id=$5 AND shipment_snapshot->>'shipment_id'=$5 AND order_id<>$6
ORDER BY (shipment_snapshot->>'last_updated')::timestamptz DESC NULLS LAST,
(shipment_snapshot->>'checked_at')::timestamptz DESC NULLS LAST, order_id
LIMIT 1`,
		job.StoreID, order.scope.Site, order.scope.Logistic, order.seller,
		order.record.ShipmentID, order.record.ID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	var snapshot OrderShipment
	if json.Unmarshal(raw, &snapshot) != nil || snapshot.ID != order.record.ShipmentID || snapshot.CheckedAt.IsZero() {
		return ErrOrderInvalidResponse
	}
	// A newer local order may have moved to a different shipment; in that
	// case the guarded update affects zero rows and leaves the newer order intact.
	_, err = writeOrderShipmentSnapshot(ctx, tx, job.OwnerID, job.StoreID, order.scope,
		order.seller, order.record.ID, job.ParentSellerID, snapshot)
	return err
}

func shipmentAssociationFallbackAllowed(err error) bool {
	var deferred *notifications.Deferred
	return errors.As(err, &deferred) &&
		(deferred.Reason == notifications.DeferAuthorization || deferred.Reason == notifications.DeferAssociation)
}

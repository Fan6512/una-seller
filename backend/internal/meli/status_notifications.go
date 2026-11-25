package meli

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
)

// Separate minimal projections. Never overwrite listing drafts or claim
// decisions, and never infer an order ID from a claim's resource type.
type marketplaceStatusNotice struct {
	id, sourceSeller, seller, site, status string

	resourceType, related, stage, kind, reason string

	quantity *int

	updated, checked time.Time
}

func decodeMarketplaceStatusNotice(raw json.RawMessage, job notifications.Job, auth AuthorizedStore, checked time.Time) (marketplaceStatusNotice, error) {
	n := marketplaceStatusNotice{id: job.ResourceID, checked: checked}
	switch job.Topic {
	case "marketplace_claims":
		var p struct {
			ID        json.Number `json:"id"`
			Resource  string      `json:"resource"`
			RelatedID json.Number `json:"resource_id"`
			Status    string      `json:"status"`
			Stage     string      `json:"stage"`
			Kind      string      `json:"type"`
			Reason    string      `json:"reason_id"`
			SiteID    string      `json:"site_id"`
			Updated   time.Time   `json:"last_updated"`
			Players   []struct {
				Type string      `json:"type"`
				Role string      `json:"role"`
				ID   json.Number `json:"user_id"`
			} `json:"players"`
		}
		if json.Unmarshal(raw, &p) != nil || p.ID.String() != n.id ||
			!orderNumber(p.RelatedID.String()) || p.RelatedID.String() == "0" {
			return n, ErrOrderInvalidResponse
		}
		for _, player := range p.Players {
			if player.Type != "seller" {
				continue
			}
			if n.sourceSeller != "" || (player.Role != "respondent" && player.Role != "complainant") {
				return n, notifications.ErrScope
			}
			n.sourceSeller = player.ID.String()
		}
		n.site, n.status, n.updated = p.SiteID, p.Status, p.Updated
		n.resourceType, n.related, n.stage, n.kind, n.reason = p.Resource, p.RelatedID.String(), p.Stage, p.Kind, p.Reason
		if !ValidOrderStatus(n.resourceType) || n.resourceType == "" || !ValidOrderStatus(n.stage) ||
			!ValidOrderStatus(n.kind) || !ValidOrderStatus(strings.ToLower(n.reason)) {
			return n, ErrOrderInvalidResponse
		}
	case "marketplace_items":
		var p struct {
			ID       string      `json:"id"`
			SellerID json.Number `json:"seller_id"`
			SiteID   string      `json:"site_id"`
			Status   string      `json:"status"`
			Quantity *int        `json:"available_quantity"`
			Updated  time.Time   `json:"last_updated"`
		}
		if json.Unmarshal(raw, &p) != nil || p.ID != n.id || !strings.HasPrefix(p.ID, p.SiteID) ||
			(p.Quantity != nil && (*p.Quantity < 0 || *p.Quantity > 2147483647)) {
			return n, ErrOrderInvalidResponse
		}
		n.sourceSeller, n.site, n.status, n.updated, n.quantity = p.SellerID.String(), p.SiteID, p.Status, p.Updated, p.Quantity
	default:
		return n, notifications.ErrScope
	}
	if n.updated.IsZero() || n.status == "" || !ValidOrderStatus(n.status) {
		return n, ErrOrderInvalidResponse
	}
	var err error
	n.seller, err = customerSeller(auth, job.NotificationUserID, n.sourceSeller, n.site)
	return n, err
}

func (s *Service) prepareMarketplaceStatusNotification(ctx context.Context, job notifications.Job) (notifications.VerifiedResource, error) {
	ctx = backgroundOrderRead(ctx)
	if !notifications.ValidResourceID(job.Topic, job.ResourceID) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	var prefix string
	switch job.Topic {
	case "marketplace_claims":
		prefix = "/marketplace/v2/claims/"
	case "marketplace_items":
		prefix = "/marketplace/items/"
	default:
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	auth, err := s.RequireStoreConnection(ctx, job.OwnerID, job.StoreID)
	if err != nil {
		return notifications.VerifiedResource{}, notificationAuthError(err)
	}
	if !notificationStoreMatches(auth, job) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	checked := time.Now().UTC().Truncate(time.Microsecond)
	var raw json.RawMessage
	status, err := s.orderGET(ctx, auth.AccessToken, prefix+job.ResourceID, &raw)
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return notifications.VerifiedResource{}, &notifications.Deferred{Cause: ErrOrderShipmentDenied, Reason: notifications.DeferAuthorization, Delay: time.Minute}
	}
	if err != nil {
		return notifications.VerifiedResource{}, notificationAuthError(err)
	}
	if status != http.StatusOK {
		return notifications.VerifiedResource{}, ErrOrderInvalidResponse
	}
	n, err := decodeMarketplaceStatusNotice(raw, job, auth, checked)
	if err != nil {
		return notifications.VerifiedResource{}, err
	}
	return notifications.VerifiedResource{
		Topic: job.Topic, ResourceID: n.id, SellerID: n.seller,
		Apply: func(ctx context.Context, tx notifications.LocalWriter) error {
			if err := recheckNotificationSiteSeller(ctx, tx, job, n.sourceSeller, n.seller, n.site); err != nil {
				return err
			}
			return writeMarketplaceStatusNotice(ctx, tx, job, n)
		},
	}, nil
}

func writeMarketplaceStatusNotice(ctx context.Context, tx notifications.LocalWriter, job notifications.Job, n marketplaceStatusNotice) error {
	switch job.Topic {
	case "marketplace_claims":
		_, err := tx.Exec(ctx, `INSERT INTO app.meli_claim_notifications AS c
(store_id,owner_user_id,parent_seller_id,seller_id,site_id,claim_id,resource_type,related_id,status,stage,claim_type,reason_code,last_updated,checked_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
ON CONFLICT(store_id,site_id,seller_id,claim_id) DO UPDATE SET
resource_type=EXCLUDED.resource_type,related_id=EXCLUDED.related_id,status=EXCLUDED.status,
stage=EXCLUDED.stage,claim_type=EXCLUDED.claim_type,reason_code=EXCLUDED.reason_code,
last_updated=EXCLUDED.last_updated,checked_at=EXCLUDED.checked_at
WHERE c.owner_user_id=EXCLUDED.owner_user_id AND c.parent_seller_id=EXCLUDED.parent_seller_id
AND (c.last_updated,c.checked_at)<=(EXCLUDED.last_updated,EXCLUDED.checked_at)`,
			job.StoreID, job.OwnerID, job.ParentSellerID, n.seller, n.site, n.id, n.resourceType, n.related, n.status, n.stage, n.kind, n.reason, n.updated, n.checked)
		return err
	case "marketplace_items":
		_, err := tx.Exec(ctx, `INSERT INTO app.meli_item_notifications AS c
(store_id,owner_user_id,parent_seller_id,seller_id,site_id,item_id,status,available_quantity,last_updated,checked_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
ON CONFLICT(store_id,site_id,seller_id,item_id) DO UPDATE SET
status=EXCLUDED.status,available_quantity=COALESCE(EXCLUDED.available_quantity,c.available_quantity),
last_updated=EXCLUDED.last_updated,checked_at=EXCLUDED.checked_at
WHERE c.owner_user_id=EXCLUDED.owner_user_id AND c.parent_seller_id=EXCLUDED.parent_seller_id
AND (c.last_updated,c.checked_at)<=(EXCLUDED.last_updated,EXCLUDED.checked_at)`,
			job.StoreID, job.OwnerID, job.ParentSellerID, n.seller, n.site, n.id, n.status, n.quantity, n.updated, n.checked)
		return err
	default:
		return notifications.ErrScope
	}
}

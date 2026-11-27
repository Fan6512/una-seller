package meli

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strconv"
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

	price, actions string

	subStatus *string

	dueAt *time.Time

	updated, checked time.Time
}

var (
	claimActionPattern = regexp.MustCompile(`^[a-z][a-z_]{0,63}$`)
	itemSubStatus      = regexp.MustCompile(`^[a-z][a-z_]{0,39}$`)
	itemPricePattern   = regexp.MustCompile(`^[0-9]{1,9}(\.[0-9]{1,6})?$`)
)

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
				Role             string      `json:"role"`
				ID               json.Number `json:"user_id"`
				AvailableActions []struct {
					Action    string `json:"action"`
					Mandatory bool   `json:"mandatory"`
					DueDate   string `json:"due_date"`
				} `json:"available_actions"`
			} `json:"players"`
		}
		if json.Unmarshal(raw, &p) != nil || p.ID.String() != n.id ||
			!orderNumber(p.RelatedID.String()) || p.RelatedID.String() == "0" {
			return n, ErrOrderInvalidResponse
		}
		// Resolve the seller independently of the notification's user. A hint
		// cannot disambiguate two mapped participants. Claim roles alone do not
		// distinguish buyer from seller, particularly for shipment claims.
		selected := -1
		for i, player := range p.Players {
			if player.Role != "respondent" && player.Role != "complainant" {
				continue
			}
			id := player.ID.String()
			if _, err := customerSeller(auth, auth.SellerID, id, p.SiteID); err != nil {
				continue
			}
			if selected >= 0 {
				return n, notifications.ErrScope
			}
			selected = i
			n.sourceSeller = id
		}
		if selected < 0 {
			return n, notifications.ErrScope
		}
		if _, err := customerSeller(auth, job.NotificationUserID, n.sourceSeller, p.SiteID); err != nil {
			return n, err
		}
		// Store only the identified seller's advertised action names, as
		// read-only metadata. A deadline is urgent only when mandatory=true.
		actions := map[string]bool{}
		for _, action := range p.Players[selected].AvailableActions {
			if !claimActionPattern.MatchString(action.Action) {
				return n, ErrOrderInvalidResponse
			}
			actions[action.Action] = true
			if !action.Mandatory || action.DueDate == "" {
				continue
			}
			due, err := time.Parse(time.RFC3339, action.DueDate)
			if err != nil {
				return n, ErrOrderInvalidResponse
			}
			if n.dueAt == nil || due.Before(*n.dueAt) {
				n.dueAt = &due
			}
		}
		names := make([]string, 0, len(actions))
		for name := range actions {
			names = append(names, name)
		}
		sort.Strings(names)
		n.actions = strings.Join(names, ",")
		n.site, n.status, n.updated = p.SiteID, p.Status, p.Updated
		n.resourceType, n.related, n.stage, n.kind, n.reason = p.Resource, p.RelatedID.String(), p.Stage, p.Kind, p.Reason
		if !ValidOrderStatus(n.resourceType) || n.resourceType == "" || !ValidOrderStatus(n.stage) ||
			!ValidOrderStatus(n.kind) || !ValidOrderStatus(strings.ToLower(n.reason)) || len(n.actions) > 512 {
			return n, ErrOrderInvalidResponse
		}
	case "marketplace_items":
		var p struct {
			ID        string      `json:"id"`
			SellerID  json.Number `json:"seller_id"`
			SiteID    string      `json:"site_id"`
			Status    string      `json:"status"`
			SubStatus []string    `json:"sub_status"`
			Price     json.Number `json:"price"`
			Quantity  *int        `json:"available_quantity"`
			Updated   time.Time   `json:"last_updated"`
		}
		if json.Unmarshal(raw, &p) != nil || p.ID != n.id || !strings.HasPrefix(p.ID, p.SiteID) ||
			(p.Quantity != nil && (*p.Quantity < 0 || *p.Quantity > 2147483647)) {
			return n, ErrOrderInvalidResponse
		}
		if price := p.Price.String(); price != "" {
			if !itemPricePattern.MatchString(price) {
				return n, ErrOrderInvalidResponse
			}
			if value, err := strconv.ParseFloat(price, 64); err != nil || value < 0 || value > 1e12 {
				return n, ErrOrderInvalidResponse
			}
			n.price = price
		}
		for _, sub := range p.SubStatus {
			if !itemSubStatus.MatchString(sub) {
				return n, ErrOrderInvalidResponse
			}
		}
		// Missing/null sub_status is not proof that moderation has cleared.
		// An explicit empty array is authoritative and clears the old value.
		if p.SubStatus != nil {
			subStatus := strings.Join(p.SubStatus, ",")
			if len(subStatus) > 160 {
				return n, ErrOrderInvalidResponse
			}
			n.subStatus = &subStatus
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
(store_id,owner_user_id,parent_seller_id,seller_id,site_id,claim_id,resource_type,related_id,status,stage,claim_type,reason_code,due_at,allowed_actions,last_updated,checked_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
ON CONFLICT(store_id,site_id,seller_id,claim_id) DO UPDATE SET
resource_type=EXCLUDED.resource_type,related_id=EXCLUDED.related_id,status=EXCLUDED.status,
stage=EXCLUDED.stage,claim_type=EXCLUDED.claim_type,reason_code=EXCLUDED.reason_code,
due_at=EXCLUDED.due_at,allowed_actions=EXCLUDED.allowed_actions,
last_updated=EXCLUDED.last_updated,checked_at=EXCLUDED.checked_at
WHERE c.owner_user_id=EXCLUDED.owner_user_id AND c.parent_seller_id=EXCLUDED.parent_seller_id
AND (c.last_updated,c.checked_at)<=(EXCLUDED.last_updated,EXCLUDED.checked_at)`,
			job.StoreID, job.OwnerID, job.ParentSellerID, n.seller, n.site, n.id, n.resourceType, n.related, n.status, n.stage, n.kind, n.reason, n.dueAt, n.actions, n.updated, n.checked)
		return err
	case "marketplace_items":
		_, err := tx.Exec(ctx, `INSERT INTO app.meli_item_notifications AS c
(store_id,owner_user_id,parent_seller_id,seller_id,site_id,item_id,status,sub_status,price,available_quantity,last_updated,checked_at)
VALUES($1,$2,$3,$4,$5,$6,$7,COALESCE($8::text,''),NULLIF($9,'')::numeric,$10,$11,$12)
ON CONFLICT(store_id,site_id,seller_id,item_id) DO UPDATE SET
status=EXCLUDED.status,sub_status=COALESCE($8::text,c.sub_status),
price=COALESCE(EXCLUDED.price,c.price),
available_quantity=COALESCE(EXCLUDED.available_quantity,c.available_quantity),
last_updated=EXCLUDED.last_updated,checked_at=EXCLUDED.checked_at
WHERE c.owner_user_id=EXCLUDED.owner_user_id AND c.parent_seller_id=EXCLUDED.parent_seller_id
AND (c.last_updated,c.checked_at)<=(EXCLUDED.last_updated,EXCLUDED.checked_at)`,
			job.StoreID, job.OwnerID, job.ParentSellerID, n.seller, n.site, n.id, n.status, n.subStatus, n.price, n.quantity, n.updated, n.checked)
		return err
	default:
		return notifications.ErrScope
	}
}

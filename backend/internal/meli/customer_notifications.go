package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

// A customer interaction belongs to a child/site, not to an inferred
// fulfillment mode. Multiple modes of the same child do not create copies.
type customerNotice struct {
	id, seller, sourceSeller, site, related, status, body, translated, answer, direction string

	created, checked time.Time
}

var customerItemID = regexp.MustCompile(`^CBT[0-9]{1,30}$`)

func customerSeller(auth AuthorizedStore, user, seller, site string) (string, error) {
	if !auth.Verified || !orderNumber(seller) || seller == "0" || !orderSitePattern.MatchString(site) || site == "CBT" {
		return "", notifications.ErrScope
	}
	child := ""
	for _, market := range auth.Markets {
		id := market.UserID.String()
		if market.SiteID != site || !orderNumber(id) || id == "0" ||
			(seller != auth.SellerID && seller != id) {
			continue
		}
		if child != "" && child != id {
			return "", notifications.ErrScope
		}
		child = id
	}
	if child == "" || !notificationChildMatches(auth.SellerID, user, child) {
		return "", notifications.ErrScope
	}
	return child, nil
}

func decodeCustomerNotice(raw json.RawMessage, job notifications.Job, auth AuthorizedStore, checked time.Time) (customerNotice, error) {
	var n customerNotice
	n.id, n.checked = job.ResourceID, checked
	switch job.Topic {
	case "marketplace_questions":
		var q struct {
			ID         json.Number `json:"id"`
			SellerID   json.Number `json:"seller_id"`
			SiteID     string      `json:"site_id"`
			ItemID     string      `json:"item_id"`
			Status     string      `json:"status"`
			Text       string      `json:"text"`
			Translated string      `json:"text_translated"`
			Created    time.Time   `json:"date_created"`
			Answer     *struct {
				Text string `json:"text"`
			} `json:"answer"`
		}
		if json.Unmarshal(raw, &q) != nil || q.ID.String() != n.id || !customerItemID.MatchString(q.ItemID) {
			return n, ErrOrderInvalidResponse
		}
		var err error
		n.seller, err = customerSeller(auth, job.NotificationUserID, q.SellerID.String(), q.SiteID)
		if err != nil {
			return n, err
		}
		n.sourceSeller = q.SellerID.String()
		n.site, n.related, n.status, n.created = q.SiteID, q.ItemID, q.Status, q.Created
		n.body, n.translated = q.Text, q.Translated
		if q.Answer != nil {
			n.answer = q.Answer.Text
		}
	case "marketplace_messages":
		var p struct {
			Messages []struct {
				ID         string `json:"id"`
				SiteID     string `json:"site_id"`
				Status     string `json:"status"`
				Text       string `json:"text"`
				Translated string `json:"text_translated"`
				From       struct {
					UserID json.Number `json:"user_id"`
				} `json:"from"`
				To struct {
					UserID json.Number `json:"user_id"`
				} `json:"to"`
				Date struct {
					Created time.Time `json:"created"`
				} `json:"message_date"`
				Resources []struct {
					ID   json.Number `json:"id"`
					Name string      `json:"name"`
				} `json:"message_resources"`
			} `json:"messages"`
		}
		if json.Unmarshal(raw, &p) != nil || len(p.Messages) != 1 || p.Messages[0].ID != n.id {
			return n, ErrOrderInvalidResponse
		}
		msg := p.Messages[0]
		var seller, pack string
		for _, resource := range msg.Resources {
			switch resource.Name {
			case "sellers":
				if seller != "" || !orderNumber(resource.ID.String()) || resource.ID.String() == "0" {
					return n, ErrOrderInvalidResponse
				}
				seller = resource.ID.String()
			case "packs":
				if pack != "" || !orderNumber(resource.ID.String()) || resource.ID.String() == "0" {
					return n, ErrOrderInvalidResponse
				}
				pack = resource.ID.String()
			}
		}
		if seller == "" || pack == "" {
			return n, ErrOrderInvalidResponse
		}
		var err error
		n.seller, err = customerSeller(auth, job.NotificationUserID, seller, msg.SiteID)
		if err != nil {
			return n, err
		}
		from, to := msg.From.UserID.String(), msg.To.UserID.String()
		if !orderNumber(from) || !orderNumber(to) || from == "0" || to == "0" {
			return n, ErrOrderInvalidResponse
		}
		isSeller := func(id string) bool { return id == n.seller || id == auth.SellerID }
		for _, market := range auth.Markets {
			id := market.UserID.String()
			if id != n.seller && (id == from || id == to) {
				return n, notifications.ErrScope
			}
		}
		switch {
		case isSeller(to) && !isSeller(from):
			n.direction = "incoming"
		case isSeller(from) && !isSeller(to):
			n.direction = "outgoing"
		default:
			return n, notifications.ErrScope
		}
		n.sourceSeller = seller
		n.site, n.related, n.status, n.created = msg.SiteID, pack, msg.Status, msg.Date.Created
		n.body, n.translated = msg.Text, msg.Translated
	default:
		return n, notifications.ErrScope
	}
	if n.created.IsZero() || n.status == "" || !ValidOrderStatus(strings.ToLower(n.status)) ||
		len(n.body) > 16384 || len(n.translated) > 16384 || len(n.answer) > 16384 {
		return n, ErrOrderInvalidResponse
	}
	// Unrecognized/non-visible states are retained as metadata, not exposed text.
	switch n.status {
	case "UNANSWERED", "ANSWERED", "CLOSED_UNANSWERED", "available":
	default:
		n.body, n.translated, n.answer = "", "", ""
	}
	return n, nil
}

func (s *Service) prepareCustomerNotification(ctx context.Context, job notifications.Job) (notifications.VerifiedResource, error) {
	ctx = backgroundOrderRead(ctx)
	if !notifications.ValidResourceID(job.Topic, job.ResourceID) {
		return notifications.VerifiedResource{}, notifications.ErrScope
	}
	var prefix string
	switch job.Topic {
	case "marketplace_questions":
		prefix = "/marketplace/questions/"
	case "marketplace_messages":
		prefix = "/marketplace/messages/"
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
	n, err := decodeCustomerNotice(raw, job, auth, checked)
	if err != nil {
		return notifications.VerifiedResource{}, err
	}
	return notifications.VerifiedResource{
		Topic: job.Topic, ResourceID: n.id, SellerID: n.seller,
		Apply: func(ctx context.Context, tx notifications.LocalWriter) error {
			return writeCustomerNotice(ctx, tx, job, n)
		},
	}, nil
}

func recheckNotificationSiteSeller(ctx context.Context, tx notifications.LocalWriter, job notifications.Job, sourceSeller, child, site string) error {
	var raw []byte
	err := tx.QueryRow(ctx, `SELECT marketplaces FROM app.meli_store_connections
WHERE id=$1::uuid AND owner_user_id=$2::uuid AND seller_id=$3 AND status='connected'
AND marketplaces_status='verified' FOR SHARE`, job.StoreID, job.OwnerID, job.ParentSellerID).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return notifications.ErrScope
	}
	if err != nil {
		return err
	}
	var markets []Marketplace
	if json.Unmarshal(raw, &markets) != nil {
		return notifications.ErrScope
	}
	auth := AuthorizedStore{SellerID: job.ParentSellerID, Verified: true, Markets: markets}
	if seller, err := customerSeller(auth, job.NotificationUserID, sourceSeller, site); err != nil || seller != child {
		return notifications.ErrScope
	}
	return nil
}

func writeCustomerNotice(ctx context.Context, tx notifications.LocalWriter, job notifications.Job, n customerNotice) error {
	if err := recheckNotificationSiteSeller(ctx, tx, job, n.sourceSeller, n.seller, n.site); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `INSERT INTO app.meli_customer_notifications AS c
(store_id,owner_user_id,parent_seller_id,seller_id,site_id,topic,resource_id,related_id,status,
body,translated_body,answer,direction,created_at,checked_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
ON CONFLICT(store_id,site_id,seller_id,topic,resource_id) DO UPDATE SET
related_id=EXCLUDED.related_id,status=EXCLUDED.status,body=EXCLUDED.body,
translated_body=EXCLUDED.translated_body,answer=EXCLUDED.answer,direction=EXCLUDED.direction,
created_at=EXCLUDED.created_at,checked_at=EXCLUDED.checked_at
WHERE c.owner_user_id=EXCLUDED.owner_user_id AND c.parent_seller_id=EXCLUDED.parent_seller_id
AND c.checked_at<=EXCLUDED.checked_at`, job.StoreID, job.OwnerID, job.ParentSellerID, n.seller,
		n.site, job.Topic, n.id, n.related, n.status, n.body, n.translated, n.answer, n.direction, n.created, n.checked)
	return err
}

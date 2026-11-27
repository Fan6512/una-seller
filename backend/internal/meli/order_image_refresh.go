package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var ErrOrderImageUnavailable = errors.New("official order item image unavailable")

// RefreshOrderPrimaryImage retrieves at most one missing item image. It only
// writes a trusted CDN URL into an already-owned cached order. No buyer data is
// downloaded into storage and the official product resource is never modified.
func (s *Service) RefreshOrderPrimaryImage(ctx context.Context, owner, store string, scope OrderScope, id string) (OrderRecord, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !orderNumber(id) {
		return OrderRecord{}, ErrOrderNotFound
	}
	order, err := s.GetOrderSummary(ctx, owner, store, scope, id)
	if err != nil {
		return OrderRecord{}, err
	}
	var itemID string
	for _, line := range order.Items {
		if line.ImageURL == "" && listingID.MatchString(line.ItemID) &&
			len(line.ItemID) >= len(scope.Site) && line.ItemID[:len(scope.Site)] == scope.Site {
			itemID = line.ItemID
			break
		}
	}
	if itemID == "" {
		return order, nil // Nothing safe or eligible to refresh.
	}
	auth, child, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderRecord{}, err
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var remote struct {
		ID              string      `json:"id"`
		SellerID        json.Number `json:"seller_id"`
		SiteID          string      `json:"site_id"`
		SecureThumbnail string      `json:"secure_thumbnail"`
		Pictures        []struct {
			SecureURL string `json:"secure_url"`
		} `json:"pictures"`
	}
	code, err := s.orderGET(fetchCtx, auth.AccessToken, "/marketplace/items/"+itemID, &remote)
	if err != nil {
		if code == http.StatusNotFound {
			return OrderRecord{}, ErrOrderImageUnavailable
		}
		return OrderRecord{}, err
	}
	if remote.ID != itemID || remote.SiteID != scope.Site || remote.SellerID.String() != child {
		return OrderRecord{}, ErrOrderImageUnavailable
	}
	imageURL := remote.SecureThumbnail
	if !orderImageURLAllowed(imageURL) {
		imageURL = ""
		for _, picture := range remote.Pictures {
			if orderImageURLAllowed(picture.SecureURL) {
				imageURL = picture.SecureURL
				break
			}
		}
	}
	if imageURL == "" {
		return OrderRecord{}, ErrOrderImageUnavailable
	}
	// Recheck current owner/child mapping after the network read. Keep other
	// order items and order status intact even if sync updated them concurrently.
	tag, err := s.db.Exec(ctx, `WITH connection AS (`+orderConnection+`)
 UPDATE app.meli_orders o SET order_items=(
   SELECT COALESCE(jsonb_agg(
     CASE WHEN line.value->>'item_id'=$6 AND COALESCE(line.value->>'image_url','')=''
       THEN jsonb_set(line.value,'{image_url}',to_jsonb($7::text),true)
       ELSE line.value END ORDER BY line.ordinality),'[]'::jsonb)
   FROM jsonb_array_elements(o.order_items) WITH ORDINALITY AS line(value,ordinality)
 ),updated_at=now()
 FROM connection c
 WHERE c.id=o.store_id AND c.child_id=o.seller_id
 AND c.status<>'disconnected' AND o.store_id=$1::uuid
 AND o.site_id=$3 AND o.logistic_type=$4 AND o.order_id=$5`,
		store, owner, scope.Site, scope.Logistic, id, itemID, imageURL)
	if err != nil {
		return OrderRecord{}, err
	}
	if tag.RowsAffected() == 0 {
		return OrderRecord{}, ErrOrderAuthorizationChanged
	}
	return s.GetOrderSummary(ctx, owner, store, scope, id)
}

package meli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
)

// Use the established CBT pack-envelope search parser and official request
// coordinator. Only IDs are returned; every detail is re-read by the worker.
func (s *Service) SearchDiscoveryOrders(ctx context.Context, scope notifications.DiscoveryScope, from, to time.Time, offset int) (notifications.DiscoveryPage, error) {
	ctx = backgroundOrderRead(ctx)
	if !orderScopeValid(OrderScope{Site: scope.Site, Logistic: scope.Logistic}) || !from.Before(to) || offset < 0 {
		return notifications.DiscoveryPage{}, ErrOrderInvalidResponse
	}
	auth, err := s.RequireStoreConnection(ctx, scope.OwnerID, scope.StoreID)
	if err != nil {
		return notifications.DiscoveryPage{}, notificationAuthError(err)
	}
	if !auth.Verified || auth.SellerID != scope.ParentID {
		return notifications.DiscoveryPage{}, notifications.ErrScope
	}
	var childSite, siteMode int
	for _, market := range auth.Markets {
		if market.SiteID != scope.Site {
			continue
		}
		if market.UserID.String() == scope.ChildID {
			childSite++
		}
		if market.LogisticType == scope.Logistic {
			if market.UserID.String() != scope.ChildID {
				return notifications.DiscoveryPage{}, notifications.ErrScope
			}
			siteMode++
		}
	}
	if childSite != 1 || siteMode != 1 || !notifications.ValidResourceID("marketplace_orders", scope.ChildID) {
		return notifications.DiscoveryPage{}, notifications.ErrScope
	}
	query := url.Values{
		"seller.id":         {scope.ChildID},
		"site":              {scope.Site},
		"date_created.from": {from.UTC().Format(time.RFC3339)},
		"date_created.to":   {to.UTC().Format(time.RFC3339)},
		"sort":              {"date_asc"},
		"limit":             {strconv.Itoa(OrderBatchSize)},
		"offset":            {strconv.Itoa(offset)},
	}
	var raw json.RawMessage
	status, err := s.orderGET(ctx, auth.AccessToken, "/marketplace/orders/search?"+query.Encode(), &raw)
	if err != nil {
		return notifications.DiscoveryPage{}, notificationAuthError(err)
	}
	if status != http.StatusOK {
		return notifications.DiscoveryPage{}, ErrOrderInvalidResponse
	}
	// A shrinking result set is recoverable by rereading the same window
	// from offset zero; other malformed responses remain failures.
	var paging struct {
		Paging struct {
			Total *int `json:"total"`
		} `json:"paging"`
	}
	if offset > 0 && json.Unmarshal(raw, &paging) == nil && paging.Paging.Total != nil &&
		*paging.Paging.Total >= 0 && *paging.Paging.Total < offset {
		return notifications.DiscoveryPage{}, notifications.ErrDiscoveryPageChanged
	}
	ids, total, count, err := parseOrderSearch(raw, scope.ChildID, offset)
	if err != nil {
		return notifications.DiscoveryPage{}, err
	}
	return notifications.DiscoveryPage{IDs: ids, Count: count, Total: total}, nil
}

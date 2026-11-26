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

// FetchMissedFeedPage uses only an explicitly selected, current stored token.
// This operator command never refreshes credentials or changes subscriptions.
func (s *Service) FetchMissedFeedPage(ctx context.Context, owner, store, topic string, offset int) ([]byte, error) {
	if !validUUID(owner) || !validUUID(store) || !notifications.ValidReplayTopic(topic) || offset < 0 || offset > 100000 {
		return nil, ErrOrderInvalidResponse
	}
	row, err := s.snapshot(ctx, owner, store)
	if err != nil {
		return nil, err
	}
	if row.lock != nil || row.expires == nil || !row.expires.After(time.Now().Add(refreshAhead)) {
		return nil, ErrReauthorizationRequired
	}
	auth, err := s.accessFromSnapshot(owner, row)
	if err != nil {
		return nil, err
	}
	if !auth.Verified {
		return nil, ErrSiteNotAuthorized
	}
	query := url.Values{
		"app_id": {s.cfg.ClientID},
		"topic":  {topic},
		"limit":  {strconv.Itoa(notifications.MissedFeedPageSize)},
		"offset": {strconv.Itoa(offset)},
	}
	var raw json.RawMessage
	status, err := s.orderGET(backgroundOrderRead(ctx), auth.AccessToken, "/missed_feeds?"+query.Encode(), &raw)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, ErrOrderInvalidResponse
	}
	return raw, nil
}

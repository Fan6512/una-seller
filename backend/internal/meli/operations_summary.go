package meli

import (
	"context"
	"time"
)

type OperationsStoreSummary struct {
	StoreID           string `json:"store_id"`
	Nickname          string `json:"nickname"`
	Connected         bool   `json:"connected"`
	PaidOrders        int    `json:"paid_orders"`
	CachedUnanswered  int    `json:"cached_unanswered"`
	IncomingMessages  int    `json:"incoming_messages_7d"`
	OpenClaims        int    `json:"open_claims"`
	PausedItems       int    `json:"paused_items"`
	PendingJobs       int    `json:"pending_jobs"`
	DeadJobs          int    `json:"dead_jobs"`
}

type OperationsSummary struct {
	Stores []OperationsStoreSummary `json:"stores"`
	CheckedAt time.Time `json:"checked_at"`
	Truncated bool `json:"truncated"`
	CacheOnly bool `json:"cache_only"`
}

// A local, owner-scoped summary. Never calls the provider or infers unread
// status, verified delivery, or current completeness from a cached count.
func (s *Service) OperationsSummary(ctx context.Context, owner string) (OperationsSummary, error) {
	out := OperationsSummary{Stores: []OperationsStoreSummary{}, CacheOnly: true}
	rows, err := s.db.Query(ctx, `WITH stores AS (
		SELECT id,owner_user_id,seller_id,nickname,marketplaces,
		       status='connected' AND marketplaces_status='verified' AS connected
		FROM app.meli_store_connections WHERE owner_user_id=$1::uuid
		ORDER BY id LIMIT 101
	)
	SELECT c.id::text,c.nickname,c.connected,
		(SELECT count(*) FROM app.meli_orders o WHERE c.connected AND o.store_id=c.id
		  AND o.status='paid' AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		    WHERE m.value->>'user_id'=o.seller_id AND m.value->>'site_id'=o.site_id
		    AND m.value->>'logistic_type'=o.logistic_type)),
		(SELECT count(*) FROM app.meli_customer_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id AND n.topic='marketplace_questions' AND n.status='UNANSWERED'
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT count(*) FROM app.meli_customer_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id AND n.topic='marketplace_messages'
		    AND n.status='available' AND n.direction='incoming' AND n.created_at>now()-interval '7 days'
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT count(*) FROM app.meli_claim_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id AND n.status='opened'
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT count(*) FROM app.meli_item_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id AND n.status='paused'
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT count(*) FROM app.meli_notification_jobs j
		  WHERE c.connected AND j.store_id=c.id AND j.owner_user_id=c.owner_user_id
		    AND j.parent_seller_id=c.seller_id AND j.application_id=$2 AND j.dead_at IS NULL
		    AND (j.notification_user_id=c.seller_id OR EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=j.notification_user_id))),
		(SELECT count(*) FROM app.meli_notification_jobs j
		  WHERE c.connected AND j.store_id=c.id AND j.owner_user_id=c.owner_user_id
		    AND j.parent_seller_id=c.seller_id AND j.application_id=$2 AND j.dead_at IS NOT NULL
		    AND (j.notification_user_id=c.seller_id OR EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=j.notification_user_id)))
	FROM stores c ORDER BY c.id`, owner, s.cfg.ClientID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var v OperationsStoreSummary
		if err := rows.Scan(&v.StoreID, &v.Nickname, &v.Connected, &v.PaidOrders, &v.CachedUnanswered,
			&v.IncomingMessages, &v.OpenClaims, &v.PausedItems, &v.PendingJobs, &v.DeadJobs); err != nil {
			return out, err
		}
		out.Stores = append(out.Stores, v)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	if len(out.Stores) > 100 {
		out.Stores = out.Stores[:100]
		out.Truncated = true
	}
	out.CheckedAt = time.Now().UTC()
	return out, nil
}

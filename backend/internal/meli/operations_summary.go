package meli

import (
	"context"
	"time"
)

type OperationsStoreSummary struct {
	StoreID          string   `json:"store_id"`
	Nickname         string   `json:"nickname"`
	Connected        bool     `json:"connected"`
	PaidOrders       int      `json:"paid_orders"`
	CachedUnanswered *int     `json:"cached_unanswered"`
	PendingMessages  *int     `json:"pending_message_conversations"`
	OpenClaims       *int     `json:"open_claims"`
	PausedItems      *int     `json:"paused_items"`
	PendingJobs      int      `json:"pending_jobs"`
	DeadJobs         int      `json:"dead_jobs"`
	Sites            []string `json:"sites"`
}

type OperationsSummary struct {
	Stores    []OperationsStoreSummary `json:"stores"`
	CheckedAt time.Time                `json:"checked_at"`
	Truncated bool                     `json:"truncated"`
	CacheOnly bool                     `json:"cache_only"`
}

// Group available message snapshots by the exact conversation scope. Use
// provider creation time, not notification arrival/check time, so duplicates
// and out-of-order deliveries cannot reopen a replied conversation.
const operationsMessageConversationsSQL = `SELECT store_id,owner_user_id,parent_seller_id,seller_id,site_id,related_id,
	max(created_at) FILTER (WHERE direction='incoming') AS last_incoming,
	max(created_at) FILTER (WHERE direction='outgoing') AS last_outgoing
	FROM app.meli_customer_notifications
	WHERE owner_user_id=$1::uuid AND topic='marketplace_messages' AND status='available'
	  AND direction IN ('incoming','outgoing') AND related_id ~ '^[1-9][0-9]{0,29}$'
	GROUP BY store_id,owner_user_id,parent_seller_id,seller_id,site_id,related_id`

// Equal timestamps do not prove that the outgoing message followed the incoming
// one. Keep the conversation pending until a strictly later reply is observed.
const operationsMessagePendingSQL = `(n.last_incoming IS NOT NULL
	AND (n.last_outgoing IS NULL OR n.last_incoming>=n.last_outgoing))`

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
		(SELECT CASE WHEN count(*)>0 THEN count(*) FILTER (WHERE n.status='UNANSWERED') END
		  FROM app.meli_customer_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id AND n.topic='marketplace_questions'
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT CASE WHEN count(*)>0 THEN count(*) FILTER (WHERE `+operationsMessagePendingSQL+`) END
		  FROM (`+operationsMessageConversationsSQL+`) n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT CASE WHEN count(*)>0 THEN count(*) FILTER (WHERE n.status='opened') END
		  FROM app.meli_claim_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id
		    AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		      WHERE m.value->>'user_id'=n.seller_id AND m.value->>'site_id'=n.site_id)),
		(SELECT CASE WHEN count(*)>0 THEN count(*) FILTER (WHERE n.status='paused') END
		  FROM app.meli_item_notifications n
		  WHERE c.connected AND n.store_id=c.id AND n.owner_user_id=c.owner_user_id
		    AND n.parent_seller_id=c.seller_id
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
		,ARRAY(SELECT DISTINCT m.value->>'site_id' FROM jsonb_array_elements(c.marketplaces) m(value)
		  WHERE c.connected AND m.value->>'site_id' ~ '^[A-Z]{3,8}$' AND m.value->>'site_id'<>'CBT'
		  ORDER BY 1)
	FROM stores c ORDER BY c.id`, owner, s.cfg.ClientID)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var v OperationsStoreSummary
		if err := rows.Scan(&v.StoreID, &v.Nickname, &v.Connected, &v.PaidOrders, &v.CachedUnanswered,
			&v.PendingMessages, &v.OpenClaims, &v.PausedItems, &v.PendingJobs, &v.DeadJobs, &v.Sites); err != nil {
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

// OperationsOverview adds bounded read-only work lists to the existing
// store-level counters. A row describes the last locally observed state,
// not an unread notification or proof of a current platform-wide backlog.
type OperationsRow struct {
	StoreID    string     `json:"store_id"`
	Site       string     `json:"site"`
	Kind       string     `json:"kind"`
	ResourceID string     `json:"resource_id"`
	Status     string     `json:"status"`
	ObservedAt time.Time  `json:"observed_at"`
	DueAt      *time.Time `json:"due_at,omitempty"`
}

type OperationsOverview struct {
	OperationsSummary
	Sales    OperationsSales `json:"sales"`
	Activity []OperationsRow `json:"activity"`
	Todos    []OperationsRow `json:"todos"`
}

func ValidOperationsFilter(store, site string) bool {
	return (store == "" || validUUID(store)) && (site == "" || (orderSitePattern.MatchString(site) && site != "CBT"))
}

func (s *Service) OperationsOverview(ctx context.Context, owner, store, site string) (OperationsOverview, error) {
	out := OperationsOverview{Activity: []OperationsRow{}, Todos: []OperationsRow{}}
	if !ValidOperationsFilter(store, site) {
		return out, ErrOrderInvalidResponse
	}
	summary, err := s.OperationsSummary(ctx, owner)
	if err != nil {
		return out, err
	}
	out.OperationsSummary = summary
	rows, err := s.db.Query(ctx, `WITH stores AS (
		SELECT id,owner_user_id,seller_id,marketplaces FROM app.meli_store_connections
		WHERE owner_user_id=$1::uuid AND status='connected' AND marketplaces_status='verified'
		  AND ($3='' OR id::text=$3)
		ORDER BY id LIMIT 100
	), markets AS (
		SELECT DISTINCT c.id,c.owner_user_id,c.seller_id AS parent_id,
		  m.value->>'site_id' AS site,m.value->>'user_id' AS child,m.value->>'logistic_type' AS logistic
		FROM stores c CROSS JOIN LATERAL jsonb_array_elements(c.marketplaces) m(value)
		WHERE ($4='' OR m.value->>'site_id'=$4)
	), order_states AS (
		SELECT o.store_id,o.site_id,o.order_id,
		  (`+orderFulfillmentStatusSQL+`) AS status,o.updated_at,sla.expected_date
		FROM app.meli_orders o
		LEFT JOIN app.meli_order_dispatch_sla sla ON sla.store_id=o.store_id AND sla.site_id=o.site_id
		  AND sla.logistic_type=o.logistic_type AND sla.seller_id=o.seller_id
		  AND sla.order_id=o.order_id AND sla.shipment_id=o.shipment_id
		WHERE EXISTS (SELECT 1 FROM markets m WHERE m.id=o.store_id AND m.site=o.site_id
		  AND m.child=o.seller_id AND m.logistic=o.logistic_type)
	), observed AS (
		SELECT o.store_id::text AS store_id,o.site_id AS site,'order'::text AS kind,o.order_id AS resource_id,o.status,
		  o.updated_at AS observed_at,CASE WHEN o.status IN ('awaiting_dispatch','label_printed') THEN o.expected_date END AS due_at,
		  o.status IN ('awaiting_dispatch','label_printed','logistics_exception','shipment_cancelled') AS todo
		FROM order_states o
		UNION ALL
		SELECT n.store_id::text,n.site_id,'question',
		  n.resource_id,n.status,n.checked_at,NULL::timestamptz,n.status='UNANSWERED'
		FROM app.meli_customer_notifications n
		WHERE n.topic='marketplace_questions' AND EXISTS (SELECT 1 FROM markets m WHERE m.id=n.store_id AND m.owner_user_id=n.owner_user_id
		  AND m.parent_id=n.parent_seller_id AND m.site=n.site_id AND m.child=n.seller_id)
		UNION ALL
		SELECT n.store_id::text,n.site_id,'message',n.related_id,
		  CASE WHEN `+operationsMessagePendingSQL+` THEN 'pending_reply'
		       WHEN n.last_incoming IS NOT NULL THEN 'replied' ELSE 'sent' END,
		  greatest(n.last_incoming,n.last_outgoing),NULL::timestamptz,`+operationsMessagePendingSQL+`
		FROM (`+operationsMessageConversationsSQL+`) n
		WHERE EXISTS (SELECT 1 FROM markets m WHERE m.id=n.store_id AND m.owner_user_id=n.owner_user_id
		  AND m.parent_id=n.parent_seller_id AND m.site=n.site_id AND m.child=n.seller_id)
		UNION ALL
		SELECT n.store_id::text,n.site_id,'claim',n.claim_id,n.status,n.checked_at,
		  CASE WHEN n.status='opened' THEN n.due_at END,n.status='opened'
		FROM app.meli_claim_notifications n
		WHERE EXISTS (SELECT 1 FROM markets m WHERE m.id=n.store_id AND m.owner_user_id=n.owner_user_id
		  AND m.parent_id=n.parent_seller_id AND m.site=n.site_id AND m.child=n.seller_id)
		UNION ALL
		SELECT n.store_id::text,n.site_id,'item',n.item_id,n.status,n.checked_at,NULL::timestamptz,n.status='paused'
		FROM app.meli_item_notifications n
		WHERE EXISTS (SELECT 1 FROM markets m WHERE m.id=n.store_id AND m.owner_user_id=n.owner_user_id
		  AND m.parent_id=n.parent_seller_id AND m.site=n.site_id AND m.child=n.seller_id)
		UNION ALL
		SELECT j.store_id::text,'','sync_failure',j.resource_id,'failed',j.dead_at,NULL::timestamptz,true
		FROM app.meli_notification_jobs j JOIN stores c ON c.id=j.store_id
		WHERE $4='' AND j.owner_user_id=c.owner_user_id AND j.parent_seller_id=c.seller_id
		  AND j.application_id=$2 AND j.dead_at IS NOT NULL
		  AND (j.notification_user_id=c.seller_id OR EXISTS (SELECT 1 FROM markets m
		    WHERE m.id=c.id AND m.child=j.notification_user_id))
	)
	(SELECT 'activity',store_id,site,kind,resource_id,status,observed_at,due_at FROM observed
	 ORDER BY observed_at DESC,store_id,site,kind,resource_id LIMIT 30)
	UNION ALL
	(SELECT 'todo',store_id,site,kind,resource_id,status,observed_at,due_at FROM observed WHERE todo
	 ORDER BY due_at ASC NULLS LAST,observed_at DESC,store_id,site,kind,resource_id LIMIT 30)`,
		owner, s.cfg.ClientID, store, site)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var bucket string
		var row OperationsRow
		if err := rows.Scan(&bucket, &row.StoreID, &row.Site, &row.Kind, &row.ResourceID,
			&row.Status, &row.ObservedAt, &row.DueAt); err != nil {
			return out, err
		}
		if bucket == "todo" {
			out.Todos = append(out.Todos, row)
		} else {
			out.Activity = append(out.Activity, row)
		}
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	rows.Close()
	out.Sales, err = s.operationsSales(ctx, owner, store, site, out.OperationsSummary)
	return out, err
}

package notifications

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresInbox struct { DB *pgxpool.Pool }

// Resolve the currently verified CBT parent/child mapping. Ambiguous mappings
// are rejected; lock the connection while inserting. Duplicate deliveries are
// acknowledged only after the atomic SQL statement has succeeded.
func (p PostgresInbox) Enqueue(ctx context.Context, e Event) error {
	var matched int
	err := p.DB.QueryRow(ctx,`WITH candidates AS (
 SELECT c.id,c.owner_user_id,c.seller_id
 FROM app.meli_store_connections c
 WHERE c.status='connected' AND c.marketplaces_status='verified'
 AND (c.seller_id=$2 OR EXISTS (
  SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
  WHERE m.value->>'user_id'=$2
 ))
 FOR SHARE
), scoped AS (
 SELECT * FROM candidates WHERE (SELECT count(*) FROM candidates)=1
), inserted AS (
 INSERT INTO app.meli_notification_inbox
 (store_id,owner_user_id,parent_seller_id,application_id,notification_user_id,topic,resource_id,event_key,source_received_at)
 SELECT id,owner_user_id,seller_id,$1,$2,$3,$4,$5,$6 FROM scoped
 ON CONFLICT (store_id,owner_user_id,application_id,notification_user_id,topic,resource_id,event_key) DO NOTHING
 RETURNING id
)
SELECT count(*) FROM scoped`,e.ApplicationID,e.UserID,e.Topic,e.ResourceID,e.Key,e.ReceivedAt).Scan(&matched)
	if err != nil { return err }
	if matched != 1 { return ErrScope }
	return nil
}

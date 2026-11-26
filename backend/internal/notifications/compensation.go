package notifications

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"time"
)

// Known-resource compensation is independent of the new-order discovery
// cursor. It also covers payment-pending orders that may become paid.
type compensationCandidate struct {
	scope                 storeScope
	user, topic, resource string
}

func (p PostgresInbox) CompensationSweep(ctx context.Context, appID string) error {
	if !idPattern.MatchString(appID) {
		return errors.New("invalid compensation application")
	}
	rows, err := p.DB.Query(ctx, `WITH eligible AS (
		SELECT c.id AS store_id,c.owner_user_id,c.seller_id AS parent_id,
		       o.seller_id AS child_id,o.order_id,o.shipment_id,o.updated_at,o.shipment_snapshot,o.status
		FROM app.meli_orders o JOIN app.meli_store_connections c ON c.id=o.store_id
		WHERE c.status='connected' AND c.marketplaces_status='verified' AND o.status IN ('paid','payment_required','payment_in_process','confirmed')
		  AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		    WHERE m.value->>'user_id'=o.seller_id AND m.value->>'site_id'=o.site_id
		      AND m.value->>'logistic_type'=o.logistic_type)
	), candidates AS (
		SELECT store_id,owner_user_id,parent_id,child_id,'marketplace_orders' AS topic,
		       order_id AS resource,updated_at AS checked,interval '30 minutes' AS cadence FROM eligible
		WHERE updated_at<now()-interval '30 minutes'
		UNION ALL
		SELECT store_id,owner_user_id,parent_id,child_id,'marketplace_shipments',shipment_id,
		       CASE WHEN shipment_snapshot->>'shipment_id'=shipment_id
		         THEN (shipment_snapshot->>'checked_at')::timestamptz END,interval '60 minutes'
		FROM eligible WHERE status='paid' AND shipment_id ~ '^[1-9][0-9]{0,29}$'
		  AND (shipment_snapshot->>'shipment_id' IS DISTINCT FROM shipment_id
		    OR COALESCE(shipment_snapshot->>'status','') NOT IN ('delivered','cancelled'))
	), due AS (
		SELECT DISTINCT ON (store_id,owner_user_id,parent_id,topic,resource)
		  store_id,owner_user_id,parent_id,child_id,topic,resource,checked
		FROM candidates x WHERE (checked IS NULL OR checked<now()-cadence)
		  AND NOT EXISTS (SELECT 1 FROM app.meli_notification_inbox i
		    WHERE i.store_id=x.store_id AND i.owner_user_id=x.owner_user_id
		      AND i.parent_seller_id=x.parent_id AND i.application_id=$1 AND i.topic=x.topic
		      AND i.resource_id=x.resource AND i.enqueued_at>now()-x.cadence)
		  AND NOT EXISTS (SELECT 1 FROM app.meli_notification_jobs j
		    WHERE j.store_id=x.store_id AND j.owner_user_id=x.owner_user_id
		      AND j.parent_seller_id=x.parent_id AND j.application_id=$1 AND j.topic=x.topic
		      AND j.resource_id=x.resource AND (j.dead_at IS NULL OR j.dead_at>now()-interval '24 hours'))
		  AND NOT EXISTS (SELECT 1 FROM app.meli_notification_cooldowns c
		    WHERE c.store_id=x.store_id AND c.owner_user_id=x.owner_user_id
		      AND c.parent_seller_id=x.parent_id AND c.application_id=$1 AND c.available_at>now())
		ORDER BY store_id,owner_user_id,parent_id,topic,resource,checked ASC NULLS FIRST,child_id
	)
	SELECT store_id::text,owner_user_id::text,parent_id,child_id,topic,resource FROM due
	WHERE NOT EXISTS (SELECT 1 FROM app.meli_order_read_coordination WHERE application_id=$1 AND retry_until>now())
	  AND (SELECT count(*) FROM app.meli_notification_jobs WHERE application_id=$1 AND dead_at IS NULL)<32
	ORDER BY checked ASC NULLS FIRST,store_id,topic,resource LIMIT 8`, appID)
	if err != nil {
		return err
	}
	var candidates []compensationCandidate
	for rows.Next() {
		var c compensationCandidate
		if err := rows.Scan(&c.scope.storeID, &c.scope.ownerID, &c.scope.parentID, &c.user, &c.topic, &c.resource); err != nil {
			rows.Close()
			return err
		}
		candidates = append(candidates, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, c := range candidates {
		if !ValidResourceID(c.topic, c.resource) || !idPattern.MatchString(c.user) {
			continue
		}
		interval := 30 * time.Minute
		if c.topic == "marketplace_shipments" {
			interval = time.Hour
		}
		key := sha256.Sum256([]byte("compensation:" + now.Truncate(interval).Format(time.RFC3339)))
		e := Event{ApplicationID: appID, UserID: c.user, Topic: c.topic, ResourceID: c.resource,
			Key: hex.EncodeToString(key[:]), ReceivedAt: now}
		if err := p.enqueue(ctx, e, &c.scope); err != nil {
			if errors.Is(err, ErrScope) || errors.Is(err, ErrScopeUnavailable) {
				continue
			}
			if errors.Is(err, ErrCapacity) {
				return nil
			}
			return err
		}
	}
	return nil
}

// One capped sweep per minute, no catch-up storm on restart. Official GETs
// remain in the existing worker and its shared rate gate.
func (p PostgresInbox) RunCompensation(ctx context.Context, appID string, logger *log.Logger, search DiscoverySearch) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			discoveryCtx, cancelDiscovery := context.WithTimeout(ctx, 25*time.Second)
			err := p.DiscoverySweep(discoveryCtx, appID, search)
			cancelDiscovery()
			if err != nil && !errors.Is(err, ErrCapacity) {
				logger.Print("order discovery sweep unavailable; progress retained")
			}
			sweepCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err = p.CompensationSweep(sweepCtx, appID)
			cancel()
			if err != nil {
				logger.Print("notification compensation sweep unavailable")
			}
		}
	}
}

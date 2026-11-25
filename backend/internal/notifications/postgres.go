package notifications

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queueCapacity      = 20000
	queueAdmissionLock = 1410389900
)

var ErrCapacity = errors.New("notification queue capacity reached")

type PostgresInbox struct {
	DB *pgxpool.Pool
}

type storeScope struct {
	storeID, ownerID, parentID string
}

// Lock candidate mappings through admission or the final scoped write. A known
// but unavailable account is retryable; an unknown/ambiguous identity is not.
func resolveScope(ctx context.Context, tx pgx.Tx, seller string) (storeScope, error) {
	rows, err := tx.Query(ctx, `
		SELECT c.id::text, c.owner_user_id::text, c.seller_id,
		       c.status='connected' AND c.marketplaces_status='verified'
		FROM app.meli_store_connections c
		WHERE (c.seller_id=$1 OR EXISTS (
			SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
			WHERE m.value->>'user_id'=$1
		  ))
		ORDER BY c.id LIMIT 2 FOR SHARE`, seller)
	if err != nil {
		return storeScope{}, err
	}
	type candidate struct {
		scope storeScope
		ready bool
	}
	matches, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (candidate, error) {
		var c candidate
		err := row.Scan(&c.scope.storeID, &c.scope.ownerID, &c.scope.parentID, &c.ready)
		return c, err
	})
	if err != nil {
		return storeScope{}, err
	}
	if len(matches) != 1 {
		return storeScope{}, ErrScope
	}
	if !matches[0].ready {
		return storeScope{}, ErrScopeUnavailable
	}
	return matches[0].scope, nil
}

// Enqueue atomically records a unique delivery and advances its resource job.
func (p PostgresInbox) Enqueue(ctx context.Context, e Event) error {
	return pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL synchronous_commit=on`); err != nil {
			return err
		}
		// Admission serializes the capacity check across API processes.
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(queueAdmissionLock)); err != nil {
			return err
		}
		scope, err := resolveScope(ctx, tx, e.UserID)
		if err != nil {
			return err
		}
		var duplicate, full bool
		err = tx.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM app.meli_notification_inbox
				WHERE store_id=$1 AND owner_user_id=$2 AND application_id=$3
				  AND topic=$4 AND resource_id=$5 AND notification_user_id=$6
				  AND event_key=$7
			), EXISTS (
				SELECT 1 FROM app.meli_notification_inbox OFFSET $9
			) OR (EXISTS (
				SELECT 1 FROM app.meli_notification_jobs OFFSET $9
			) AND NOT EXISTS (
				SELECT 1 FROM app.meli_notification_jobs
				WHERE store_id=$1 AND owner_user_id=$2 AND application_id=$3
				  AND topic=$4 AND resource_id=$5 AND parent_seller_id=$8
				  AND dead_at IS NULL
				FOR UPDATE
			))`, scope.storeID, scope.ownerID, e.ApplicationID, e.Topic,
			e.ResourceID, e.UserID, e.Key, scope.parentID, queueCapacity-1).Scan(&duplicate, &full)
		if err != nil {
			return err
		}
		if duplicate {
			return nil
		}
		if full {
			return ErrCapacity
		}
		_, err = tx.Exec(ctx, `
			WITH receipt AS (
				INSERT INTO app.meli_notification_inbox
				    (store_id,owner_user_id,parent_seller_id,application_id,topic,resource_id,
				     notification_user_id,event_key,source_received_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
				RETURNING store_id,owner_user_id,parent_seller_id,application_id,topic,
				          resource_id,notification_user_id
			)
			INSERT INTO app.meli_notification_jobs AS j
			    (store_id,owner_user_id,parent_seller_id,application_id,topic,resource_id,notification_user_id)
			SELECT * FROM receipt
			ON CONFLICT (store_id,owner_user_id,parent_seller_id,application_id,topic,resource_id)
			WHERE dead_at IS NULL
			DO UPDATE SET generation=j.generation+1, notification_user_id=EXCLUDED.notification_user_id`,
			scope.storeID, scope.ownerID, scope.parentID, e.ApplicationID, e.Topic,
			e.ResourceID, e.UserID, e.Key, e.ReceivedAt)
		return err
	})
}

type QueueStats struct {
	Receipts, GlobalReceipts, Jobs, Pending, Leased, Dead, Unregistered, Cooling, ExpiredLeases int
	OldestPending                                                                               time.Duration
}

func (p PostgresInbox) stats(ctx context.Context, appID string, topics []string) (QueueStats, error) {
	var stats QueueStats
	var oldestSeconds int64
	err := p.DB.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM app.meli_notification_inbox WHERE application_id=$1),
		       (SELECT count(*) FROM app.meli_notification_inbox),
		       (SELECT count(*) FROM app.meli_notification_jobs),
		       count(*) FILTER (WHERE j.dead_at IS NULL AND (j.lease_until IS NULL OR j.lease_until <= now())),
		       count(*) FILTER (WHERE j.lease_until > now()),
		       count(*) FILTER (WHERE j.dead_at IS NOT NULL),
		       count(*) FILTER (WHERE j.dead_at IS NULL AND NOT (j.topic=ANY($2::text[]))),
		       count(*) FILTER (WHERE j.dead_at IS NULL AND c.available_at > now()),
		       count(*) FILTER (WHERE j.lease_until <= now()),
		       COALESCE(extract(epoch FROM now()-min(j.pending_since)
		           FILTER (WHERE j.dead_at IS NULL)),0)::bigint
		FROM app.meli_notification_jobs j
		LEFT JOIN app.meli_notification_cooldowns c
		    USING (store_id,owner_user_id,parent_seller_id,application_id)
		WHERE j.application_id=$1`, appID, topics).Scan(
		&stats.Receipts, &stats.GlobalReceipts, &stats.Jobs, &stats.Pending, &stats.Leased, &stats.Dead, &stats.Unregistered,
		&stats.Cooling, &stats.ExpiredLeases, &oldestSeconds)
	stats.OldestPending = time.Duration(oldestSeconds) * time.Second
	return stats, err
}

func (p PostgresInbox) prune(ctx context.Context) error {
	_, err := p.DB.Exec(ctx, `
		WITH old_receipts AS (
			SELECT id FROM app.meli_notification_inbox WHERE enqueued_at < now()-interval '7 days'
			ORDER BY enqueued_at,id LIMIT 500 FOR UPDATE SKIP LOCKED
		), receipts AS (
			DELETE FROM app.meli_notification_inbox WHERE id IN (SELECT id FROM old_receipts)
		), old_jobs AS (
			SELECT id FROM app.meli_notification_jobs WHERE dead_at < now()-interval '30 days'
			ORDER BY dead_at LIMIT 500 FOR UPDATE SKIP LOCKED
		), jobs AS (
			DELETE FROM app.meli_notification_jobs WHERE id IN (SELECT id FROM old_jobs)
		), old_cooldowns AS (
			SELECT store_id,owner_user_id,parent_seller_id,application_id
			FROM app.meli_notification_cooldowns WHERE available_at <= now()
			ORDER BY available_at LIMIT 500 FOR UPDATE SKIP LOCKED
		)
		DELETE FROM app.meli_notification_cooldowns c USING old_cooldowns o
		WHERE (c.store_id,c.owner_user_id,c.parent_seller_id,c.application_id)
		     =(o.store_id,o.owner_user_id,o.parent_seller_id,o.application_id)`)
	if err != nil {
		return fmt.Errorf("prune notification queue: %w", err)
	}
	return nil
}

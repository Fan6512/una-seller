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
const scopeCandidatesSQL = `
		SELECT c.id::text AS store_id, c.owner_user_id::text AS owner_id, c.seller_id AS parent_id,
		       c.status='connected' AND c.marketplaces_status='verified' AS ready
		FROM app.meli_store_connections c
		WHERE (c.seller_id=$1 OR EXISTS (
			SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
			WHERE m.value->>'user_id'=$1
		  ))
		ORDER BY c.id LIMIT 2 FOR SHARE`

func resolveScope(ctx context.Context, tx pgx.Tx, seller string) (storeScope, error) {
	rows, err := tx.Query(ctx, scopeCandidatesSQL, seller)
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
func (p PostgresInbox) Enqueue(ctx context.Context, e Event) (err error) {
	trace := newACKTrace()
	defer func() { trace.report(err) }()
	conn, err := p.DB.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	trace.next("begin")
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	trace.next("admission")
	if err = lockAdmission(ctx, tx, pgx.QueryExecModeExec); err != nil {
		return err
	}
	trace.next("receipt_sql")
	if err = enqueueReceipt(ctx, tx, e, "webhook", webhookReceiptSQL, pgx.QueryExecModeExec); err != nil {
		return err
	}
	trace.next("commit")
	err = tx.Commit(ctx)
	if err == nil {
		trace.next("committed")
	}
	return err
}

func (p PostgresInbox) enqueue(ctx context.Context, e Event, expected *storeScope) error {
	return p.enqueueScoped(ctx, e, expected, nil)
}

// Discovery must distinguish a durable outcome from temporary backpressure.
// Unlike a periodic known-resource hint, a discovered order cannot be silently
// suppressed by a full queue or a cooldown before advancing its search cursor.
func (p PostgresInbox) enqueueScoped(ctx context.Context, e Event, expected *storeScope, discovery *DiscoveryScope) error {
	return pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		return p.enqueueInTx(ctx, tx, e, expected, discovery, "webhook")
	})
}

// Replay uses the same admission and generation update in its page transaction.
func lockAdmission(ctx context.Context, tx pgx.Tx, mode pgx.QueryExecMode) error {
	// Establish synchronous commit and serialize admission before the next
	// statement takes its snapshot. The transaction holds the lock until it ends.
	_, err := tx.Exec(ctx, `SELECT set_config('synchronous_commit','on',true), pg_advisory_xact_lock($1)`, mode, int64(queueAdmissionLock))
	return err
}

func (p PostgresInbox) enqueueInTx(ctx context.Context, tx pgx.Tx, e Event, expected *storeScope, discovery *DiscoveryScope, provenance string) error {
	mode := tx.Conn().Config().DefaultQueryExecMode
	if err := lockAdmission(ctx, tx, mode); err != nil {
		return err
	}
	scope, err := resolveScope(ctx, tx, e.UserID)
	if err != nil {
		return err
	}
	source := provenance
	if expected != nil {
		if scope != *expected {
			return ErrScope
		}
		source = "compensation"
		if discovery != nil {
			if err := checkDiscoveryMapping(ctx, tx, *discovery); err != nil {
				return err
			}
			var covered, blocked bool
			err := tx.QueryRow(ctx, `SELECT
				EXISTS (SELECT 1 FROM app.meli_orders WHERE store_id=$1
				  AND site_id=$7 AND logistic_type=$8 AND seller_id=$9 AND order_id=$6)
				OR EXISTS (SELECT 1 FROM app.meli_notification_jobs
				  WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3
				    AND application_id=$4 AND topic=$5 AND resource_id=$6
				    AND notification_user_id IN ($3,$9)
				    AND (dead_at IS NULL OR dead_at>now()-interval '24 hours')),
				(SELECT count(*) FROM app.meli_notification_jobs
				  WHERE application_id=$4 AND dead_at IS NULL)>=32
				OR EXISTS (SELECT 1 FROM app.meli_order_read_coordination
				  WHERE application_id=$4 AND retry_until>now())
				OR EXISTS (SELECT 1 FROM app.meli_notification_cooldowns
				  WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3
				    AND application_id=$4 AND available_at>now())`,
				scope.storeID, scope.ownerID, scope.parentID, e.ApplicationID, e.Topic, e.ResourceID,
				discovery.Site, discovery.Logistic, discovery.ChildID).Scan(&covered, &blocked)
			if err != nil {
				return err
			}
			if covered {
				return nil
			}
			if blocked {
				return ErrCapacity
			}
		} else {
			var suppress bool
			err := tx.QueryRow(ctx, `SELECT
				EXISTS (SELECT 1 FROM app.meli_notification_jobs
				  WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3
				    AND application_id=$4 AND topic=$5 AND resource_id=$6
				    AND (dead_at IS NULL OR dead_at>now()-interval '24 hours'))
				OR EXISTS (SELECT 1 FROM app.meli_notification_inbox
				  WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3
				    AND application_id=$4 AND topic=$5 AND resource_id=$6
				    AND enqueued_at>now()-CASE WHEN $5='marketplace_orders'
				      THEN interval '30 minutes' ELSE interval '60 minutes' END)
				OR (SELECT count(*) FROM app.meli_notification_jobs
				  WHERE application_id=$4 AND dead_at IS NULL)>=32
				OR EXISTS (SELECT 1 FROM app.meli_order_read_coordination
				  WHERE application_id=$4 AND retry_until>now())
				OR EXISTS (SELECT 1 FROM app.meli_notification_cooldowns
				  WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3
				    AND application_id=$4 AND available_at>now())`,
				scope.storeID, scope.ownerID, scope.parentID, e.ApplicationID, e.Topic, e.ResourceID).Scan(&suppress)
			if err != nil {
				return err
			}
			if suppress {
				return nil
			}
		}
	}
	return enqueueReceipt(ctx, tx, e, source, scopedReceiptSQL, mode, scope.storeID, scope.ownerID, scope.parentID)
}

// The live query locks candidate mappings, checks capacity, and writes the
// receipt and job together. Its snapshot starts after admission is acquired.
const webhookReceiptSQL = `WITH candidates AS MATERIALIZED (` + scopeCandidatesSQL + `)` + receiptSQL

// Compensation and replay already resolved and locked their scope.
const scopedReceiptSQL = `WITH candidates(store_id,owner_id,parent_id,ready) AS (
	SELECT $9::text,$10::text,$11::text,true
)` + receiptSQL

const receiptSQL = `,
	scope AS MATERIALIZED (
		SELECT store_id::uuid, owner_id::uuid, parent_id FROM candidates
		WHERE ready AND (SELECT count(*) FROM candidates)=1
	), admission AS MATERIALIZED (
		SELECT scope.*, EXISTS (
			SELECT 1 FROM app.meli_notification_inbox i
			WHERE i.store_id=scope.store_id AND i.owner_user_id=scope.owner_id AND i.application_id=$2
			  AND i.topic=$3 AND i.resource_id=$4 AND i.notification_user_id=$1
			  AND i.event_key=$5
		) AS duplicate, EXISTS (
			SELECT 1 FROM app.meli_notification_inbox OFFSET $8
		) OR (EXISTS (
			SELECT 1 FROM app.meli_notification_jobs OFFSET $8
		) AND NOT EXISTS (
			SELECT 1 FROM app.meli_notification_jobs j
			WHERE j.store_id=scope.store_id AND j.owner_user_id=scope.owner_id AND j.application_id=$2
			  AND j.topic=$3 AND j.resource_id=$4 AND j.parent_seller_id=scope.parent_id
			  AND j.dead_at IS NULL
			FOR UPDATE
		)) AS at_capacity
		FROM scope
	), receipt AS (
		INSERT INTO app.meli_notification_inbox
		    (store_id,owner_user_id,parent_seller_id,application_id,topic,resource_id,
		     notification_user_id,event_key,source_received_at,source)
		SELECT store_id,owner_id,parent_id,$2,$3,$4,$1,$5,$6::timestamptz,$7
		FROM admission WHERE NOT duplicate AND NOT at_capacity
		RETURNING store_id,owner_user_id,parent_seller_id,application_id,topic,
		          resource_id,notification_user_id
	), job AS (
		INSERT INTO app.meli_notification_jobs AS j
		    (store_id,owner_user_id,parent_seller_id,application_id,topic,resource_id,notification_user_id)
		SELECT * FROM receipt
		ON CONFLICT (store_id,owner_user_id,parent_seller_id,application_id,topic,resource_id)
		WHERE dead_at IS NULL
		DO UPDATE SET generation=j.generation+1, notification_user_id=EXCLUDED.notification_user_id,
		    attempts=0, pending_since=clock_timestamp(), last_failure=NULL
	)
	SELECT CASE
		WHEN (SELECT count(*) FROM candidates)<>1 THEN 'scope_rejected'
		WHEN NOT EXISTS (SELECT 1 FROM scope) THEN 'scope_unavailable'
		WHEN EXISTS (SELECT 1 FROM admission WHERE NOT duplicate AND at_capacity) THEN 'capacity'
		ELSE 'accepted'
	END`

func enqueueReceipt(ctx context.Context, tx pgx.Tx, e Event, source, query string, mode pgx.QueryExecMode, scope ...any) error {
	args := append([]any{mode, e.UserID, e.ApplicationID, e.Topic, e.ResourceID,
		e.Key, e.ReceivedAt, source, queueCapacity - 1}, scope...)
	var status string
	if err := tx.QueryRow(ctx, query, args...).Scan(&status); err != nil {
		return err
	}
	switch status {
	case "scope_rejected":
		return ErrScope
	case "scope_unavailable":
		return ErrScopeUnavailable
	case "capacity":
		return ErrCapacity
	case "accepted":
		return nil
	default:
		return errors.New("unknown notification admission result")
	}
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

package notifications

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const maxAttempts = 8

var errLeaseLost = errors.New("notification lease lost")

// ErrPermanentScope is reserved for proven identity contradictions, not a
// temporarily unavailable or ambiguous marketplace mapping.
var ErrPermanentScope = fmt.Errorf("permanent notification identity mismatch: %w", ErrScope)

// Job identifies the authorized scope for the handler's official read.
type Job struct {
	ID, StoreID, OwnerID, ParentSellerID, ApplicationID string
	NotificationUserID, Topic, ResourceID               string
	generation                                          int64
	attempts                                            int
	leaseToken                                          string
}

// Prepare must use the connected store's current OAuth authorization to GET the
// official resource. Return identity from that response and only local writes
// in Apply. The worker runs Apply in the fenced PostgreSQL transaction.
type Prepare func(context.Context, Job) (VerifiedResource, error)

type VerifiedResource struct {
	Topic, ResourceID, SellerID string
	Apply                       func(context.Context, LocalWriter) error
}

// LocalWriter keeps transaction completion under the worker's control.
type LocalWriter interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// RateLimited tells the worker to persist a store-wide cooldown across topics.
// RetryAfter is the duration derived from the upstream Retry-After header.
type RateLimited struct {
	RetryAfter time.Duration
	// Shared means the application-wide coordinator already persisted the limit.
	Shared bool
}

func (*RateLimited) Error() string { return "notification upstream rate limited" }

const (
	DeferBusy          = "local_busy"
	DeferAuthorization = "scope_unavailable"
	DeferAssociation   = "association_pending"
)

// Deferred is a bounded wait, not a failed upstream attempt. The lease is
// released and the claim refunded. Old pending jobs still reach dead letter.
type Deferred struct {
	Cause  error
	Reason string
	Delay  time.Duration
}

func (*Deferred) Error() string { return "notification deferred" }

func (e *Deferred) Unwrap() error { return e.Cause }

type Worker struct {
	inbox    PostgresInbox
	appID    string
	handlers map[string]Prepare
	topics   []string
}

func NewWorker(inbox PostgresInbox, appID string, handlers map[string]Prepare) *Worker {
	w := &Worker{inbox: inbox, appID: appID, handlers: make(map[string]Prepare), topics: []string{}}
	for _, topic := range []string{"marketplace_orders", "marketplace_shipments"} {
		if prepare := handlers[topic]; prepare != nil {
			w.handlers[topic] = prepare
			w.topics = append(w.topics, topic)
		}
	}
	return w
}

// ValidateActivation is required before exposing the receiver. Every currently
// accepted topic must have a real handler; otherwise a disabled deployment
// could fill its global queue with jobs that cannot be consumed.
func (w *Worker) ValidateActivation() error {
	for _, topic := range []string{"marketplace_orders", "marketplace_shipments"} {
		if w.handlers[topic] == nil {
			return errors.New("notification activation requires all accepted topic handlers")
		}
	}
	return nil
}

func (w *Worker) Run(ctx context.Context, logger *log.Logger) {
	// Deliberately single-flight, at most one claim per second per process.
	// Keep a single consumer until real traffic validates pacing and capacity.
	poll := time.NewTicker(time.Second)
	defer poll.Stop()
	maintenance := time.NewTicker(time.Minute)
	defer maintenance.Stop()
	report := func() {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := w.inbox.prune(ctx); err != nil {
			logger.Printf("notification queue maintenance failed: %s", failureCode(err))
		}
		stats, err := w.inbox.stats(ctx, w.appID, w.topics)
		if err != nil {
			logger.Printf("notification queue diagnostics failed: %s", failureCode(err))
			return
		}
		logger.Printf("notification queue receipts=%d receipts_global=%d jobs_total=%d pending=%d leased=%d dead=%d unregistered=%d cooling=%d expired_leases=%d oldest_pending=%s capacity=%d",
			stats.Receipts, stats.GlobalReceipts, stats.Jobs, stats.Pending, stats.Leased, stats.Dead, stats.Unregistered,
			stats.Cooling, stats.ExpiredLeases, stats.OldestPending, queueCapacity)
		if stats.Unregistered > 0 || stats.GlobalReceipts >= queueCapacity*8/10 || stats.Jobs >= queueCapacity*8/10 {
			logger.Print("WARNING: notification queue backlog or capacity requires operator review")
		}
	}
	report()
	for {
		select {
		case <-ctx.Done():
			return
		case <-maintenance.C:
			report()
		case <-poll.C:
			if err := w.runOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
				logger.Printf("notification queue dispatch failed: %s", failureCode(err))
			}
		}
	}
}

func failureCode(err error) string {
	var postgres *pgconn.PgError
	switch {
	case errors.As(err, &postgres):
		return "postgres_" + postgres.Code
	case errors.Is(err, errLeaseLost):
		return "lease_lost"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline"
	default:
		return "operation_failed"
	}
}

func (w *Worker) claim(ctx context.Context) (Job, error) {
	var j Job
	if len(w.topics) == 0 {
		return j, pgx.ErrNoRows
	}
	err := w.inbox.DB.QueryRow(ctx, `
		WITH exhausted AS (
			SELECT id FROM app.meli_notification_jobs
			WHERE application_id=$1 AND topic=ANY($2::text[])
			  AND dead_at IS NULL AND attempts >= $3 AND lease_until <= now()
			ORDER BY lease_until LIMIT 100 FOR UPDATE SKIP LOCKED
		), dead AS (
			UPDATE app.meli_notification_jobs
			SET dead_at=clock_timestamp(), lease_token=NULL, lease_until=NULL, last_failure='lease_expired'
			WHERE id IN (SELECT id FROM exhausted)
		), next AS (
			SELECT j.id FROM app.meli_notification_jobs j
			WHERE j.application_id=$1 AND j.topic=ANY($2::text[]) AND j.dead_at IS NULL
			  AND j.available_at <= now() AND (j.lease_until IS NULL OR j.lease_until <= now())
			  AND j.attempts < $3
			  AND NOT EXISTS (
				SELECT 1 FROM app.meli_order_read_coordination c
				WHERE c.application_id=j.application_id AND c.retry_until > now()
			  )
			  AND NOT EXISTS (
				SELECT 1 FROM app.meli_notification_cooldowns c
				WHERE (c.store_id,c.owner_user_id,c.parent_seller_id,c.application_id)
				     =(j.store_id,j.owner_user_id,j.parent_seller_id,j.application_id)
				  AND c.available_at > now()
			  )
			ORDER BY j.available_at,j.pending_since,j.id LIMIT 1 FOR UPDATE SKIP LOCKED
		)
		UPDATE app.meli_notification_jobs j
		SET lease_token=gen_random_uuid(), lease_until=clock_timestamp()+interval '45 seconds',
		    attempts=j.attempts+1
		FROM next WHERE next.id=j.id
		RETURNING j.id::text,j.store_id::text,j.owner_user_id::text,j.parent_seller_id,
		          j.application_id,j.notification_user_id,j.topic,j.resource_id,j.generation,
		          j.attempts,j.lease_token::text`, w.appID, w.topics, maxAttempts).Scan(
		&j.ID, &j.StoreID, &j.OwnerID, &j.ParentSellerID, &j.ApplicationID,
		&j.NotificationUserID, &j.Topic, &j.ResourceID, &j.generation, &j.attempts, &j.leaseToken)
	return j, err
}

func (w *Worker) runOnce(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 35*time.Second)
	defer cancel()
	j, err := w.claim(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	workCtx, stop := context.WithTimeout(ctx, 30*time.Second)
	defer stop()
	err = pgx.BeginFunc(workCtx, w.inbox.DB, func(tx pgx.Tx) error {
		return j.checkScope(workCtx, tx, j.NotificationUserID)
	})
	var resource VerifiedResource
	if err == nil {
		resource, err = w.handlers[j.Topic](workCtx, j)
	}
	if err == nil {
		err = w.finish(workCtx, j, resource)
	}
	if err == nil || errors.Is(err, errLeaseLost) {
		return err
	}
	if failErr := w.retry(ctx, j, err); failErr != nil {
		return failErr
	}
	return nil
}

func (j Job) checkScope(ctx context.Context, tx pgx.Tx, seller string) error {
	scope, err := resolveScope(ctx, tx, seller)
	if err != nil {
		return err
	}
	if scope != (storeScope{j.StoreID, j.OwnerID, j.ParentSellerID}) {
		return ErrScope
	}
	return nil
}

func (w *Worker) finish(ctx context.Context, j Job, resource VerifiedResource) error {
	if resource.Topic != j.Topic || resource.ResourceID != j.ResourceID ||
		!idPattern.MatchString(resource.SellerID) || resource.Apply == nil {
		return ErrScope
	}
	return pgx.BeginFunc(ctx, w.inbox.DB, func(tx pgx.Tx) error {
		sellers := []string{j.NotificationUserID}
		if resource.SellerID != j.NotificationUserID {
			sellers = append(sellers, resource.SellerID)
			slices.Sort(sellers)
		}
		for _, seller := range sellers {
			if err := j.checkScope(ctx, tx, seller); err != nil {
				return err
			}
		}
		var generation int64
		err := tx.QueryRow(ctx, `
			SELECT generation FROM app.meli_notification_jobs
			WHERE id=$1 AND lease_token=$2 AND lease_until > clock_timestamp()
			FOR UPDATE`, j.ID, j.leaseToken).Scan(&generation)
		if errors.Is(err, pgx.ErrNoRows) {
			return errLeaseLost
		}
		if err != nil {
			return err
		}
		if err = resource.Apply(ctx, struct{ LocalWriter }{tx}); err != nil {
			return err
		}
		// Leave a fresh pass if a new receipt arrived during the official read.
		query := `DELETE FROM app.meli_notification_jobs
		          WHERE id=$1 AND lease_token=$2 AND lease_until > clock_timestamp()`
		if generation != j.generation {
			query = `UPDATE app.meli_notification_jobs
			         SET attempts=0, lease_token=NULL, lease_until=NULL, last_failure=NULL,
			             pending_since=clock_timestamp(), available_at=clock_timestamp()
			         WHERE id=$1 AND lease_token=$2 AND lease_until > clock_timestamp()`
		}
		tag, err := tx.Exec(ctx, query, j.ID, j.leaseToken)
		if err == nil && tag.RowsAffected() != 1 {
			return errLeaseLost
		}
		return err
	})
}

func (w *Worker) retry(ctx context.Context, j Job, cause error) error {
	delay := 5 * time.Second << (j.attempts - 1)
	var limited *RateLimited
	var cooldown time.Duration
	var deferred *Deferred
	refund := false
	terminal := false
	backoff := false
	reason := "handler_failed"
	switch {
	case errors.As(cause, &deferred):
		switch deferred.Reason {
		case DeferBusy, DeferAuthorization, DeferAssociation:
			reason = deferred.Reason
			delay = min(max(deferred.Delay, time.Second), time.Hour)
			refund = true
			// Never poll a pending upstream state at a fixed rate: grow the
			// wait with the time already spent waiting, capped at one hour.
			backoff = deferred.Reason != DeferBusy
		}
	case errors.Is(cause, ErrScopeUnavailable):
		reason = DeferAuthorization
		delay = time.Minute
		refund = true
		backoff = true
	case errors.Is(cause, ErrPermanentScope):
		reason = "scope_rejected"
		terminal = true
	case errors.Is(cause, ErrScope):
		reason = "scope_rejected"
	case errors.As(cause, &limited):
		reason = "rate_limited"
		if limited.Shared {
			delay = max(time.Second, limited.RetryAfter)
		} else {
			cooldown = max(time.Minute, limited.RetryAfter)
			delay = cooldown
		}
		refund = true
	case errors.Is(cause, context.DeadlineExceeded):
		reason = "deadline"
	}
	return pgx.BeginFunc(ctx, w.inbox.DB, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE app.meli_notification_jobs
			SET lease_token=NULL, lease_until=NULL, last_failure=$3,
			    available_at=clock_timestamp()+CASE WHEN $8::boolean
			      THEN least(interval '1 hour',
			                 interval '1 minute' * power(2, least(floor(extract(epoch from clock_timestamp()-pending_since)/600), 6)))
			      ELSE $4*interval '1 second' END,
			    attempts=CASE WHEN $6::boolean THEN greatest(attempts-1,0) ELSE attempts END,
			    dead_at=CASE WHEN $7::boolean OR ($6::boolean AND pending_since <= clock_timestamp()-interval '24 hours')
			                      OR (NOT $6::boolean AND attempts >= $5) THEN clock_timestamp() END
			WHERE id=$1 AND lease_token=$2 AND lease_until > clock_timestamp()`,
			j.ID, j.leaseToken, reason, delay.Seconds(), maxAttempts, refund, terminal, backoff)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return errLeaseLost
		}
		if cooldown == 0 {
			return nil
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO app.meli_notification_cooldowns AS c
			    (store_id,owner_user_id,parent_seller_id,application_id,available_at)
			VALUES ($1,$2,$3,$4,clock_timestamp()+$5*interval '1 second')
			ON CONFLICT (store_id,owner_user_id,parent_seller_id,application_id)
			DO UPDATE SET available_at=greatest(c.available_at,EXCLUDED.available_at)`,
			j.StoreID, j.OwnerID, j.ParentSellerID, j.ApplicationID, cooldown.Seconds())
		return err
	})
}

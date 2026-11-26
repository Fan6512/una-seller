package notifications

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Search is injected by the existing Mercado Libre service. The scheduler
// stores only safe IDs and progress, never access tokens or search payloads.
type DiscoveryScope struct {
	StoreID, OwnerID, ParentID, ChildID, Site, Logistic string
}

type DiscoveryPage struct {
	IDs []string

	Count, Total int
}

type DiscoverySearch func(context.Context, DiscoveryScope, time.Time, time.Time, int) (DiscoveryPage, error)

// The provider can shrink a result set during offset pagination. Rewind the
// same fixed window, never advance its watermark on this signal.
var ErrDiscoveryPageChanged = errors.New("order discovery result set changed")

type discoveryCursor struct {
	DiscoveryScope

	from, to time.Time

	offset, count int

	more bool

	ids []string

	revision int64
}

func checkDiscoveryMapping(ctx context.Context, tx pgx.Tx, s DiscoveryScope) error {
	resolved, err := resolveScope(ctx, tx, s.ChildID)
	if err != nil {
		return err
	}
	if resolved != (storeScope{storeID: s.StoreID, ownerID: s.OwnerID, parentID: s.ParentID}) {
		return ErrScope
	}
	var ok bool
	err = tx.QueryRow(ctx, `SELECT
		(SELECT count(*) FROM jsonb_array_elements(c.marketplaces) m(value)
		 WHERE m.value->>'site_id'=$4 AND m.value->>'logistic_type'=$5)=1
		AND (SELECT count(*) FROM jsonb_array_elements(c.marketplaces) m(value)
		 WHERE m.value->>'site_id'=$4 AND m.value->>'user_id'=$6)=1
		AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
		 WHERE m.value->>'site_id'=$4 AND m.value->>'logistic_type'=$5 AND m.value->>'user_id'=$6)
	FROM app.meli_store_connections c WHERE c.id=$1 AND c.owner_user_id=$2
	  AND c.seller_id=$3 AND c.status='connected' AND c.marketplaces_status='verified' FOR SHARE`,
		s.StoreID, s.OwnerID, s.ParentID, s.Site, s.Logistic, s.ChildID).Scan(&ok)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !ok) {
		return ErrScope
	}
	return err
}

// One page or at most eight pending order IDs per tick. The current deployment
// has one scheduler; the row reservation/revision also fences stale commits
// without holding a connection or lock across an official network call.
func (p PostgresInbox) DiscoverySweep(ctx context.Context, appID string, search DiscoverySearch) error {
	if !idPattern.MatchString(appID) || search == nil {
		return errors.New("invalid order discovery configuration")
	}
	c, err := p.reserveDiscovery(ctx, appID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(c.ids) == 0 {
		page, err := search(ctx, c.DiscoveryScope, c.from, c.to, c.offset)
		if errors.Is(err, ErrDiscoveryPageChanged) {
			return p.rewindDiscovery(ctx, appID, c)
		}
		if err != nil {
			return err
		}
		if err := validateDiscoveryPage(page, c.offset); err != nil {
			return err
		}
		c.ids, c.count, c.more = page.IDs, page.Count, c.offset+page.Count < page.Total
		if len(c.ids) == 0 {
			return p.advanceDiscovery(ctx, appID, c, 0)
		}
		// Persist the entire page before admitting IDs; a restart never needs
		// to refetch a partially drained page whose membership may have changed.
		if err := p.saveDiscoveryPage(ctx, appID, c); err != nil {
			return err
		}
	}
	n := 0
	var admissionErr error
	for _, id := range c.ids[:min(len(c.ids), 8)] {
		key := sha256.Sum256([]byte("discovery:" + c.Site + ":" + c.Logistic + ":" + c.to.UTC().Format(time.RFC3339Nano)))
		e := Event{ApplicationID: appID, UserID: c.ChildID, Topic: "marketplace_orders", ResourceID: id,
			Key: hex.EncodeToString(key[:]), ReceivedAt: time.Now().UTC()}
		expected := storeScope{storeID: c.StoreID, ownerID: c.OwnerID, parentID: c.ParentID}
		if err := p.enqueueScoped(ctx, e, &expected, &c.DiscoveryScope); err != nil {
			admissionErr = err
			break
		}
		n++
	}
	// Every admitted ID is already durable, or covered by a local order/job.
	// If this checkpoint fails, the same page is retried idempotently.
	if n > 0 {
		if err := p.advanceDiscovery(ctx, appID, c, n); err != nil {
			return err
		}
	}
	return admissionErr
}

func validateDiscoveryPage(p DiscoveryPage, offset int) error {
	if offset < 0 || p.Count < 0 || p.Count > 3 || p.Total < offset || p.Count > p.Total-offset ||
		len(p.IDs) > 120 || (p.Count == 0) != (len(p.IDs) == 0) ||
		(p.Count == 0 && offset < p.Total) {
		return errors.New("invalid order discovery page")
	}
	seen := make(map[string]bool)
	for _, id := range p.IDs {
		if !ValidResourceID("marketplace_orders", id) || seen[id] {
			return errors.New("invalid order discovery ID")
		}
		seen[id] = true
	}
	return nil
}

func (p PostgresInbox) reserveDiscovery(ctx context.Context, appID string) (discoveryCursor, error) {
	var c discoveryCursor
	err := pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		// Create checkpoints only for current verified mappings. Reauthorization
		// with a different owner/parent/child creates a separate checkpoint.
		_, err := tx.Exec(ctx, `INSERT INTO app.meli_order_discovery
			(store_id,owner_user_id,parent_seller_id,application_id,site_id,logistic_type,seller_id)
			SELECT DISTINCT c.id,c.owner_user_id,c.seller_id,$1,m.value->>'site_id',m.value->>'logistic_type',m.value->>'user_id'
			FROM app.meli_store_connections c CROSS JOIN LATERAL jsonb_array_elements(c.marketplaces) m(value)
			WHERE c.status='connected' AND c.marketplaces_status='verified'
			  AND c.seller_id ~ '^[1-9][0-9]{0,29}$'
			  AND m.value->>'user_id' ~ '^[1-9][0-9]{0,29}$'
			  AND m.value->>'site_id' ~ '^[A-Z]{3,8}$' AND m.value->>'site_id'<>'CBT'
			  AND m.value->>'logistic_type' ~ '^[a-z_]{1,40}$'
			ON CONFLICT DO NOTHING`, appID)
		if err != nil {
			return err
		}
		// Oldest due scope first, one reservation per sweep. Fixed creation
		// windows use a two-minute indexing lag and 30-minute overlap.
		return tx.QueryRow(ctx, `WITH due AS (
			SELECT d.store_id,d.owner_user_id,d.parent_seller_id,d.application_id,d.site_id,d.logistic_type,d.seller_id
			FROM app.meli_order_discovery d JOIN app.meli_store_connections c ON c.id=d.store_id
			WHERE d.application_id=$1 AND d.next_due_at<=now()
			  AND c.owner_user_id=d.owner_user_id AND c.seller_id=d.parent_seller_id
			  AND c.status='connected' AND c.marketplaces_status='verified'
			  AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
			    WHERE m.value->>'site_id'=d.site_id AND m.value->>'logistic_type'=d.logistic_type AND m.value->>'user_id'=d.seller_id)
			  AND NOT EXISTS (SELECT 1 FROM app.meli_order_read_coordination
			    WHERE application_id=$1 AND retry_until>now())
			  AND NOT EXISTS (SELECT 1 FROM app.meli_notification_cooldowns x
			    WHERE x.store_id=d.store_id AND x.owner_user_id=d.owner_user_id AND x.parent_seller_id=d.parent_seller_id
			      AND x.application_id=$1 AND x.available_at>now())
			  AND (SELECT count(*) FROM app.meli_notification_jobs WHERE application_id=$1 AND dead_at IS NULL)<32
			  AND (d.window_to IS NOT NULL OR d.completed_to<date_trunc('second',now())-interval '2 minutes')
			ORDER BY d.next_due_at,d.store_id,d.site_id,d.logistic_type,d.seller_id
			LIMIT 1 FOR UPDATE OF d SKIP LOCKED
		)
		UPDATE app.meli_order_discovery d SET
		  window_from=COALESCE(d.window_from,d.completed_to-interval '30 minutes'),
		  window_to=COALESCE(d.window_to,least(d.completed_to+interval '24 hours',date_trunc('second',now())-interval '2 minutes')),
		  next_due_at=now()+interval '2 minutes',revision=d.revision+1
		FROM due WHERE (d.store_id,d.owner_user_id,d.parent_seller_id,d.application_id,d.site_id,d.logistic_type,d.seller_id)
		  =(due.store_id,due.owner_user_id,due.parent_seller_id,due.application_id,due.site_id,due.logistic_type,due.seller_id)
		RETURNING d.store_id::text,d.owner_user_id::text,d.parent_seller_id,d.seller_id,d.site_id,d.logistic_type,
		  d.window_from,d.window_to,d.next_offset,d.pending_ids,d.pending_count,d.pending_more,d.revision`, appID).
			Scan(&c.StoreID, &c.OwnerID, &c.ParentID, &c.ChildID, &c.Site, &c.Logistic,
				&c.from, &c.to, &c.offset, &c.ids, &c.count, &c.more, &c.revision)
	})
	return c, err
}

func (p PostgresInbox) saveDiscoveryPage(ctx context.Context, appID string, c discoveryCursor) error {
	return pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL synchronous_commit=on`); err != nil {
			return err
		}
		if err := checkDiscoveryMapping(ctx, tx, c.DiscoveryScope); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE app.meli_order_discovery
			SET pending_ids=$9,pending_count=$10,pending_more=$11
			WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3 AND application_id=$4
			  AND site_id=$5 AND logistic_type=$6 AND seller_id=$7 AND revision=$8 AND cardinality(pending_ids)=0`,
			c.StoreID, c.OwnerID, c.ParentID, appID, c.Site, c.Logistic, c.ChildID, c.revision, c.ids, c.count, c.more)
		if err == nil && tag.RowsAffected() != 1 {
			return ErrScopeUnavailable
		}
		return err
	})
}

func (p PostgresInbox) advanceDiscovery(ctx context.Context, appID string, c discoveryCursor, n int) error {
	return pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL synchronous_commit=on`); err != nil {
			return err
		}
		if err := checkDiscoveryMapping(ctx, tx, c.DiscoveryScope); err != nil {
			return err
		}
		remaining := c.ids[n:]
		if remaining == nil {
			remaining = []string{}
		}
		pageDone := len(remaining) == 0
		windowDone := pageDone && !c.more
		tag, err := tx.Exec(ctx, `UPDATE app.meli_order_discovery SET
		  pending_ids=$9,pending_count=CASE WHEN $10 THEN 0 ELSE pending_count END,
		  pending_more=CASE WHEN $10 THEN false ELSE pending_more END,
		  next_offset=CASE WHEN $11 THEN 0 WHEN $10 THEN next_offset+$12 ELSE next_offset END,
		  completed_to=CASE WHEN $11 THEN window_to ELSE completed_to END,
		  window_from=CASE WHEN $11 THEN NULL ELSE window_from END,
		  window_to=CASE WHEN $11 THEN NULL ELSE window_to END,
		  last_completed_at=CASE WHEN $11 THEN now() ELSE last_completed_at END,
		  next_due_at=now()+CASE WHEN $11 AND window_to>=now()-interval '32 minutes'
		    THEN interval '30 minutes' ELSE interval '1 minute' END,
		  revision=revision+1
		WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3 AND application_id=$4
		  AND site_id=$5 AND logistic_type=$6 AND seller_id=$7 AND revision=$8`,
			c.StoreID, c.OwnerID, c.ParentID, appID, c.Site, c.Logistic, c.ChildID, c.revision,
			remaining, pageDone, windowDone, c.count)
		if err == nil && tag.RowsAffected() != 1 {
			return ErrScopeUnavailable
		}
		return err
	})
}

func (p PostgresInbox) rewindDiscovery(ctx context.Context, appID string, c discoveryCursor) error {
	return pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL synchronous_commit=on`); err != nil {
			return err
		}
		if err := checkDiscoveryMapping(ctx, tx, c.DiscoveryScope); err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE app.meli_order_discovery
			SET next_offset=0,next_due_at=now()+interval '2 minutes',revision=revision+1
			WHERE store_id=$1 AND owner_user_id=$2 AND parent_seller_id=$3 AND application_id=$4
			  AND site_id=$5 AND logistic_type=$6 AND seller_id=$7 AND revision=$8
			  AND cardinality(pending_ids)=0 AND next_offset>0`,
			c.StoreID, c.OwnerID, c.ParentID, appID, c.Site, c.Logistic, c.ChildID, c.revision)
		if err == nil && tag.RowsAffected() != 1 {
			return ErrScopeUnavailable
		}
		return err
	})
}

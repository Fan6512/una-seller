package meli

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collectors/billing"
	"github.com/jackc/pgx/v5"
)

const billingCollectorLock int64 = 2026112702

type billingCollectionTarget struct {
	owner, store, order, pack, shipment string
	scope                               OrderScope
	evidence, revision                  string
	lastUpdated                         time.Time
	snapshot, attempt                   []byte
}

type billingCollectionAttempt struct {
	Evidence         string               `json:"evidence"`
	State            string               `json:"state"`
	BillingAttempts  int                  `json:"billing_attempts"`
	RelationAttempts int                  `json:"relation_attempts"`
	RetryUnix        int64                `json:"retry_unix"`
	Preview          *BillingOrderPreview `json:"preview,omitempty"`
}

type billingCollectionWork struct {
	target  billingCollectionTarget
	attempt billingCollectionAttempt
}

// One scope, bounded batch, explicit USD source. No full account billing
// export, no realtime per-order polling, no repeat of a processed observation.
func (s *Service) billingCollectionTargets(ctx context.Context, now time.Time) ([]billingCollectionTarget, error) {
	rows, err := s.db.Query(ctx, `WITH candidates AS (
	SELECT c.owner_user_id::text AS owner_id,o.store_id::text AS store_id,o.site_id,o.logistic_type,
		o.order_id,o.pack_id,o.shipment_id,o.date_created,o.last_updated,
		`+expectedProfitEvidence+` AS evidence,`+expectedProfitRevision+` AS revision,
		o.billing_snapshot,o.expected_profit_billing_attempt AS attempt
	FROM app.meli_orders o JOIN app.meli_store_connections c ON c.id=o.store_id
	WHERE c.status='connected' AND c.marketplaces_status='verified'
	  AND o.expected_profit_confirmation IS NULL AND `+expectedProfitEstimateEligible+`
	  AND o.date_created BETWEEN $1::timestamptz-interval '60 days' AND $1::timestamptz-interval '1 hour'
	  AND EXISTS (SELECT 1 FROM jsonb_array_elements(c.marketplaces) m(value)
	    WHERE m.value->>'site_id'=o.site_id AND m.value->>'logistic_type'=o.logistic_type
	      AND m.value->>'user_id'=o.seller_id)
	), due AS (
	SELECT * FROM candidates WHERE attempt->>'evidence' IS DISTINCT FROM evidence
	OR (attempt->>'state' IN ('pending','billing_read')
	  AND CASE WHEN attempt->>'retry_unix' ~ '^[0-9]{1,12}$'
	    THEN (attempt->>'retry_unix')::bigint ELSE 0 END <= $2
	  AND CASE WHEN attempt->>'state'='pending' THEN
	    CASE WHEN attempt->>'billing_attempts' ~ '^[0-9]{1,3}$'
	      THEN (attempt->>'billing_attempts')::int ELSE $3 END
	    ELSE CASE WHEN attempt->>'relation_attempts' ~ '^[0-9]{1,3}$'
	      THEN (attempt->>'relation_attempts')::int ELSE $3 END END < $3)
	), first_scope AS (
	SELECT owner_id,store_id,site_id,logistic_type FROM due
	ORDER BY date_created,store_id,site_id,order_id LIMIT 1
	)
	SELECT d.owner_id,d.store_id,d.site_id,d.logistic_type,d.order_id,d.pack_id,d.shipment_id,
		d.last_updated,d.evidence,d.revision,d.billing_snapshot,d.attempt
	FROM due d JOIN first_scope f USING (owner_id,store_id,site_id,logistic_type)
	ORDER BY d.date_created,d.order_id LIMIT $4`, now, now.Unix(), billing.MaxAttempts, billing.MaxBatch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var targets []billingCollectionTarget
	for rows.Next() {
		var t billingCollectionTarget
		if err := rows.Scan(&t.owner, &t.store, &t.scope.Site, &t.scope.Logistic, &t.order, &t.pack, &t.shipment,
			&t.lastUpdated, &t.evidence, &t.revision, &t.snapshot, &t.attempt); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

// CAS binds the captured full order and authorization to the durable attempt.
// A new cost or manual confirmation may win the race; we never overwrite it.
func (s *Service) saveBillingCollectionState(ctx context.Context, w billingCollectionWork, finish bool, net *BillingEstimatedUSDNet) error {
	t := w.target
	raw, err := json.Marshal(w.attempt)
	if err != nil {
		return err
	}
	var estimate any
	if net != nil {
		b, err := json.Marshal(storedBillingEstimate{BillingEstimatedUSDNet: *net, Evidence: t.evidence})
		if err != nil {
			return err
		}
		estimate = string(b)
	}
	var id string
	err = s.db.QueryRow(ctx, `WITH authorized AS (
		SELECT c.* FROM app.meli_store_connections c
		WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid FOR SHARE
	)
	UPDATE app.meli_orders o SET expected_profit_billing_attempt=$7::jsonb,
		expected_profit_billing_estimate=CASE WHEN $8::boolean THEN $9::jsonb ELSE o.expected_profit_billing_estimate END
	FROM authorized c WHERE o.store_id=c.id AND `+expectedProfitScope+`
	AND o.expected_profit_confirmation IS NULL AND `+expectedProfitEstimateEligible+`
	AND `+expectedProfitRevision+`=$6 RETURNING o.order_id`,
		t.store, t.owner, t.scope.Site, t.scope.Logistic, t.order, t.revision, string(raw), finish, estimate).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrderAuthorizationChanged
	}
	return err
}

func reusableBillingPreview(t billingCollectionTarget) *BillingOrderPreview {
	var snap storedOrderBilling
	if json.Unmarshal(t.snapshot, &snap) != nil || snap.ShipmentID != t.shipment ||
		snap.Preview.OrderID != t.order || !snap.Preview.Available || snap.Preview.CheckedAt == nil ||
		snap.Preview.CheckedAt.Before(t.lastUpdated) {
		return nil
	}
	return &snap.Preview
}

func (s *Service) refreshBillingEstimates(ctx context.Context, now time.Time) error {
	ctx = backgroundOrderRead(ctx)
	// One active pass even when two API processes overlap. The dedicated
	// session holds only an advisory lock, not a database transaction.
	conn, err := s.db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", billingCollectorLock).Scan(&locked); err != nil || !locked {
		return err
	}
	defer func() {
		unlock, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var released bool
		if err := conn.QueryRow(unlock, "SELECT pg_advisory_unlock($1)", billingCollectorLock).Scan(&released); err != nil || !released {
			_ = conn.Conn().Close(unlock)
		}
	}()
	targets, err := s.billingCollectionTargets(ctx, now)
	if err != nil || len(targets) == 0 {
		return err
	}
	var ready, fresh []billingCollectionWork
	retry := now.Add(billing.Interval).Unix()
	for _, t := range targets {
		a := billingCollectionAttempt{}
		if json.Unmarshal(t.attempt, &a) != nil || a.Evidence != t.evidence {
			a = billingCollectionAttempt{Evidence: t.evidence}
		}
		w := billingCollectionWork{target: t, attempt: a}
		switch {
		case a.State == "billing_read":
			if a.Preview == nil || a.Preview.OrderID != t.order || !a.Preview.Available || a.Preview.CheckedAt == nil {
				w.attempt.State, w.attempt.Preview = "review", nil
				if err := s.saveBillingCollectionState(ctx, w, true, nil); err != nil && !errors.Is(err, ErrOrderAuthorizationChanged) {
					return err
				}
				continue
			}
			ready = append(ready, w)
		default:
			if cached := reusableBillingPreview(t); cached != nil {
				w.attempt.State, w.attempt.Preview = "billing_read", cached
				w.attempt.RetryUnix = now.Unix()
				if err := s.saveBillingCollectionState(ctx, w, false, nil); errors.Is(err, ErrOrderAuthorizationChanged) {
					continue
				} else if err != nil {
					return err
				}
				ready = append(ready, w)
				continue
			}
			w.attempt.State, w.attempt.Preview = "pending", nil
			w.attempt.BillingAttempts++
			w.attempt.RetryUnix = retry
			if err := s.saveBillingCollectionState(ctx, w, false, nil); errors.Is(err, ErrOrderAuthorizationChanged) {
				continue
			} else if err != nil {
				return err
			}
			fresh = append(fresh, w)
		}
	}
	if len(fresh) > 0 {
		first := fresh[0].target
		ids := make([]string, len(fresh))
		for i, w := range fresh {
			ids[i] = w.target.order
		}
		page, err := s.readOrderBillingMode(ctx, first.owner, first.store, first.scope, ids, false, false)
		if err != nil {
			return err // Durable pending attempts prevent immediate repeat, including 429.
		}
		if len(page.Items) != len(fresh) {
			return ErrOrderInvalidResponse
		}
		observed := time.Now().UTC()
		for i, w := range fresh {
			item := page.Items[i]
			if item.OrderID != w.target.order {
				return ErrOrderInvalidResponse
			}
			if !item.Available {
				if w.attempt.BillingAttempts >= billing.MaxAttempts {
					w.attempt.State = "review"
				}
				if err := s.saveBillingCollectionState(ctx, w, true, nil); err != nil && !errors.Is(err, ErrOrderAuthorizationChanged) {
					return err
				}
				continue
			}
			item.CheckedAt = &observed
			w.attempt.State, w.attempt.Preview = "billing_read", &item
			w.attempt.RetryUnix = now.Unix()
			if err := s.saveBillingCollectionState(ctx, w, false, nil); errors.Is(err, ErrOrderAuthorizationChanged) {
				continue
			} else if err != nil {
				return err
			}
			ready = append(ready, w)
		}
	}
	// A successful Billing response is persisted before any relationship reads.
	// If these fail, the next pass resumes the saved preview, not Billing.
	for _, w := range ready {
		t := w.target
		w.attempt.RelationAttempts++
		w.attempt.RetryUnix = retry
		if err := s.saveBillingCollectionState(ctx, w, false, nil); errors.Is(err, ErrOrderAuthorizationChanged) {
			continue
		} else if err != nil {
			return err
		}
		report, err := s.probeBillingByPack(ctx, t.owner, t.store, t.scope, t.pack, w.attempt.Preview)
		if err != nil {
			return err
		}
		w.attempt.Preview = nil
		w.attempt.State = "review"
		if report.BillingEstimate != nil {
			w.attempt.State = "done"
		}
		if err := s.saveBillingCollectionState(ctx, w, true, report.BillingEstimate); err != nil && !errors.Is(err, ErrOrderAuthorizationChanged) {
			return err
		}
	}
	return nil
}

func (s *Service) RunBillingCollector(ctx context.Context, logger *log.Logger) {
	billing.Run(ctx, logger, s.refreshBillingEstimates)
}

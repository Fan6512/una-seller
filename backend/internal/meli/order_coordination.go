package meli

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

type backgroundOrderReadKey struct{}

func backgroundOrderRead(ctx context.Context) context.Context {
	return context.WithValue(ctx, backgroundOrderReadKey{}, true)
}

func isBackgroundOrderRead(ctx context.Context) bool {
	background, _ := ctx.Value(backgroundOrderReadKey{}).(bool)
	return background
}

// Foreground waits for an in-flight background read, but preserves the old
// busy result for foreground/foreground contention. Background never jumps a
// waiting foreground request. No in-flight HTTP request is forcibly canceled.
type orderRequestGate struct {
	mu         sync.Mutex
	busy       bool
	background bool
	waiters    int
	released   chan struct{}
}

func (g *orderRequestGate) acquire(ctx context.Context, background bool) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	waiting := false
	defer func() {
		if waiting {
			g.waiters--
		}
	}()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !g.busy && (!background || g.waiters == 0) {
			g.busy, g.background = true, background
			g.released = make(chan struct{})
			return nil
		}
		if background || (g.busy && !g.background) {
			return ErrOrderBusy
		}
		if !waiting {
			g.waiters++
			waiting = true
		}
		released := g.released
		g.mu.Unlock()
		select {
		case <-ctx.Done():
		case <-released:
		}
		g.mu.Lock()
	}
}

// Lock remains available for existing internal contention tests.
func (g *orderRequestGate) Lock() {
	if err := g.acquire(context.Background(), false); err != nil {
		panic(err)
	}
}

func (g *orderRequestGate) Unlock() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.busy, g.background = false, false
	close(g.released)
}

func laterOrderTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// EnableSharedOrderCoordination is called before the HTTP server starts, only
// when Webhook is enabled. Do not mutate this setting during request handling.
func (s *Service) EnableSharedOrderCoordination(ctx context.Context) error {
	if s.db == nil || s.cfg.ClientID == "" {
		return errors.New("order coordination requires database and application")
	}
	var exists bool
	if err := s.db.QueryRow(ctx, `SELECT to_regclass('app.meli_order_read_coordination') IS NOT NULL`).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return errors.New("order coordination migration required")
	}
	s.sharedOrderCoordination = true
	return nil
}

// Use a transaction-scoped advisory lock, compatible with transaction poolers.
// One bounded official GET holds the lock, not an entire notification or batch.
// All processes must opt into this coordinator before sharing an application.
// A separate cleanup context preserves 429 cooldown even after client cancel.
func (s *Service) beginSharedOrderRead(ctx context.Context) (func() error, error) {
	if !s.sharedOrderCoordination {
		return nil, nil
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	rollback := func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}
	var locked bool
	if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtextextended('una-seller:order-read:' || $1::text,0))`, s.cfg.ClientID).Scan(&locked); err != nil {
		rollback()
		return nil, err
	}
	if !locked {
		rollback()
		return nil, ErrOrderBusy
	}
	var next, search, retry time.Time
	err = tx.QueryRow(ctx, `SELECT next_request_at,next_search_at,retry_until
FROM app.meli_order_read_coordination WHERE application_id=$1`, s.cfg.ClientID).Scan(&next, &search, &retry)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		rollback()
		return nil, err
	}
	s.orderNextRequest = laterOrderTime(s.orderNextRequest, next)
	s.orderNextSearch = laterOrderTime(s.orderNextSearch, search)
	s.orderRetryUntil = laterOrderTime(s.orderRetryUntil, retry)
	return func() error {
		defer rollback()
		cleanup, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := tx.Exec(cleanup, `INSERT INTO app.meli_order_read_coordination AS c
(application_id,next_request_at,next_search_at,retry_until)
VALUES($1,$2,$3,$4)
ON CONFLICT(application_id) DO UPDATE SET
next_request_at=greatest(c.next_request_at,EXCLUDED.next_request_at),
next_search_at=greatest(c.next_search_at,EXCLUDED.next_search_at),
retry_until=greatest(c.retry_until,EXCLUDED.retry_until)`,
			s.cfg.ClientID, s.orderNextRequest, s.orderNextSearch, s.orderRetryUntil)
		if err != nil {
			return err
		}
		return tx.Commit(cleanup)
	}, nil
}

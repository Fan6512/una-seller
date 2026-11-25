package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrderReadForegroundPriority(t *testing.T) {
	var g orderRequestGate
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := g.acquire(ctx, true); err != nil {
		t.Fatal(err)
	}
	acquired := make(chan error, 1)
	go func() { acquired <- g.acquire(ctx, false) }()
	for {
		g.mu.Lock()
		waiting := g.waiters
		g.mu.Unlock()
		if waiting == 1 {
			break
		}
		select {
		case <-ctx.Done():
			g.Unlock()
			t.Fatal("foreground did not wait for background")
		case <-time.After(time.Millisecond):
		}
	}
	g.Unlock()
	if err := g.acquire(ctx, true); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("background jumped foreground: %v", err)
	}
	if err := <-acquired; err != nil {
		t.Fatal(err)
	}
	if err := g.acquire(ctx, false); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("foreground contention changed: %v", err)
	}
	g.Unlock()
	if err := g.acquire(ctx, true); err != nil {
		t.Fatal(err)
	}
	g.Unlock()
}

func TestSharedOrderReadCoordinationPostgres(t *testing.T) {
	s, db, _, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	// Each isolated test uses a distinct synthetic application identifier.
	s.cfg.ClientID = store
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), "DELETE FROM app.meli_order_read_coordination WHERE application_id=$1", store)
	})
	if err := s.EnableSharedOrderCoordination(ctx); err != nil {
		t.Fatal(err)
	}
	other, err := NewService(db, s.cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := other.EnableSharedOrderCoordination(ctx); err != nil {
		t.Fatal(err)
	}
	finish, err := s.beginSharedOrderRead(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.beginSharedOrderRead(ctx); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("second process obtained the advisory lock: %v", err)
	}
	if err := finish(); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()
	s.apiURL, other.apiURL = server.URL, server.URL
	var raw json.RawMessage
	status, err := s.orderGET(ctx, "fixture", "/marketplace/orders/123", &raw)
	var limited *OrderRateLimitError
	if status != 429 || !errors.As(err, &limited) || time.Until(limited.until) < 58*time.Second {
		t.Fatalf("shared cooldown floor: %d %v", status, err)
	}
	status, err = other.orderGET(backgroundOrderRead(ctx), "fixture", "/marketplace/orders/456", &raw)
	if status != 429 || !errors.As(err, &limited) || calls.Load() != 1 {
		t.Fatalf("new service bypassed durable cooldown: %d %v calls=%d", status, err, calls.Load())
	}
	var persisted time.Time
	if err := db.QueryRow(ctx, "SELECT retry_until FROM app.meli_order_read_coordination WHERE application_id=$1", store).Scan(&persisted); err != nil || time.Until(persisted) < 58*time.Second {
		t.Fatalf("cooldown not persisted: %v %v", persisted, err)
	}
}

// A cancelled request must release the advisory lock and must not reach the
// provider, so the next process is free to read.
func TestSharedOrderReadCancellationReleasesLock(t *testing.T) {
	s, db, _, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s.cfg.ClientID = store
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), "DELETE FROM app.meli_order_read_coordination WHERE application_id=$1", store)
	})
	if err := s.EnableSharedOrderCoordination(ctx); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	s.apiURL = server.URL
	// Force a pacing wait so the caller cancels while it holds the lock.
	s.orderNextRequest = time.Now().Add(5 * time.Second)
	short, stop := context.WithTimeout(ctx, 200*time.Millisecond)
	defer stop()
	var raw json.RawMessage
	if _, err := s.orderGET(short, "fixture", "/marketplace/orders/123", &raw); err == nil {
		t.Fatal("cancelled pacing wait reported success")
	}
	if calls.Load() != 0 {
		t.Fatalf("cancelled request reached the provider: %d", calls.Load())
	}
	other, err := NewService(db, s.cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := other.EnableSharedOrderCoordination(ctx); err != nil {
		t.Fatal(err)
	}
	finish, err := other.beginSharedOrderRead(ctx)
	if err != nil {
		t.Fatalf("advisory lock leaked after cancellation: %v", err)
	}
	if err := finish(); err != nil {
		t.Fatal(err)
	}
}

// When the coordination database is unavailable the official read must fail
// closed without touching the provider.
func TestSharedOrderReadRequiresDatabase(t *testing.T) {
	s, db, _, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	s.cfg.ClientID = store
	if err := s.EnableSharedOrderCoordination(ctx); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	s.apiURL = server.URL
	db.Close()
	var raw json.RawMessage
	if _, err := s.orderGET(ctx, "fixture", "/marketplace/orders/123", &raw); err == nil {
		t.Fatal("read succeeded without the coordination database")
	}
	if calls.Load() != 0 {
		t.Fatalf("database outage still reached the provider: %d", calls.Load())
	}
}

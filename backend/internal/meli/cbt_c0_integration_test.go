package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// This test uses the same PG17 migration and encrypted-store fixture as the
// OAuth refresh integration tests. All external CBT calls are synthetic.
func TestCBTC0PublicationModeOwnerUpsertReauthAndDisconnectPG(t *testing.T) {
	s, db, owner, storeID := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var seller string
	if err := db.QueryRow(ctx, `SELECT seller_id FROM app.meli_store_connections WHERE id=$1::uuid`, storeID).Scan(&seller); err != nil {
		t.Fatal(err)
	}
	const probeToken = "synthetic-c0-access"
	access, err := s.cipher.Seal(probeToken, tokenAAD(owner, seller, "access"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections
 SET access_token_ciphertext=$2,access_token_expires_at=now()+interval '2 hours'
 WHERE id=$1::uuid`, storeID, access); err != nil {
		t.Fatal(err)
	}

	var managed atomic.Bool
	var disconnectDuringProbe atomic.Bool
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Error("CBT C0 must not send query strings")
		}
		if r.URL.Path == "/oauth/token" {
			if r.Method != http.MethodPost {
				t.Error("OAuth reauthorization must use POST")
			}
			_, _ = w.Write([]byte(`{"access_token":"synthetic-reauth-access","refresh_token":"synthetic-reauth-refresh","expires_in":7200}`))
			return
		}
		calls.Add(1)
		if r.Method != http.MethodGet {
			t.Errorf("CBT probe unexpectedly wrote to %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer "+probeToken &&
			r.Header.Get("Authorization") != "Bearer synthetic-reauth-access" {
			t.Error("unexpected CBT authorization header")
		}
		switch r.URL.Path {
		case "/users/me":
			fmt.Fprintf(w, `{"id":%s,"site_id":"CBT","status":{"sell":{"allow":true}}}`, seller)
		case "/users/" + seller:
			fmt.Fprint(w, `{"tags":["user_product_seller"]}`)
		case "/marketplace/users/" + seller:
			business := ""
			if managed.Load() {
				business = "CBT CN Fulfillment Managed"
			}
			if disconnectDuringProbe.Load() {
				if err := s.Disconnect(ctx, owner, storeID); err != nil {
					t.Errorf("disconnect during probe: %v", err)
					http.Error(w, "fixture failed", http.StatusInternalServerError)
					return
				}
			}
			fmt.Fprintf(w, `{"user_id":%s,"site_id":"CBT","marketplaces":[{"user_id":987654,"site_id":"MLB","logistic_type":"remote","business_model":%q,"pricing_model":"net_proceeds","user_product":true}]}`, seller, business)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	s.client = server.Client()

	first, err := s.ProbePublicationMode(ctx, owner, storeID)
	if err != nil || first.Mode != "standard_up" {
		t.Fatalf("first C0 mode=%+v err=%v", first, err)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected exactly three GET calls, got %d", calls.Load())
	}

	var savedOwner, savedMode string
	var savedJSON []byte
	var count int
	err = db.QueryRow(ctx, `SELECT count(*),min(owner_user_id::text),min(publication_mode)
 FROM app.meli_cbt_c0_evidence WHERE store_id=$1::uuid`, storeID).Scan(&count, &savedOwner, &savedMode)
	if err != nil || count != 1 || savedOwner != owner || savedMode != "standard_up" {
		t.Fatalf("first SQL insert: count=%d owner-matches=%t mode=%s err=%v", count, savedOwner == owner, savedMode, err)
	}
	if err := db.QueryRow(ctx, `SELECT evidence FROM app.meli_cbt_c0_evidence WHERE store_id=$1::uuid`, storeID).Scan(&savedJSON); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(savedJSON), seller) || strings.Contains(string(savedJSON), probeToken) {
		t.Fatal("private account identity or token was persisted in C0 evidence")
	}

	var ownerB string
	if err := db.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(&ownerB); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ProbePublicationMode(ctx, ownerB, storeID); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("owner B accessed A's C0 probe: %v", err)
	}
	if _, err := s.ProbePublicationMode(ctx, owner, storeID); !errors.Is(err, ErrCBTProbeCooldown) {
		t.Fatalf("repeated C0 did not respect 10s cooldown: %v", err)
	}
	if _, err := s.ProbePublicationMode(ctx, owner, strings.ToUpper(storeID)); !errors.Is(err, ErrCBTProbeCooldown) {
		t.Fatalf("uppercase UUID bypassed the seller cooldown: %v", err)
	}
	if calls.Load() != 3 {
		t.Fatal("denied requests still called Mercado Libre")
	}

	// Expire the synthetic deadline without waiting ten seconds. The next
	// request must pass the real cooldown logic and update the same row.
	s.probeMu.Lock()
	s.probeNext[seller] = time.Now().Add(-time.Second)
	s.probeMu.Unlock()
	managed.Store(true)
	second, err := s.ProbePublicationMode(ctx, owner, storeID)
	if err != nil || second.Mode != "fully_managed" {
		t.Fatalf("second C0 mode=%+v err=%v", second, err)
	}
	err = db.QueryRow(ctx, `SELECT count(*),min(publication_mode)
 FROM app.meli_cbt_c0_evidence WHERE store_id=$1::uuid`, storeID).Scan(&count, &savedMode)
	if err != nil || count != 1 || savedMode != "fully_managed" {
		t.Fatalf("upsert must replace the same row: count=%d mode=%s err=%v", count, savedMode, err)
	}
	stores, err := s.List(ctx, owner)
	if err != nil || len(stores) != 1 || stores[0].PublicationMode != "fully_managed" || stores[0].PublicationEvidence == nil {
		t.Fatalf("C0 List did not read saved owner evidence: %+v err=%v", stores, err)
	}
	if other, err := s.List(ctx, ownerB); err != nil || len(other) != 0 {
		t.Fatalf("owner B saw another owner's store: %+v err=%v", other, err)
	}

	// Probe directly with a previously authorized snapshot. Change only the
	// local row so all three official mock GETs still match that snapshot.
	auth, err := s.RequireStoreConnection(ctx, owner, storeID)
	if err != nil {
		t.Fatal(err)
	}
	changedAccess, err := s.cipher.Seal("new-synthetic-token", tokenAAD(owner, seller, "access"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		change  string
		args    []any
		restore string
		oldArgs []any
	}{
		{"access_cipher_changed", "UPDATE app.meli_store_connections SET access_token_ciphertext=$2 WHERE id=$1::uuid", []any{storeID, changedAccess}, "UPDATE app.meli_store_connections SET access_token_ciphertext=$2 WHERE id=$1::uuid", []any{storeID, auth.accessCipher}},
		{"marketplaces_changed", "UPDATE app.meli_store_connections SET marketplaces=marketplaces || '[{\"user_id\":765432,\"site_id\":\"MLM\",\"logistic_type\":\"remote\"}]'::jsonb WHERE id=$1::uuid", []any{storeID}, "UPDATE app.meli_store_connections SET marketplaces=$2::jsonb WHERE id=$1::uuid", []any{storeID, string(auth.marketsJSON)}},
		{"newer_evidence", "UPDATE app.meli_cbt_c0_evidence SET checked_at=now()+interval '1 hour' WHERE store_id=$1::uuid", []any{storeID}, "UPDATE app.meli_cbt_c0_evidence SET checked_at=$2 WHERE store_id=$1::uuid", []any{storeID, second.CheckedAt}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := db.Exec(ctx, tc.change, tc.args...); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if _, err := db.Exec(ctx, tc.restore, tc.oldArgs...); err != nil {
					t.Fatal(err)
				}
			}()
			var before, after []byte
			query := "SELECT to_jsonb(e) FROM app.meli_cbt_c0_evidence e WHERE store_id=$1::uuid"
			if err := db.QueryRow(ctx, query, storeID).Scan(&before); err != nil {
				t.Fatal(err)
			}
			managed.Store(false)
			defer managed.Store(true)
			beforeCalls := calls.Load()
			if _, err := s.probeAuthorizedPublicationMode(ctx, owner, storeID, auth); !errors.Is(err, ErrCBTProbeStale) {
				t.Fatalf("stale snapshot was not rejected: %v", err)
			}
			if calls.Load()-beforeCalls != 3 {
				t.Fatal("probe did not reach the persistence guard")
			}
			if err := db.QueryRow(ctx, query, storeID).Scan(&after); err != nil || string(before) != string(after) {
				t.Fatalf("stale probe changed stored evidence: %v", err)
			}
		})
	}

	// Reauthorization replaces old evidence and attempts C0 once using
	// the new grant. The old Fully Managed mode must not survive.
	managed.Store(false)
	beforeReauthCalls := calls.Load()
	link, err := s.Start(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	reauth, err := s.Complete(ctx, parsed.Query().Get("state"), "synthetic-code", "")
	if err != nil || reauth.ID != storeID {
		t.Fatalf("reauthorization failed or changed store: %+v err=%v", reauth, err)
	}
	if calls.Load()-beforeReauthCalls != 5 || reauth.PublicationMode != "standard_up" || reauth.PublicationEvidence == nil {
		t.Fatalf("reauth did not automatically perform one C0 probe: calls=%d mode=%s", calls.Load()-beforeReauthCalls, reauth.PublicationMode)
	}
	if err := db.QueryRow(ctx, `SELECT count(*),min(publication_mode) FROM app.meli_cbt_c0_evidence
 WHERE store_id=$1::uuid`, storeID).Scan(&count, &savedMode); err != nil || count != 1 || savedMode != "standard_up" {
		t.Fatalf("reauth did not replace old evidence: count=%d mode=%s err=%v", count, savedMode, err)
	}
	stores, err = s.List(ctx, owner)
	if err != nil || len(stores) != 1 || stores[0].PublicationMode != "standard_up" || stores[0].PublicationEvidence == nil {
		t.Fatalf("reauth did not expose current C0: %+v err=%v", stores, err)
	}
	if calls.Load()-beforeReauthCalls != 5 {
		t.Fatal("reading the store list repeated the automatic C0 probe")
	}

	s.probeMu.Lock()
	delete(s.probeNext, seller)
	s.probeMu.Unlock()
	managed.Store(false)
	last, err := s.ProbePublicationMode(ctx, owner, storeID)
	if err != nil {
		t.Fatalf("new authorization C0 probe failed: %v", err)
	}
	s.probeMu.Lock()
	s.probeNext[seller] = time.Now().Add(-time.Second)
	s.probeMu.Unlock()
	managed.Store(true)
	disconnectDuringProbe.Store(true)
	if _, err := s.ProbePublicationMode(ctx, owner, storeID); !errors.Is(err, ErrCBTProbeStale) {
		t.Fatalf("probe must not write after disconnect: %v", err)
	}
	var checkedAt time.Time
	if err := db.QueryRow(ctx, `SELECT publication_mode,checked_at FROM app.meli_cbt_c0_evidence
 WHERE store_id=$1::uuid`, storeID).Scan(&savedMode, &checkedAt); err != nil ||
		savedMode != last.Mode || !checkedAt.Equal(last.CheckedAt.Truncate(time.Microsecond)) {
		t.Fatalf("stale probe changed stored evidence: mode=%s checked_at=%s err=%v", savedMode, checkedAt, err)
	}
	stores, err = s.List(ctx, owner)
	if err != nil || len(stores) != 1 || stores[0].Status != "disconnected" ||
		stores[0].PublicationMode != "undetermined" || stores[0].PublicationEvidence != nil ||
		stores[0].PublicationModeCheckedAt != nil {
		t.Fatalf("disconnected store exposed C0 evidence: %+v err=%v", stores, err)
	}

	// The persistent row may be retained for audit; disconnected UI must
	// never display it or claim active publishing capability.
	var evidence map[string]any
	if err := json.Unmarshal(savedJSON, &evidence); err != nil {
		t.Fatal(err)
	}
	if _, ok := evidence["publication_mode"]; !ok {
		t.Fatal("sanitized persisted evidence missing mode")
	}
}

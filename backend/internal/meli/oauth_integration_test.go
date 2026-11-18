package meli

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"

	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOAuthAuthorizationAndStoreIsolation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration requires TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}

	// Both seller and owner IDs are synthetic and unique per test run.
	var owner1, owner2 string
	if err := db.QueryRow(ctx, "SELECT gen_random_uuid()::text").Scan(&owner1); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, "SELECT gen_random_uuid()::text").Scan(&owner2); err != nil {
		t.Fatal(err)
	}
	sellerID := strconv.FormatInt(time.Now().UnixNano(), 10)
	const accessToken = "fake-access-token"
	const refreshToken = "fake-refresh-token"
	const callback = "https://example.test" + CallbackPath

	var challenge string
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth/token":
			if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				http.Error(w, "incorrect token request", 400)
				return
			}
			if err := r.ParseForm(); err != nil ||
				r.PostForm.Get("client_id") != "client_test" ||
				r.PostForm.Get("client_secret") != "server-secret" ||
				r.PostForm.Get("redirect_uri") != callback ||
				r.PostForm.Get("code") != "code_test" {
				http.Error(w, "bad grant", 400)
				return
			}
			verifier := r.PostForm.Get("code_verifier")
			proof := sha256.Sum256([]byte(verifier))
			if verifier == "" || base64.RawURLEncoding.EncodeToString(proof[:]) != challenge {
				http.Error(w, "incorrect PKCE proof", 400)
				return
			}
			_, _ = w.Write([]byte(`{"access_token":"fake-access-token","refresh_token":"fake-refresh-token","expires_in":10800}`))
		case "/users/me":
			if r.Header.Get("Authorization") != "Bearer "+accessToken {
				http.Error(w, "invalid token", 401)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": json.Number(sellerID), "nickname": "Synthetic CBT Seller"})
		case "/marketplace/users/" + sellerID:
			if r.Header.Get("Authorization") != "Bearer "+accessToken {
				http.Error(w, "invalid marketplace token", http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"user_id": json.Number(sellerID),
				"site_id": "CBT",
				"marketplaces": []map[string]any{
					{"user_id": 123456, "site_id": "MLM", "logistic_type": "remote"},
					{"user_id": 345678, "site_id": "MLB", "logistic_type": "remote"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer mock.Close()

	secret := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	service, err := NewService(db, Config{
		ClientID: "client_test", ClientSecret: "server-secret", RedirectURI: callback,
		TokenEncryptionKey: secret, PKCE: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	service.apiURL = mock.URL

	authorization, err := service.Start(ctx, owner1)
	if err != nil {
		t.Fatal(err)
	}
	u, err := urlParse(authorization)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("client_id") != "client_test" ||
		u.Query().Get("redirect_uri") != callback ||
		u.Query().Get("code_challenge_method") != "S256" {
		t.Fatal("authorization URL missing expected parameters")
	}
	state := u.Query().Get("state")
	challenge = u.Query().Get("code_challenge")
	if len(state) < 40 || challenge == "" {
		t.Fatal("missing CSRF state or PKCE challenge")
	}

	store, err := service.Complete(ctx, state, "code_test", "")
	if err != nil {
		t.Fatal(err)
	}
	if store.SellerID != sellerID || store.Nickname != "Synthetic CBT Seller" || store.Status != "connected" {
		t.Fatalf("store data mismatch: %+v", store)
	}
	if store.MarketsStatus != "verified" || len(store.Marketplaces) != 2 ||
		store.Marketplaces[0].SiteID != "MLM" || store.Marketplaces[0].UserID.String() != "123456" ||
		store.Marketplaces[1].SiteID != "MLB" {
		t.Fatalf("CBT parent/child marketplace mapping missing: %+v", store)
	}
	if _, err := service.Complete(ctx, state, "code_test", ""); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("replayed state accepted: %v", err)
	}

	var accessCipher, refreshCipher []byte
	if err := db.QueryRow(ctx, `SELECT access_token_ciphertext, refresh_token_ciphertext
 FROM app.meli_store_connections WHERE id=$1::uuid`, store.ID).Scan(&accessCipher, &refreshCipher); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(accessCipher), accessToken) || strings.Contains(string(refreshCipher), refreshToken) {
		t.Fatal("stored OAuth token was not encrypted")
	}
	if got, err := service.cipher.Open(refreshCipher, tokenAAD(owner1, sellerID, "refresh")); err != nil || got != refreshToken {
		t.Fatal("stored refresh token could not be decrypted for its owner")
	}
	if _, err := service.cipher.Open(refreshCipher, tokenAAD(owner2, sellerID, "refresh")); err == nil {
		t.Fatal("another owner's context decrypted the token")
	}

	stores, err := service.List(ctx, owner2)
	if err != nil || len(stores) != 0 {
		t.Fatalf("another owner can see the store: %v, %v", stores, err)
	}
	if err := service.Disconnect(ctx, owner2, store.ID); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("another owner disconnected the store: %v", err)
	}
	other, err := service.Start(ctx, owner2)
	if err != nil {
		t.Fatal(err)
	}
	otherURL, err := urlParse(other)
	if err != nil {
		t.Fatal(err)
	}
	challenge = otherURL.Query().Get("code_challenge")
	if _, err := service.Complete(ctx, otherURL.Query().Get("state"), "code_test", ""); !errors.Is(err, ErrSellerOwned) {
		t.Fatalf("seller ownership transfer was not refused: %v", err)
	}

	if err := service.Disconnect(ctx, owner1, store.ID); err != nil {
		t.Fatal(err)
	}
	stores, err = service.List(ctx, owner1)
	if err != nil || len(stores) != 1 || stores[0].Status != "disconnected" {
		t.Fatalf("disconnect did not persist: %v, %v", stores, err)
	}
	if err := db.QueryRow(ctx, `SELECT access_token_ciphertext IS NULL AND refresh_token_ciphertext IS NULL
 FROM app.meli_store_connections WHERE id=$1::uuid`, store.ID).Scan(new(bool)); err != nil {
		t.Fatal(err)
	}
	// Clean only the synthetic fixtures belonging to this test.
	_, _ = db.Exec(ctx, "DELETE FROM app.meli_store_connections WHERE id=$1::uuid", store.ID)
	_, _ = db.Exec(ctx, "DELETE FROM app.meli_oauth_states WHERE owner_user_id IN ($1::uuid,$2::uuid)", owner1, owner2)
	_, _ = db.Exec(ctx, "DELETE FROM app.users WHERE id IN ($1::uuid,$2::uuid)", owner1, owner2)
}

func TestInvalidStoreID(t *testing.T) {
	if validUUID("bad-store-id") || !validUUID("550e8400-e29b-41d4-a716-446655440000") {
		t.Fatal("unexpected store UUID validation")
	}
}

func urlParse(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return u, nil
}

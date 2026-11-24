package httpapi

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const fastTestUser = "6cbe41df-69cb-4594-9bad-a5473aa1a5fb"

func testClaims(issuer string) jwt.MapClaims {
	return jwt.MapClaims{
		"iss": issuer, "sub": fastTestUser, "aud": "authenticated", "role": "authenticated",
		"iat": time.Now().Add(-time.Minute).Unix(), "exp": time.Now().Add(time.Hour).Unix(),
	}
}

func testToken(t *testing.T, method jwt.SigningMethod, kid string, key any, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func testECJWK(kid string, key *ecdsa.PrivateKey) map[string]string {
	return map[string]string{
		"kid": kid, "alg": "ES256", "kty": "EC", "use": "sig", "crv": "P-256",
		"x": base64.RawURLEncoding.EncodeToString(key.X.FillBytes(make([]byte, 32))),
		"y": base64.RawURLEncoding.EncodeToString(key.Y.FillBytes(make([]byte, 32))),
	}
}

func TestAuthFastReadRouteAllowlist(t *testing.T) {
	id := "6cbe41df-69cb-4594-9bad-a5473aa1a5fb"
	cases := []struct {
		method, path string
		want         bool
	}{
		{"GET", "/api/v1/collections", true},
		{"GET", "/api/v1/products", true},
		{"GET", "/api/v1/collections/" + id, true},
		{"GET", "/api/v1/products/" + id, true},
		{"DELETE", "/api/v1/collections/" + id, false},
		{"POST", "/api/v1/collections", false},
		{"GET", "/api/v1/collections/" + id + "/working-draft", false},
		{"GET", "/api/v1/products/" + id + "/cost", false},
		{"GET", "/api/v1/stores", false},
		{"POST", "/api/v1/stores/oauth/start", false},
		{"GET", "/api/v1/collections/bad-id", false},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if got := authFastReadRoute(req); got != tc.want {
			t.Errorf("%s %s: fast=%v want %v", tc.method, tc.path, got, tc.want)
		}
	}
}

func TestFastAuthHS256PositiveCacheAndSensitiveBypass(t *testing.T) {
	var remoteCount atomic.Int32
	var acceptedToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/user" {
			http.NotFound(w, r)
			return
		}
		remoteCount.Add(1)
		if r.Header.Get("Authorization") != "Bearer "+acceptedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"` + fastTestUser + `"}`))
	}))
	defer srv.Close()
	v, err := NewSupabaseVerifier(srv.URL, "sb_publishable_test")
	if err != nil {
		t.Fatal(err)
	}
	token := testToken(t, jwt.SigningMethodHS256, "legacy", []byte("private-signing-secret"), testClaims(srv.URL+"/auth/v1"))
	acceptedToken = token
	for i := 0; i < 2; i++ {
		if id, err := v.VerifyFast(context.Background(), token); err != nil || id != fastTestUser {
			t.Fatalf("cached verification #%d: id=%q err=%v", i, id, err)
		}
	}
	if remoteCount.Load() != 1 {
		t.Fatalf("expected one remote request, got %d", remoteCount.Load())
	}
	// Mutations use the original Verify method and never read this cache.
	if _, err := v.Verify(context.Background(), token); err != nil || remoteCount.Load() != 2 {
		t.Fatalf("remote validation bypassed: err=%v count=%d", err, remoteCount.Load())
	}
	digest := sha256.Sum256([]byte(token))
	v.fast.mu.Lock()
	v.fast.positive[digest] = verifiedIdentity{userID: fastTestUser, until: time.Now().Add(-time.Second)}
	v.fast.mu.Unlock()
	if _, err := v.VerifyFast(context.Background(), token); err != nil || remoteCount.Load() != 3 {
		t.Fatalf("expired cache accepted: err=%v count=%d", err, remoteCount.Load())
	}

	// Expired JWTs are rejected locally; changed HS256 tokens miss cache and are rejected by Auth.
	expired := testClaims(srv.URL + "/auth/v1")
	expired["exp"] = time.Now().Add(-time.Minute).Unix()
	old := testToken(t, jwt.SigningMethodHS256, "legacy", []byte("private-signing-secret"), expired)
	if _, err := v.VerifyFast(context.Background(), old); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expired token: %v", err)
	}
	if _, err := v.VerifyFast(context.Background(), token+"broken"); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("forged token: %v", err)
	}
}

func TestFastAuthHS256ConcurrentRequestsCollapse(t *testing.T) {
	var remoteCount atomic.Int32
	var accepted string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteCount.Add(1)
		time.Sleep(25 * time.Millisecond)
		if r.Header.Get("Authorization") != "Bearer "+accepted {
			w.WriteHeader(401)
			return
		}
		_, _ = w.Write([]byte(`{"id":"` + fastTestUser + `"}`))
	}))
	defer srv.Close()
	v, _ := NewSupabaseVerifier(srv.URL, "sb_publishable_test")
	accepted = testToken(t, jwt.SigningMethodHS256, "legacy", []byte("local-secret"), testClaims(srv.URL+"/auth/v1"))
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := v.VerifyFast(context.Background(), accepted); err != nil {
				t.Errorf("concurrent VerifyFast: %v", err)
			}
		}()
	}
	group.Wait()
	if count := remoteCount.Load(); count != 1 {
		t.Fatalf("20 requests used %d Auth calls, want 1", count)
	}
}

func TestFastAuthJWKSValidClaimsAndRotation(t *testing.T) {
	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	keys := []map[string]string{testECJWK("key-1", key1)}
	var jwksHits atomic.Int32
	var authHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/v1/.well-known/jwks.json":
			jwksHits.Add(1)
			mu.Lock()
			snapshot := append([]map[string]string{}, keys...)
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": snapshot})
		case "/auth/v1/user":
			authHits.Add(1)
			w.WriteHeader(401)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	v, err := NewSupabaseVerifier(srv.URL, "sb_publishable_test")
	if err != nil {
		t.Fatal(err)
	}
	claims := testClaims(srv.URL + "/auth/v1")
	token1 := testToken(t, jwt.SigningMethodES256, "key-1", key1, claims)
	for i := 0; i < 3; i++ {
		if id, err := v.VerifyFast(context.Background(), token1); err != nil || id != fastTestUser {
			t.Fatalf("ES256 verify: id=%s err=%v", id, err)
		}
	}
	if jwksHits.Load() != 1 || authHits.Load() != 0 {
		t.Fatalf("unexpected requests jwks=%d auth=%d", jwksHits.Load(), authHits.Load())
	}
	// A small issuance clock skew is permitted, but larger drift is rejected.
	skew := testClaims(srv.URL + "/auth/v1")
	skew["iat"] = time.Now().Add(5 * time.Second).Unix()
	if id, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key1, skew)); err != nil || id != fastTestUser {
		t.Fatalf("5s future iat should be accepted: id=%q err=%v", id, err)
	}
	skew["iat"] = time.Now().Add(45 * time.Second).Unix()
	if _, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key1, skew)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("45s future iat should be rejected: %v", err)
	}
	bad := testClaims(srv.URL + "/auth/v1")
	bad["aud"] = "service_role"
	if _, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key1, bad)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("bad aud accepted: %v", err)
	}
	bad = testClaims("https://another-project.supabase.co/auth/v1")
	if _, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key1, bad)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("bad issuer accepted: %v", err)
	}
	bad = testClaims(srv.URL + "/auth/v1")
	bad["role"] = "service_role"
	if _, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key1, bad)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("service role accepted: %v", err)
	}
	bad = testClaims(srv.URL + "/auth/v1")
	bad["exp"] = time.Now().Add(-time.Minute).Unix()
	if _, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key1, bad)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expired JWT accepted: %v", err)
	}
	if _, err := v.VerifyFast(context.Background(), testToken(t, jwt.SigningMethodES256, "key-1", key2, claims)); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("forged signature accepted: %v", err)
	}
	// Rotation: a new signing kid is learned on one forced refresh.
	mu.Lock()
	keys = append(keys, testECJWK("key-2", key2))
	mu.Unlock()
	v.fast.mu.Lock()
	v.fast.lastJWKSFetch = time.Now().Add(-jwksMissRetryDelay - time.Second)
	v.fast.mu.Unlock()
	token2 := testToken(t, jwt.SigningMethodES256, "key-2", key2, claims)
	if id, err := v.VerifyFast(context.Background(), token2); err != nil || id != fastTestUser {
		t.Fatalf("rotated JWT: id=%q err=%v", id, err)
	}
	if jwksHits.Load() != 2 || authHits.Load() != 0 {
		t.Fatalf("JWKS rotation hits %d, Auth hits %d", jwksHits.Load(), authHits.Load())
	}
	for i := 0; i < 3; i++ {
		bogus := testToken(t, jwt.SigningMethodES256, "unknown-key-"+string(rune('A'+i)), key1, claims)
		if _, err := v.VerifyFast(context.Background(), bogus); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("unknown kid accepted: %v", err)
		}
	}
	if jwksHits.Load() != 2 {
		t.Fatalf("unknown kids triggered excess JWKS calls: %d", jwksHits.Load())
	}
}

func TestFastAuthJWKSEmptySetPreservesKeysAndRecovers(t *testing.T) {
	key1, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	// JWK "alg" is optional; key type and curve determine the allowed method.
	jwk1 := testECJWK("key-1", key1)
	delete(jwk1, "alg")
	jwk2 := testECJWK("key-2", key2)
	delete(jwk2, "alg")

	var phase atomic.Int32
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/.well-known/jwks.json" {
			http.NotFound(w, r)
			return
		}
		hits.Add(1)
		switch phase.Load() {
		case 0:
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{jwk1}})
		case 1:
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{}})
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{jwk1, jwk2}})
		}
	}))
	defer srv.Close()
	v, err := NewSupabaseVerifier(srv.URL, "sb_publishable_test")
	if err != nil {
		t.Fatal(err)
	}
	claims := testClaims(srv.URL + "/auth/v1")
	oldToken := testToken(t, jwt.SigningMethodES256, "key-1", key1, claims)
	newToken := testToken(t, jwt.SigningMethodES256, "key-2", key2, claims)
	if _, err := v.VerifyFast(context.Background(), oldToken); err != nil {
		t.Fatalf("optional JWK alg rejected: %v", err)
	}
	v.fast.mu.Lock()
	originalExpiry := v.fast.keysUntil
	v.fast.lastJWKSFetch = time.Now().Add(-jwksMissRetryDelay - time.Second)
	v.fast.mu.Unlock()

	// Empty JWKS is an outage, not a valid key rotation. Preserve old keys.
	phase.Store(1)
	if _, err := v.VerifyFast(context.Background(), newToken); !errors.Is(err, ErrAuthUnavailable) {
		t.Fatalf("empty JWKS should fail closed with service unavailable: %v", err)
	}
	if _, err := v.VerifyFast(context.Background(), oldToken); err != nil {
		t.Fatalf("previously cached key was discarded: %v", err)
	}
	for i := 0; i < 3; i++ {
		if _, err := v.VerifyFast(context.Background(), newToken); !errors.Is(err, ErrAuthUnavailable) {
			t.Fatalf("retry during backoff must fail closed: %v", err)
		}
	}
	if hits.Load() != 2 {
		t.Fatalf("empty JWKS caused repeated fetches: %d, want 2", hits.Load())
	}
	v.fast.mu.Lock()
	expiryUnchanged := v.fast.keysUntil.Equal(originalExpiry)
	v.fast.jwksRetryAfter = time.Time{}
	v.fast.lastJWKSFetch = time.Now().Add(-jwksMissRetryDelay - time.Second)
	v.fast.mu.Unlock()
	if !expiryUnchanged {
		t.Fatal("empty JWKS incorrectly prolonged key cache lifetime")
	}

	phase.Store(2)
	if id, err := v.VerifyFast(context.Background(), newToken); err != nil || id != fastTestUser {
		t.Fatalf("valid JWKS after recovery: id=%q err=%v", id, err)
	}
	if hits.Load() != 3 {
		t.Fatalf("JWKS retry after recovery: %d hits, want 3", hits.Load())
	}
}

func TestFastAuthJWKSRS256(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	var serverURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kid": "rsa-1", "alg": "RS256", "kty": "RSA", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
			"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.PublicKey.E)).Bytes()),
		}}})
	}))
	defer srv.Close()
	serverURL = srv.URL
	v, _ := NewSupabaseVerifier(serverURL, "sb_publishable_test")
	token := testToken(t, jwt.SigningMethodRS256, "rsa-1", key, testClaims(serverURL+"/auth/v1"))
	if id, err := v.VerifyFast(context.Background(), token); err != nil || id != fastTestUser {
		t.Fatalf("RSA JWKS verify: id=%q err=%v", id, err)
	}
}

func TestFastAuthJWKSOutageFailsClosed(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	v, _ := NewSupabaseVerifier(srv.URL, "sb_publishable_test")
	token := testToken(t, jwt.SigningMethodES256, "missing", key, testClaims(srv.URL+"/auth/v1"))
	for i := 0; i < 3; i++ {
		if _, err := v.VerifyFast(context.Background(), token); !errors.Is(err, ErrAuthUnavailable) {
			t.Fatalf("JWKS outage must fail closed with 503: %v", err)
		}
	}
	if hits.Load() != 1 {
		t.Fatalf("outage caused sequential JWKS request storm: %d, want 1", hits.Load())
	}
}

func TestFastAuthOversizedTokenRejected(t *testing.T) {
	v, _ := NewSupabaseVerifier("https://example.com", "sb_publishable_test")
	raw := strings.Repeat("x", maxAccessTokenSize+1)
	if _, err := v.VerifyFast(context.Background(), raw); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("oversized token was not rejected: %v", err)
	}
}

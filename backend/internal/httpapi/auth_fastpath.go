package httpapi

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/sync/singleflight"
)

const (
	positiveAuthTTL    = 30 * time.Second
	jwksCacheTTL       = 10 * time.Minute
	jwksMissRetryDelay = 30 * time.Second
	maxPositiveTokens  = 2048
	maxAccessTokenSize = 16 << 10
)

type verifiedIdentity struct {
	userID string
	until  time.Time
}

type publicSigningKey struct {
	alg string
	key any
}

// fastAuth is process-local and bounded. No access token, signing secret or
// user metadata is persisted, logged or exposed to the desktop.
type fastAuth struct {
	jwksURL       string
	issuer        string
	client        *http.Client
	mu            sync.Mutex
	keys          map[string]publicSigningKey
	keysUntil     time.Time
	lastJWKSFetch time.Time
	positive      map[[32]byte]verifiedIdentity
	tokenFlight   singleflight.Group
	jwksFlight    singleflight.Group
}

func newFastAuth(base string, client *http.Client) *fastAuth {
	return &fastAuth{
		jwksURL:  base + "/auth/v1/.well-known/jwks.json",
		issuer:   base + "/auth/v1",
		client:   client,
		keys:     make(map[string]publicSigningKey),
		positive: make(map[[32]byte]verifiedIdentity),
	}
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

func tokenParts(raw string) ([]string, tokenHeader, bool) {
	if raw == "" || len(raw) > maxAccessTokenSize {
		return nil, tokenHeader{}, false
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, tokenHeader{}, false
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || len(b) > 4096 {
		return nil, tokenHeader{}, false
	}
	var header tokenHeader
	if json.Unmarshal(b, &header) != nil {
		return nil, tokenHeader{}, false
	}
	return parts, header, true
}

func unverifiedExpiration(part string) time.Time {
	b, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil || len(b) > maxAccessTokenSize {
		return time.Time{}
	}
	var payload struct {
		Expires int64 `json:"exp"`
	}
	if json.Unmarshal(b, &payload) != nil || payload.Expires <= 0 {
		return time.Time{}
	}
	return time.Unix(payload.Expires, 0)
}

// verify is only called for explicitly allowlisted, read-only routes. The
// unverified JWT header only chooses a verification mechanism, never identity.
// Unsupported JWT algorithms fail closed; non-JWT legacy inputs keep using Auth.
func (f *fastAuth) verify(ctx context.Context, raw string, remote func(context.Context, string) (string, error)) (string, error) {
	if raw == "" || len(raw) > maxAccessTokenSize {
		return "", ErrInvalidSession
	}
	parts, header, ok := tokenParts(raw)
	if !ok {
		return remote(ctx, raw)
	}
	switch header.Algorithm {
	case "ES256", "RS256":
		if header.KeyID == "" {
			return "", ErrInvalidSession
		}
		return f.verifyAsymmetric(ctx, raw, header)
	case "HS256":
		return f.verifySymmetric(ctx, raw, parts[1], remote)
	default:
		return "", ErrInvalidSession
	}
}

func (f *fastAuth) verifySymmetric(ctx context.Context, raw, payload string, remote func(context.Context, string) (string, error)) (string, error) {
	exp := unverifiedExpiration(payload)
	// Never accept a locally expired JWT. Invalid or missing exp is handed
	// to the Auth server and, even on success, cannot enter the cache.
	if !exp.IsZero() && !time.Now().Before(exp) {
		return "", ErrInvalidSession
	}
	if exp.IsZero() {
		return remote(ctx, raw)
	}
	digest := sha256.Sum256([]byte(raw))
	f.mu.Lock()
	cached, ok := f.positive[digest]
	if ok && !time.Now().Before(cached.until) {
		delete(f.positive, digest)
		ok = false
	}
	f.mu.Unlock()
	if ok {
		return cached.userID, nil
	}
	// Requests sharing a token cannot independently stampede the Auth server.
	idKey := hex.EncodeToString(digest[:])
	result, err, _ := f.tokenFlight.Do(idKey, func() (any, error) {
		f.mu.Lock()
		cached, ok := f.positive[digest]
		if ok && !time.Now().Before(cached.until) {
			delete(f.positive, digest)
			ok = false
		}
		f.mu.Unlock()
		if ok {
			return cached.userID, nil
		}
		id, err := remote(ctx, raw)
		if err != nil {
			return "", err
		}
		now := time.Now()
		until := now.Add(positiveAuthTTL)
		if exp.Before(until) {
			until = exp
		}
		if until.After(now) {
			f.mu.Lock()
			if len(f.positive) >= maxPositiveTokens {
				for k, v := range f.positive {
					if !now.Before(v.until) {
						delete(f.positive, k)
					}
				}
				if len(f.positive) >= maxPositiveTokens {
					for k := range f.positive {
						delete(f.positive, k)
						break
					}
				}
			}
			f.positive[digest] = verifiedIdentity{userID: id, until: until}
			f.mu.Unlock()
		}
		return id, nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (f *fastAuth) verifyAsymmetric(ctx context.Context, raw string, header tokenHeader) (string, error) {
	publicKey, err := f.lookupKey(ctx, header.KeyID, header.Algorithm)
	if err != nil {
		return "", err
	}
	claims := jwt.MapClaims{}
	token, err := jwt.NewParser(
		jwt.WithValidMethods([]string{header.Algorithm}),
		jwt.WithIssuer(f.issuer),
		jwt.WithAudience("authenticated"),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	).ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != header.Algorithm || t.Header["kid"] != header.KeyID {
			return nil, ErrInvalidSession
		}
		return publicKey, nil
	})
	if err != nil || token == nil || !token.Valid || claims["role"] != "authenticated" {
		return "", ErrInvalidSession
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil || !time.Now().Before(exp.Time) {
		return "", ErrInvalidSession
	}
	id, err := claims.GetSubject()
	if err != nil || !collections.ValidUserID(id) {
		return "", ErrInvalidSession
	}
	return id, nil
}

func (f *fastAuth) lookupKey(ctx context.Context, kid, alg string) (any, error) {
	now := time.Now()
	f.mu.Lock()
	key, found := f.keys[kid]
	fresh := now.Before(f.keysUntil)
	recent := now.Sub(f.lastJWKSFetch) < jwksMissRetryDelay
	f.mu.Unlock()
	if fresh && found {
		if key.alg != alg {
			return nil, ErrInvalidSession
		}
		return key.key, nil
	}
	if fresh && recent {
		return nil, ErrInvalidSession
	}
	// Force one discovery refresh for unknown kid (throttled against attacker
	// supplied kids). Expired sets must refresh even when no key is missing.
	_, err, _ := f.jwksFlight.Do("jwks", func() (any, error) {
		f.mu.Lock()
		now := time.Now()
		key, found := f.keys[kid]
		fresh := now.Before(f.keysUntil)
		if fresh && (found || now.Sub(f.lastJWKSFetch) < jwksMissRetryDelay) {
			f.mu.Unlock()
			return nil, nil
		}
		f.lastJWKSFetch = now
		f.mu.Unlock()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.jwksURL, nil)
		if err != nil {
			return nil, ErrAuthUnavailable
		}
		resp, err := f.client.Do(req)
		if err != nil {
			return nil, ErrAuthUnavailable
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, ErrAuthUnavailable
		}
		var set struct {
			Keys []struct {
				KeyType   string `json:"kty"`
				Algorithm string `json:"alg"`
				KeyID     string `json:"kid"`
				Use       string `json:"use"`
				Curve     string `json:"crv"`
				X         string `json:"x"`
				Y         string `json:"y"`
				N         string `json:"n"`
				E         string `json:"e"`
			} `json:"keys"`
		}
		if json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&set) != nil {
			return nil, ErrAuthUnavailable
		}
		keys := make(map[string]publicSigningKey)
		for _, k := range set.Keys {
			if k.KeyID == "" || (k.Use != "" && k.Use != "sig") {
				continue
			}
			var pk any
			switch {
			case k.KeyType == "EC" && k.Algorithm == "ES256" && k.Curve == "P-256":
				xb, ex := base64.RawURLEncoding.DecodeString(k.X)
				yb, ey := base64.RawURLEncoding.DecodeString(k.Y)
				if ex != nil || ey != nil || len(xb) != 32 || len(yb) != 32 {
					continue
				}
				x, y := new(big.Int).SetBytes(xb), new(big.Int).SetBytes(yb)
				if !elliptic.P256().IsOnCurve(x, y) {
					continue
				}
				pk = &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
			case k.KeyType == "RSA" && k.Algorithm == "RS256":
				nb, en := base64.RawURLEncoding.DecodeString(k.N)
				eb, ee := base64.RawURLEncoding.DecodeString(k.E)
				if en != nil || ee != nil || len(nb) == 0 || len(eb) == 0 || len(eb) > 4 {
					continue
				}
				n := new(big.Int).SetBytes(nb)
				e := new(big.Int).SetBytes(eb).Int64()
				if n.BitLen() < 2048 || n.BitLen() > 4096 || e < 3 || e > 2147483647 || e%2 == 0 {
					continue
				}
				pk = &rsa.PublicKey{N: n, E: int(e)}
			}
			if pk != nil {
				keys[k.KeyID] = publicSigningKey{alg: k.Algorithm, key: pk}
			}
		}
		f.mu.Lock()
		f.keys = keys
		f.keysUntil = time.Now().Add(jwksCacheTTL)
		f.mu.Unlock()
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	key, found = f.keys[kid]
	fresh = time.Now().Before(f.keysUntil)
	f.mu.Unlock()
	if !fresh || !found || key.alg != alg {
		return nil, ErrInvalidSession
	}
	return key.key, nil
}

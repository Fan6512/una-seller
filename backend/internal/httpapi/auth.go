package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
)

// IdentityVerifier returns an authenticated Supabase Auth user ID, never a client-supplied owner ID.
type IdentityVerifier interface {
	Verify(ctx context.Context, accessToken string) (string, error)
}

var ErrAuthUnavailable = errors.New("authentication provider unavailable")
var ErrInvalidSession = errors.New("invalid session")

type SupabaseVerifier struct {
	userURL string
	apiKey  string
	client  *http.Client
	fast    *fastAuth
}

// Supabase Auth validates bearer tokens server-side; no JWT signing secret is needed.
func NewSupabaseVerifier(rawURL, publishableKey string) (*SupabaseVerifier, error) {
	u, err := url.Parse(strings.TrimRight(rawURL, "/"))
	if err != nil || u == nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return nil, errors.New("invalid SUPABASE_URL")
	}
	if u.Scheme != "https" {
		host := u.Hostname()
		if u.Scheme != "http" || (host != "localhost" && net.ParseIP(host) == nil) || !isLoopback(host) {
			return nil, errors.New("SUPABASE_URL must be HTTPS")
		}
	}
	if strings.TrimSpace(publishableKey) == "" {
		return nil, errors.New("SUPABASE_PUBLISHABLE_KEY is required")
	}
	v := &SupabaseVerifier{userURL: u.String() + "/auth/v1/user", apiKey: publishableKey, client: &http.Client{Timeout: 6 * time.Second}}
	v.fast = newFastAuth(u.String(), v.client)
	return v, nil
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (v *SupabaseVerifier) Verify(ctx context.Context, accessToken string) (string, error) {
	if v == nil || strings.TrimSpace(accessToken) == "" {
		return "", ErrInvalidSession
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.userURL, nil)
	if err != nil {
		return "", ErrAuthUnavailable
	}
	req.Header.Set("apikey", v.apiKey)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := v.client.Do(req)
	if err != nil {
		return "", ErrAuthUnavailable
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return "", ErrInvalidSession
	}
	if resp.StatusCode != http.StatusOK {
		return "", ErrAuthUnavailable
	}
	var user struct {
		ID string `json:"id"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&user); err != nil || !collections.ValidUserID(user.ID) {
		return "", ErrInvalidSession
	}
	return user.ID, nil
}

// VerifyFast is available only to the explicitly allowlisted read-only routes.
// Writes and sensitive routes continue using Verify() against Supabase Auth.
func (v *SupabaseVerifier) VerifyFast(ctx context.Context, accessToken string) (string, error) {
	if v == nil {
		return "", ErrInvalidSession
	}
	if v.fast == nil {
		return v.Verify(ctx, accessToken)
	}
	return v.fast.verify(ctx, accessToken, v.Verify)
}

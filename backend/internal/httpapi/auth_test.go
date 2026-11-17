package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSupabaseUserVerification(t *testing.T) {
	const userID = "6cbe41df-69cb-4594-9bad-a5473aa1a5fb"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/auth/v1/user" {
			t.Errorf("unexpected Supabase route: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
			return
		}
		if r.Header.Get("apikey") != "sb_publishable_test" {
			t.Errorf("publishable key missing")
		}
		switch r.Header.Get("Authorization") {
		case "Bearer valid":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"` + userID + `","email":"someone@example.com"}`))
		case "Bearer invalid-id":
			_, _ = w.Write([]byte(`{"id":"not-a-uuid"}`))
		case "Bearer outage":
			w.WriteHeader(503)
		case "Bearer sb_publishable_test", "Bearer legacy-anon-test", "Bearer forged", "Bearer expired":
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer srv.Close()

	v, err := NewSupabaseVerifier(srv.URL, "sb_publishable_test")
	if err != nil {
		t.Fatal(err)
	}
	if id, err := v.Verify(context.Background(), "valid"); err != nil || id != userID {
		t.Fatalf("valid session: id=%q err=%v", id, err)
	}
	for _, token := range []string{"", "invalid", "invalid-id", "sb_publishable_test", "legacy-anon-test", "forged", "expired"} {
		if _, err := v.Verify(context.Background(), token); !errors.Is(err, ErrInvalidSession) {
			t.Fatalf("token %q should be invalid: %v", token, err)
		}
	}
	if _, err := v.Verify(context.Background(), "outage"); !errors.Is(err, ErrAuthUnavailable) {
		t.Fatalf("provider outage must fail closed: %v", err)
	}
}

func TestSupabaseVerifierRejectsUntrustedURL(t *testing.T) {
	for _, raw := range []string{"", "http://example.org", "https://example.org/other", "https://user:pass@example.org", "https://example.org?bad=1"} {
		if _, err := NewSupabaseVerifier(raw, "sb_publishable_test"); err == nil {
			t.Fatalf("URL %q unexpectedly accepted", raw)
		}
	}
	if _, err := NewSupabaseVerifier("https://example.org", ""); err == nil {
		t.Fatal("missing API key unexpectedly accepted")
	}
}

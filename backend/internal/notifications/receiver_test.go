package notifications

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type captureSink struct {
	events []Event
	err    error
}

func (s *captureSink) Enqueue(_ context.Context, e Event) error {
	s.events = append(s.events, e)
	return s.err
}

const token = "synthetic-private-ingress-token-for-tests"
const sample = `{"_id":"event123","application_id":2069392825111111,"user_id":"465432224","topic":"marketplace_orders","resource":"/marketplace/orders/1499111111","received":"2019-10-09T13:58:23.329Z","attempts":1}`

func TestReceiverValidatesBeforeDurableACK(t *testing.T) {
	s := &captureSink{}
	h, err := NewReceiver("2069392825111111", token, s)
	if err != nil {
		t.Fatal(err)
	}
	post := func(method, body, auth string) int {
		r := httptest.NewRequest(method, Path, strings.NewReader(body))
		r.Header.Set(ingressHeader, auth)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	for _, tc := range []struct {
		method, body, auth string
		want               int
	}{
		{"GET", sample, token, 405},
		{"POST", sample, "", 403},
		{"POST", strings.Replace(sample, "2069392825111111", "999", 1), token, 400},
		{"POST", strings.Replace(sample, "/marketplace/orders/1499111111", "https://evil.example/1499111111", 1), token, 400},
		{"POST", strings.Replace(sample, "/marketplace/orders/1499111111", "/marketplace/orders/1499111111?x=1", 1), token, 400},
		{"POST", strings.Replace(sample, "marketplace_orders", "marketplace_questions", 1), token, 400},
		{"POST", sample + " {}", token, 400},
		{"POST", sample + strings.Repeat(" ", maxBody), token, 400},
	} {
		if got := post(tc.method, tc.body, tc.auth); got != tc.want {
			t.Fatalf("status %d want %d", got, tc.want)
		}
	}
	if len(s.events) != 0 {
		t.Fatal("invalid notification reached sink")
	}
	if got := post("POST", sample, token); got != http.StatusOK {
		t.Fatalf("valid status %d", got)
	}
	if got := post("POST", strings.Replace(sample, `"attempts":1`, `"attempts":2`, 1), token); got != 200 {
		t.Fatal(got)
	}
	if len(s.events) != 2 || s.events[0].Key != s.events[1].Key || s.events[0].ResourceID != "1499111111" {
		t.Fatal("retry identity changed")
	}
	s.err = errors.New("storage unavailable")
	if got := post("POST", sample, token); got != 503 {
		t.Fatal("storage failure acknowledged")
	}
	s.err = ErrScopeUnavailable
	if got := post("POST", sample, token); got != 503 {
		t.Fatal("temporarily unavailable seller must be retryable")
	}
	s.err = ErrScope
	if got := post("POST", sample, token); got != 403 {
		t.Fatal("unmapped seller acknowledged")
	}
}

func TestReceiverRequiresExplicitConfiguration(t *testing.T) {
	if _, err := NewReceiver("123", "", &captureSink{}); err == nil {
		t.Fatal("empty token accepted")
	}
	if _, err := NewReceiver("", "synthetic-private-ingress-token-for-tests", &captureSink{}); err == nil {
		t.Fatal("empty app accepted")
	}
}

func TestOpaqueNotificationID(t *testing.T) {
	body := strings.Replace(sample, "event123", "event:/v2.测试+id=", 1)
	first, err := parse(strings.NewReader(body), "2069392825111111")
	if err != nil {
		t.Fatal(err)
	}
	retry, err := parse(strings.NewReader(strings.Replace(body, `"attempts":1`, `"attempts":2`, 1)), "2069392825111111")
	if err != nil || retry.Key != first.Key || len(first.Key) != 64 {
		t.Fatal("opaque ID must hash consistently on retry")
	}
}

func TestWorkerActivationRequiresAllAcceptedTopics(t *testing.T) {
	prepare := func(context.Context, Job) (VerifiedResource, error) {
		return VerifiedResource{}, nil
	}
	for _, handlers := range []map[string]Prepare{
		nil,
		{"marketplace_orders": prepare},
		{"marketplace_shipments": prepare},
		{"marketplace_orders": prepare, "marketplace_shipments": nil},
	} {
		if err := NewWorker(PostgresInbox{}, "123", handlers).ValidateActivation(); err == nil {
			t.Fatal("incomplete handler registry accepted")
		}
	}
	if err := NewWorker(PostgresInbox{}, "123", map[string]Prepare{
		"marketplace_orders": prepare, "marketplace_shipments": prepare,
		"marketplace_questions": prepare, "marketplace_messages": prepare,
		"marketplace_claims": prepare, "marketplace_items": prepare,
	}).ValidateActivation(); err != nil {
		t.Fatal(err)
	}
}

func TestCustomerNotificationResourceAllowlist(t *testing.T) {
	for _, tc := range []struct {
		topic, path, id string
	}{
		{"marketplace_questions", "/marketplace/questions/123", "123"},
		{"marketplace_messages", "/marketplace/messages/a35cf79864a845ca9a3bf6aee59bb4d7", "a35cf79864a845ca9a3bf6aee59bb4d7"},
		{"marketplace_messages", "/marketplace/messages/../../orders/1", ""},
		{"marketplace_messages", "/marketplace/messages/abc123abc123abc123?mark_as_read=true", ""},
		{"marketplace_questions", "/marketplace/questions/123/answers", ""},
		{"marketplace_claims", "/marketplace/claims/123", "123"},
		{"marketplace_items", "/marketplace/items/MLB123", "MLB123"},
		{"marketplace_items", "/marketplace/items/CBT123", ""},
		{"marketplace_items", "/marketplace/items/MLB123/price_to_win", ""},
	} {
		body := strings.Replace(strings.Replace(sample, "marketplace_orders", tc.topic, 1), "/marketplace/orders/1499111111", tc.path, 1)
		got, err := parse(strings.NewReader(body), "2069392825111111")
		if tc.id == "" {
			if err == nil {
				t.Fatalf("unsafe resource accepted: %s", tc.path)
			}
		} else if err != nil || got.ResourceID != tc.id {
			t.Fatalf("valid resource rejected: %s %v", tc.path, err)
		}
	}
}

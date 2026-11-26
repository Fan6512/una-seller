package notifications

import (
	"context"
	"testing"
)

func TestExplicitEnabledTopics(t *testing.T) {
	for _, v := range []string{"", "marketplace_orders,", "*", "marketplace_orders,marketplace_orders", "unknown"} {
		if _, err := ParseEnabledTopics(v); err == nil {
			t.Fatalf("accepted invalid topics: %q", v)
		}
	}
	if got, err := ParseEnabledTopics("marketplace_orders, marketplace_shipments"); err != nil || len(got) != 2 {
		t.Fatalf("valid subset rejected: %v", err)
	}
}

func TestActivationSubset(t *testing.T) {
	w := NewWorker(PostgresInbox{}, "123", map[string]Prepare{
		"marketplace_orders": func(_ context.Context, _ Job) (VerifiedResource, error) { return VerifiedResource{}, nil },
	})
	if err := w.ValidateTopicsActivation([]string{"marketplace_orders"}); err != nil {
		t.Fatal(err)
	}
	if err := w.ValidateTopicsActivation([]string{"marketplace_orders", "marketplace_shipments"}); err == nil {
		t.Fatal("accepted unregistered topic")
	}
}

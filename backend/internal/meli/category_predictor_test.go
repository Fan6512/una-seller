package meli

import (
	"context"
	"errors"
	"testing"
)

func TestCBTCategoryPredictionRequiresEditedDescriptiveTitle(t *testing.T) {
	s := &Service{}
	for _, title := range []string{"", "shoes", "X", "   "} {
		_, err := s.PredictCBTCategories(context.Background(), "ignored", title)
		if !errors.Is(err, ErrCBTPredictorInput) {
			t.Fatalf("invalid title %q should be rejected before any token lookup: %v", title, err)
		}
	}
}
func TestCBTCategorySuggestionRejectsSiteIDs(t *testing.T) {
	for _, test := range []struct {
		id   string
		want bool
	}{
		{"CBT12345", true}, {"MLB12345", false}, {"MLM999", false},
		{"CBTabc", false}, {"CBT", false}, {"CBT123/other", false},
	} {
		if got := cbtCategoryID.MatchString(test.id); got != test.want {
			t.Fatalf("category %s: got %v want %v", test.id, got, test.want)
		}
	}
}

func TestPreferredCBTStoreRefreshesExpiredAccessToken(t *testing.T) {
	stores := []Store{
		{ID: "disconnected-id", Status: "disconnected"},
		{ID: "expired-id", Status: "token_expired"},
		{ID: "connected-id", Status: "connected"},
	}
	if got := preferredCBTStore(stores); got != "expired-id" {
		t.Fatalf("expected first refreshable CBT store, got %q", got)
	}
	if got := preferredCBTStore([]Store{{ID: "bad", Status: "reauth_required"}}); got != "" {
		t.Fatalf("reauth required must not trigger refresh: %q", got)
	}
}

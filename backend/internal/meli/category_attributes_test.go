package meli

import "testing"

func TestCategoryForAuthorizedSite(t *testing.T) {
	cases := []struct {
		category, site string
		want           bool
	}{
		{"MLB9891", "MLB", true},
		{"MLM123456", "MLM", true},
		{"MLB9891", "MLM", false},
		{"CBT9891", "CBT", false},
		{"MLB../../items", "MLB", false},
		{"MLB", "MLB", false},
		{"MLB123?foo=1", "MLB", false},
		{"mlb9891", "MLB", false},
		{"MLB9891", "", false},
		{"MLB1234567890123456", "MLB", false},
	}
	for _, tc := range cases {
		if got := validateCategoryForSite(tc.category, tc.site); got != tc.want {
			t.Errorf("validateCategoryForSite(%q,%q)=%v want %v", tc.category, tc.site, got, tc.want)
		}
	}
}

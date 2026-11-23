package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

const (
	mockToken = "synthetic-test-token"
	mockSelf  = 700000000000000001
	mockUser  = 700000000000000002
)

func TestProbeRedactsAccountEvidence(t *testing.T) {
	var mu sync.Mutex
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()
		if r.Method != http.MethodGet || r.URL.RawQuery != "" || r.Header.Get("Authorization") != "Bearer "+mockToken {
			t.Error("probe did not use a bearer-authenticated GET")
		}
		switch r.URL.Path {
		case "/users/me":
			fmt.Fprintf(w, `{"id":%d,"site_id":"CBT","status":{"sell":{"allow":true}},"email":"private-email"}`, mockSelf)
		case fmt.Sprintf("/users/%d", mockSelf):
			fmt.Fprint(w, `{"tags":["user_product_seller","global_seller","warehouse_management","private-tag"],"nickname":"private-nickname"}`)
		case fmt.Sprintf("/marketplace/users/%d", mockSelf):
			fmt.Fprintf(w, `{
                "user_id":%d,"site_id":"CBT","access_token":"private-token",
                "marketplaces":[
                    {"user_id":%d,"site_id":"MLM","logistic_type":"fulfillment","business_model":"CBT CN Fulfillment Managed","pricing_model":"net_proceeds","user_product":true},
                    {"user_id":%d,"site_id":"MLM","logistic_type":"remote","business_model":"private-business","pricing_model":"private-pricing","user_product":false},
                    {"user_id":%d,"site_id":"MCO","logistic_type":"remote","pricing_model":"listing_price"},
                    {"user_id":%d,"site_id":"MLB","logistic_type":"private-mode"}
                ]
            }`, mockSelf, mockUser, mockUser+1, mockUser+2, mockUser+3)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	got, err := probe(probeClient(), server.URL, mockToken)
	if err != nil {
		t.Fatal(err)
	}
	want := accountEvidence{
		SiteID:            "CBT",
		UserProductSeller: "present",
		GlobalSeller:      "present",
		WarehouseAccess:   "present",
		SellAllowed:       "enabled",
		Marketplaces: []marketplaceEvidence{
			{SellerAlias: "seller-1", SiteID: "MLM", LogisticType: "fulfillment", BusinessModel: "CBT CN Fulfillment Managed", PricingModel: "net_proceeds", UserProduct: "enabled"},
			{SellerAlias: "seller-2", SiteID: "MLM", LogisticType: "remote", BusinessModel: "unrecognized", PricingModel: "unrecognized", UserProduct: "disabled"},
			{SellerAlias: "seller-3", SiteID: "MCO", LogisticType: "remote", BusinessModel: unverified, PricingModel: "listing_price", UserProduct: unverified},
			{SellerAlias: "seller-4", SiteID: "MLB", LogisticType: "unrecognized", BusinessModel: unverified, PricingModel: unverified, UserProduct: unverified},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected sanitized evidence: %+v", got)
	}
	wantPaths := []string{"/users/me", fmt.Sprintf("/users/%d", mockSelf), fmt.Sprintf("/marketplace/users/%d", mockSelf)}
	mu.Lock()
	seenPaths := append([]string(nil), paths...)
	mu.Unlock()
	if !reflect.DeepEqual(seenPaths, wantPaths) {
		t.Fatalf("request paths = %v, want %v", seenPaths, wantPaths)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"private-", mockToken, fmt.Sprint(mockSelf), fmt.Sprint(mockUser), fmt.Sprint(mockUser + 1), fmt.Sprint(mockUser + 2), fmt.Sprint(mockUser + 3)} {
		if strings.Contains(string(encoded), secret) {
			t.Error("sanitized evidence contains a private value")
		}
	}
}

func TestProbeDoesNotTreatMissingTagsAsAbsent(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
		want string
	}{
		{"omitted", `{}`, unverified},
		{"null", `{"tags":null}`, unverified},
		{"empty", `{"tags":[]}`, "absent"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/users/me":
					fmt.Fprint(w, `{"id":1,"site_id":"CBT"}`)
				case "/users/1":
					fmt.Fprint(w, tt.body)
				case "/marketplace/users/1":
					fmt.Fprint(w, `{"user_id":1,"site_id":"CBT","marketplaces":[]}`)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			got, err := probe(probeClient(), server.URL, mockToken)
			if err != nil || got.UserProductSeller != tt.want || got.SellAllowed != unverified || got.Marketplaces == nil {
				t.Fatalf("evidence = %+v, err = %v", got, err)
			}
		})
	}
}

func TestProbeEnumAndNumericIDs(t *testing.T) {
	if got := knownValue("", "remote"); got != unverified {
		t.Fatalf("missing enum = %q", got)
	}
	if got := knownValue("self_service", "self_service"); got != "self_service" {
		t.Fatalf("valid enum = %q", got)
	}
	if got := knownValue("Bearer private-token", "remote"); got != "unrecognized" {
		t.Fatalf("unsafe enum leaked: %q", got)
	}
	if got := knownValue("remote_v2", "remote"); got != "unrecognized:remote_v2" {
		t.Fatalf("new public enum was lost: %q", got)
	}
	for _, unsafe := range []string{"private-token", "account_secret", "user_123", "TOKEN", "x@example.com", strings.Repeat("x", 40)} {
		if got := knownValue(unsafe, "remote"); got != "unrecognized" {
			t.Fatalf("unsafe enum echoed: %q", got)
		}
	}
	for _, input := range []string{`{"id":123}`, `{"id":"123"}`} {
		var item struct {
			ID json.Number `json:"id"`
		}
		if err := json.Unmarshal([]byte(input), &item); err != nil || item.ID.String() != "123" {
			t.Fatalf("numeric identifier: %s: %v", input, err)
		}
	}
	for _, input := range []string{`{"id":"abc"}`, `{"id":"123456789012345678901234567890123456789"}`} {
		var item struct {
			ID json.Number `json:"id"`
		}
		err := json.Unmarshal([]byte(input), &item)
		if strings.Contains(input, "abc") && err == nil {
			t.Fatal("nonnumeric ID accepted")
		}
		if err == nil && userIDFormat.MatchString(item.ID.String()) {
			t.Fatal("oversized numeric ID accepted")
		}
	}
}

func TestProbeBusinessModelValuesAndSensitiveEnumRedaction(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  string
	}{
		{"", unverified},
		{"CBT CN Fulfillment Managed", "CBT CN Fulfillment Managed"},
		{"CBT CN Fulfillment Non-Managed", "unrecognized:CBT CN Fulfillment Non-Managed"},
		{"seller_managed", "unrecognized:seller_managed"},
		{"user_defined_price", "unrecognized:user_defined_price"},
		{"CBT CN API Key", "unrecognized"},
		{"user_id_12345", "unrecognized"},
		{"CBT CN private-token", "unrecognized"},
		{"CBT CN Fulfillment\nNon-Managed", "unrecognized"},
		{"https://example.com", "unrecognized"},
		{strings.Repeat("X", 49), "unrecognized"},
	} {
		if got := knownBusinessModel(tc.value); got != tc.want {
			t.Errorf("business_model %q: got %q, want %q", tc.value, got, tc.want)
		}
	}
}

func TestProbeStopsOnInvalidAccountEvidence(t *testing.T) {
	for _, tt := range []struct {
		name    string
		user    string
		mapping string
		want    string
	}{
		{"wrong seller tags", `{"id":2,"tags":[]}`, "", "seller tags: account identity mismatch"},
		{"wrong mapping", `{"tags":[]}`, `{"user_id":2,"site_id":"CBT","marketplaces":[]}`, "marketplace mapping: expected the same CBT merchant and a marketplace array"},
		{"missing mapping", `{"tags":[]}`, `{"user_id":1,"site_id":"CBT"}`, "marketplace mapping: expected the same CBT merchant and a marketplace array"},
		{"invalid child", `{"tags":[]}`, `{"user_id":1,"site_id":"CBT","marketplaces":[{"user_id":1,"site_id":"MLM"}]}`, "marketplace mapping: invalid child identity"},
		{"invalid child site", `{"tags":[]}`, `{"user_id":1,"site_id":"CBT","marketplaces":[{"user_id":2,"site_id":"CBT"}]}`, "marketplace mapping: invalid child identity"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/users/me":
					fmt.Fprint(w, `{"id":1,"site_id":"CBT"}`)
				case "/users/1":
					fmt.Fprint(w, tt.user)
				case "/marketplace/users/1":
					fmt.Fprint(w, tt.mapping)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			got, err := probe(probeClient(), server.URL, mockToken)
			if err == nil || err.Error() != tt.want || got.Marketplaces != nil {
				t.Fatal("invalid account response produced evidence or unexpected error")
			}
		})
	}
}

func TestProbeDoesNotLeakFailedResponseOrFollowRedirect(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"forbidden", http.StatusForbidden, "private-upstream-body", "account identity: HTTP 403"},
		{"redirect", http.StatusFound, "private-upstream-body", "account identity: HTTP 302"},
		{"oversized", http.StatusOK, `{"id":1,"site_id":"CBT"}` + strings.Repeat(" ", maxBody), "account identity: invalid or oversized JSON response"},
		{"non-CBT", http.StatusOK, `{"id":1,"site_id":"MLM"}`, "account identity: expected a CBT merchant"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/users/me" {
					t.Error("probe sent an unexpected request")
				}
				w.Header().Set("Location", "/must-not-follow")
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			got, err := probe(probeClient(), server.URL, mockToken)
			if err == nil || err.Error() != tt.want || got.Marketplaces != nil {
				t.Fatal("unsafe or unrecognized response produced evidence or unexpected error")
			}
		})
	}
}

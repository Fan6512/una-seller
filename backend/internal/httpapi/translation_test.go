package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func TestTranslationServiceBatchesAndCachesThreeLanguages(t *testing.T) {
	var calls atomic.Int32
	translator := NewAzureTranslationService("mock-key", "mock-region")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		q := r.URL.Query()
		if r.Method != "POST" || q.Get("to") != "zh-Hans" || q.Get("api-version") != "3.0" ||
			r.Header.Get("Ocp-Apim-Subscription-Key") != "mock-key" ||
			r.Header.Get("Ocp-Apim-Subscription-Region") != "mock-region" ||
			(q.Get("from") != "en" && q.Get("from") != "es" && q.Get("from") != "pt") {
			t.Error("invalid provider request")
		}
		var batch []struct{ Text string }
		if json.NewDecoder(r.Body).Decode(&batch) != nil || len(batch) != 2 ||
			batch[0].Text == batch[1].Text || batch[1].Text != "Brand" {
			t.Error("provider did not receive two unique original strings in order")
		}
		_, _ = io.WriteString(w, `[{"translations":[{"text":"连衣裙","to":"zh-Hans"}]},{"translations":[{"text":"品牌","to":"zh-Hans"}]}]`)
	}))
	defer server.Close()
	translator.endpoint = server.URL
	for source, original := range map[string]string{"en": "Dress", "es": "Vestido", "pt": "Vestido curto"} {
		translations := translator.translate(context.Background(), source, []string{original, original, "Brand"})
		if len(translations) != 2 || translations[original] != "连衣裙" || translations["Brand"] != "品牌" {
			t.Fatalf("%s: unexpected translations: %v", source, translations)
		}
		hit := translator.translate(context.Background(), source, []string{original, "Brand"})
		if hit[original] != "连衣裙" {
			t.Fatalf("cache lost original key %q", original)
		}
	}
	if calls.Load() != 3 {
		t.Fatalf("every language should use one batch, called %d times", calls.Load())
	}
}

func TestTranslationServiceFallsBackAndHonorsCooldown(t *testing.T) {
	if service := NewAzureTranslationService("", ""); service != nil || len(service.translate(context.Background(), "en", []string{"Brand"})) != 0 {
		t.Fatal("unconfigured translation must return no translations")
	}
	for _, status := range []int{http.StatusTooManyRequests, http.StatusForbidden, http.StatusUnauthorized} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(status)
			}))
			defer server.Close()
			service := NewAzureTranslationService("mock-key", "")
			service.endpoint = server.URL
			for range 2 {
				if result := service.translate(context.Background(), "en", []string{"Brand"}); len(result) != 0 {
					t.Fatalf("failure manufactured translations: %v", result)
				}
			}
			if calls.Load() != 1 {
				t.Fatalf("provider cooldown did not stop repeated requests")
			}
		})
	}
	t.Run("timeout and offline", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		service := NewAzureTranslationService("mock-key", "")
		service.endpoint = server.URL
		service.client.Timeout = 20 * time.Millisecond
		if got := service.translate(context.Background(), "", []string{"Product title"}); len(got) != 0 {
			t.Fatalf("timeout returned translations: %v", got)
		}
		server.Close()
		if got := service.translate(context.Background(), "", []string{"Product title"}); len(got) != 0 {
			t.Fatalf("offline provider returned translations: %v", got)
		}
	})
}

func TestOrderTitleTranslationsAuthAndPayload(t *testing.T) {
	const route = "/api/v1/stores/order-title-translations"
	h := HandlerWithServices(&collections.Store{}, stubVerifier{id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb"}, &meli.Service{}, nil)
	for _, tc := range []struct {
		token, body string
		status      int
	}{
		{"", `[{"order_id":"10001"}]`, http.StatusUnauthorized},
		{"Bearer synthetic-session", `[{"order_id":"10001","title":"do not send arbitrary text"}]`, http.StatusBadRequest},
		{"Bearer synthetic-session", `[{"order_id":"10001"}]`, http.StatusOK},
	} {
		req := httptest.NewRequest(http.MethodPost, route, strings.NewReader(tc.body))
		if tc.token != "" {
			req.Header.Set("Authorization", tc.token)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != tc.status || tc.status == http.StatusOK && strings.TrimSpace(rec.Body.String()) != "{}" {
			t.Fatalf("route status %d, want %d: %s", rec.Code, tc.status, rec.Body.String())
		}
	}
}

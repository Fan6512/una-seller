package httpapi

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

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
	translator.provider.(*azureTranslationProvider).endpoint = server.URL
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
			service.provider.(*azureTranslationProvider).endpoint = server.URL
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
		service.provider.(*azureTranslationProvider).endpoint = server.URL
		service.provider.(*azureTranslationProvider).client.Timeout = 20 * time.Millisecond
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
	h := HandlerWithServices(&collections.Store{}, stubVerifier{id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb"}, &meli.Service{}, nil, nil)
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

func TestTranslationServiceProviderSelection(t *testing.T) {
	for _, tc := range []struct {
		selected, azureKey, baiduID, baiduSecret, want string
	}{
		{"", "", "", "", ""},
		{"", "mock-azure", "mock-baidu", "mock-value", "azure"},
		{"azure", "mock-azure", "mock-baidu", "mock-value", "azure"},
		{"baidu", "mock-azure", "mock-baidu", "mock-value", "baidu"},
		{"doubao", "mock-doubao", "", "", "doubao"},
		{"doubao", "", "", "", ""},
		{"volcengine", "mock-secret", "mock-access-id", "", "volcengine"},
		{"volcengine", "", "mock-access-id", "", ""},
		{"", "", "mock-baidu", "mock-value", ""},
		{"baidu", "mock-azure", "", "mock-value", ""},
		{"baidu", "mock-azure", "mock-baidu", "", ""},
		{"disabled", "mock-azure", "mock-baidu", "mock-value", ""},
		{"unexpected", "mock-azure", "mock-baidu", "mock-value", ""},
	} {
		service := NewTranslationService(tc.selected, tc.azureKey, "mock-region", tc.baiduID, tc.baiduSecret)
		actual := ""
		if service != nil {
			actual = service.providerName
		}
		if actual != tc.want {
			t.Errorf("selector %q returned %q, want %q", tc.selected, actual, tc.want)
		}
	}
}

func TestTranslationServiceBaiduSignatureAndSharedFrequency(t *testing.T) {
	const appID, secret = "mock-app", "mock-value"
	first := strings.Repeat("界", 989)
	const normalized, decomposed, third = "Café + %&?", "Cafe\u0301 + %&?", "Otro producto"
	called := make(chan time.Time, 4)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called <- time.Now()
		current := requests.Add(1)
		if r.Method != http.MethodPost || r.URL.RawQuery != "" ||
			r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || r.ParseForm() != nil {
			t.Error("expected credentials and metadata via the POST form")
			return
		}
		q, salt := r.PostForm.Get("q"), r.PostForm.Get("salt")
		wantSign := fmt.Sprintf("%x", md5.Sum([]byte(appID+q+salt+secret)))
		if salt == "" || r.PostForm.Get("sign") != wantSign || r.PostForm.Get("appid") != appID ||
			r.PostForm.Get("from") != "spa" || r.PostForm.Get("to") != "zh" || r.Header.Get("Authorization") != "" ||
			strings.Contains(r.PostForm.Encode(), secret) {
			t.Error("incorrect provider parameters")
		}
		if current == 1 && (q != first+"\n"+normalized || utf8.RuneCountInString(q) != 1000) ||
			current == 2 && q != third || current > 2 {
			t.Error("provider received an unexpected batch")
		}
		rows := make([]map[string]string, 0, 2)
		for i, original := range strings.Split(q, "\n") {
			rows = append(rows, map[string]string{"src": original, "dst": fmt.Sprintf("翻译 %d-%d", current, i)})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"from": "spa", "to": "zh", "trans_result": rows})
	}))
	defer server.Close()
	service := NewTranslationService("baidu", "", "", appID, secret)
	service.provider.(*baiduTranslationProvider).endpoint = server.URL

	overlong := strings.Repeat("界", 1001)
	result := service.translate(context.Background(), "es", []string{first, decomposed, normalized, overlong, "untrusted\nsplit"})
	if len(result) != 3 || result[first] != "翻译 1-0" || result[decomposed] != "翻译 1-1" || result[normalized] != "翻译 1-1" {
		t.Fatalf("first read mixed results or lost original keys: %#v", result)
	}
	result = service.translate(context.Background(), "es", []string{third, normalized})
	if len(result) != 2 || result[third] != "翻译 2-0" || result[normalized] != "翻译 1-1" || requests.Load() != 2 {
		t.Fatalf("second read failed or unexpectedly repeated a cached request: %#v", result)
	}
	firstRequest, secondRequest := <-called, <-called
	if delay := secondRequest.Sub(firstRequest); delay < time.Second {
		t.Fatalf("request interval %s exceeded 1 QPS", delay)
	}
}

func TestTranslationServiceBaiduAutoSeparatesLanguagesAndBoundsWait(t *testing.T) {
	var count atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		original := r.FormValue("q")
		labels := map[string][2]string{"Glass jar": {"en", "玻璃罐"}, "Pote de vidro": {"pt", "玻璃容器"}}
		if labels[original][0] == "" || r.FormValue("from") != "auto" {
			t.Error("auto-detection did not receive an individual known title")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"from": labels[original][0], "to": "zh",
			"trans_result": []map[string]string{{"src": original, "dst": labels[original][1]}}})
	}))
	defer server.Close()
	service := NewTranslationService("baidu", "", "", "mock-id", "mock-value")
	p := service.provider.(*baiduTranslationProvider)
	p.endpoint = server.URL
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	result := service.translate(ctx, "", []string{"Glass jar", "Pote de vidro"})
	if len(result) != 1 || result["Glass jar"] != "玻璃罐" || count.Load() != 1 {
		t.Fatalf("deadline did not preserve only the first completed translation: %#v", result)
	}
	result = service.translate(context.Background(), "", []string{"Glass jar", "Pote de vidro"})
	if len(result) != 2 || result["Pote de vidro"] != "玻璃容器" || count.Load() != 2 {
		t.Fatalf("auto-language results mixed or sent cached labels again: %#v", result)
	}
}

func TestTranslationServiceBaiduErrorsAndMapping(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		cooldown   bool
	}{
		{"frequency string", `{"error_code":"54003"}`, 200, true},
		{"frequency number", `{"error_code":54005}`, 200, true},
		{"HTTP rate limit", `{}`, 429, true},
		{"provider error", `{"error_code":"52001"}`, 200, false},
		{"invalid JSON", `{"from":`, 200, false},
		{"wrong target", `{"from":"en","to":"en","trans_result":[{"src":"Brand","dst":"品牌"},{"src":"Bag","dst":"包"}]}`, 200, false},
		{"wrong source language", `{"from":"pt","to":"zh","trans_result":[{"src":"Brand","dst":"品牌"},{"src":"Bag","dst":"包"}]}`, 200, false},
		{"swapped rows", `{"from":"en","to":"zh","trans_result":[{"src":"Bag","dst":"包"},{"src":"Brand","dst":"品牌"}]}`, 200, false},
		{"wrong count", `{"from":"en","to":"zh","trans_result":[{"src":"Brand","dst":"品牌"}]}`, 200, false},
		{"blank translation", `{"from":"en","to":"zh","trans_result":[{"src":"Brand","dst":"品牌"},{"src":"Bag","dst":" "}]}`, 200, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.Header().Set("Retry-After", "5")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			service := NewTranslationService("baidu", "", "", "mock-id", "mock-value")
			p := service.provider.(*baiduTranslationProvider)
			p.endpoint = server.URL
			if result := service.translate(context.Background(), "en", []string{"Brand", "Bag"}); len(result) != 0 || len(service.cache) != 0 {
				t.Fatalf("invalid result surfaced or cached a partial translation: %#v", result)
			}
			if tc.cooldown {
				if result := service.translate(context.Background(), "en", []string{"Another"}); len(result) != 0 || calls.Load() != 1 ||
					time.Until(p.retryAfter) < 4*time.Second {
					t.Fatal("provider cooldown did not apply to later requests")
				}
			}
		})
	}
	t.Run("timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()
		service := NewTranslationService("baidu", "", "", "mock-id", "mock-value")
		p := service.provider.(*baiduTranslationProvider)
		p.endpoint = server.URL
		p.client.Timeout = 20 * time.Millisecond
		if result := service.translate(context.Background(), "", []string{"Glass jar"}); len(result) != 0 || len(service.cache) != 0 {
			t.Fatalf("timed out call surfaced or cached a translation: %#v", result)
		}
	})
}

func TestTranslationServiceDedupesNFCAndBoundsCache(t *testing.T) {
	const composed, decomposed = "Café", "Cafe\u0301"
	var calls atomic.Int32
	release := make(chan struct{})
	entered := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}
		var texts []struct{ Text string }
		if json.NewDecoder(r.Body).Decode(&texts) != nil || len(texts) != 1 || texts[0].Text != composed {
			t.Error("expected one NFC input")
		}
		_, _ = io.WriteString(w, `[{"translations":[{"text":"咖啡","to":"zh-Hans"}]}]`)
	}))
	defer server.Close()
	service := NewAzureTranslationService("mock-key", "")
	service.provider.(*azureTranslationProvider).endpoint = server.URL
	first, second := make(chan map[string]string, 1), make(chan map[string]string, 1)
	go func() { first <- service.translate(context.Background(), "", []string{composed}) }()
	<-entered
	go func() { second <- service.translate(context.Background(), "", []string{decomposed}) }()
	close(release)
	for result, expected := range map[chan map[string]string]string{first: composed, second: decomposed} {
		if got := <-result; len(got) != 1 || got[expected] != "咖啡" {
			t.Fatalf("wrong original key or translated value: %#v", got)
		}
	}
	if calls.Load() != 1 {
		t.Fatal("concurrent NFC-equivalent translations repeated upstream work")
	}

	key := translationCacheKey{"azure", "", translationTarget, composed}
	if remaining := time.Until(service.cache[key].expires); remaining > translationCacheTTL || remaining < translationCacheTTL-time.Minute {
		t.Fatalf("unexpected cache TTL %s", remaining)
	}
	for i := 0; i < translationCacheLimit-1; i++ {
		otherKey := translationCacheKey{"azure", "", translationTarget, fmt.Sprint(i)}
		service.cache[otherKey] = translationCacheEntry{text: "旧", expires: time.Now().Add(time.Minute)}
	}
	// A full cache evicts an entry. An expired entry must be fetched again.
	service.cache[key] = translationCacheEntry{text: "过期", expires: time.Now().Add(-time.Second)}
	if got := service.translate(context.Background(), "", []string{composed}); got[composed] != "咖啡" || calls.Load() != 2 {
		t.Fatalf("expired cache entry reused: %#v", got)
	}
	if got := service.translate(context.Background(), "es", []string{composed}); got[composed] != "咖啡" ||
		calls.Load() != 3 || len(service.cache) != translationCacheLimit {
		t.Fatalf("cache crossed source languages or exceeded its bound: %#v", got)
	}
}

func TestDoubaoTranslationProtocolAndBatching(t *testing.T) {
	service := NewDoubaoTranslationService("mock-doubao-key")
	p := service.provider.(*doubaoTranslationProvider)
	calls := 0
	ids := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		id := r.Header.Get("X-Api-Request-Id")
		if r.Method != http.MethodPost || r.Header.Get("X-Api-Key") != "mock-doubao-key" ||
			r.Header.Get("X-Api-Resource-Id") != "volc.speech.mt" || len(id) != 36 || ids[id] ||
			r.Header.Get("Content-Type") != "application/json" {
			t.Error("invalid Doubao request")
		}
		ids[id] = true
		var body struct {
			Source string   `json:"source_language"`
			Target string   `json:"target_language"`
			Texts  []string `json:"text_list"`
		}
		if json.NewDecoder(r.Body).Decode(&body) != nil || body.Source != "en" ||
			body.Target != "zh" || len(body.Texts) == 0 || len(body.Texts) > 16 {
			t.Error("invalid Doubao body")
		}
		rows := make([]map[string]string, 0, len(body.Texts))
		for _, text := range body.Texts {
			rows = append(rows, map[string]string{"translation": "译：" + text})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 20000000, "data": map[string]any{"translation_list": rows}})
	}))
	defer server.Close()
	p.endpoint = server.URL
	texts := make([]string, 17)
	for i := range texts {
		texts[i] = fmt.Sprintf("category %d", i)
	}
	out := service.TranslateCategoryBatch(context.Background(), texts)
	if calls != 2 || len(out) != 17 || service.CategoryBatchSize() != 16 {
		t.Fatalf("invalid batch result: calls=%d count=%d", calls, len(out))
	}
	for _, text := range texts {
		if out[text] != "译："+text {
			t.Fatal("translation alignment changed")
		}
	}
	if len(service.TranslateCategoryBatch(context.Background(), texts)) != 17 || calls != 2 {
		t.Fatal("cache did not reuse completed translations")
	}
}

func TestDoubaoRejectsFailedBusinessResponse(t *testing.T) {
	for _, body := range []string{
		`{"code":45000001,"data":{"translation_list":[{"translation":"bad"}]}}`,
		`{"code":20000000,"data":{"translation_list":[]}}`,
	} {
		service := NewDoubaoTranslationService("mock-key")
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, body)
		}))
		service.provider.(*doubaoTranslationProvider).endpoint = server.URL
		if out := service.translate(context.Background(), "es", []string{"Marca"}); len(out) != 0 {
			t.Error("invalid response accepted")
		}
		server.Close()
	}
}

func TestVolcengineTranslateFixedSignature(t *testing.T) {
	body := []byte(`{"TargetLanguage":"zh","TextList":["Hello world"]}`)
	req, err := http.NewRequest(http.MethodPost, volcTranslateURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	signVolcTranslate(req, body, "mock-access-id", "mock-secret", time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC))
	want := "HMAC-SHA256 Credential=mock-access-id/20261127/cn-north-1/translate/request, SignedHeaders=content-type;host;x-content-sha256;x-date, Signature=164286a193e5c959bcc84ad8eeb9eb77c50a47afcba92978cdf8b26d09299162"
	if req.Header.Get("Authorization") != want || req.Header.Get("X-Date") != "20261127T000000Z" {
		t.Fatal("fixed signature mismatch")
	}
}

func TestVolcengineTranslationBatchAndBusinessError(t *testing.T) {
	service := NewVolcengineTranslationService("mock-access-id", "mock-secret")
	p := service.provider.(*volcengineTranslationProvider)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body struct {
			Source string   `json:"SourceLanguage"`
			Target string   `json:"TargetLanguage"`
			Texts  []string `json:"TextList"`
		}
		if r.Method != http.MethodPost || !strings.HasPrefix(r.Header.Get("Authorization"), "HMAC-SHA256 Credential=mock-access-id/") ||
			json.NewDecoder(r.Body).Decode(&body) != nil || body.Source != "en" || body.Target != "zh" || len(body.Texts) > 16 {
			t.Error("invalid Volcengine request")
		}
		if calls == 3 {
			_, _ = io.WriteString(w, `{"ResponseMetadata":{"Error":{"Code":"-429"}},"TranslationList":[{"Translation":"must not accept"}]}`)
			return
		}
		rows := make([]map[string]string, 0, len(body.Texts))
		for _, text := range body.Texts {
			rows = append(rows, map[string]string{"Translation": "译：" + text})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ResponseMetadata": map[string]any{"Error": nil}, "TranslationList": rows})
	}))
	defer server.Close()
	p.endpoint = server.URL
	texts := make([]string, 17)
	for i := range texts {
		texts[i] = fmt.Sprintf("label %d", i)
	}
	out := service.TranslateCategoryBatch(context.Background(), texts)
	if len(out) != 17 || calls != 2 || service.CategoryBatchSize() != 16 {
		t.Fatalf("unexpected result: %d labels, %d requests", len(out), calls)
	}
	for _, text := range texts {
		if out[text] != "译："+text {
			t.Fatal("translation alignment changed")
		}
	}
	if out := service.TranslateCategoryBatch(context.Background(), []string{"untranslated"}); len(out) != 0 {
		t.Fatal("business error accepted")
	}
	if out := service.TranslateCategoryBatch(context.Background(), []string{"another"}); len(out) != 0 || calls != 3 {
		t.Fatal("rate-limit cooldown ignored")
	}
}

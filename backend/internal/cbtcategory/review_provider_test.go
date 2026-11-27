package cbtcategory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLLMProviderBearerAndQuota(t *testing.T) {
	dir := t.TempDir()
	if _, err := ReconcileLLMUsage(dir, 0); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls++
		var body llmReviewRequest
		if req.Method != http.MethodPost || req.Header.Get("Authorization") != "Bearer mock-key" ||
			req.Header.Get("Content-Type") != "application/json" || json.NewDecoder(req.Body).Decode(&body) != nil ||
			body.AppID != "mock-app" || body.From != "en" || body.To != "zh" || body.Query != "Tools | Axes" ||
			body.ModelType != "llm" || body.Reference != llmCommerceReference || body.NeedIntervene != 0 {
			t.Error("incorrect LLM request")
		}
		state, err := loadLLMUsage(dir)
		if err != nil || state.Chars != 12 {
			t.Errorf("request not reserved before send: %+v %v", state, err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"from": "en", "to": "zh", "trans_result": []map[string]string{{"src": body.Query, "dst": "工具 | 斧头"}}})
	}))
	defer server.Close()
	p := llmReviewProvider{appID: "mock-app", key: "mock-key", endpoint: server.URL, client: server.Client()}
	if got, err := p.translate(context.Background(), dir, "Tools | Axes"); err != nil || got != "斧头" {
		t.Fatalf("unexpected result %q %v", got, err)
	}
	if _, err := ReconcileLLMUsage(dir, LLMTrialLimit); err != nil {
		t.Fatal(err)
	}
	if _, err := p.translate(context.Background(), dir, "Tools | Axes"); err == nil || calls != 1 {
		t.Fatal("quota overage reached upstream")
	}
}

func TestLLMProviderOfficialErrorAndResponseValidation(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{"official signature error", `{"error_code":"54001","error_msg":"Invalid Sign secret-must-not-leak"}`, "code=54001"},
		{"missing languages", `{"trans_result":[{"src":"Tools | Axes","dst":"工具 | 斧头"}]}`, "requested source"},
		{"wrong source", `{"from":"en","to":"zh","trans_result":[{"src":"Other","dst":"其他"}]}`, "requested source"},
		{"lost separator", `{"from":"en","to":"zh","trans_result":[{"src":"Tools | Axes","dst":"斧头"}]}`, "path separators"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if _, err := ReconcileLLMUsage(dir, 0); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			p := llmReviewProvider{appID: "mock-app", key: "mock-key", endpoint: server.URL, client: server.Client()}
			got, err := p.translate(context.Background(), dir, "Tools | Axes")
			if got != "" || err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "secret-must-not-leak") {
				t.Fatalf("unexpected result %q %v", got, err)
			}
			state, err := loadLLMUsage(dir)
			if err != nil || state.Chars != 12 {
				t.Fatalf("failed request lost reserved quota: %+v %v", state, err)
			}
		})
	}
}

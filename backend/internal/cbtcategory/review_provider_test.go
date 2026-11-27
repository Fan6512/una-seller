package cbtcategory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		var body map[string]string
		if req.Method != http.MethodPost || req.Header.Get("Authorization") != "Bearer mock-key" ||
			req.Header.Get("Content-Type") != "application/json" || json.NewDecoder(req.Body).Decode(&body) != nil ||
			body["appid"] != "mock-app" || body["from"] != "en" || body["to"] != "zh" || body["q"] != "Tools | Axes" {
			t.Error("incorrect LLM request")
		}
		state, err := loadLLMUsage(dir)
		if err != nil || state.Chars != 12 {
			t.Errorf("request not reserved before send: %+v %v", state, err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"from": "en", "to": "zh", "trans_result": []map[string]string{{"src": body["q"], "dst": "工具 | 斧头"}}})
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

package cbtcategory

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLLMBuildCheckpointAndDeferred(t *testing.T) {
	dir := t.TempDir()
	fixture := generatedVersion(t, version1, fixtureTree())
	s, err := readDump(bytes.NewReader(fixture.body), fixture.version)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, stagingFile, s); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, activeFile, s); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileLLMUsage(dir, 50001); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls++
		if _, err := os.Stat(filepath.Join(dir, llmBuildPendingFile)); err != nil {
			t.Error("request sent without pending receipt")
		}
		var body llmReviewRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		parts := strings.Split(body.Query, " | ")
		for i := range parts {
			parts[i] = "商品"
		}
		if calls == 2 {
			parts[len(parts)-1] = "Unverified English"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"from": "en", "to": "zh", "trans_result": []map[string]string{{"src": body.Query, "dst": strings.Join(parts, " | ")}}})
	}))
	defer server.Close()
	p := llmReviewProvider{appID: "mock", key: "mock", endpoint: server.URL, client: server.Client(), limit: LLMReviewHardLimit}
	plan, err := PlanLLMBuild(dir, "fixed", p.limit)
	if err != nil || !plan.BudgetFits || plan.Selected < 2 {
		t.Fatalf("bad build plan %+v %v", plan, err)
	}
	tooSmall := p
	tooSmall.limit = plan.Usage + plan.SourceChars - 1
	if _, err := runLLMBuild(context.Background(), dir, "fixed", 1, tooSmall); err == nil || calls != 0 {
		t.Fatal("full remaining budget was not enforced")
	}
	first, err := runLLMBuild(context.Background(), dir, "fixed", 1, p)
	if err != nil || first.Completed != 1 || first.ReservedChars != first.PlannedSourceChars || first.Usage == nil || *first.Usage <= 50001 {
		t.Fatalf("bad first checkpoint %+v %v", first, err)
	}
	second, err := runLLMBuild(context.Background(), dir, "fixed", 1, p)
	if err != nil || second.Completed != 0 || second.Deferred != 1 || len(second.Attempts) != 1 || first.Attempts[0].ID == second.Attempts[0].ID {
		t.Fatalf("bad deferred checkpoint %+v %v", second, err)
	}
	next, err := PlanLLMBuild(dir, "fixed", p.limit)
	if err != nil || next.Selected != plan.Selected-2 || next.Deferred != 1 || next.Pending || calls != 2 {
		t.Fatalf("bad resumed plan %+v %v", next, err)
	}
	journal, err := os.ReadFile(filepath.Join(dir, llmBuildReceiptsFile))
	if err != nil || bytes.Count(journal, []byte("\n")) != 2 || !bytes.Contains(journal, []byte("deferred_english")) {
		t.Fatalf("missing durable receipts: %s %v", journal, err)
	}
	if err := os.WriteFile(filepath.Join(dir, llmBuildPendingFile), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := runLLMBuild(context.Background(), dir, "fixed", 1, p); err == nil || calls != 2 {
		t.Fatal("unresolved attempt did not block requests")
	}
	trial := p
	trial.limit = LLMTrialLimit
	if _, err := runLLMTrial(context.Background(), dir, "all", "fixed", 1, true, trial); err == nil || !strings.Contains(err.Error(), "pending attempt") || calls != 2 {
		t.Fatalf("trial bypassed pending attempt: %v", err)
	}
	runLock, err := LockFile(filepath.Join(dir, "llm-review-run.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileLLMUsage(dir, 60000); err == nil {
		t.Error("reconciliation raced active run")
	}
	_ = runLock.Close()
	after, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("build modified active dictionary")
	}
}

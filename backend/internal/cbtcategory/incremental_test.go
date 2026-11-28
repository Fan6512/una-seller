package cbtcategory

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func prepareIncrementalFixture(t *testing.T) (string, snapshot, []byte, *httptest.Server, *fixtureVersion) {
	t.Helper()
	dir := t.TempDir()
	initialTree := fixtureTree()
	oldLeaf := PathNode{ID: "CBT202", Name: "Old"}
	initialTree["CBT200"] = fixtureNode(petsPath[:1], []PathNode{petsPath[1], oldLeaf}, "enabled")
	initialTree["CBT202"] = fixtureNode([]PathNode{petsPath[0], oldLeaf}, nil, "enabled")
	initial := generatedVersion(t, version1, initialTree)
	active, err := readDump(bytes.NewReader(initial.body), initial.version)
	if err != nil {
		t.Fatal(err)
	}
	active.Categories["CBT100"] = categoryWithLabel(active.Categories["CBT100"], "家居")
	active.Categories["CBT101"] = categoryWithLabel(active.Categories["CBT101"], "其他")
	active.Categories["CBT200"] = categoryWithLabel(active.Categories["CBT200"], "宠物")
	active.Categories["CBT201"] = categoryWithLabel(active.Categories["CBT201"], "其他")
	active.Categories["CBT202"] = categoryWithLabel(active.Categories["CBT202"], "旧分类")
	if err := saveJSON(dir, activeFile, active); err != nil {
		t.Fatal(err)
	}
	activeRaw, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, llmReleaseManagedFile), []byte("{\"managed\":true}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileLLMUsage(dir, 1000); err != nil {
		t.Fatal(err)
	}
	tree := fixtureTree()
	jugs := PathNode{ID: "CBT102", Name: "Jugs"}
	tree["CBT100"] = fixtureNode(homePath[:1], []PathNode{homePath[1], jugs}, "enabled")
	tree["CBT102"] = fixtureNode([]PathNode{homePath[0], jugs}, nil, "enabled")
	cats := PathNode{ID: "CBT201", Name: "Cats"}
	tree["CBT200"] = fixtureNode(petsPath[:1], []PathNode{cats}, "enabled")
	tree["CBT201"] = fixtureNode([]PathNode{petsPath[0], cats}, nil, "enabled")
	next := generatedVersion(t, version2, tree)
	server, latest := fixtureServer(t)
	latest.Store(next)
	manual := []override{{Path: []PathNode{homePath[0], jugs}, Zh: "水壶"}}
	if err := saveJSON(dir, overrideFile, manual); err != nil {
		t.Fatal(err)
	}
	return dir, active, activeRaw, server, next
}

func categoryWithLabel(item category, label string) category {
	item.Zh = label
	return item
}

func runIncrementalPlanFixture(t *testing.T, dir string, server *httptest.Server) IncrementalPlanReport {
	t.Helper()
	result, err := planIncremental(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all")
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestIncrementalPlanProtectsManagedSnapshotAndDetectsDelta(t *testing.T) {
	dir, _, activeRaw, server, _ := prepareIncrementalFixture(t)
	result := runIncrementalPlanFixture(t, dir, server)
	if result.Counts.Added != 1 || result.Counts.ChangedPath != 1 || result.Counts.Removed != 1 ||
		result.Counts.TranslationNeeded != 1 || result.Counts.ReviewRequired != 2 ||
		!slices.Contains(result.AddedIDs, "CBT102") || !slices.Contains(result.ChangedPathIDs, "CBT201") ||
		!slices.Contains(result.RemovedIDs, "CBT202") {
		t.Fatalf("unexpected incremental plan: %+v", result)
	}
	if result.SourceChars != len([]rune("Pets | Cats")) || result.BudgetRemaining != LLMReviewHardLimit-1000 || !result.BudgetFits {
		t.Fatalf("incorrect incremental quota estimate: %+v", result)
	}
	after, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil || !bytes.Equal(activeRaw, after) {
		t.Fatal("incremental plan modified active dictionary")
	}
	if _, err := os.Stat(filepath.Join(dir, stagingFile)); !os.IsNotExist(err) {
		t.Fatal("incremental plan reused ordinary staging state")
	}
	if _, err := os.Stat(filepath.Join(dir, progressFile)); !os.IsNotExist(err) {
		t.Fatal("incremental plan reused ordinary progress journal")
	}
	var manifest incrementalManifest
	if err := readJSONFile(filepath.Join(dir, incrementalManifestFile), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Rows[0].ID != "CBT102" || manifest.Rows[0].Origin != "manual" ||
		manifest.Rows[0].NeedsTranslation || manifest.Rows[1].ID != "CBT201" ||
		manifest.Rows[1].Origin != "pending" || !manifest.Rows[1].NeedsTranslation {
		t.Fatalf("manual override priority or path delta incorrect: %+v", manifest.Rows)
	}
	if err := os.WriteFile(filepath.Join(dir, incrementalCandidatesFile), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := planIncremental(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all"); err == nil ||
		!strings.Contains(err.Error(), "already in progress") {
		t.Fatalf("a new plan overwrote an in-progress incremental run: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, incrementalCandidatesFile)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, llmReleaseManagedFile)); err != nil {
		t.Fatal(err)
	}
	if _, err := planIncremental(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all"); err == nil {
		t.Fatal("incremental plan accepted missing managed marker")
	}
}

func TestIncrementalBuildAssemblePublishAndRollback(t *testing.T) {
	dir, _, activeRaw, server, _ := prepareIncrementalFixture(t)
	plan := runIncrementalPlanFixture(t, dir, server)
	var calls int
	translator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls++
		var body llmReviewRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.Query != "Pets | Cats" || body.From != "en" || body.To != "zh" || body.ModelType != "llm" {
			t.Errorf("unexpected incremental LLM request: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"from": "en", "to": "zh",
			"trans_result": []map[string]string{{"src": body.Query, "dst": "宠物 | 猫"}},
		})
	}))
	defer translator.Close()
	provider := llmReviewProvider{
		appID: "fixture", key: "fixture", endpoint: translator.URL,
		client: translator.Client(), limit: LLMReviewHardLimit,
	}
	build, err := runIncrementalBuild(context.Background(), dir, 200, LLMReviewHardLimit, provider)
	if err != nil || calls != 1 || build.Completed != 1 || build.RemainingRows != 0 ||
		build.UsageChars != 1000+len([]rune("Pets | Cats")) {
		t.Fatalf("incremental build failed: %+v calls=%d err=%v", build, calls, err)
	}
	assembled, err := AssembleIncremental(dir)
	if err != nil || assembled.ReviewRows != 2 || assembled.DeferredEnglish != 0 {
		t.Fatalf("incremental assemble failed: %+v %v", assembled, err)
	}
	if assembled.ActiveSHA256 != plan.ActiveSHA256 || assembled.SourceSHA256 != plan.SourceSHA256 {
		t.Fatal("assembled evidence does not bind to plan")
	}
	var approval incrementalApproval
	if err := readJSONFile(filepath.Join(dir, incrementalApprovalFile), &approval); err != nil {
		t.Fatal(err)
	}
	approval.Reviewer = "fixture-reviewer"
	for i := range approval.Rows {
		approval.Rows[i].Accepted = true
	}
	approvalRaw, err := incrementalJSON(approval)
	if err != nil {
		t.Fatal(err)
	}
	approvalPath := filepath.Join(dir, "approved-incremental.json")
	if err := os.WriteFile(approvalPath, approvalRaw, 0600); err != nil {
		t.Fatal(err)
	}
	input := IncrementalPublishInput{
		ApprovalPath: approvalPath, ApprovalSHA256: incrementalHash(approvalRaw),
		ActiveSHA256: plan.ActiveSHA256,
	}
	dryRun, err := PublishIncremental(dir, input, false)
	if err != nil || dryRun.Applied || dryRun.ReviewRows != 2 {
		t.Fatalf("incremental publish dry-run failed: %+v %v", dryRun, err)
	}
	if _, err := os.Stat(filepath.Join(dir, releaseBackupName(plan.ActiveSHA256))); !os.IsNotExist(err) {
		t.Fatal("incremental dry-run created a backup")
	}
	published, err := PublishIncremental(dir, input, true)
	if err != nil || !published.Applied {
		t.Fatalf("incremental publish failed: %+v %v", published, err)
	}
	active, err := loadSnapshot(filepath.Join(dir, activeFile))
	if err != nil {
		t.Fatal(err)
	}
	if active.Version.Created != version2 || active.Categories["CBT101"].Zh != "其他" ||
		active.Categories["CBT102"].Zh != "水壶" || active.Categories["CBT201"].Zh != "猫" ||
		active.Categories["CBT202"].ID != "" || len(active.Retired["CBT202"]) != 1 ||
		active.Retired["CBT202"][0].Zh != "旧分类" {
		t.Fatalf("published snapshot lost translations or official version: %+v", active)
	}
	if err := ensureIncrementalManaged(dir); err != nil {
		t.Fatalf("managed marker was removed: %v", err)
	}
	if _, err := update(context.Background(), dir, nil, "", nil, 0); err == nil ||
		!strings.Contains(err.Error(), "ordinary CBT dictionary updater disabled") {
		t.Fatalf("ordinary updater was not blocked after incremental release: %v", err)
	}
	if err := prepareIncrementalPlanRotation(dir, active, published.AfterSHA256); err != nil {
		t.Fatalf("next incremental cycle could not archive completed evidence: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, incrementalManifestFile)); !os.IsNotExist(err) {
		t.Fatal("completed incremental manifest was not archived for the next cycle")
	}
	if _, err := os.Stat(filepath.Join(dir, "incremental-archive")); err != nil {
		t.Fatal("completed incremental evidence archive was not created")
	}
	rolled, err := RollbackIncremental(dir, published.AfterSHA256, plan.ActiveSHA256, true)
	if err != nil || !rolled.Applied {
		t.Fatalf("incremental rollback failed: %+v %v", rolled, err)
	}
	restored, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil || !bytes.Equal(restored, activeRaw) {
		t.Fatal("incremental rollback did not restore exact original bytes")
	}
	if err := ensureIncrementalManaged(dir); err != nil {
		t.Fatalf("rollback removed managed marker: %v", err)
	}
}

func TestIncrementalBuildBudgetAndPendingFailClosed(t *testing.T) {
	dir, _, _, server, _ := prepareIncrementalFixture(t)
	runIncrementalPlanFixture(t, dir, server)
	var calls int
	translator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		calls++
		_, _ = w.Write([]byte("{}"))
	}))
	defer translator.Close()
	provider := llmReviewProvider{
		appID: "fixture", key: "fixture", endpoint: translator.URL,
		client: translator.Client(), limit: 1001,
	}
	if _, err := runIncrementalBuild(context.Background(), dir, 1, 1001, provider); err == nil || calls != 0 {
		t.Fatal("incomplete remaining budget was not rejected before the first request")
	}
	if err := os.WriteFile(filepath.Join(dir, incrementalPendingFile), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	provider.limit = LLMReviewHardLimit
	if _, err := runIncrementalBuild(context.Background(), dir, 1, LLMReviewHardLimit, provider); err == nil || calls != 0 {
		t.Fatal("unresolved incremental pending attempt did not block a retry")
	}
	if _, err := AssembleIncremental(dir); err == nil {
		t.Fatal("assemble accepted unresolved incremental pending attempt")
	}
}

func TestIncrementalPlanNoOpAndSameTimestampCollision(t *testing.T) {
	dir := t.TempDir()
	fixture := generatedVersion(t, version1, fixtureTree())
	active, err := readDump(bytes.NewReader(fixture.body), fixture.version)
	if err != nil {
		t.Fatal(err)
	}
	active.Categories["CBT101"] = categoryWithLabel(active.Categories["CBT101"], "其他")
	if err := saveJSON(dir, activeFile, active); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, llmReleaseManagedFile), []byte("{\"managed\":true}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileLLMUsage(dir, 100); err != nil {
		t.Fatal(err)
	}
	server, latest := fixtureServer(t)
	result, err := planIncremental(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all")
	if err != nil || !result.NoOp || result.Counts.Added+result.Counts.ChangedPath+
		result.Counts.MetadataChanged+result.Counts.Removed+result.Counts.TranslationNeeded != 0 {
		t.Fatalf("same official version should be a no-op: %+v %v", result, err)
	}
	tree := fixtureTree()
	leaf := tree["CBT101"]
	allowed := false
	leaf.Settings.ListingAllowed = &allowed
	tree["CBT101"] = leaf
	latest.Store(generatedVersion(t, version1, tree))
	if _, err := planIncremental(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all"); err == nil ||
		!strings.Contains(err.Error(), "same creation time but a different MD5") {
		t.Fatalf("same timestamp with different official dump must fail closed: %v", err)
	}
}

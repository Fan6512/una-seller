package cbtcategory

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestReleaseSwitchAndRollback(t *testing.T) {
	dir := t.TempDir()
	fixture := generatedVersion(t, version1, fixtureTree())
	s, err := readDump(bytes.NewReader(fixture.body), fixture.version)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, activeFile, s); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil {
		t.Fatal(err)
	}
	item := s.Categories["CBT101"]
	item.Zh = "测试分类"
	s.Categories[item.ID] = item
	if err := saveJSON(dir, llmReleaseCandidateFile, s); err != nil {
		t.Fatal(err)
	}
	next, err := os.ReadFile(filepath.Join(dir, llmReleaseCandidateFile))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := switchReleaseLocked(dir, releaseHash(next), next, ReleaseSwitchReport{}, true); err == nil {
		t.Fatal("stale active hash accepted")
	}
	plan, err := switchReleaseLocked(dir, releaseHash(before), next, ReleaseSwitchReport{}, false)
	if err != nil || plan.Applied {
		t.Fatalf("dry-run failed: %+v %v", plan, err)
	}
	if _, err := os.Stat(filepath.Join(dir, plan.Backup)); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote backup")
	}
	if _, err := os.Stat(filepath.Join(dir, llmReleaseManagedFile)); !os.IsNotExist(err) {
		t.Fatal("dry-run enabled managed guard")
	}
	result, err := switchReleaseLocked(dir, releaseHash(before), next, ReleaseSwitchReport{}, true)
	if err != nil || !result.Applied {
		t.Fatalf("switch failed: %+v %v", result, err)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, activeFile)); err != nil || !bytes.Equal(raw, next) {
		t.Fatal("published bytes differ")
	}
	// No HTTP client is necessary: the legacy updater must stop before fetch.
	if _, err := update(context.Background(), dir, nil, "", nil, 0); err == nil || err.Error() != "ordinary CBT dictionary updater disabled after managed LLM release; use the reviewed incremental workflow" {
		t.Fatalf("legacy updater not blocked before fetch: %v", err)
	}
	if _, err := RollbackLLMRelease(dir, releaseHash(before), releaseHash(before), true); err == nil {
		t.Fatal("stale rollback accepted")
	}
	result, err = RollbackLLMRelease(dir, releaseHash(next), releaseHash(before), true)
	if err != nil || !result.Applied {
		t.Fatalf("rollback failed: %+v %v", result, err)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, activeFile)); err != nil || !bytes.Equal(raw, before) {
		t.Fatal("rollback bytes differ")
	}
	if err := rejectOrdinaryUpdateAfterRelease(dir); err == nil {
		t.Fatal("rollback removed managed release guard")
	}
}

func TestPublishRequiresApprovalEvidence(t *testing.T) {
	dir := t.TempDir()
	if _, err := PublishLLMRelease(dir, ReleasePublishInput{}, true); err == nil {
		t.Fatal("publish accepted absent approval")
	}
	hash := releaseHash([]byte("not approved"))
	if _, err := PublishLLMRelease(dir, ReleasePublishInput{ActiveSHA256: hash, ApprovalSHA256: hash}, true); err == nil {
		t.Fatal("publish accepted missing approval file")
	}
	if _, err := os.Stat(filepath.Join(dir, activeFile)); !os.IsNotExist(err) {
		t.Fatal("publish wrote active without evidence")
	}
}

func TestPublishFrozenAuditWithManualCorrection(t *testing.T) {
	dir := t.TempDir()
	fixture := generatedVersion(t, version1, fixtureTree())
	s := snapshot{Version: fixture.version, Categories: make(map[string]category)}
	var rows []ReviewRow
	for i := 0; i < 304; i++ {
		id := fmt.Sprintf("CBT%d", 10000+i)
		name := fmt.Sprintf("Category %d", i)
		item := category{ID: id, Name: name, Path: []PathNode{{ID: id, Name: name}}, Depth: 1, ListingAllowed: true, Status: "enabled"}
		s.Categories[id] = item
		row := reviewRow(item, true)
		row.Candidate = "测试分类"
		rows = append(rows, row)
	}
	for _, file := range []string{stagingFile, activeFile} {
		if err := saveJSON(dir, file, s); err != nil {
			t.Fatal(err)
		}
	}
	if err := saveJSON(dir, reviewFile, ReviewSet{Version: s.Version, Rows: rows}); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileLLMUsage(dir, 100); err != nil {
		t.Fatal(err)
	}
	known := s.Categories["CBT10000"]
	manual := []override{{Path: known.Path, Zh: "已知纠错"}}
	if err := saveJSON(dir, overrideFile, manual); err != nil {
		t.Fatal(err)
	}
	exclusions := map[string]bool{"CBT10001": true}
	if err := saveJSON(dir, "excluded.json", []string{"CBT10001"}); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, "corrections.json", []releaseCorrection{{ID: known.ID, PathNames: []string{known.Name}, Label: "已知纠错"}}); err != nil {
		t.Fatal(err)
	}
	audit, err := AssembleLLMCandidate(dir, "fixed", 300, exclusions)
	if err != nil {
		t.Fatal(err)
	}
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	for _, pair := range [][2]string{{llmReleaseCandidateFile, "baseline.json"}, {llmReleaseAuditFile, "audit.json"}} {
		if err := os.WriteFile(filepath.Join(dir, pair[1]), read(pair[0]), 0600); err != nil {
			t.Fatal(err)
		}
	}
	bad := audit.AuditRows[0]
	manual = append(manual, override{Path: bad.Path, Zh: "新人工纠错"})
	if err := saveJSON(dir, overrideFile, manual); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleLLMCandidate(dir, "fixed", 300, exclusions); err != nil {
		t.Fatal(err)
	}
	approval := releaseApproval{
		CandidateSHA256: releaseHash(read(llmReleaseCandidateFile)), AuditSHA256: releaseHash(read("audit.json")),
		CorrectionsSHA256: releaseHash(read("corrections.json")), Reviewer: "fixture",
	}
	for _, row := range audit.AuditRows {
		ok := row.ID != bad.ID
		label := row.Label
		if !ok {
			label = "新人工纠错"
		}
		approval.Rows = append(approval.Rows, releaseJudgment{ID: row.ID, OriginalOK: &ok, FinalOK: true, FinalLabel: label})
	}
	if err := saveJSON(dir, "approval.json", approval); err != nil {
		t.Fatal(err)
	}
	input := ReleasePublishInput{
		Approval: filepath.Join(dir, "approval.json"), ApprovalSHA256: releaseHash(read("approval.json")),
		ActiveSHA256: releaseHash(read(activeFile)), Audit: filepath.Join(dir, "audit.json"),
		Baseline: filepath.Join(dir, "baseline.json"), Exclusions: filepath.Join(dir, "excluded.json"),
		Corrections: filepath.Join(dir, "corrections.json"),
	}
	plan, err := PublishLLMRelease(dir, input, false)
	if err != nil || plan.Applied || plan.Audited != 300 || plan.OriginalOK != 299 {
		t.Fatalf("valid frozen audit rejected: %+v %v", plan, err)
	}
	// A stale approval hash cannot authorize a changed manifest.
	badInput := input
	badInput.ApprovalSHA256 = releaseHash([]byte("different approval"))
	if _, err := PublishLLMRelease(dir, badInput, true); err == nil {
		t.Fatal("stale approval accepted")
	}
	result, err := PublishLLMRelease(dir, input, true)
	if err != nil || !result.Applied || releaseHash(read(activeFile)) != approval.CandidateSHA256 {
		t.Fatalf("approved publication failed: %+v %v", result, err)
	}
}

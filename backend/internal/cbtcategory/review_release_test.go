package cbtcategory

import (
	"bytes"
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
	result, err := switchReleaseLocked(dir, releaseHash(before), next, ReleaseSwitchReport{}, true)
	if err != nil || !result.Applied {
		t.Fatalf("switch failed: %+v %v", result, err)
	}
	if raw, err := os.ReadFile(filepath.Join(dir, activeFile)); err != nil || !bytes.Equal(raw, next) {
		t.Fatal("published bytes differ")
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

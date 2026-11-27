package cbtcategory

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAssembleLLMCandidateOffline(t *testing.T) {
	dir := t.TempDir()
	fixture := generatedVersion(t, version1, fixtureTree())
	s, err := readDump(bytes.NewReader(fixture.body), fixture.version)
	if err != nil {
		t.Fatal(err)
	}
	item := s.Categories["CBT101"]
	item.Zh = "未经复核普通译文"
	s.Categories[item.ID] = item
	for _, file := range []string{stagingFile, activeFile} {
		if err := saveJSON(dir, file, s); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ReconcileLLMUsage(dir, 123); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, overrideFile, []override{{Path: petsPath, Zh: "宠物配件"}}); err != nil {
		t.Fatal(err)
	}
	excluded := map[string]bool{"CBT100": true}
	if _, err := AssembleLLMCandidate(dir, "fixed", 2, excluded); err == nil {
		t.Fatal("incomplete candidate build accepted")
	}
	rows := make([]ReviewRow, 0, 2)
	for _, id := range []string{"CBT100", "CBT200"} {
		row := reviewRow(s.Categories[id], true)
		row.Candidate = "分类"
		rows = append(rows, row)
	}
	if err := saveJSON(dir, reviewFile, ReviewSet{Version: s.Version, Rows: rows}); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, llmBuildDeferredFile, buildDeferred{Version: s.Version, Rows: map[string]string{item.ID: sourceHash(reviewSource(item.Path, true))}}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil {
		t.Fatal(err)
	}
	report, err := AssembleLLMCandidate(dir, "fixed", 2, excluded)
	if err != nil || report.Manual != 1 || report.LLM != 2 || report.Deferred != 1 || report.PriorExcluded != 1 || len(report.AuditRows) != 2 {
		t.Fatalf("bad release candidate %+v %v", report, err)
	}
	assembled, err := loadSnapshot(filepath.Join(dir, llmReleaseCandidateFile))
	if err != nil || assembled.Categories[item.ID].Zh != "" || assembled.Categories["CBT201"].Zh != "宠物配件" {
		t.Fatalf("bad assembled labels: %+v %v", assembled, err)
	}
	if err := validateReleaseIdentity(s, assembled); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, llmReleaseCandidateFile))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != report.SnapshotSHA256 {
		t.Fatal("audit not bound to exact snapshot bytes")
	}
	again, err := AssembleLLMCandidate(dir, "fixed", 2, excluded)
	if err != nil || !reflect.DeepEqual(report, again) {
		t.Fatal("audit not deterministic")
	}
	after, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("assembly modified active dictionary")
	}
}

package cbtcategory

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLLMReviewOfflineCandidates(t *testing.T) {
	dir := t.TempDir()
	fixture := generatedVersion(t, version1, fixtureTree())
	s, err := readDump(bytes.NewReader(fixture.body), fixture.version)
	if err != nil {
		t.Fatal(err)
	}
	item := s.Categories["CBT101"]
	item.Zh = "其他"
	s.Categories[item.ID] = item
	if err := saveJSON(dir, stagingFile, s); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, activeFile, s); err != nil {
		t.Fatal(err)
	}
	if err := saveJSON(dir, overrideFile, []override{{Path: petsPath, Zh: "宠物配件"}}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil {
		t.Fatal(err)
	}
	plan, err := PlanReview(dir, "translated", "fixed", 200, false, true)
	if err != nil || plan.Selected != 1 || plan.Manual != 1 || plan.Rows[0].ID != item.ID || plan.Rows[0].Original != "其他" {
		t.Fatalf("bad review plan: %+v err=%v", plan, err)
	}
	again, err := PlanReview(dir, "translated", "fixed", 200, false, true)
	if err != nil || !reflect.DeepEqual(plan, again) {
		t.Fatal("sample not reproducible")
	}
	plan.Rows[0].Candidate = "家居其他用品"
	in := ReviewSet{Version: plan.Version, Rows: plan.Rows}
	if err := saveJSON(dir, "input.json", in); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if n, err := ImportReview(dir, filepath.Join(dir, "input.json")); err != nil || n != 1 {
			t.Fatalf("import failed: %d %v", n, err)
		}
	}
	next, err := PlanReview(dir, "all", "fixed", 0, true, false)
	if err != nil || next.Completed != 1 || next.Manual != 1 || next.Selected != 2 {
		t.Fatalf("resume plan failed: %+v %v", next, err)
	}
	bad := reviewRow(s.Categories["CBT201"], false)
	bad.Candidate = "不可覆盖"
	if err := saveJSON(dir, "input.json", ReviewSet{Version: s.Version, Rows: []ReviewRow{bad}}); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportReview(dir, filepath.Join(dir, "input.json")); err == nil {
		t.Fatal("manual override accepted")
	}
	after, err := os.ReadFile(filepath.Join(dir, activeFile))
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("candidate work changed active dictionary")
	}
}

func TestLLMCumulativeQuotaNeverResets(t *testing.T) {
	dir := t.TempDir()
	if err := ReserveLLMReviewChars(dir, 50000, 1); err == nil {
		t.Fatal("missing usage state allowed requests")
	}
	if _, err := ReconcileLLMUsage(dir, 100); err != nil {
		t.Fatal(err)
	}
	if err := ReserveLLMReviewChars(dir, 150, 40); err != nil {
		t.Fatal(err)
	}
	if state, err := ReconcileLLMUsage(dir, 20); err != nil || state.Chars != 140 {
		t.Fatalf("usage decreased: %+v %v", state, err)
	}
	if err := ReserveLLMReviewChars(dir, 150, 11); err == nil {
		t.Fatal("overage accepted")
	}
	if err := ReserveLLMReviewChars(dir, 900001, 1); err == nil {
		t.Fatal("hard cap bypassed")
	}
	if err := os.WriteFile(filepath.Join(dir, llmUsageFile), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReconcileLLMUsage(dir, 0); err == nil {
		t.Fatal("corrupt state silently reset")
	}
}

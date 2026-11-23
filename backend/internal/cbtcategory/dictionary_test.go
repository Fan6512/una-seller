package cbtcategory

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const version1 = "2026-10-14T21:12:35.000Z"
const version2 = "2026-11-20T20:12:18.000Z"

var homePath = []PathNode{{ID: "CBT100", Name: "Home"}, {ID: "CBT101", Name: "Other"}}
var petsPath = []PathNode{{ID: "CBT200", Name: "Pets"}, {ID: "CBT201", Name: "Other"}}

func fixtureNode(path []PathNode, children []PathNode, status string) dumpNode {
	node := dumpNode{ID: path[len(path)-1].ID, Name: path[len(path)-1].Name, Path: path, Children: slices.Clone(children)}
	if node.Children == nil {
		node.Children = []PathNode{}
	}
	allowed := len(children) == 0
	node.Settings.ListingAllowed = &allowed
	node.Settings.Status = status
	return node
}

func fixtureTree() map[string]dumpNode {
	return map[string]dumpNode{
		"CBT100": fixtureNode(homePath[:1], homePath[1:], "enabled"),
		"CBT101": fixtureNode(homePath, nil, "enabled"),
		"CBT200": fixtureNode(petsPath[:1], petsPath[1:], "enabled"),
		"CBT201": fixtureNode(petsPath, nil, "enabled"),
	}
}

func gzipFixture(t *testing.T, raw []byte) []byte {
	t.Helper()
	var dst bytes.Buffer
	writer := gzip.NewWriter(&dst)
	if _, err := writer.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return dst.Bytes()
}

type fixtureVersion struct {
	version version
	body    []byte
}

func generatedVersion(t *testing.T, created string, tree map[string]dumpNode) *fixtureVersion {
	t.Helper()
	raw, err := json.Marshal(tree)
	if err != nil {
		t.Fatal(err)
	}
	body := gzipFixture(t, raw)
	return &fixtureVersion{version: version{MD5: fmt.Sprintf("%x", md5.Sum(body)), Created: created}, body: body}
}

func fixtureServer(t *testing.T) (*httptest.Server, *atomic.Pointer[fixtureVersion]) {
	t.Helper()
	var latest atomic.Pointer[fixtureVersion]
	latest.Store(generatedVersion(t, version1, fixtureTree()))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/sites/CBT/categories/all" || r.URL.RawQuery != "" ||
			r.Header.Get("Accept-Encoding") != "gzip" || r.Header.Get("Authorization") != "" {
			t.Error("unexpected official dump request")
		}
		current := latest.Load()
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-MD5", current.version.MD5)
		w.Header().Set("X-Content-Created", current.version.Created)
		_, _ = w.Write(current.body)
	}))
	t.Cleanup(server.Close)
	return server, &latest
}

func TestDumpParserRejectsCorruptionAndBrokenIdentity(t *testing.T) {
	fixture := generatedVersion(t, version1, fixtureTree())
	out, err := readDump(bytes.NewReader(fixture.body), fixture.version)
	if err != nil {
		t.Fatal(err)
	}
	leaf, root := out.Categories["CBT101"], out.Categories["CBT100"]
	if leaf.Depth != 2 || leaf.ParentID != "CBT100" || leaf.HasChildren || !leaf.ListingAllowed ||
		!root.HasChildren || root.ParentID != "" || !slices.Equal(leaf.Path, homePath) {
		t.Fatalf("missing official graph: %#v %#v", root, leaf)
	}
	if _, err = readDump(bytes.NewReader(fixture.body[:len(fixture.body)-1]), fixture.version); err == nil {
		t.Fatal("truncated gzip accepted")
	}

	for _, modify := range []func(map[string]dumpNode){
		func(tree map[string]dumpNode) { tree["MLB101"] = tree["CBT101"] },
		func(tree map[string]dumpNode) { delete(tree, "CBT100") },
		func(tree map[string]dumpNode) {
			tree["CBT101"] = fixtureNode([]PathNode{{ID: "CBT100", Name: "Wrong"}, homePath[1]}, nil, "enabled")
		},
		func(tree map[string]dumpNode) {
			tree["CBT100"] = fixtureNode(homePath[:1], []PathNode{homePath[1], homePath[1]}, "enabled")
		},
		func(tree map[string]dumpNode) { tree["CBT101"] = fixtureNode(homePath, nil, "unknown") },
	} {
		tree := fixtureTree()
		modify(tree)
		broken := generatedVersion(t, version1, tree)
		if _, err = readDump(bytes.NewReader(broken.body), broken.version); err == nil {
			t.Fatal("broken category dump accepted")
		}
	}
	entry, err := json.Marshal(fixtureTree()["CBT101"])
	if err != nil {
		t.Fatal(err)
	}
	duplicate := fmt.Appendf(nil, `{"CBT101":%s,"CBT101":%s}`, entry, entry)
	if _, err = parseDump(duplicate, fixture.version); err == nil || !strings.Contains(err.Error(), "duplicate category ID") {
		t.Fatalf("duplicate category ID not rejected: %v", err)
	}
}

func TestDictionaryResumeIncrementalUpdatesAndManualPaths(t *testing.T) {
	server, latest := fixtureServer(t)
	dir := t.TempDir()
	reader := NewReader(dir)
	run := func(ctx context.Context, translator TranslateBatch, limit int) Summary {
		t.Helper()
		result, err := update(ctx, dir, server.Client(), server.URL+"/sites/CBT/categories/all", translator, limit)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	if result := run(context.Background(), nil, 0); result.Official != 4 || result.Unprocessed != 4 || result.Translated != 0 ||
		!result.Published || result.State != "provider_disabled" {
		t.Fatalf("invalid raw-only bootstrap: %+v", result)
	}
	if got := reader.Load().PublishedBreadcrumb("CBT101"); got == nil || !slices.Equal(got.Original, []string{"Home", "Other"}) || len(got.Chinese) != 0 {
		t.Fatalf("no provider should preserve official path: %+v", got)
	}
	manual := []override{{Path: homePath, Zh: "人工配件"}}
	if err := saveJSON(dir, overrideFile, manual); err != nil {
		t.Fatal(err)
	}
	manualBefore, err := os.ReadFile(filepath.Join(dir, overrideFile))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	result := run(ctx, func(_ context.Context, texts []string) map[string]string {
		if slices.Contains(texts, "Home | Other") {
			t.Fatal("manual override should not consume provider quota")
		}
		cancel()
		return map[string]string{"Home": "家居"}
	}, 0)
	if result.State != "paused" || result.Published || result.Translated != 2 || result.Failed != 2 || result.Unprocessed != 0 {
		t.Fatalf("unexpected interrupted update outcome: %+v", result)
	}
	if path := reader.Load().Localize(homePath); !slices.Equal(path, []string{"Home", "人工配件"}) {
		t.Fatalf("unpublished machine result leaked or manual label lost: %v", path)
	}

	progress, err := os.OpenFile(filepath.Join(dir, progressFile), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = progress.WriteString(`{"version":`); err != nil {
		t.Fatal(err)
	}
	progress.Close()
	result = run(context.Background(), func(_ context.Context, texts []string) map[string]string {
		if !reflect.DeepEqual(texts, []string{"Pets", "Pets | Other"}) {
			t.Fatalf("resume repeated completed rows or used ambiguous input: %q", texts)
		}
		return map[string]string{"Pets": "宠物", "Pets | Other": "宠物｜其他"}
	}, 0)
	if result.State != "published" || result.Translated != 4 || result.Failed != 0 || !result.Published {
		t.Fatalf("resume failed: %+v", result)
	}
	if got := reader.Load().Localize(homePath); !slices.Equal(got, []string{"家居", "人工配件"}) {
		t.Fatalf("home category not localized accurately: %v", got)
	}
	if got := reader.Load().Localize(petsPath); !slices.Equal(got, []string{"宠物", "其他"}) {
		t.Fatalf("same name crossed paths: %v", got)
	}
	if got := reader.Load().Localize([]PathNode{homePath[0], petsPath[1]}); got != nil {
		t.Fatalf("mis-matched canonical path localized: %v", got)
	}
	if result := run(context.Background(), func(context.Context, []string) map[string]string {
		t.Fatal("unchanged version called the provider")
		return nil
	}, 0); result.State != "unchanged" || result.Published {
		t.Fatalf("unexpected unchanged status: %+v", result)
	}

	tree := fixtureTree()
	delete(tree, "CBT201")
	tree["CBT200"] = fixtureNode(petsPath[:1], nil, "enabled")
	newPath := []PathNode{homePath[0], {ID: "CBT102", Name: "Jugs"}}
	tree["CBT100"] = fixtureNode(homePath[:1], []PathNode{homePath[1], newPath[1]}, "enabled")
	tree["CBT102"] = fixtureNode(newPath, nil, "enabled")
	latest.Store(generatedVersion(t, version2, tree))
	result = run(context.Background(), func(_ context.Context, texts []string) map[string]string {
		if !slices.Equal(texts, []string{"Jugs"}) {
			t.Fatalf("incremental update sent unchanged or retained names: %v", texts)
		}
		return nil
	}, 0)
	if result.Published || result.Removed != 1 || result.Failed != 1 || reader.Load().UnavailableForNewListing("CBT201") {
		t.Fatalf("paused update replaced previous active version: %+v", result)
	}

	result = run(context.Background(), nil, 0)
	if !result.Published || result.State != "provider_disabled" || result.Removed != 1 || result.Failed != 1 ||
		!reader.Load().UnavailableForNewListing("CBT201") || !reader.Load().UnavailableForNewListing("CBT100") {
		t.Fatalf("disabled provider failed to publish valid official changes: %+v", result)
	}
	if got := reader.Load().PublishedBreadcrumb("CBT201"); got == nil || !slices.Equal(got.Chinese, []string{"宠物", "其他"}) {
		t.Fatalf("removed path or old translation lost: %+v", got)
	}
	if got := reader.Load().Localize(newPath); !slices.Equal(got, []string{"家居", "Jugs"}) {
		t.Fatalf("new unprocessed path should preserve original leaf: %v", got)
	}
	result = run(context.Background(), func(_ context.Context, texts []string) map[string]string {
		if !slices.Equal(texts, []string{"Jugs"}) {
			t.Fatalf("retry failed to isolate remainder: %v", texts)
		}
		return map[string]string{"Jugs": "水壶"}
	}, 1)
	if result.Translated != 4 || !result.Published || result.Batches != 1 {
		t.Fatalf("latest version did not finish: %+v", result)
	}
	manualAfter, err := os.ReadFile(filepath.Join(dir, overrideFile))
	if err != nil || !bytes.Equal(manualBefore, manualAfter) || reader.Load().Localize(homePath)[1] != "人工配件" {
		t.Fatal("provider overwrote manual override")
	}
}

func TestDictionarySingleWriterLockAndReaderKeepsLastValidSnapshot(t *testing.T) {
	server, _ := fixtureServer(t)
	dir := t.TempDir()
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		t.Fatal(err)
	}
	run := func() (Summary, error) {
		return update(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all", nil, 0)
	}
	if _, err = run(); err == nil || !strings.Contains(err.Error(), "writer lock") {
		lock.Close()
		t.Fatalf("concurrent writer was not rejected: %v", err)
	}
	if err = lock.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = run(); err != nil {
		t.Fatal(err)
	}
	reader := NewReader(dir)
	if reader.Load().PublishedBreadcrumb("CBT101") == nil {
		t.Fatal("reader failed to load active snapshot")
	}
	if err = saveJSON(dir, activeFile, map[string]any{"bad": "malformed"}); err != nil {
		t.Fatal(err)
	}
	if reader.Load().PublishedBreadcrumb("CBT101") == nil {
		t.Fatal("reader did not hold last valid active snapshot")
	}
	if _, err = run(); err == nil {
		t.Fatal("writer accepted corrupt active dictionary as an empty one")
	}
}

func TestSkippedCategoryDoesNotBlockAndCanBeRetried(t *testing.T) {
	server, latest := fixtureServer(t)
	dir := t.TempDir()
	tree := fixtureTree()
	long := strings.Repeat("L", 501)
	path := []PathNode{{ID: "CBT300", Name: long}}
	tree["CBT300"] = fixtureNode(path, nil, "enabled")
	latest.Store(generatedVersion(t, version1, tree))
	translate := func(_ context.Context, texts []string) map[string]string {
		out := make(map[string]string)
		for _, text := range texts {
			parts := strings.Split(text, " | ")
			translated := make([]string, len(parts))
			for i := range parts {
				translated[i] = "中文"
			}
			out[text] = strings.Join(translated, "｜")
		}
		return out
	}
	result, err := update(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all", translate, 0)
	if err != nil || result.State != "published" || !result.Published || result.Skipped != 1 || result.Failed != 0 || result.Unprocessed != 0 {
		t.Fatalf("skipped leaf must not stop publication: %+v %v", result, err)
	}
	result, err = update(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all", translate, 0)
	if err != nil || result.Published || result.Skipped != 1 || result.Failed != 0 || result.Unprocessed != 0 {
		t.Fatalf("skipped must remain visible and retryable: %+v %v", result, err)
	}
}

func TestSpecialSeparatorFallsBackToLeaf(t *testing.T) {
	path := []PathNode{{ID: "CBT900", Name: "Out | doors"}, {ID: "CBT901", Name: "Other"}}
	categories := map[string]category{
		"CBT900": {ID: "CBT900", Name: "Out | doors", Path: path[:1], Status: "enabled"},
		"CBT901": {ID: "CBT901", Name: "Other", Path: path, Status: "enabled"},
		"CBT777": {ID: "CBT777", Name: "Other", Path: []PathNode{{ID: "CBT777", Name: "Other"}}, Status: "enabled"},
	}
	inputs := makeBatches(snapshot{Categories: categories}, nil, nil)
	found := false
	for _, batch := range inputs {
		for _, item := range batch {
			if slices.Contains(item.ids, "CBT901") {
				found = true
				if item.text != "Other" || item.segments != 1 {
					t.Fatalf("bad fallback: %+v", item)
				}
			}
		}
	}
	if !found {
		t.Fatal("missing same-name leaf")
	}
}

func TestDictionaryFallbackForFullWidthAndLongAncestors(t *testing.T) {
	for _, tt := range []struct {
		name     string
		path     []PathNode
		want     string
		segments int
	}{
		{
			name: "full width separator",
			path: []PathNode{{ID: "CBT700", Name: "Home｜House"}, {ID: "CBT701", Name: "Other"}},
			want: "Other", segments: 1,
		},
		{
			name: "long ancestor",
			path: []PathNode{{ID: "CBT700", Name: strings.Repeat("A", 501)}, {ID: "CBT701", Name: "Kitchen"}, {ID: "CBT702", Name: "Other"}},
			want: "Kitchen | Other", segments: 2,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			leaf := tt.path[len(tt.path)-1]
			categories := map[string]category{
				leaf.ID:  {ID: leaf.ID, Name: leaf.Name, Path: tt.path, Status: "enabled"},
				"CBT999": {ID: "CBT999", Name: "Other", Path: []PathNode{{ID: "CBT999", Name: "Other"}}, Status: "enabled"},
			}
			found := false
			for _, batch := range makeBatches(snapshot{Categories: categories}, nil, nil) {
				for _, item := range batch {
					if slices.Contains(item.ids, leaf.ID) {
						found = true
						if item.text != tt.want || item.segments != tt.segments ||
							strings.Count(item.text, " | ")+1 != item.segments {
							t.Fatalf("unexpected fallback: %+v", item)
						}
					}
				}
			}
			if !found {
				t.Fatal("category was omitted")
			}
		})
	}
}

func TestReaderBackoffAndChangedFileRecovery(t *testing.T) {
	server, _ := fixtureServer(t)
	dir := t.TempDir()
	_, err := update(context.Background(), dir, server.Client(), server.URL+"/sites/CBT/categories/all", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, activeFile)
	valid, err := loadSnapshot(path)
	if err != nil {
		t.Fatal(err)
	}
	reader := NewReader(dir)
	if reader.Status().State != "loaded_untranslated" {
		t.Fatalf("unexpected initial state: %+v", reader.Status())
	}
	if err = saveJSON(dir, activeFile, map[string]any{"bad": true}); err != nil {
		t.Fatal(err)
	}
	if reader.Load().PublishedBreadcrumb("CBT101") == nil || reader.activeError != "invalid_file" ||
		reader.activeBackoff != 5*time.Second {
		t.Fatal("corrupted file discarded last valid snapshot or missed initial backoff")
	}
	retryAt := reader.activeRetryAt
	reader.Load()
	if !reader.activeRetryAt.Equal(retryAt) {
		t.Fatal("unchanged corrupt file retried before backoff")
	}
	if err = saveJSON(dir, activeFile, valid); err != nil {
		t.Fatal(err)
	}
	if reader.Load().PublishedBreadcrumb("CBT101") == nil || reader.activeError != "" ||
		!reader.activeRetryAt.IsZero() || reader.activeBackoff != 0 {
		t.Fatal("repaired file was not immediately reloaded")
	}
}

package cbtcategory

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const stagingFile = "staging.json"
const progressFile = "progress.jsonl"

type TranslateBatch func(context.Context, []string) map[string]string

type Summary struct {
	Version     version  `json:"version"`
	Official    int      `json:"official"`
	Disabled    int      `json:"disabled"`
	Removed     int      `json:"removed"`
	Translated  int      `json:"translated"`
	Unchanged   int      `json:"unchanged"`
	Failed      int      `json:"failed"`
	Unprocessed int      `json:"unprocessed"`
	Skipped     int      `json:"skipped"`
	SkipSamples []string `json:"skip_samples,omitempty"`
	Batches     int      `json:"batches"`
	Published   bool     `json:"published"`
	State       string   `json:"state"`
}

type progressRecord struct {
	Version    version           `json:"version"`
	Translated map[string]string `json:"translated,omitempty"`
	Failed     []string          `json:"failed,omitempty"`
	Skipped    map[string]string `json:"skipped,omitempty"`
}

func Update(ctx context.Context, dir string, translate TranslateBatch, maxBatches int, maxTexts ...int) (Summary, error) {
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return update(ctx, dir, client, DumpURL, translate, maxBatches, maxTexts...)
}

func update(ctx context.Context, dir string, client *http.Client, endpoint string, translate TranslateBatch, maxBatches int, maxTexts ...int) (Summary, error) {
	if strings.TrimSpace(dir) == "" || maxBatches < 0 {
		return Summary{}, errors.New("configure a CBT dictionary directory and non-negative batch limit")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return Summary{}, err
	}
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return Summary{}, fmt.Errorf("could not acquire CBT dictionary writer lock: %w", err)
	}
	defer lock.Close()
	if err := rejectOrdinaryUpdateAfterRelease(dir); err != nil {
		return Summary{}, err
	}
	active, err := loadSnapshot(filepath.Join(dir, activeFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Summary{}, err
	}
	overrides, err := loadOverrides(filepath.Join(dir, overrideFile))
	if err != nil {
		return Summary{}, err
	}
	resp, latest, err := fetchDump(ctx, client, endpoint)
	if err != nil {
		return Summary{}, err
	}
	defer resp.Body.Close()
	latestTime, _ := time.Parse(time.RFC3339Nano, latest.Created)
	activeTime, _ := time.Parse(time.RFC3339Nano, active.Version.Created)
	if latestTime.Before(activeTime) {
		return Summary{}, errors.New("official CBT dump version is older than the active dictionary")
	}
	activeSummary := summarize(active, overrides, nil)
	if active.Version == latest && activeSummary.Unprocessed == 0 && activeSummary.Skipped == 0 {
		// A new deterministic correction can update an already complete
		// snapshot without calling the provider or waiting for a new dump.
		if applyCategoryGlossary(&active, overrides, nil) {
			if err = saveJSON(dir, activeFile, active); err != nil {
				return activeSummary, err
			}
			activeSummary = summarize(active, overrides, nil)
			activeSummary.State, activeSummary.Published = "published", true
			return activeSummary, nil
		}
		activeSummary.State = "unchanged"
		return activeSummary, nil
	}
	staged, err := loadSnapshot(filepath.Join(dir, stagingFile))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Summary{}, err
	}
	stagedTime, _ := time.Parse(time.RFC3339Nano, staged.Version.Created)
	if latestTime.Before(stagedTime) {
		return Summary{}, errors.New("official CBT dump version is older than staged progress")
	}
	journal, failed, err := openProgress(filepath.Join(dir, progressFile), &staged)
	if err != nil {
		return Summary{}, err
	}
	defer journal.Close()

	if staged.Version != latest {
		var next snapshot
		if active.Version == latest {
			next = active
		} else {
			next, err = readDump(resp.Body, latest)
			if err != nil {
				return Summary{}, err
			}
			seed(&next, active, staged)
		}
		if err = saveJSON(dir, stagingFile, next); err != nil {
			return Summary{}, err
		}
		// An interrupted rollover leaves old version-tagged lines safely ignorable.
		if err = journal.Truncate(0); err != nil {
			return Summary{}, err
		}
		if _, err = journal.Seek(0, io.SeekStart); err != nil {
			return Summary{}, err
		}
		staged, failed = next, make(map[string]bool)
	}
	resp.Body.Close()
	applyCategoryGlossary(&staged, overrides, failed)
	summary := summarize(staged, overrides, failed)
	if summary.Failed+summary.Unprocessed > 0 && translate == nil {
		if active.Version != latest {
			if err = saveJSON(dir, activeFile, staged); err != nil {
				return summary, err
			}
			summary.Published = true
		}
		summary.State = "provider_disabled"
		return summary, nil
	}

	batches := makeBatches(staged, overrides, failed, maxTexts...)
	for _, batch := range batches {
		if ctx.Err() != nil || maxBatches > 0 && summary.Batches >= maxBatches {
			break
		}
		texts := make([]string, 0, len(batch))
		for _, input := range batch {
			if input.text != "" {
				texts = append(texts, input.text)
			}
		}
		var labels map[string]string
		if len(texts) > 0 {
			summary.Batches++
			labels = translate(ctx, texts)
		}
		record := progressRecord{Version: latest, Translated: make(map[string]string), Skipped: make(map[string]string)}
		for _, input := range batch {
			label := translatedLabel(labels[input.text], input.segments)
			for _, id := range input.ids {
				switch {
				case input.text == "":
					record.Skipped[id] = "category name exceeds safe translation input"
				case label == "":
					record.Failed = append(record.Failed, id)
				default:
					record.Translated[id] = label
				}
			}
		}
		if err = json.NewEncoder(journal).Encode(record); err != nil {
			return summarize(staged, overrides, failed), err
		}
		if err = journal.Sync(); err != nil {
			return summarize(staged, overrides, failed), err
		}
		applyProgress(&staged, failed, record)
		if len(texts) > 0 && len(labels) == 0 {
			// A provider-wide failure may be a quota or cooldown. Keep completed batches.
			break
		}
	}
	final := summarize(staged, overrides, failed)
	final.Batches = summary.Batches
	if final.Failed+final.Unprocessed > 0 {
		final.State = "paused"
		return final, nil
	}
	if reflect.DeepEqual(active, staged) {
		final.Published, final.State = false, "unchanged"
		return final, nil
	}
	if err = saveJSON(dir, activeFile, staged); err != nil {
		return final, err
	}
	final.Published, final.State = true, "published"
	return final, nil
}

func seed(next *snapshot, active, stage snapshot) {
	next.Retired = active.Retired
	if next.Retired == nil {
		next.Retired = make(map[string][]category)
	}
	for id, old := range active.Categories {
		if _, found := next.Categories[id]; !found &&
			!slices.ContainsFunc(next.Retired[id], func(retired category) bool { return slices.Equal(retired.Path, old.Path) }) {
			next.Retired[id] = append(next.Retired[id], old)
		}
	}
	for id, item := range next.Categories {
		for _, source := range []snapshot{active, stage} {
			if old, found := source.Categories[id]; found && old.Zh != "" && slices.Equal(old.Path, item.Path) {
				item.Zh = old.Zh
			}
		}
		next.Categories[id] = item
	}
}

func summarize(current snapshot, overrides map[string]string, failed map[string]bool) Summary {
	out := Summary{Version: current.Version, Official: len(current.Categories)}
	for id, item := range current.Categories {
		if item.Status != "enabled" {
			out.Disabled++
			continue
		}
		label := overrides[pathKey(item.Path)]
		if label == "" {
			label = item.Zh
		}
		switch {
		case label == item.Name:
			out.Unchanged++
		case label != "":
			out.Translated++
		case failed[id]:
			out.Failed++
		case current.Skipped[id] != "":
			out.Skipped++
			if len(out.SkipSamples) < 20 {
				out.SkipSamples = append(out.SkipSamples, id+": "+current.Skipped[id])
			}
		default:
			out.Unprocessed++
		}
	}
	slices.Sort(out.SkipSamples)
	for id := range current.Retired {
		if _, found := current.Categories[id]; !found {
			out.Removed++
		}
	}
	return out
}

func openProgress(path string, staged *snapshot) (*os.File, map[string]bool, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, maxJSON+1))
	if err != nil || len(data) > maxJSON {
		file.Close()
		return nil, nil, errors.New("invalid CBT dictionary progress journal")
	}
	// Sync follows complete newline-delimited records. A crashed writer can leave a torn tail.
	end := bytes.LastIndexByte(data, '\n') + 1
	failed, err := replayProgress(data[:end], staged)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	if end != len(data) {
		if err = file.Truncate(int64(end)); err != nil {
			file.Close()
			return nil, nil, err
		}
	}
	if _, err = file.Seek(0, io.SeekEnd); err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, failed, nil
}

func replayProgress(data []byte, staged *snapshot) (map[string]bool, error) {
	failed := make(map[string]bool)
	for line := range bytes.SplitSeq(data, []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var record progressRecord
		if json.Unmarshal(line, &record) != nil || !record.Version.valid() {
			return nil, errors.New("invalid CBT dictionary progress record")
		}
		applyProgress(staged, failed, record)
	}
	return failed, nil
}

func applyProgress(staged *snapshot, failed map[string]bool, record progressRecord) {
	if staged.Version != record.Version {
		return
	}
	if staged.Skipped == nil {
		staged.Skipped = make(map[string]string)
	}
	for id, label := range record.Translated {
		item, found := staged.Categories[id]
		if found && validName(label) {
			item.Zh = label
			staged.Categories[id] = item
			delete(failed, id)
			delete(staged.Skipped, id)
		}
	}
	for id, reason := range record.Skipped {
		if item, found := staged.Categories[id]; found && item.Zh == "" {
			staged.Skipped[id] = reason
			delete(failed, id)
		}
	}
	for _, id := range record.Failed {
		if item, found := staged.Categories[id]; found && item.Zh == "" {
			failed[id] = true
			delete(staged.Skipped, id)
		}
	}
}

type translationInput struct {
	text     string
	segments int
	ids      []string
}

func makeBatches(current snapshot, overrides map[string]string, failed map[string]bool, maxTexts ...int) [][]translationInput {
	limit := 120
	if len(maxTexts) > 0 && maxTexts[0] > 0 && maxTexts[0] < limit {
		limit = maxTexts[0]
	}
	names := make(map[string]int)
	var ids []string
	for id, item := range current.Categories {
		names[item.Name]++
		if item.Status == "enabled" && item.Zh == "" && overrides[pathKey(item.Path)] == "" {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, func(a, b string) int {
		if failed[a] != failed[b] {
			if failed[a] {
				return 1
			}
			return -1
		}
		return strings.Compare(a, b)
	})

	inputs := make([]translationInput, 0, len(ids))
	inputIndex := make(map[string]int)
	for _, id := range ids {
		item := current.Categories[id]
		text, segments := item.Name, 1
		if len(item.Path) > 1 {
			parts := make([]string, len(item.Path))
			for i, node := range item.Path {
				parts[i] = node.Name
			}
			// Use the actual parent for unique names and the full path for duplicates.
			// Fall back to the leaf without reusing a different path's context.
			first := len(parts) - 2
			if names[item.Name] > 1 {
				first = 0
			}
			for start := first; start < len(parts)-1; start++ {
				candidate := strings.Join(parts[start:], " | ")
				if strings.Count(candidate, "|") == len(parts)-start-1 &&
					!strings.ContainsRune(candidate, '｜') && utf8.RuneCountInString(candidate) <= 500 {
					text, segments = candidate, len(parts)-start
					break
				}
			}
		}
		if utf8.RuneCountInString(text) > 500 {
			text, segments = "", 1
		}
		if index, found := inputIndex[text]; found {
			inputs[index].ids = append(inputs[index].ids, id)
		} else {
			inputIndex[text] = len(inputs)
			inputs = append(inputs, translationInput{text: text, segments: segments, ids: []string{id}})
		}
	}

	var batches [][]translationInput
	start, chars := 0, 0
	for i, input := range inputs {
		count := utf8.RuneCountInString(input.text) + 1
		if i > start && (chars+count > 1001 || i-start >= limit) {
			batches = append(batches, inputs[start:i])
			start, chars = i, 0
		}
		chars += count
	}
	if len(inputs) > start {
		batches = append(batches, inputs[start:])
	}
	return batches
}

func translatedLabel(text string, segments int) string {
	if segments > 1 {
		parts := strings.Split(strings.ReplaceAll(text, "｜", "|"), "|")
		if len(parts) != segments {
			return ""
		}
		for _, part := range parts {
			if !validName(part) {
				return ""
			}
		}
		text = parts[len(parts)-1]
	}
	if text = strings.TrimSpace(text); validName(text) {
		return text
	}
	return ""
}

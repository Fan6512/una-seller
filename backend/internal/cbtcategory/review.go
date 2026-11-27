package cbtcategory

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

const reviewFile = "llm-review-candidates.json"

type ReviewRow struct {
	ID        string     `json:"id"`
	Path      []PathNode `json:"path"`
	Source    string     `json:"source"`
	Original  string     `json:"original"`
	Candidate string     `json:"candidate,omitempty"`
}

type ReviewSet struct {
	Version version     `json:"version"`
	Rows    []ReviewRow `json:"rows"`
}

type ReviewPlan struct {
	Version          version     `json:"version"`
	Snapshot         string      `json:"snapshot_source"`
	Scope            string      `json:"scope"`
	FullPath         bool        `json:"full_path"`
	Eligible         int         `json:"eligible"`
	Manual           int         `json:"manual_excluded"`
	Completed        int         `json:"candidate_completed"`
	Selected         int         `json:"selected"`
	SourceChars      int         `json:"source_chars"`
	SkippedUnsafe    int         `json:"skipped_unsafe"`
	GlossaryExcluded int         `json:"glossary_excluded"`
	PriorExcluded    int         `json:"prior_excluded"`
	Rows             []ReviewRow `json:"rows,omitempty"`
}

// reviewSnapshot never downloads, publishes, or modifies the ordinary journal.
func reviewSnapshot(dir string) (snapshot, string, map[string]string, error) {
	source := stagingFile
	s, err := loadSnapshot(filepath.Join(dir, source))
	if errors.Is(err, os.ErrNotExist) {
		source = activeFile
		s, err = loadSnapshot(filepath.Join(dir, source))
	}
	if err != nil {
		return snapshot{}, "", nil, err
	}
	manual, err := loadOverrides(filepath.Join(dir, overrideFile))
	if err != nil {
		return snapshot{}, "", nil, err
	}
	f, err := os.Open(filepath.Join(dir, progressFile))
	if err == nil {
		defer f.Close()
		data, err := io.ReadAll(io.LimitReader(f, maxJSON+1))
		if err != nil || len(data) > maxJSON {
			return snapshot{}, "", nil, errors.New("invalid CBT dictionary progress journal")
		}
		if _, err = replayProgress(data[:bytes.LastIndexByte(data, '\n')+1], &s); err != nil {
			return snapshot{}, "", nil, err
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return snapshot{}, "", nil, err
	}
	applyCategoryGlossary(&s, manual, make(map[string]bool))
	return s, source, manual, nil
}

func reviewSource(path []PathNode, full bool) string {
	if !full && len(path) > 2 {
		path = path[len(path)-2:]
	}
	parts := make([]string, len(path))
	for i, node := range path {
		parts[i] = node.Name
	}
	return strings.Join(parts, " | ")
}

func safeReviewSource(source string) bool {
	return strings.TrimSpace(source) != "" && utf8.ValidString(source) &&
		utf8.RuneCountInString(source) <= llmReviewSourceLimit && !strings.ContainsAny(source, "\r\n")
}

func reviewGlossaryProtected(item category) bool {
	term, ok := categoryTerms[item.Name]
	return ok && item.Zh == term.label
}

func reviewRow(item category, full bool) ReviewRow {
	return ReviewRow{ID: item.ID, Path: item.Path, Source: reviewSource(item.Path, full), Original: item.Zh}
}

func loadReview(dir string, s snapshot) (map[string]ReviewRow, error) {
	var set ReviewSet
	err := readJSONFile(filepath.Join(dir, reviewFile), &set)
	out := make(map[string]ReviewRow)
	if errors.Is(err, os.ErrNotExist) {
		return out, nil
	}
	if err != nil {
		return nil, err
	}
	if set.Version != s.Version {
		return nil, errors.New("review candidate version differs from local snapshot; archive candidates before changing versions")
	}
	for _, row := range set.Rows {
		item, ok := s.Categories[row.ID]
		if !ok || !validReviewRow(row, item) || out[row.ID].ID != "" {
			return nil, fmt.Errorf("invalid, duplicate or stale review candidate row ID %q; archive and reimport valid rows", row.ID)
		}
		out[row.ID] = row
	}
	return out, nil
}

func validReviewRow(row ReviewRow, item category) bool {
	return row.ID == item.ID && item.Status == "enabled" &&
		slices.Equal(row.Path, item.Path) && row.Original == item.Zh && safeReviewSource(row.Source) &&
		(row.Source == reviewSource(item.Path, false) || row.Source == reviewSource(item.Path, true)) &&
		validName(row.Candidate) && !strings.ContainsAny(row.Candidate, "\r\n|｜") &&
		validReviewLanguage(row.Candidate, item)
}

// Permit only verified proper-name exceptions, not arbitrary untranslated English.
func validReviewLanguage(label string, item category) bool {
	if strings.ContainsFunc(label, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
		return true
	}
	if item.ID == "CBT106855" && item.Name == "PS Vita - PlayStation Vita" {
		return label == item.Name || label == "PS Vita" || label == "PlayStation Vita"
	}
	return false
}

// ReadReviewExclusions accepts an explicit union of previously inspected IDs.
func ReadReviewExclusions(path string) (map[string]bool, error) {
	var ids []string
	if err := readJSONFile(path, &ids); err != nil {
		return nil, err
	}
	if len(ids) == 0 || len(ids) > maxCategories {
		return nil, errors.New("review exclusion list must contain category IDs")
	}
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		if !idPattern.MatchString(id) {
			return nil, errors.New("invalid CBT review exclusion ID")
		}
		out[id] = true
	}
	return out, nil
}

// PlanReview uses a stable hash within each root and round-robin across roots
// for a reproducible stratified sample. Sample zero selects all eligible rows.
// SourceChars counts one source per row, without batching delimiters or retries.
func PlanReview(dir, scope, seed string, sample int, full, export bool, exclusions ...map[string]bool) (ReviewPlan, error) {
	if strings.TrimSpace(dir) == "" || sample < 0 || (scope != "all" && scope != "translated" && scope != "missing") {
		return ReviewPlan{}, errors.New("invalid CBT review plan arguments")
	}
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return ReviewPlan{}, fmt.Errorf("could not acquire CBT dictionary writer lock: %w", err)
	}
	defer lock.Close()
	s, source, manual, err := reviewSnapshot(dir)
	if err != nil {
		return ReviewPlan{}, err
	}
	completed, err := loadReview(dir, s)
	if err != nil {
		return ReviewPlan{}, err
	}
	out := ReviewPlan{Version: s.Version, Snapshot: source, Scope: scope, FullPath: full}
	groups := make(map[string][]ReviewRow)
	for _, item := range s.Categories {
		if item.Status != "enabled" {
			continue
		}
		if manual[pathKey(item.Path)] != "" {
			out.Manual++
			continue
		}
		if reviewGlossaryProtected(item) {
			out.GlossaryExcluded++
			continue
		}
		if completed[item.ID].ID != "" {
			out.Completed++
			continue
		}
		excluded := false
		for _, ids := range exclusions {
			if ids[item.ID] {
				excluded = true
				break
			}
		}
		if excluded {
			out.PriorExcluded++
			continue
		}
		translated := item.Zh != "" && item.Zh != item.Name
		if scope == "translated" && !translated || scope == "missing" && translated {
			continue
		}
		row := reviewRow(item, full)
		if !safeReviewSource(row.Source) {
			out.SkippedUnsafe++
			continue
		}
		root := item.Path[0].ID
		groups[root] = append(groups[root], row)
		out.Eligible++
	}
	roots := make([]string, 0, len(groups))
	for root, rows := range groups {
		roots = append(roots, root)
		slices.SortFunc(rows, func(a, b ReviewRow) int {
			ha := sha256.Sum256([]byte(seed + "\x00" + a.ID))
			hb := sha256.Sum256([]byte(seed + "\x00" + b.ID))
			return bytes.Compare(ha[:], hb[:])
		})
	}
	slices.Sort(roots)
	for offset := 0; out.Selected < out.Eligible && (sample == 0 || out.Selected < sample); offset++ {
		for _, root := range roots {
			if offset >= len(groups[root]) || sample > 0 && out.Selected >= sample {
				continue
			}
			row := groups[root][offset]
			out.Selected++
			out.SourceChars += utf8.RuneCountInString(row.Source)
			if export {
				out.Rows = append(out.Rows, row)
			}
		}
	}
	return out, nil
}

// ImportReview saves validated candidate leaf labels only. It never writes
// active.json, staging.json, overrides.json, or progress.jsonl.
func ImportReview(dir, input string) (int, error) {
	if strings.TrimSpace(dir) == "" || strings.TrimSpace(input) == "" {
		return 0, errors.New("configure review directory and candidate input")
	}
	var incoming ReviewSet
	if err := readJSONFile(input, &incoming); err != nil {
		return 0, err
	}
	return storeReview(dir, incoming)
}

func storeReview(dir string, incoming ReviewSet) (int, error) {
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return 0, fmt.Errorf("could not acquire CBT dictionary writer lock: %w", err)
	}
	defer lock.Close()
	s, _, manual, err := reviewSnapshot(dir)
	if err != nil {
		return 0, err
	}
	existing, err := loadReview(dir, s)
	if err != nil {
		return 0, err
	}
	if incoming.Version != s.Version || len(incoming.Rows) == 0 {
		return 0, errors.New("review input version mismatch or empty candidates")
	}
	for _, row := range incoming.Rows {
		item, ok := s.Categories[row.ID]
		if !ok || !validReviewRow(row, item) || manual[pathKey(item.Path)] != "" || reviewGlossaryProtected(item) {
			return 0, fmt.Errorf("invalid review candidate or protected manual/glossary override, row ID %q", row.ID)
		}
		if old, ok := existing[row.ID]; ok && old.Candidate != row.Candidate {
			return 0, fmt.Errorf("conflicting candidate row ID %q; archive previous review before replacing", row.ID)
		}
		existing[row.ID] = row
	}
	set := ReviewSet{Version: s.Version, Rows: make([]ReviewRow, 0, len(existing))}
	for _, row := range existing {
		set.Rows = append(set.Rows, row)
	}
	slices.SortFunc(set.Rows, func(a, b ReviewRow) int { return strings.Compare(a.ID, b.ID) })
	if err = saveJSON(dir, reviewFile, set); err != nil {
		return 0, err
	}
	return len(set.Rows), nil
}

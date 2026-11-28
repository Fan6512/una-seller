package cbtcategory

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

type incrementalCandidateRow struct {
	ID           string `json:"id"`
	SourceSHA256 string `json:"source_sha256"`
	Label        string `json:"label"`
}

type incrementalCandidateSet struct {
	ActiveSHA256   string                            `json:"active_sha256"`
	SourceSHA256   string                            `json:"source_sha256"`
	ManifestSHA256 string                            `json:"manifest_sha256"`
	Rows           map[string]incrementalCandidateRow `json:"rows"`
}

type incrementalDeferredSet struct {
	ActiveSHA256   string            `json:"active_sha256"`
	SourceSHA256   string            `json:"source_sha256"`
	ManifestSHA256 string            `json:"manifest_sha256"`
	Rows           map[string]string `json:"source_sha256_by_id"`
}

type incrementalPending struct {
	ActiveSHA256   string    `json:"active_sha256"`
	SourceSHA256   string    `json:"source_sha256"`
	ManifestSHA256 string    `json:"manifest_sha256"`
	ID             string    `json:"id"`
	SourceSHA256Row string   `json:"row_source_sha256"`
	SourceChars    int       `json:"source_chars"`
	UsageBefore    int       `json:"usage_before_chars"`
	StartedAt      time.Time `json:"started_at"`
}

type incrementalReceipt struct {
	ActiveSHA256   string    `json:"active_sha256"`
	SourceSHA256   string    `json:"source_sha256"`
	ManifestSHA256 string    `json:"manifest_sha256"`
	ID             string    `json:"id"`
	SourceSHA256Row string   `json:"row_source_sha256"`
	SourceChars    int       `json:"source_chars"`
	ReservedChars  int       `json:"reserved_chars"`
	UsageBefore    int       `json:"usage_before_chars"`
	Outcome        string    `json:"outcome"`
	CompletedAt    time.Time `json:"completed_at"`
}

type IncrementalBuildReport struct {
	Selected                 int               `json:"selected"`
	Completed                int               `json:"completed"`
	Deferred                 int               `json:"deferred_english"`
	RemainingRows            int               `json:"remaining_rows"`
	RemainingSourceChars     int               `json:"remaining_source_chars"`
	PlannedRemainingChars    int               `json:"planned_remaining_chars"`
	ReservedChars            int               `json:"reserved_chars"`
	UsageBefore              int               `json:"usage_before_chars"`
	UsageChars               int               `json:"usage_chars"`
	BudgetChars              int               `json:"budget_chars"`
	BudgetFits               bool              `json:"budget_fits"`
	Attempts                 []incrementalReceipt `json:"attempts,omitempty"`
}

type IncrementalAssembleReport struct {
	ActiveSHA256       string            `json:"active_sha256"`
	SourceSHA256       string            `json:"source_sha256"`
	ManifestSHA256     string            `json:"manifest_sha256"`
	CandidateSHA256    string            `json:"candidate_sha256"`
	ReviewSHA256       string            `json:"review_sha256"`
	ApprovalTemplate   string            `json:"approval_template"`
	Counts             incrementalCounts `json:"counts"`
	ReviewRows         int               `json:"review_rows"`
	DeferredEnglish    int               `json:"deferred_english"`
}

func verifyIncrementalReceipts(dir string) error {
	path := filepath.Join(dir, incrementalReceiptsFile)
	raw, err := readIncrementalFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		return errors.New("incremental receipt journal has an incomplete tail; archive and inspect before continuing")
	}
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		var receipt incrementalReceipt
		if json.Unmarshal(scanner.Bytes(), &receipt) != nil ||
			!idPattern.MatchString(receipt.ID) || !validReleaseHash(receipt.ActiveSHA256) ||
			!validReleaseHash(receipt.SourceSHA256) || !validReleaseHash(receipt.ManifestSHA256) ||
			!validReleaseHash(receipt.SourceSHA256Row) || receipt.SourceChars <= 0 ||
			receipt.ReservedChars < 0 || receipt.UsageBefore < 0 ||
			(receipt.Outcome != "stored" && receipt.Outcome != "deferred_english") ||
			receipt.CompletedAt.IsZero() {
			return errors.New("incremental receipt journal is invalid; archive and inspect before continuing")
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.New("incremental receipt journal could not be read")
	}
	return nil
}

func appendIncrementalReceipt(dir string, receipt incrementalReceipt) error {
	if err := verifyIncrementalReceipts(dir); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, incrementalReceiptsFile), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() > 0 {
		var last [1]byte
		if _, err := f.ReadAt(last[:], info.Size()-1); err != nil || last[0] != '\n' {
			return errors.New("incremental receipt journal has an incomplete tail")
		}
	}
	if err := json.NewEncoder(f).Encode(receipt); err != nil {
		return err
	}
	return f.Sync()
}

func syncIncrementalDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}

func loadIncrementalPlanLocked(dir string) (snapshot, []byte, snapshot, incrementalManifest, string, error) {
	var empty snapshot
	activeRaw, err := readIncrementalFile(filepath.Join(dir, activeFile))
	if err != nil {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("active CBT dictionary is unavailable")
	}
	active, err := loadSnapshot(filepath.Join(dir, activeFile))
	if err != nil {
		return empty, nil, empty, incrementalManifest{}, "", err
	}
	sourceRaw, err := readIncrementalFile(filepath.Join(dir, incrementalSourceFile))
	if err != nil {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("incremental source snapshot is unavailable; run incremental plan again")
	}
	source, err := loadSnapshot(filepath.Join(dir, incrementalSourceFile))
	if err != nil {
		return empty, nil, empty, incrementalManifest{}, "", err
	}
	manifestRaw, err := readIncrementalFile(filepath.Join(dir, incrementalManifestFile))
	if err != nil {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("incremental manifest is unavailable; run incremental plan again")
	}
	var manifest incrementalManifest
	if json.Unmarshal(manifestRaw, &manifest) != nil {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("incremental manifest is invalid")
	}
	activeHash, sourceHash, manifestHash := incrementalHash(activeRaw), incrementalHash(sourceRaw), incrementalHash(manifestRaw)
	if manifest.ActiveSHA256 != activeHash || manifest.SourceSHA256 != sourceHash ||
		manifest.ActiveVersion != active.Version || manifest.SourceVersion != source.Version {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("incremental plan is stale; active, source or manifest changed")
	}
	if _, err := compareIncrementalVersions(active.Version, source.Version); err != nil {
		return empty, nil, empty, incrementalManifest{}, "", err
	}
	overrideRaw, overrideErr := readIncrementalFile(filepath.Join(dir, overrideFile))
	if overrideErr != nil && !errors.Is(overrideErr, os.ErrNotExist) {
		return empty, nil, empty, incrementalManifest{}, "", overrideErr
	}
	if manifest.OverridesSHA256 != incrementalHash(overrideRaw) {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("manual override file changed after incremental plan; run incremental plan again")
	}
	overrides, err := loadOverrides(filepath.Join(dir, overrideFile))
	if err != nil {
		return empty, nil, empty, incrementalManifest{}, "", err
	}
	if err := validateIncrementalManifest(active, source, manifest, overrides); err != nil {
		return empty, nil, empty, incrementalManifest{}, "", err
	}
	usage, err := loadLLMUsage(dir)
	if err != nil || usage.Chars < manifest.UsageChars || usage.Chars > LLMReviewHardLimit {
		return empty, nil, empty, incrementalManifest{}, "", errors.New("cumulative LLM usage is missing, decreased or exceeds the hard limit")
	}
	return active, activeRaw, source, manifest, manifestHash, nil
}

func validateIncrementalManifest(active, source snapshot, manifest incrementalManifest, overrides map[string]string) error {
	if manifest.SourceVersion != source.Version || manifest.ActiveVersion != active.Version ||
		manifest.SourceChars < 0 || manifest.UsageChars < 0 || manifest.OverridesSHA256 == "" {
		return errors.New("invalid incremental manifest metadata")
	}
	expected := buildIncrementalManifest(active, source, manifest.ActiveSHA256, overrides)
	expected.SourceSHA256 = manifest.SourceSHA256
	expected.OverridesSHA256 = manifest.OverridesSHA256
	expected.UsageChars = manifest.UsageChars
	expected.BudgetRemaining = max(0, LLMReviewHardLimit-manifest.UsageChars)
	expected.BudgetFits = expected.SourceChars <= expected.BudgetRemaining
	if !reflect.DeepEqual(expected, manifest) {
		return errors.New("incremental manifest does not match the frozen official source, active snapshot and overrides")
	}
	return nil
}

func countIncrementalTranslations(rows []incrementalRow) int {
	count := 0
	for _, row := range rows {
		if row.NeedsTranslation {
			count++
		}
	}
	return count
}

func loadIncrementalBuildState(dir string, activeHash, sourceHash, manifestHash string, manifest incrementalManifest) (incrementalCandidateSet, incrementalDeferredSet, error) {
	candidates := incrementalCandidateSet{
		ActiveSHA256: activeHash, SourceSHA256: sourceHash, ManifestSHA256: manifestHash,
		Rows: make(map[string]incrementalCandidateRow),
	}
	deferred := incrementalDeferredSet{
		ActiveSHA256: activeHash, SourceSHA256: sourceHash, ManifestSHA256: manifestHash,
		Rows: make(map[string]string),
	}
	candidatePath := filepath.Join(dir, incrementalCandidatesFile)
	var savedCandidates incrementalCandidateSet
	if err := readJSONFile(candidatePath, &savedCandidates); err == nil {
		if savedCandidates.ActiveSHA256 != activeHash || savedCandidates.SourceSHA256 != sourceHash ||
			savedCandidates.ManifestSHA256 != manifestHash || savedCandidates.Rows == nil {
			return candidates, deferred, errors.New("incremental candidates do not match the current plan")
		}
		candidates = savedCandidates
	} else if !errors.Is(err, os.ErrNotExist) {
		return candidates, deferred, err
	}
	deferredPath := filepath.Join(dir, incrementalDeferredFile)
	var savedDeferred incrementalDeferredSet
	if err := readJSONFile(deferredPath, &savedDeferred); err == nil {
		if savedDeferred.ActiveSHA256 != activeHash || savedDeferred.SourceSHA256 != sourceHash ||
			savedDeferred.ManifestSHA256 != manifestHash || savedDeferred.Rows == nil {
			return candidates, deferred, errors.New("incremental deferred state does not match the current plan")
		}
		deferred = savedDeferred
	} else if !errors.Is(err, os.ErrNotExist) {
		return candidates, deferred, err
	}
	allowed := make(map[string]incrementalRow, len(manifest.Rows))
	for _, row := range manifest.Rows {
		if row.NeedsTranslation {
			allowed[row.ID] = row
		}
	}
	for id, row := range candidates.Rows {
		sourceRow, ok := allowed[id]
		item := category{ID: id, Name: sourceRow.Original}
		if !ok || row.ID != id || row.SourceSHA256 != sourceRow.SourceSHA256 ||
			!validReviewLanguage(row.Label, item) {
			return candidates, deferred, fmt.Errorf("invalid or stale incremental candidate %s", id)
		}
	}
	for id, hash := range deferred.Rows {
		sourceRow, ok := allowed[id]
		if !ok || hash != sourceRow.SourceSHA256 || candidates.Rows[id].ID != "" {
			return candidates, deferred, fmt.Errorf("invalid or stale incremental deferred row %s", id)
		}
	}
	return candidates, deferred, nil
}

func remainingIncrementalRows(manifest incrementalManifest, candidates incrementalCandidateSet, deferred incrementalDeferredSet) ([]incrementalRow, int) {
	rows := make([]incrementalRow, 0)
	chars := 0
	for _, row := range manifest.Rows {
		if !row.NeedsTranslation || candidates.Rows[row.ID].ID != "" || deferred.Rows[row.ID] != "" {
			continue
		}
		rows = append(rows, row)
		chars += utf8.RuneCountInString(row.Source)
	}
	return rows, chars
}

func RunIncrementalBuild(ctx context.Context, dir string, maxRows, budget int, appID, key string) (IncrementalBuildReport, error) {
	appID, key = strings.TrimSpace(appID), strings.TrimSpace(key)
	if appID == "" || key == "" || strings.ContainsAny(key, "\r\n") {
		return IncrementalBuildReport{}, errors.New("explicit Baidu LLM credentials required")
	}
	provider := llmReviewProvider{
		appID: appID, key: key, endpoint: baiduLLMEndpoint, limit: budget,
		client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}
	return runIncrementalBuild(ctx, dir, maxRows, budget, provider)
}

func runIncrementalBuild(ctx context.Context, dir string, maxRows, budget int, provider llmReviewProvider) (out IncrementalBuildReport, runErr error) {
	if strings.TrimSpace(dir) == "" || maxRows <= 0 || maxRows > LLMBuildMaxRows ||
		budget <= 0 || budget > LLMReviewHardLimit || provider.limit != budget {
		return out, errors.New("incremental build requires 1-200 rows and a cumulative budget <= 900000")
	}
	runLock, err := LockFile(filepath.Join(dir, "llm-review-run.lock"))
	if err != nil {
		return out, err
	}
	defer runLock.Close()
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return out, err
	}
	defer lock.Close()
	if err := ensureIncrementalManaged(dir); err != nil {
		return out, err
	}
	if err := ensureNoIncrementalPending(dir); err != nil {
		return out, err
	}
	if err := verifyIncrementalReceipts(dir); err != nil {
		return out, err
	}
	active, _, source, manifest, manifestHash, err := loadIncrementalPlanLocked(dir)
	if err != nil {
		return out, err
	}
	_ = active
	candidates, deferred, err := loadIncrementalBuildState(dir, manifest.ActiveSHA256, manifest.SourceSHA256, manifestHash, manifest)
	if err != nil {
		return out, err
	}
	remaining, remainingChars := remainingIncrementalRows(manifest, candidates, deferred)
	usage, err := loadLLMUsage(dir)
	if err != nil {
		return out, err
	}
	out.UsageBefore, out.UsageChars, out.BudgetChars = usage.Chars, usage.Chars, budget
	out.RemainingRows, out.RemainingSourceChars, out.PlannedRemainingChars = len(remaining), remainingChars, remainingChars
	out.BudgetFits = usage.Chars <= budget && remainingChars <= budget-usage.Chars
	if !out.BudgetFits {
		return out, errors.New("complete remaining incremental translation exceeds the cumulative budget; no request was sent")
	}
	if manifest.SourceVersion != source.Version {
		return out, errors.New("incremental source version changed")
	}
	if len(remaining) == 0 {
		return out, nil
	}
	rows := remaining[:min(maxRows, len(remaining))]
	out.Selected = len(rows)
	for i, row := range rows {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		if i > 0 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return out, ctx.Err()
			case <-timer.C:
			}
		}
		usage, err := loadLLMUsage(dir)
		if err != nil {
			return out, err
		}
		pending := incrementalPending{
			ActiveSHA256: manifest.ActiveSHA256, SourceSHA256: manifest.SourceSHA256,
			ManifestSHA256: manifestHash, ID: row.ID, SourceSHA256Row: row.SourceSHA256,
			SourceChars: utf8.RuneCountInString(row.Source), UsageBefore: usage.Chars,
			StartedAt: time.Now().UTC(),
		}
		if _, err := writeIncrementalJSON(dir, incrementalPendingFile, pending); err != nil {
			return out, err
		}
		label, reserved, err := provider.translate(ctx, dir, row.Source)
		if err != nil {
			return out, fmt.Errorf("incremental LLM row %s failed; pending marker retained and automatic retry blocked: %w", row.ID, err)
		}
		receipt := incrementalReceipt{
			ActiveSHA256: manifest.ActiveSHA256, SourceSHA256: manifest.SourceSHA256,
			ManifestSHA256: manifestHash, ID: row.ID, SourceSHA256Row: row.SourceSHA256,
			SourceChars: pending.SourceChars, ReservedChars: reserved, UsageBefore: usage.Chars,
			CompletedAt: time.Now().UTC(),
		}
		item := source.Categories[row.ID]
		switch {
		case validReviewLanguage(label, item):
			candidates.Rows[row.ID] = incrementalCandidateRow{ID: row.ID, SourceSHA256: row.SourceSHA256, Label: label}
			if _, err := writeIncrementalJSON(dir, incrementalCandidatesFile, candidates); err != nil {
				return out, err
			}
			receipt.Outcome = "stored"
			out.Completed++
		case validName(label) && !strings.ContainsAny(label, "\r\n|｜"):
			deferred.Rows[row.ID] = row.SourceSHA256
			if _, err := writeIncrementalJSON(dir, incrementalDeferredFile, deferred); err != nil {
				return out, err
			}
			receipt.Outcome = "deferred_english"
			out.Deferred++
		default:
			return out, fmt.Errorf("incremental LLM row %s returned an invalid label; pending marker retained", row.ID)
		}
		if err := appendIncrementalReceipt(dir, receipt); err != nil {
			return out, err
		}
		if err := os.Remove(filepath.Join(dir, incrementalPendingFile)); err != nil {
			return out, err
		}
		if err := syncIncrementalDir(dir); err != nil {
			return out, err
		}
		out.ReservedChars += reserved
		out.Attempts = append(out.Attempts, receipt)
	}
	usage, err = loadLLMUsage(dir)
	if err != nil {
		return out, err
	}
	candidates, deferred, err = loadIncrementalBuildState(dir, manifest.ActiveSHA256, manifest.SourceSHA256, manifestHash, manifest)
	if err != nil {
		return out, err
	}
	remaining, remainingChars = remainingIncrementalRows(manifest, candidates, deferred)
	out.UsageChars, out.RemainingRows, out.RemainingSourceChars = usage.Chars, len(remaining), remainingChars
	return out, nil
}

type incrementalReviewRow struct {
	ID              string     `json:"id"`
	Kind            string     `json:"kind"`
	Path            []PathNode `json:"path"`
	Source          string     `json:"source"`
	SourceSHA256    string     `json:"source_sha256"`
	Original        string     `json:"original"`
	Candidate       string     `json:"candidate"`
	Origin          string     `json:"origin"`
	FallbackEnglish bool       `json:"fallback_english"`
}

type incrementalReview struct {
	ActiveSHA256   string                 `json:"active_sha256"`
	SourceSHA256   string                 `json:"source_sha256"`
	ManifestSHA256 string                 `json:"manifest_sha256"`
	CandidateSHA256 string                `json:"candidate_sha256"`
	Rows           []incrementalReviewRow `json:"rows"`
}

type incrementalApprovalRow struct {
	ID                   string `json:"id"`
	SourceSHA256         string `json:"source_sha256"`
	Accepted             bool   `json:"accepted"`
	FinalLabel           string `json:"final_label"`
	AcceptEnglishFallback bool  `json:"accept_english_fallback"`
}

type incrementalApproval struct {
	ActiveSHA256    string                  `json:"active_sha256"`
	SourceSHA256    string                  `json:"source_sha256"`
	ManifestSHA256  string                  `json:"manifest_sha256"`
	CandidateSHA256 string                  `json:"candidate_sha256"`
	ReviewSHA256    string                  `json:"review_sha256"`
	Reviewer        string                  `json:"reviewer"`
	Rows            []incrementalApprovalRow `json:"rows"`
}

func expectedIncrementalRetired(active snapshot, manifest incrementalManifest) map[string][]category {
	out := make(map[string][]category, len(active.Retired)+len(manifest.RemovedIDs))
	for id, rows := range active.Retired {
		out[id] = slices.Clone(rows)
	}
	for _, id := range manifest.RemovedIDs {
		old, ok := active.Categories[id]
		if !ok {
			continue
		}
		duplicate := false
		for _, existing := range out[id] {
			if existing.ID == old.ID && slices.Equal(existing.Path, old.Path) {
				duplicate = true
				break
			}
		}
		if !duplicate {
			out[id] = append(out[id], old)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func validateIncrementalCandidateIdentity(active, source, candidate snapshot, manifest incrementalManifest) error {
	if candidate.Version != source.Version || len(candidate.Categories) != len(source.Categories) ||
		!reflect.DeepEqual(candidate.Retired, expectedIncrementalRetired(active, manifest)) {
		return errors.New("incremental candidate version, category count or retired history mismatch")
	}
	rows := make(map[string]incrementalRow, len(manifest.Rows))
	for _, row := range manifest.Rows {
		rows[row.ID] = row
	}
	for id, official := range source.Categories {
		item, ok := candidate.Categories[id]
		if !ok || item.ID != official.ID || item.Name != official.Name || item.ParentID != official.ParentID ||
			item.Depth != official.Depth || item.HasChildren != official.HasChildren ||
			item.ListingAllowed != official.ListingAllowed || item.Status != official.Status ||
			!slices.Equal(item.Path, official.Path) || (item.Zh != "" && !validName(item.Zh)) {
			return fmt.Errorf("incremental candidate changed official identity for %s", id)
		}
		old, existed := active.Categories[id]
		row, hasRow := rows[id]
		if existed && slices.Equal(old.Path, official.Path) {
			if !hasRow || !row.NeedsTranslation {
				if item.Zh != old.Zh {
					return fmt.Errorf("incremental candidate overwrote unchanged translation for %s", id)
				}
			}
		} else if official.Status == "enabled" && (hasRow || (existed && !slices.Equal(old.Path, official.Path))) {
			if !hasRow && item.Zh != "" {
				return fmt.Errorf("incremental candidate has untracked translation for %s", id)
			}
		}
	}
	for _, id := range manifest.RemovedIDs {
		if _, ok := candidate.Categories[id]; ok {
			return fmt.Errorf("removed category remains active: %s", id)
		}
		history := candidate.Retired[id]
		if len(history) == 0 || !slices.Equal(history[len(history)-1].Path, active.Categories[id].Path) {
			return fmt.Errorf("removed category history missing: %s", id)
		}
	}
	return nil
}

func AssembleIncremental(dir string) (IncrementalAssembleReport, error) {
	var out IncrementalAssembleReport
	if strings.TrimSpace(dir) == "" {
		return out, errors.New("configure a CBT dictionary directory")
	}
	runLock, err := LockFile(filepath.Join(dir, "llm-review-run.lock"))
	if err != nil {
		return out, err
	}
	defer runLock.Close()
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return out, err
	}
	defer lock.Close()
	if err := ensureIncrementalManaged(dir); err != nil {
		return out, err
	}
	if err := ensureNoIncrementalPending(dir); err != nil {
		return out, err
	}
	if err := verifyIncrementalReceipts(dir); err != nil {
		return out, err
	}
	active, _, source, manifest, manifestHash, err := loadIncrementalPlanLocked(dir)
	if err != nil {
		return out, err
	}
	candidates, deferred, err := loadIncrementalBuildState(dir, manifest.ActiveSHA256, manifest.SourceSHA256, manifestHash, manifest)
	if err != nil {
		return out, err
	}
	remaining, _ := remainingIncrementalRows(manifest, candidates, deferred)
	if len(remaining) != 0 {
		return out, errors.New("incremental translation is incomplete; finish all remaining rows before assemble")
	}
	target := source
	target.Retired = expectedIncrementalRetired(active, manifest)
	rows := make(map[string]incrementalRow, len(manifest.Rows))
	for _, row := range manifest.Rows {
		rows[row.ID] = row
	}
	deferredCount := 0
	for id, official := range source.Categories {
		item := official
		old, existed := active.Categories[id]
		if existed && slices.Equal(old.Path, official.Path) {
			item.Zh = old.Zh
		}
		if row, found := rows[id]; found {
			switch row.Origin {
			case "manual", "glossary":
				item.Zh = row.Label
			case "existing":
				item.Zh = row.ExistingLabel
			case "pending":
				if candidate, ok := candidates.Rows[id]; ok {
					item.Zh = candidate.Label
				} else if deferred.Rows[id] == row.SourceSHA256 {
					item.Zh = ""
					deferredCount++
				} else {
					return out, fmt.Errorf("incremental row %s has no candidate or deferred receipt", id)
				}
			default:
				return out, fmt.Errorf("unknown incremental origin for %s", id)
			}
		}
		target.Categories[id] = item
	}
	if err := validateIncrementalCandidateIdentity(active, source, target, manifest); err != nil {
		return out, err
	}
	candidateHash, err := writeIncrementalJSON(dir, incrementalCandidateFile, target)
	if err != nil {
		return out, err
	}
	review := incrementalReview{
		ActiveSHA256: manifest.ActiveSHA256, SourceSHA256: manifest.SourceSHA256,
		ManifestSHA256: manifestHash, CandidateSHA256: candidateHash,
		Rows: make([]incrementalReviewRow, 0, len(manifest.Rows)),
	}
	for _, row := range manifest.Rows {
		item := target.Categories[row.ID]
		label := item.Zh
		fallback := label == "" || label == item.Name
		if fallback {
			label = item.Name
		}
		origin := row.Origin
		if row.NeedsTranslation {
			switch {
			case candidates.Rows[row.ID].ID != "":
				origin = "llm"
			case deferred.Rows[row.ID] != "":
				origin = "deferred_english"
			}
		}
		review.Rows = append(review.Rows, incrementalReviewRow{
			ID: row.ID, Kind: row.Kind, Path: slices.Clone(row.Path), Source: row.Source,
			SourceSHA256: row.SourceSHA256, Original: row.Original, Candidate: label,
			Origin: origin, FallbackEnglish: fallback,
		})
	}
	reviewHash, err := writeIncrementalJSON(dir, incrementalReviewFile, review)
	if err != nil {
		return out, err
	}
	approval := incrementalApproval{
		ActiveSHA256: manifest.ActiveSHA256, SourceSHA256: manifest.SourceSHA256,
		ManifestSHA256: manifestHash, CandidateSHA256: candidateHash, ReviewSHA256: reviewHash,
		Rows: make([]incrementalApprovalRow, 0, len(review.Rows)),
	}
	for _, row := range review.Rows {
		approval.Rows = append(approval.Rows, incrementalApprovalRow{
			ID: row.ID, SourceSHA256: row.SourceSHA256, FinalLabel: row.Candidate,
			AcceptEnglishFallback: row.FallbackEnglish,
		})
	}
	if _, err := writeIncrementalJSON(dir, incrementalApprovalFile, approval); err != nil {
		return out, err
	}
	out = IncrementalAssembleReport{
		ActiveSHA256: manifest.ActiveSHA256, SourceSHA256: manifest.SourceSHA256,
		ManifestSHA256: manifestHash, CandidateSHA256: candidateHash,
		ReviewSHA256: reviewHash, ApprovalTemplate: incrementalApprovalFile,
		Counts: manifest.Counts, ReviewRows: len(review.Rows), DeferredEnglish: deferredCount,
	}
	return out, nil
}

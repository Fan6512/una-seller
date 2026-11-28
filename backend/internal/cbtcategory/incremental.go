package cbtcategory

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	incrementalSourceFile     = "llm-incremental-source.json"
	incrementalManifestFile   = "llm-incremental-manifest.json"
	incrementalCandidatesFile = "llm-incremental-candidates.json"
	incrementalDeferredFile   = "llm-incremental-deferred.json"
	incrementalPendingFile    = "llm-incremental-pending.json"
	incrementalReceiptsFile   = "llm-incremental-receipts.jsonl"
	incrementalCandidateFile  = "llm-incremental-candidate.json"
	incrementalReviewFile     = "llm-incremental-review.json"
	incrementalApprovalFile   = "llm-incremental-approval.json"
)

type incrementalCounts struct {
	Added             int `json:"added"`
	ChangedPath       int `json:"changed_path"`
	MetadataChanged   int `json:"metadata_changed"`
	Removed           int `json:"removed"`
	Unchanged         int `json:"unchanged"`
	NewlyEnabled      int `json:"newly_enabled"`
	ReviewRequired    int `json:"review_required"`
	TranslationNeeded int `json:"translation_required"`
}

type incrementalRow struct {
	ID               string     `json:"id"`
	Kind             string     `json:"kind"`
	Path             []PathNode `json:"path"`
	Source           string     `json:"source"`
	SourceSHA256     string     `json:"source_sha256"`
	Original         string     `json:"original"`
	Origin           string     `json:"origin"`
	Label            string     `json:"label,omitempty"`
	ExistingLabel    string     `json:"existing_label,omitempty"`
	NeedsTranslation bool       `json:"needs_translation"`
}

type incrementalManifest struct {
	ActiveSHA256       string            `json:"active_sha256"`
	ActiveVersion      version           `json:"active_version"`
	SourceSHA256       string            `json:"source_sha256"`
	SourceVersion      version           `json:"source_version"`
	Counts             incrementalCounts `json:"counts"`
	AddedIDs           []string          `json:"added_ids,omitempty"`
	ChangedPathIDs     []string          `json:"changed_path_ids,omitempty"`
	MetadataChangedIDs []string          `json:"metadata_changed_ids,omitempty"`
	RemovedIDs         []string          `json:"removed_ids,omitempty"`
	UnchangedIDs       []string          `json:"unchanged_ids,omitempty"`
	NewlyEnabledIDs    []string          `json:"newly_enabled_ids,omitempty"`
	Rows               []incrementalRow  `json:"rows,omitempty"`
	SourceChars        int               `json:"source_chars"`
	UsageChars         int               `json:"usage_chars"`
	BudgetRemaining    int               `json:"budget_remaining_chars"`
	BudgetFits         bool              `json:"budget_fits"`
	NoOp               bool              `json:"no_op"`
}

type IncrementalPlanReport struct {
	ActiveSHA256       string            `json:"active_sha256"`
	SourceSHA256       string            `json:"source_sha256"`
	ManifestSHA256     string            `json:"manifest_sha256"`
	ActiveVersion      version           `json:"active_version"`
	SourceVersion      version           `json:"source_version"`
	Counts             incrementalCounts `json:"counts"`
	AddedIDs           []string          `json:"added_ids,omitempty"`
	ChangedPathIDs     []string          `json:"changed_path_ids,omitempty"`
	MetadataChangedIDs []string          `json:"metadata_changed_ids,omitempty"`
	RemovedIDs         []string          `json:"removed_ids,omitempty"`
	UnchangedIDs       []string          `json:"unchanged_ids,omitempty"`
	NewlyEnabledIDs    []string          `json:"newly_enabled_ids,omitempty"`
	SourceChars        int               `json:"source_chars"`
	UsageChars         int               `json:"usage_chars"`
	BudgetRemaining    int               `json:"budget_remaining_chars"`
	BudgetFits         bool              `json:"budget_fits"`
	NoOp               bool              `json:"no_op"`
}

func incrementalHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum[:])
}

func incrementalJSON(value any) ([]byte, string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, "", err
	}
	raw = append(raw, '\n')
	return raw, incrementalHash(raw), nil
}

func writeIncrementalJSON(dir, name string, value any) (string, error) {
	raw, hash, err := incrementalJSON(value)
	if err != nil {
		return "", err
	}
	if err := writeReleaseBytes(dir, name, raw); err != nil {
		return "", err
	}
	return hash, nil
}

func readIncrementalFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxJSON {
		return nil, errors.New("invalid incremental evidence file")
	}
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) > maxJSON || !utf8.Valid(raw) {
		return nil, errors.New("invalid incremental evidence bytes")
	}
	return raw, nil
}

func ensureIncrementalManaged(dir string) error {
	path := filepath.Join(dir, llmReleaseManagedFile)
	info, err := os.Lstat(path)
	if err != nil {
		return errors.New("managed CBT dictionary release marker is required")
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 1024 {
		return errors.New("managed CBT dictionary release marker is invalid")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return errors.New("managed CBT dictionary release marker is unreadable")
	}
	var marker struct {
		Managed *bool `json:"managed"`
	}
	if json.Unmarshal(raw, &marker) != nil || marker.Managed == nil || !*marker.Managed {
		return errors.New("managed CBT dictionary release marker must contain managed=true")
	}
	return nil
}

func ensureNoIncrementalPending(dir string) error {
	for _, name := range []string{llmBuildPendingFile, incrementalPendingFile} {
		_, err := os.Lstat(filepath.Join(dir, name))
		if err == nil {
			return fmt.Errorf("unresolved CBT dictionary LLM pending attempt: %s", name)
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("could not inspect CBT dictionary pending state: %s", name)
		}
	}
	return nil
}

func compareIncrementalVersions(active, source version) (int, error) {
	activeTime, err := time.Parse(time.RFC3339Nano, active.Created)
	if err != nil {
		return 0, errors.New("active CBT dictionary has an invalid version timestamp")
	}
	sourceTime, err := time.Parse(time.RFC3339Nano, source.Created)
	if err != nil {
		return 0, errors.New("official CBT dump has an invalid version timestamp")
	}
	switch {
	case sourceTime.Before(activeTime):
		return -1, errors.New("official CBT dump version is older than the active dictionary")
	case sourceTime.Equal(activeTime) && source.MD5 != active.MD5:
		return 0, errors.New("official CBT dump has the same creation time but a different MD5; manual investigation required")
	case source == active:
		return 0, nil
	case sourceTime.Equal(activeTime):
		return 0, errors.New("official CBT dump version is ambiguous")
	default:
		return 1, nil
	}
}

func categoryMetadataEqual(a, b category) bool {
	return a.ParentID == b.ParentID && a.Depth == b.Depth &&
		a.HasChildren == b.HasChildren && a.ListingAllowed == b.ListingAllowed &&
		a.Status == b.Status
}

func incrementalGlossaryLabel(item category) string {
	term, ok := categoryTerms[item.Name]
	if !ok {
		return ""
	}
	if item.Zh != "" && item.Zh != item.Name && item.Zh != term.bad && item.Zh != term.label {
		return ""
	}
	return term.label
}

func buildIncrementalManifest(active, source snapshot, activeHash string, overrides map[string]string) incrementalManifest {
	out := incrementalManifest{
		ActiveSHA256: activeHash, ActiveVersion: active.Version,
		SourceVersion: source.Version,
	}
	ids := make([]string, 0, len(source.Categories))
	for id := range source.Categories {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		item := source.Categories[id]
		old, found := active.Categories[id]
		kind := "added"
		pathChanged := false
		newlyEnabled := false
		if found {
			pathChanged = !slices.Equal(old.Path, item.Path)
			newlyEnabled = old.Status != "enabled" && item.Status == "enabled"
			switch {
			case pathChanged:
				kind = "changed_path"
				out.ChangedPathIDs = append(out.ChangedPathIDs, id)
			case !categoryMetadataEqual(old, item):
				kind = "metadata_changed"
				out.MetadataChangedIDs = append(out.MetadataChangedIDs, id)
			default:
				kind = "unchanged"
				out.UnchangedIDs = append(out.UnchangedIDs, id)
			}
		} else {
			out.AddedIDs = append(out.AddedIDs, id)
		}
		if newlyEnabled {
			out.NewlyEnabledIDs = append(out.NewlyEnabledIDs, id)
		}

		needsTranslation := item.Status == "enabled" && (!found || pathChanged ||
			(newlyEnabled && (old.Zh == "" || old.Zh == old.Name)))
		if !needsTranslation && found && !pathChanged {
			item.Zh = old.Zh
		}
		row := incrementalRow{
			ID: id, Kind: kind, Path: slices.Clone(item.Path),
			Source: reviewSource(item.Path, true), Original: item.Name,
			ExistingLabel: old.Zh,
		}
		row.SourceSHA256 = sourceHash(row.Source)
		if item.Status == "enabled" && kind != "unchanged" {
			switch {
			case overrides[pathKey(item.Path)] != "":
				row.Origin = "manual"
				row.Label = overrides[pathKey(item.Path)]
				row.NeedsTranslation = false
			case incrementalGlossaryLabel(item) != "":
				row.Origin = "glossary"
				row.Label = incrementalGlossaryLabel(item)
				row.NeedsTranslation = false
			case needsTranslation:
				row.Origin = "pending"
				row.NeedsTranslation = true
			default:
				row.Origin = "existing"
				row.Label = old.Zh
				row.NeedsTranslation = false
			}
			out.Rows = append(out.Rows, row)
		} else if needsTranslation {
			row.Origin = "pending"
			row.NeedsTranslation = true
			out.Rows = append(out.Rows, row)
		}
		if row.NeedsTranslation {
			out.SourceChars += utf8.RuneCountInString(row.Source)
		}
	}
	for id := range active.Categories {
		if _, found := source.Categories[id]; !found {
			out.RemovedIDs = append(out.RemovedIDs, id)
		}
	}
	slices.Sort(out.RemovedIDs)
	out.Counts = incrementalCounts{
		Added: len(out.AddedIDs), ChangedPath: len(out.ChangedPathIDs),
		MetadataChanged: len(out.MetadataChangedIDs), Removed: len(out.RemovedIDs),
		Unchanged: len(out.UnchangedIDs), NewlyEnabled: len(out.NewlyEnabledIDs),
		ReviewRequired: len(out.Rows),
	}
	for _, row := range out.Rows {
		if row.NeedsTranslation {
			out.Counts.TranslationNeeded++
		}
	}
	out.NoOp = source.Version == active.Version
	return out
}

func PlanIncremental(ctx context.Context, dir string) (IncrementalPlanReport, error) {
	client := &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return planIncremental(ctx, dir, client, DumpURL)
}

func planIncremental(ctx context.Context, dir string, client *http.Client, endpoint string) (IncrementalPlanReport, error) {
	var out IncrementalPlanReport
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
	activeRaw, err := readIncrementalFile(filepath.Join(dir, activeFile))
	if err != nil {
		return out, errors.New("active CBT dictionary is unavailable")
	}
	active, err := loadSnapshot(filepath.Join(dir, activeFile))
	if err != nil {
		return out, err
	}
	activeHash := incrementalHash(activeRaw)
	resp, officialVersion, err := fetchDump(ctx, client, endpoint)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	source, err := readDump(resp.Body, officialVersion)
	if err != nil {
		return out, err
	}
	if _, err := compareIncrementalVersions(active.Version, source.Version); err != nil {
		return out, err
	}
	overrides, err := loadOverrides(filepath.Join(dir, overrideFile))
	if err != nil {
		return out, err
	}
	manifest := buildIncrementalManifest(active, source, activeHash, overrides)
	sourceHash := ""
	if _, sourceHash, err = incrementalJSON(source); err != nil {
		return out, err
	}
	manifest.SourceSHA256 = sourceHash
	usage, err := loadLLMUsage(dir)
	if err != nil {
		return out, err
	}
	manifest.UsageChars = usage.Chars
	manifest.BudgetRemaining = max(0, LLMReviewHardLimit-usage.Chars)
	manifest.BudgetFits = manifest.SourceChars <= manifest.BudgetRemaining
	if _, err := writeIncrementalJSON(dir, incrementalSourceFile, source); err != nil {
		return out, err
	}
	manifestHash, err := writeIncrementalJSON(dir, incrementalManifestFile, manifest)
	if err != nil {
		return out, err
	}
	out = IncrementalPlanReport{
		ActiveSHA256: activeHash, SourceSHA256: sourceHash, ManifestSHA256: manifestHash,
		ActiveVersion: active.Version, SourceVersion: source.Version, Counts: manifest.Counts,
		AddedIDs: manifest.AddedIDs, ChangedPathIDs: manifest.ChangedPathIDs,
		MetadataChangedIDs: manifest.MetadataChangedIDs, RemovedIDs: manifest.RemovedIDs,
		UnchangedIDs: manifest.UnchangedIDs, NewlyEnabledIDs: manifest.NewlyEnabledIDs,
		SourceChars: manifest.SourceChars, UsageChars: manifest.UsageChars,
		BudgetRemaining: manifest.BudgetRemaining, BudgetFits: manifest.BudgetFits,
		NoOp: manifest.NoOp,
	}
	return out, nil
}

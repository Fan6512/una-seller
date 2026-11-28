package cbtcategory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

const llmReleaseManagedFile = "llm-release-managed.json"

// Once a managed release is attempted, ordinary updates must fail closed.
// This marker intentionally survives rollback and partial publication failures.
func rejectOrdinaryUpdateAfterRelease(dir string) error {
	_, err := os.Lstat(filepath.Join(dir, llmReleaseManagedFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("ordinary CBT dictionary updater disabled after managed LLM release; use the reviewed incremental workflow")
}

type ReleasePublishInput struct {
	Approval       string
	ApprovalSHA256 string
	ActiveSHA256   string
	Audit          string
	Baseline       string
	Exclusions     string
	Corrections    string
}

type releaseJudgment struct {
	ID         string `json:"id"`
	OriginalOK *bool  `json:"original_ok"`
	FinalOK    bool   `json:"final_ok"`
	FinalLabel string `json:"final_label"`
}

type releaseApproval struct {
	CandidateSHA256   string            `json:"candidate_sha256"`
	AuditSHA256       string            `json:"audit_sha256"`
	CorrectionsSHA256 string            `json:"corrections_sha256"`
	Reviewer          string            `json:"reviewer"`
	Rows              []releaseJudgment `json:"rows"`
}

type releaseCorrection struct {
	ID        string   `json:"id"`
	PathNames []string `json:"path_names"`
	Label     string   `json:"suggested_zh"`
}

type ReleaseSwitchReport struct {
	Applied      bool   `json:"applied"`
	BeforeSHA256 string `json:"before_sha256"`
	AfterSHA256  string `json:"after_sha256"`
	Backup       string `json:"backup"`
	OriginalOK   int    `json:"original_ok,omitempty"`
	Audited      int    `json:"audited,omitempty"`
}

func releaseHash(raw []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

func validReleaseHash(value string) bool {
	b, err := hex.DecodeString(value)
	return err == nil && len(b) == sha256.Size && value == strings.ToLower(value)
}

func releaseRead(path string) ([]byte, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("release evidence file required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxJSON {
		return nil, errors.New("invalid release evidence file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) > maxJSON {
		return nil, errors.New("release evidence exceeds size limit")
	}
	return raw, nil
}

func releaseLabel(item category) string {
	if item.Zh != "" {
		return item.Zh
	}
	return item.Name
}

func releaseBackupName(hash string) string {
	return "llm-release-backup-" + hash + ".json"
}

// Only the explicit publish/rollback commands may switch active.json here.
// The regular updater must remain disabled until the LLM incremental workflow exists.
func PublishLLMRelease(dir string, input ReleasePublishInput, apply bool) (out ReleaseSwitchReport, err error) {
	if strings.TrimSpace(dir) == "" || !validReleaseHash(input.ApprovalSHA256) || !validReleaseHash(input.ActiveSHA256) {
		return out, errors.New("publish requires expected approval and active SHA256")
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

	approvalRaw, err := releaseRead(input.Approval)
	if err != nil || releaseHash(approvalRaw) != input.ApprovalSHA256 {
		return out, errors.New("approval file hash mismatch or unreadable")
	}
	var approval releaseApproval
	if err := json.Unmarshal(approvalRaw, &approval); err != nil || strings.TrimSpace(approval.Reviewer) == "" || len(approval.Rows) != 300 {
		return out, errors.New("approval must identify reviewer and all 300 judgments")
	}
	auditRaw, err := releaseRead(input.Audit)
	if err != nil || releaseHash(auditRaw) != approval.AuditSHA256 {
		return out, errors.New("frozen audit report hash mismatch")
	}
	var audit ReleaseCandidateReport
	if err := json.Unmarshal(auditRaw, &audit); err != nil || audit.AuditSelected != 300 || len(audit.AuditRows) != 300 || audit.AuditEligible < 300 || strings.TrimSpace(audit.Seed) == "" {
		return out, errors.New("invalid frozen audit report")
	}
	baselineRaw, err := releaseRead(input.Baseline)
	if err != nil || releaseHash(baselineRaw) != audit.SnapshotSHA256 {
		return out, errors.New("frozen baseline hash mismatch")
	}
	baseline, err := loadSnapshot(input.Baseline)
	if err != nil || baseline.Version != audit.Version {
		return out, errors.New("invalid frozen baseline snapshot")
	}
	excluded, err := ReadReviewExclusions(input.Exclusions)
	if err != nil || releaseExclusionsHash(excluded) != audit.ExclusionsSHA256 {
		return out, errors.New("historical exclusion set hash mismatch")
	}
	report, expected, origins, err := releaseSnapshotLocked(dir, audit.Seed)
	if err != nil {
		return out, err
	}
	if report.Usage < audit.Usage {
		return out, errors.New("LLM usage counter decreased after audit")
	}
	finalRaw, err := releaseRead(filepath.Join(dir, llmReleaseCandidateFile))
	if err != nil || releaseHash(finalRaw) != approval.CandidateSHA256 {
		return out, errors.New("final candidate hash mismatch")
	}
	expectedRaw, err := json.Marshal(expected)
	if err != nil || !bytes.Equal(finalRaw, append(expectedRaw, '\n')) {
		return out, errors.New("candidate does not match current build and manual corrections; reassemble offline")
	}
	if err := validateReleaseIdentity(baseline, expected); err != nil {
		return out, err
	}
	for id, item := range expected.Categories {
		if baseline.Categories[id].Zh != item.Zh && origins[id] != "manual" {
			return out, fmt.Errorf("non-manual candidate change after frozen audit: %s", id)
		}
	}

	correctionRaw, err := releaseRead(input.Corrections)
	if err != nil || releaseHash(correctionRaw) != approval.CorrectionsSHA256 {
		return out, errors.New("known correction file hash mismatch")
	}
	var corrections []releaseCorrection
	if err := json.Unmarshal(correctionRaw, &corrections); err != nil || len(corrections) == 0 || len(corrections) > maxCategories {
		return out, errors.New("invalid known correction file")
	}
	known := make(map[string]bool)
	for _, fix := range corrections {
		item, ok := expected.Categories[fix.ID]
		if !ok || known[fix.ID] || origins[fix.ID] != "manual" || item.Zh != fix.Label || len(item.Path) != len(fix.PathNames) {
			return out, fmt.Errorf("known correction not resolved: %s", fix.ID)
		}
		for i, node := range item.Path {
			if node.Name != fix.PathNames[i] {
				return out, fmt.Errorf("known correction path changed: %s", fix.ID)
			}
		}
		known[fix.ID] = true
	}

	judgments := make(map[string]releaseJudgment)
	for _, row := range approval.Rows {
		if row.OriginalOK == nil || !row.FinalOK || !validName(row.FinalLabel) || judgments[row.ID].ID != "" || !idPattern.MatchString(row.ID) {
			return out, errors.New("incomplete or duplicate final judgment")
		}
		judgments[row.ID] = row
	}
	for _, row := range audit.AuditRows {
		judgment, ok := judgments[row.ID]
		old, found := baseline.Categories[row.ID]
		item := expected.Categories[row.ID]
		if !ok || !found || excluded[row.ID] || old.Status != "enabled" || !slices.Equal(old.Path, row.Path) ||
			releaseLabel(old) != row.Label || row.Original != old.Name || judgment.FinalLabel != releaseLabel(item) ||
			(row.Origin != "llm" && row.Origin != "deferred_english" && row.Origin != "unsafe_english") {
			return out, fmt.Errorf("audit evidence mismatch: %s", row.ID)
		}
		if *judgment.OriginalOK {
			out.OriginalOK++
		} else if origins[row.ID] != "manual" || releaseLabel(item) == row.Label {
			return out, fmt.Errorf("audit error not corrected: %s", row.ID)
		}
		delete(judgments, row.ID)
		out.Audited++
	}
	if len(judgments) != 0 || out.Audited != 300 || out.OriginalOK < 285 {
		return out, errors.New("original independent audit below 95 percent or incomplete")
	}
	return switchReleaseLocked(dir, input.ActiveSHA256, finalRaw, out, apply)
}

func RollbackLLMRelease(dir, expectedActive, backupHash string, apply bool) (out ReleaseSwitchReport, err error) {
	if strings.TrimSpace(dir) == "" || !validReleaseHash(expectedActive) || !validReleaseHash(backupHash) {
		return out, errors.New("rollback requires expected active and backup SHA256")
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
	pending, err := buildPending(dir)
	if err != nil || pending {
		return out, errors.New("unresolved LLM build pending attempt")
	}
	raw, err := releaseRead(filepath.Join(dir, releaseBackupName(backupHash)))
	if err != nil || releaseHash(raw) != backupHash {
		return out, errors.New("rollback backup missing or hash mismatch")
	}
	return switchReleaseLocked(dir, expectedActive, raw, out, apply)
}

func switchReleaseLocked(dir, expectedActive string, next []byte, out ReleaseSwitchReport, apply bool) (ReleaseSwitchReport, error) {
	activePath := filepath.Join(dir, activeFile)
	current, err := releaseRead(activePath)
	if err != nil || releaseHash(current) != expectedActive {
		return out, errors.New("current active snapshot hash changed")
	}
	active, err := loadSnapshot(activePath)
	if err != nil {
		return out, err
	}
	var target snapshot
	if err := json.Unmarshal(next, &target); err != nil {
		return out, err
	}
	if err := validateReleaseIdentity(active, target); err != nil {
		return out, err
	}
	// A translation release/rollback cannot change the official version or
	// disabled category labels. It never touches overrides or source files.
	for id, item := range active.Categories {
		if item.Status != "enabled" && !reflect.DeepEqual(item, target.Categories[id]) {
			return out, fmt.Errorf("disabled category changed: %s", id)
		}
		if item := target.Categories[id]; item.Zh != "" && !validName(item.Zh) {
			return out, fmt.Errorf("invalid target label: %s", id)
		}
	}
	out.BeforeSHA256, out.AfterSHA256 = expectedActive, releaseHash(next)
	out.Backup = releaseBackupName(expectedActive)
	if out.BeforeSHA256 == out.AfterSHA256 {
		return out, nil
	}
	if !apply {
		return out, nil
	}
	backupPath := filepath.Join(dir, out.Backup)
	if previous, err := releaseRead(backupPath); err == nil {
		if !bytes.Equal(previous, current) {
			return out, errors.New("existing release backup is invalid")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return out, err
	} else if err := writeReleaseBytes(dir, out.Backup, current); err != nil {
		return out, err
	}
	// Persist the guard before active can be replaced. If publication fails
	// after this point, keep the guard in place until an operator investigates.
	if err := writeReleaseBytes(dir, llmReleaseManagedFile, []byte("{\"managed\":true}\n")); err != nil {
		return out, err
	}
	if err := writeReleaseBytes(dir, activeFile, next); err != nil {
		return out, err
	}
	out.Applied = true
	return out, nil
}

// Preserve exact bytes for SHA-bound evidence and rollback. Run as the
// dictionary service account, with the same private directory permissions.
func writeReleaseBytes(dir, name string, raw []byte) error {
	f, err := os.CreateTemp(dir, ".llm-release-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(raw); err == nil {
		err = f.Sync()
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	if err := os.Rename(f.Name(), filepath.Join(dir, name)); err != nil {
		return err
	}
	// On production Linux, fsync the directory entry before reporting success.
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

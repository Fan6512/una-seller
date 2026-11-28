package cbtcategory

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type IncrementalPublishInput struct {
	ApprovalPath   string
	ApprovalSHA256 string
	ActiveSHA256   string
}

type IncrementalSwitchReport struct {
	Applied      bool   `json:"applied"`
	BeforeSHA256 string `json:"before_sha256"`
	AfterSHA256  string `json:"after_sha256"`
	Backup       string `json:"backup"`
	ReviewRows   int    `json:"review_rows"`
}

func PublishIncremental(dir string, input IncrementalPublishInput, apply bool) (out IncrementalSwitchReport, err error) {
	if strings.TrimSpace(dir) == "" || !validReleaseHash(input.ActiveSHA256) ||
		!validReleaseHash(input.ApprovalSHA256) || strings.TrimSpace(input.ApprovalPath) == "" {
		return out, errors.New("incremental publish requires expected active SHA256, approval path and approval SHA256")
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
	active, activeRaw, source, manifest, manifestHash, err := loadIncrementalPlanLocked(dir)
	if err != nil {
		return out, err
	}
	if manifest.ActiveSHA256 != input.ActiveSHA256 {
		return out, errors.New("publish active SHA256 does not match the frozen plan")
	}
	candidateRaw, err := readIncrementalFile(filepath.Join(dir, incrementalCandidateFile))
	if err != nil {
		return out, errors.New("incremental candidate is unavailable; run assemble again")
	}
	candidateHash := incrementalHash(candidateRaw)
	var candidate snapshot
	if json.Unmarshal(candidateRaw, &candidate) != nil {
		return out, errors.New("incremental candidate JSON is invalid")
	}
	reviewRaw, err := readIncrementalFile(filepath.Join(dir, incrementalReviewFile))
	if err != nil {
		return out, errors.New("incremental review is unavailable; run assemble again")
	}
	reviewHash := incrementalHash(reviewRaw)
	var review incrementalReview
	if json.Unmarshal(reviewRaw, &review) != nil ||
		review.ActiveSHA256 != manifest.ActiveSHA256 || review.SourceSHA256 != manifest.SourceSHA256 ||
		review.ManifestSHA256 != manifestHash || review.CandidateSHA256 != candidateHash {
		return out, errors.New("incremental review hash binding does not match current candidate")
	}
	if len(review.Rows) != len(manifest.Rows) {
		return out, errors.New("incremental review does not include every enabled delta row")
	}
	if err := validateIncrementalCandidateIdentity(active, source, candidate, manifest); err != nil {
		return out, err
	}

	approvalRaw, err := readIncrementalFile(input.ApprovalPath)
	if err != nil || incrementalHash(approvalRaw) != input.ApprovalSHA256 {
		return out, errors.New("incremental approval file hash mismatch or unreadable")
	}
	var approval incrementalApproval
	if json.Unmarshal(approvalRaw, &approval) != nil ||
		strings.TrimSpace(approval.Reviewer) == "" ||
		approval.ActiveSHA256 != manifest.ActiveSHA256 ||
		approval.SourceSHA256 != manifest.SourceSHA256 ||
		approval.ManifestSHA256 != manifestHash ||
		approval.CandidateSHA256 != candidateHash ||
		approval.ReviewSHA256 != reviewHash ||
		len(approval.Rows) != len(review.Rows) {
		return out, errors.New("incremental approval is incomplete or bound to stale evidence")
	}
	judgments := make(map[string]incrementalApprovalRow, len(approval.Rows))
	for _, judgment := range approval.Rows {
		if !idPattern.MatchString(judgment.ID) || !validReleaseHash(judgment.SourceSHA256) ||
			judgments[judgment.ID].ID != "" || !judgment.Accepted {
			return out, errors.New("incremental approval has a missing, duplicate or rejected judgment")
		}
		judgments[judgment.ID] = judgment
	}
	for _, row := range review.Rows {
		judgment, ok := judgments[row.ID]
		if !ok || judgment.SourceSHA256 != row.SourceSHA256 {
			return out, fmt.Errorf("incremental approval row missing or stale: %s", row.ID)
		}
		item, found := candidate.Categories[row.ID]
		if !found || item.Name != row.Original || !slices.Equal(item.Path, row.Path) ||
			row.Source != reviewSource(row.Path, true) || row.SourceSHA256 != sourceHash(row.Source) {
			return out, fmt.Errorf("incremental review identity mismatch: %s", row.ID)
		}
		if row.FallbackEnglish {
			if !judgment.AcceptEnglishFallback || judgment.FinalLabel != row.Original {
				return out, fmt.Errorf("English fallback requires explicit acceptance for %s", row.ID)
			}
			item.Zh = ""
		} else {
			if judgment.AcceptEnglishFallback || !validReviewLanguage(judgment.FinalLabel, item) {
				return out, fmt.Errorf("invalid final Chinese label for %s", row.ID)
			}
			switch row.Origin {
			case "manual", "glossary", "existing":
				if judgment.FinalLabel != row.Candidate {
					return out, fmt.Errorf("protected %s label cannot be changed in approval: %s", row.Origin, row.ID)
				}
			case "llm":
				// The reviewer may correct an LLM candidate, but must approve
				// the exact final label that will be published.
			default:
				return out, fmt.Errorf("unknown incremental review origin for %s", row.ID)
			}
			item.Zh = judgment.FinalLabel
		}
		candidate.Categories[row.ID] = item
		delete(judgments, row.ID)
	}
	if len(judgments) != 0 {
		return out, errors.New("incremental approval contains unreviewed extra rows")
	}
	if err := validateIncrementalCandidateIdentity(active, source, candidate, manifest); err != nil {
		return out, err
	}
	finalRaw, finalHash, err := incrementalJSON(candidate)
	if err != nil {
		return out, err
	}
	out = IncrementalSwitchReport{
		BeforeSHA256: manifest.ActiveSHA256, AfterSHA256: finalHash,
		Backup: releaseBackupName(manifest.ActiveSHA256), ReviewRows: len(review.Rows),
	}
	if out.BeforeSHA256 == out.AfterSHA256 {
		return out, nil
	}
	if !apply {
		return out, nil
	}
	if err := switchIncrementalLocked(dir, manifest.ActiveSHA256, activeRaw, finalRaw); err != nil {
		return out, err
	}
	out.Applied = true
	return out, nil
}

func switchIncrementalLocked(dir, expectedActive string, currentRaw, nextRaw []byte) error {
	activePath := filepath.Join(dir, activeFile)
	current, err := readIncrementalFile(activePath)
	if err != nil || incrementalHash(current) != expectedActive {
		return errors.New("current active snapshot hash changed")
	}
	if len(currentRaw) > 0 && !slices.Equal(current, currentRaw) {
		return errors.New("active snapshot bytes changed during incremental publish")
	}
	var target snapshot
	if json.Unmarshal(nextRaw, &target) != nil || !target.Version.valid() ||
		len(target.Categories) == 0 || len(target.Categories) > maxCategories {
		return errors.New("invalid incremental target snapshot")
	}
	if _, err := os.Stat(filepath.Join(dir, llmReleaseManagedFile)); err != nil {
		return errors.New("managed release marker disappeared before active switch")
	}
	backupName := releaseBackupName(expectedActive)
	backupPath := filepath.Join(dir, backupName)
	if previous, err := readIncrementalFile(backupPath); err == nil {
		if !slices.Equal(previous, current) {
			return errors.New("existing content-addressed incremental backup is invalid")
		}
	} else if !os.IsNotExist(err) {
		return err
	} else if err := writeReleaseBytes(dir, backupName, current); err != nil {
		return err
	}
	if err := writeReleaseBytes(dir, activeFile, nextRaw); err != nil {
		return err
	}
	return nil
}

func RollbackIncremental(dir, expectedActive, backupHash string, apply bool) (out IncrementalSwitchReport, err error) {
	if strings.TrimSpace(dir) == "" || !validReleaseHash(expectedActive) || !validReleaseHash(backupHash) {
		return out, errors.New("incremental rollback requires expected active and backup SHA256")
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
	current, err := readIncrementalFile(filepath.Join(dir, activeFile))
	if err != nil || incrementalHash(current) != expectedActive {
		return out, errors.New("current active snapshot hash changed before incremental rollback")
	}
	backupName := releaseBackupName(backupHash)
	backupPath := filepath.Join(dir, backupName)
	backup, err := readIncrementalFile(backupPath)
	if err != nil || incrementalHash(backup) != backupHash {
		return out, errors.New("incremental rollback backup missing or hash mismatch")
	}
	if _, err := loadSnapshot(backupPath); err != nil {
		return out, errors.New("incremental rollback backup is not a valid dictionary snapshot")
	}
	out = IncrementalSwitchReport{
		BeforeSHA256: expectedActive, AfterSHA256: backupHash,
		Backup: releaseBackupName(expectedActive),
	}
	if expectedActive == backupHash {
		return out, nil
	}
	if !apply {
		return out, nil
	}
	if err := switchIncrementalLocked(dir, expectedActive, current, backup); err != nil {
		return out, err
	}
	out.Applied = true
	return out, nil
}

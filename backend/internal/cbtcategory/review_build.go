package cbtcategory

import (
	"context"
	"encoding/json"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const llmBuildDeferredFile = "llm-build-deferred.json"
const llmBuildPendingFile = "llm-build-pending.json"
const llmBuildReceiptsFile = "llm-build-receipts.jsonl"
const LLMBuildMaxRows = 200

type buildDeferred struct {
	Version version           `json:"version"`
	Rows    map[string]string `json:"source_sha256_by_id"`
}

type LLMBuildPlan struct {
	ReviewPlan
	Deferred   int  `json:"deferred_english"`
	Usage      int  `json:"usage_chars"`
	Budget     int  `json:"budget_chars"`
	BudgetFits bool `json:"budget_fits"`
	Pending    bool `json:"pending_attempt"`
}

type LLMBuildResult struct {
	LLMTrialResult
	RemainingSourceChars int `json:"remaining_source_chars_before"`
	Deferred             int `json:"deferred_english"`
	Budget               int `json:"budget_chars"`
}

type buildReceipt struct {
	Version     version         `json:"version"`
	UsageBefore int             `json:"usage_before_chars"`
	Attempt     LLMTrialAttempt `json:"attempt"`
}

func buildPending(dir string) (bool, error) {
	_, err := os.Stat(filepath.Join(dir, llmBuildPendingFile))
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

// A complete durable receipt precedes removal of the pending marker.
// A torn tail must be archived and inspected by the operator, not appended to.
func appendBuildReceipt(dir string, receipt buildReceipt) error {
	f, err := os.OpenFile(filepath.Join(dir, llmBuildReceiptsFile), os.O_CREATE|os.O_RDWR|os.O_APPEND, 0600)
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
		if _, err := f.ReadAt(last[:], info.Size()-1); err != nil {
			return err
		}
		if last[0] != '\n' {
			return errors.New("incomplete LLM build receipt tail")
		}
	}
	if err := json.NewEncoder(f).Encode(receipt); err != nil {
		return err
	}
	return f.Sync()
}

func sourceHash(source string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(source)))
}

func loadBuildDeferred(dir string, s snapshot) (buildDeferred, error) {
	d := buildDeferred{Version: s.Version, Rows: make(map[string]string)}
	err := readJSONFile(filepath.Join(dir, llmBuildDeferredFile), &d)
	if errors.Is(err, os.ErrNotExist) {
		return d, nil
	}
	if err != nil {
		return d, err
	}
	if d.Version != s.Version || d.Rows == nil || len(d.Rows) > maxCategories {
		return d, errors.New("invalid or stale LLM build deferred state")
	}
	for id, hash := range d.Rows {
		item, ok := s.Categories[id]
		if !ok || item.Status != "enabled" || hash != sourceHash(reviewSource(item.Path, true)) {
			return d, fmt.Errorf("stale deferred category %q", id)
		}
	}
	return d, nil
}

// buildPlanLocked holds the ordinary dictionary writer lock for the entire
// batch. The complete remaining budget is checked before a single request.
func buildPlanLocked(dir, seed string, budget int, export bool) (LLMBuildPlan, snapshot, map[string]ReviewRow, buildDeferred, error) {
	var out LLMBuildPlan
	s, source, manual, err := reviewSnapshot(dir)
	if err != nil {
		return out, s, nil, buildDeferred{}, err
	}
	completed, err := loadReview(dir, s)
	if err != nil {
		return out, s, nil, buildDeferred{}, err
	}
	for id, row := range completed {
		item := s.Categories[id]
		if manual[pathKey(item.Path)] == "" && !reviewGlossaryProtected(item) && row.Source != reviewSource(item.Path, true) {
			return out, s, nil, buildDeferred{}, fmt.Errorf("candidate %s was not built with full path; archive and review before full build", id)
		}
	}
	deferred, err := loadBuildDeferred(dir, s)
	if err != nil {
		return out, s, nil, deferred, err
	}
	exclude := make(map[string]bool, len(deferred.Rows))
	for id := range deferred.Rows {
		exclude[id] = true
	}
	plan, err := planReviewSnapshot(s, source, manual, completed, "all", seed, 0, true, export, exclude)
	if err != nil {
		return out, s, nil, deferred, err
	}
	out.ReviewPlan = plan
	out.Deferred, out.PriorExcluded = plan.PriorExcluded, 0
	state, err := loadLLMUsage(dir)
	if err != nil {
		return out, s, nil, deferred, err
	}
	out.Usage, out.Budget = state.Chars, budget
	out.BudgetFits = state.Chars <= budget && plan.SourceChars <= budget-state.Chars
	out.Pending, err = buildPending(dir)
	if err != nil {
		return out, s, nil, deferred, err
	}
	return out, s, completed, deferred, nil
}

func PlanLLMBuild(dir, seed string, budget int) (LLMBuildPlan, error) {
	if strings.TrimSpace(dir) == "" || budget <= 0 || budget > LLMReviewHardLimit {
		return LLMBuildPlan{}, errors.New("invalid LLM build budget or directory")
	}
	lock, err := lockUpdate(filepath.Join(dir, "update.lock"))
	if err != nil {
		return LLMBuildPlan{}, err
	}
	defer lock.Close()
	plan, _, _, _, err := buildPlanLocked(dir, seed, budget, false)
	return plan, err
}

// RunLLMBuild is an explicit, separate maintenance command, never called by
// normal translation or the trial command. It cannot publish a dictionary.
func RunLLMBuild(ctx context.Context, dir, seed string, maxRows, budget int, appID, key string) (LLMBuildResult, error) {
	appID, key = strings.TrimSpace(appID), strings.TrimSpace(key)
	if appID == "" || key == "" || strings.ContainsAny(key, "\r\n") {
		return LLMBuildResult{}, errors.New("explicit LLM credentials required")
	}
	return runLLMBuild(ctx, dir, seed, maxRows, llmReviewProvider{
		appID: appID, key: key, endpoint: baiduLLMEndpoint, limit: budget,
		client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	})
}

func runLLMBuild(ctx context.Context, dir, seed string, maxRows int, p llmReviewProvider) (out LLMBuildResult, runErr error) {
	if strings.TrimSpace(dir) == "" || maxRows <= 0 || maxRows > LLMBuildMaxRows || p.limit <= 0 || p.limit > LLMReviewHardLimit {
		return out, errors.New("build requires 1-200 rows and explicit budget <= 900000")
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
	defer func() {
		state, err := loadLLMUsage(dir)
		if err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("LLM usage report unavailable: %w", err))
			return
		}
		out.Usage = &state.Chars
	}()
	plan, s, completed, deferred, err := buildPlanLocked(dir, seed, p.limit, true)
	if err != nil {
		return out, err
	}
	out.Version, out.UsageBefore, out.Budget = plan.Version, &plan.Usage, p.limit
	out.RemainingSourceChars = plan.SourceChars
	if plan.Pending {
		return out, errors.New("unresolved LLM build pending attempt; reconcile usage and archive pending evidence before resuming")
	}
	if !plan.BudgetFits {
		return out, errors.New("complete remaining LLM build exceeds cumulative budget")
	}
	rows := plan.Rows[:min(maxRows, len(plan.Rows))]
	out.Selected = len(rows)
	for _, row := range rows {
		out.PlannedSourceChars += utf8.RuneCountInString(row.Source)
	}
	for i, row := range rows {
		if i > 0 {
			timer := time.NewTimer(time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return out, ctx.Err()
			case <-timer.C:
			}
		}
		if err := ctx.Err(); err != nil {
			return out, err
		}
		attempt := LLMTrialAttempt{ID: row.ID, SourceSHA256: sourceHash(row.Source),
			SourceChars: utf8.RuneCountInString(row.Source), Outcome: "pending"}
		state, err := loadLLMUsage(dir)
		if err != nil {
			return out, err
		}
		// Write intent before reserving/sending. A crash or unknown response
		// blocks automatic retry until an operator reconciles and archives it.
		receipt := buildReceipt{Version: s.Version, UsageBefore: state.Chars, Attempt: attempt}
		if err := saveJSON(dir, llmBuildPendingFile, receipt); err != nil {
			return out, err
		}
		label, reserved, err := p.translate(ctx, dir, row.Source)
		attempt.ReservedChars = reserved
		out.ReservedChars += reserved
		if err != nil {
			attempt.Outcome = "request_failed"
			out.Attempts = append(out.Attempts, attempt)
			return out, fmt.Errorf("LLM build row %s: %w", row.ID, err)
		}
		row.Candidate = label
		item := s.Categories[row.ID]
		if !validReviewLanguage(label, item) && validName(label) && !strings.ContainsAny(label, "\r\n|｜") {
			// Never trust an arbitrary English label. Keep the official English
			// fallback separately; this is not a translated candidate.
			deferred.Rows[row.ID] = attempt.SourceSHA256
			err = saveJSON(dir, llmBuildDeferredFile, deferred)
			attempt.Outcome = "deferred_english"
			if err == nil {
				out.Deferred++
			}
		} else if validReviewRow(row, item) {
			completed[row.ID] = row
			err = saveReviewMap(dir, s.Version, completed)
			attempt.Outcome = "stored"
			if err == nil {
				out.Completed++
			}
		} else {
			err = errors.New("invalid LLM build candidate")
			attempt.Outcome = "candidate_rejected"
		}
		if err != nil {
			attempt.Outcome = "candidate_not_committed"
			out.Attempts = append(out.Attempts, attempt)
			return out, fmt.Errorf("LLM build row %s: %w", row.ID, err)
		}
		out.Attempts = append(out.Attempts, attempt)
		receipt.Attempt = attempt
		if err := appendBuildReceipt(dir, receipt); err != nil {
			return out, err
		}
		if err := os.Remove(filepath.Join(dir, llmBuildPendingFile)); err != nil {
			return out, err
		}
	}
	return out, nil
}

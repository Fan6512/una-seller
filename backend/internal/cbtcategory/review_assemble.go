package cbtcategory

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

const llmReleaseCandidateFile = "llm-release-candidate.json"
const llmReleaseAuditFile = "llm-release-audit.json"

type releaseAuditRow struct {
	ID       string     `json:"id"`
	Path     []PathNode `json:"path"`
	Label    string     `json:"label"`
	Origin   string     `json:"origin"`
	Original string     `json:"original"`
}

type ReleaseCandidateReport struct {
	Version        version           `json:"version"`
	SnapshotSHA256 string            `json:"snapshot_sha256"`
	Seed           string            `json:"seed"`
	Enabled        int               `json:"enabled"`
	Disabled       int               `json:"disabled"`
	Manual         int               `json:"manual"`
	Glossary       int               `json:"glossary"`
	LLM            int               `json:"llm"`
	Deferred       int               `json:"deferred_english"`
	Unsafe         int               `json:"unsafe_english"`
	PriorExcluded    int               `json:"prior_excluded"`
	AuditEligible    int               `json:"audit_eligible"`
	AuditSelected    int               `json:"audit_selected"`
	ExclusionsSHA256 string            `json:"exclusions_sha256"`
	Usage            int               `json:"usage_chars"`
	AuditRows      []releaseAuditRow `json:"audit_rows"`
}

// AssembleLLMCandidate is intentionally offline. It does not write active,
// staging, progress, overrides or the quota counter, and cannot publish.
func AssembleLLMCandidate(dir, seed string, sample int, exclusions map[string]bool) (ReleaseCandidateReport, error) {
	var out ReleaseCandidateReport
	if strings.TrimSpace(dir) == "" || strings.TrimSpace(seed) == "" || sample <= 0 || sample > 300 || len(exclusions) == 0 {
		return out, errors.New("invalid release candidate audit arguments")
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
	plan, s, completed, deferred, err := buildPlanLocked(dir, seed, LLMReviewHardLimit, false)
	if err != nil {
		return out, err
	}
	if plan.Pending || !plan.BudgetFits || plan.Selected != 0 {
		return out, errors.New("LLM build not complete, unresolved pending or quota invalid")
	}
	_, _, manual, err := reviewSnapshot(dir)
	if err != nil {
		return out, err
	}
	out.Version, out.Seed, out.Usage = s.Version, seed, plan.Usage
	// Work only on local values. Keep official category identities, status,
	// listing permissions and retired history unchanged.
	s.Skipped = nil
	origins := make(map[string]string)
	for id, item := range s.Categories {
		if item.Status != "enabled" {
			out.Disabled++
			continue
		}
		out.Enabled++
		switch {
		case manual[pathKey(item.Path)] != "":
			item.Zh = manual[pathKey(item.Path)]
			out.Manual++
			origins[id] = "manual"
		case reviewGlossaryProtected(item):
			out.Glossary++
			origins[id] = "glossary"
		case completed[id].ID != "":
			item.Zh = completed[id].Candidate
			out.LLM++
			origins[id] = "llm"
		case deferred.Rows[id] != "":
			item.Zh = ""
			out.Deferred++
			origins[id] = "deferred_english"
		case !safeReviewSource(reviewSource(item.Path, true)):
			item.Zh = ""
			out.Unsafe++
			origins[id] = "unsafe_english"
		default:
			return out, fmt.Errorf("unaccounted enabled category %s", id)
		}
		s.Categories[id] = item
	}
	if out.Enabled != out.Manual+out.Glossary+out.LLM+out.Deferred+out.Unsafe {
		return out, errors.New("release candidate coverage mismatch")
	}
	// Compare official identities to the underlying local snapshot, including
	// disabled nodes and the complete parent graph. Only enabled Zh may change.
	official, err := loadSnapshot(filepath.Join(dir, plan.Snapshot))
	if err != nil {
		return out, err
	}
	if err := validateReleaseIdentity(official, s); err != nil {
		return out, err
	}
	// Round-robin root strata, deterministic hash order, no prior reviewed
	// IDs or manual/glossary rows. Include English fallbacks in audit scope.
	groups := make(map[string][]releaseAuditRow)
	for id, item := range s.Categories {
		origin := origins[id]
		if origin == "" || origin == "manual" || origin == "glossary" {
			continue
		}
		if exclusions[id] {
			out.PriorExcluded++
			continue
		}
		label := item.Zh
		if label == "" {
			label = item.Name
		}
		root := item.Path[0].ID
		groups[root] = append(groups[root], releaseAuditRow{
			ID: id, Path: item.Path, Label: label, Origin: origin, Original: item.Name,
		})
	}
	roots := make([]string, 0, len(groups))
	eligible := 0
	for root, rows := range groups {
		roots = append(roots, root)
		eligible += len(rows)
		slices.SortFunc(rows, func(a, b releaseAuditRow) int {
			return strings.Compare(sourceHash(seed+"\x00"+a.ID), sourceHash(seed+"\x00"+b.ID))
		})
	}
	if eligible < sample {
		return out, errors.New("insufficient independent audit rows")
	}
	out.AuditEligible = eligible
	ids := make([]string, 0, len(exclusions))
	for id := range exclusions {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var exclusionBytes strings.Builder
	for _, id := range ids {
		exclusionBytes.WriteString(id)
		exclusionBytes.WriteByte('\n')
	}
	exclusionHash := sha256.Sum256([]byte(exclusionBytes.String()))
	out.ExclusionsSHA256 = fmt.Sprintf("%x", exclusionHash)
	slices.Sort(roots)
	for offset := 0; len(out.AuditRows) < sample; offset++ {
		for _, root := range roots {
			if offset < len(groups[root]) && len(out.AuditRows) < sample {
				out.AuditRows = append(out.AuditRows, groups[root][offset])
			}
		}
	}
	out.AuditSelected = len(out.AuditRows)
	// saveJSON uses json.Encoder.Encode: canonical map keys and trailing LF.
	// Compute the exact bytes with the same encoder (including HTML escaping).
	raw, err := json.Marshal(s)
	if err != nil {
		return out, err
	}
	out.SnapshotSHA256 = fmt.Sprintf("%x", sha256.Sum256(append(raw, '\n')))
	if err := saveJSON(dir, llmReleaseCandidateFile, s); err != nil {
		return out, err
	}
	if err := saveJSON(dir, llmReleaseAuditFile, out); err != nil {
		return out, err
	}
	return out, nil
}

func validateReleaseIdentity(official, candidate snapshot) error {
	if official.Version != candidate.Version || len(official.Categories) != len(candidate.Categories) || !reflect.DeepEqual(official.Retired, candidate.Retired) {
		return errors.New("release candidate version or category count changed")
	}
	for id, a := range official.Categories {
		b, ok := candidate.Categories[id]
		if !ok || a.ID != b.ID || a.Name != b.Name || a.ParentID != b.ParentID || a.Depth != b.Depth ||
			a.HasChildren != b.HasChildren || a.ListingAllowed != b.ListingAllowed || a.Status != b.Status ||
			!slices.Equal(a.Path, b.Path) {
			return fmt.Errorf("official category identity changed for %s", id)
		}
		if len(b.Path) > 1 {
			parent, ok := candidate.Categories[b.ParentID]
			if !ok || !slices.Equal(parent.Path, b.Path[:len(b.Path)-1]) {
				return fmt.Errorf("official parent chain mismatch for %s", id)
			}
		} else if b.ParentID != "" {
			return fmt.Errorf("unexpected root parent for %s", id)
		}
	}
	return nil
}

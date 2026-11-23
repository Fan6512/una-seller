package cbtcategory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const activeFile = "active.json"
const overrideFile = "overrides.json"

type override struct {
	Path []PathNode `json:"path"`
	Zh   string     `json:"zh"`
}

// Reader exposes the latest complete file, retaining its last valid copy after a read error.
type Reader struct {
	dir                      string
	mu                       sync.Mutex
	activeInfo, overrideInfo os.FileInfo
	view                     View
	activeError              string
	activeRetryAt            time.Time
	activeBackoff            time.Duration
}

func NewReader(dir string) *Reader {
	if dir = strings.TrimSpace(dir); dir == "" {
		return nil
	}
	return &Reader{dir: dir}
}

type View struct {
	snapshot  snapshot
	overrides map[string]string
}

type Breadcrumb struct {
	Original []string `json:"original"`
	Chinese  []string `json:"zh,omitempty"`
}

func (r *Reader) Load() View {
	if r == nil {
		return View{}
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	path := filepath.Join(r.dir, activeFile)
	if info, err := os.Stat(path); err == nil {
		changed := fileChanged(r.activeInfo, info)
		if changed || (!r.activeRetryAt.IsZero() && !time.Now().Before(r.activeRetryAt)) {
			if value, loadErr := loadSnapshot(path); loadErr == nil {
				r.view.snapshot, r.activeInfo = value, info
				r.activeError, r.activeRetryAt, r.activeBackoff = "", time.Time{}, 0
			} else {
				r.activeInfo, r.activeError = info, "invalid_file"
				if changed {
					r.activeBackoff = 0
				}
				if r.activeBackoff == 0 {
					r.activeBackoff = 5 * time.Second
				} else {
					r.activeBackoff = min(2*r.activeBackoff, 60*time.Second)
				}
				r.activeRetryAt = time.Now().Add(r.activeBackoff)
			}
		}
	} else if errors.Is(err, os.ErrNotExist) {
		r.activeError = "missing_file"
	} else {
		r.activeError = "stat_failed"
	}
	path = filepath.Join(r.dir, overrideFile)
	if info, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		r.view.overrides, r.overrideInfo = nil, nil
	} else if err == nil && fileChanged(r.overrideInfo, info) {
		if values, err := loadOverrides(path); err == nil {
			r.view.overrides, r.overrideInfo = values, info
		}
	}
	return r.view
}

type ReaderStatus struct {
	State      string
	Version    string
	Translated int
}

// Status reports the currently loaded dictionary without exposing its private location.
func (r *Reader) Status() ReaderStatus {
	if r == nil {
		return ReaderStatus{State: "not_configured"}
	}
	view := r.Load()
	r.mu.Lock()
	defer r.mu.Unlock()
	if !view.snapshot.Version.valid() {
		if r.activeError != "" {
			return ReaderStatus{State: r.activeError}
		}
		return ReaderStatus{State: "not_initialized"}
	}
	status := ReaderStatus{State: "loaded", Version: view.snapshot.Version.Created}
	for _, item := range view.snapshot.Categories {
		if item.Status != "enabled" {
			continue
		}
		label := view.overrides[pathKey(item.Path)]
		if label == "" {
			label = item.Zh
		}
		if label != "" && label != item.Name {
			status.Translated++
		}
	}
	if status.Translated == 0 {
		status.State = "loaded_untranslated"
	}
	if r.activeError != "" {
		status.State = "stale_" + r.activeError
	}
	return status
}

func fileChanged(old, current os.FileInfo) bool {
	return old == nil || !os.SameFile(old, current) || old.Size() != current.Size() || !old.ModTime().Equal(current.ModTime())
}

func (v View) UnavailableForNewListing(id string) bool {
	if item, ok := v.snapshot.Categories[id]; ok {
		return item.Status != "enabled" || !item.ListingAllowed || item.HasChildren
	}
	return len(v.snapshot.Retired[id]) > 0
}

// Localize checks the entire official path before applying any labels.
func (v View) Localize(path []PathNode) []string {
	if len(path) == 0 {
		return nil
	}
	item, ok := v.snapshot.Categories[path[len(path)-1].ID]
	if !ok || !slices.Equal(path, item.Path) {
		return nil
	}
	return v.breadcrumb(item.Path, false).Chinese
}

// PublishedBreadcrumb accepts only a CBT category ID from a verified listing snapshot.
func (v View) PublishedBreadcrumb(id string) *Breadcrumb {
	item, found := v.snapshot.Categories[id]
	if !found {
		history := v.snapshot.Retired[id]
		if len(history) == 0 {
			return nil
		}
		item = history[len(history)-1]
	}
	result := v.breadcrumb(item.Path, true)
	return &result
}

func (v View) breadcrumb(path []PathNode, includeRetired bool) Breadcrumb {
	result := Breadcrumb{Original: make([]string, len(path))}
	labels := make([]string, len(path))
	hasChinese := false
	for i, node := range path {
		result.Original[i] = node.Name
		prefix := path[:i+1]
		item, found := v.snapshot.Categories[node.ID]
		if !found || !slices.Equal(prefix, item.Path) {
			item = category{}
			if includeRetired {
				for _, retired := range v.snapshot.Retired[node.ID] {
					if slices.Equal(prefix, retired.Path) {
						item = retired
						break
					}
				}
			}
		}
		label := v.overrides[pathKey(prefix)]
		if label == "" {
			label = item.Zh
		}
		if label == "" {
			label = node.Name
		}
		labels[i] = label
		hasChinese = hasChinese || label != node.Name
	}
	if hasChinese {
		result.Chinese = labels
	}
	return result
}

func pathKey(path []PathNode) string {
	hash := sha256.New()
	for _, node := range path {
		fmt.Fprintf(hash, "%d:%s%d:%s", len(node.ID), node.ID, len(node.Name), node.Name)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func loadSnapshot(path string) (snapshot, error) {
	var value snapshot
	if err := readJSONFile(path, &value); err != nil {
		return snapshot{}, err
	}
	if !value.Version.valid() || len(value.Categories) == 0 || len(value.Categories) > maxCategories {
		return snapshot{}, errors.New("invalid CBT dictionary snapshot")
	}
	for id, item := range value.Categories {
		if !idPattern.MatchString(id) || item.ID != id || !validName(item.Name) ||
			len(item.Path) == 0 || item.Path[len(item.Path)-1] != (PathNode{ID: id, Name: item.Name}) ||
			item.Depth != len(item.Path) || (item.Zh != "" && !validName(item.Zh)) {
			return snapshot{}, errors.New("invalid category in CBT dictionary snapshot")
		}
	}
	return value, nil
}

func loadOverrides(path string) (map[string]string, error) {
	var records []override
	err := readJSONFile(path, &records)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(records))
	for _, record := range records {
		if len(record.Path) == 0 || len(record.Path) > 32 || !validName(record.Zh) {
			return nil, errors.New("invalid CBT dictionary override")
		}
		for _, node := range record.Path {
			if !idPattern.MatchString(node.ID) || !validName(node.Name) {
				return nil, errors.New("invalid CBT dictionary override path")
			}
		}
		key := pathKey(record.Path)
		if _, found := values[key]; found {
			return nil, errors.New("duplicate CBT dictionary override path")
		}
		values[key] = record.Zh
	}
	return values, nil
}

func readJSONFile(path string, value any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxJSON+1))
	if err != nil || len(data) > maxJSON || !utf8.Valid(data) || json.Unmarshal(data, value) != nil {
		return errors.New("invalid CBT dictionary file")
	}
	return nil
}

func saveJSON(dir, name string, value any) error {
	f, err := os.CreateTemp(dir, ".cbt-dictionary-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	err = json.NewEncoder(f).Encode(value)
	if err == nil {
		info, statErr := f.Stat()
		switch {
		case statErr != nil:
			err = statErr
		case info.Size() > maxJSON:
			err = errors.New("CBT dictionary exceeds its read size limit")
		default:
			err = f.Sync()
		}
	}
	closeErr := f.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(f.Name(), filepath.Join(dir, name))
}

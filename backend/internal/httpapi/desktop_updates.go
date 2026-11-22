package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const desktopReleasesURL = "https://api.github.com/repos/Fan6512/una-seller/releases?per_page=50"
const desktopAssetPrefix = "https://github.com/Fan6512/una-seller/releases/download/"

var (
	officialTag = regexp.MustCompile(`^desktop-v([1-9][0-9]*)\.(0|[1-9][0-9]*)\.0$`)
	testTag = regexp.MustCompile(`^desktop-test-v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.0(?:-beta\.([1-9][0-9]*))?$`)
)

type desktopRelease struct {
	TagName    string `json:"tag_name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	Assets     []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type desktopManifest struct {
	Version   string `json:"version"`
	Notes     string `json:"notes,omitempty"`
	Platforms map[string]struct {
		URL       string `json:"url"`
		Signature string `json:"signature"`
	} `json:"platforms"`
}

// This is the normal SemVer ordering: a beta for the same major/minor is older
// than the eventual stable release. All publicly displayed versions are MAJOR.MINOR.
type desktopVersion struct {
	major, minor, betaIndex int
	beta                   bool
}

func parseDesktopRelease(release desktopRelease) (desktopVersion, string, bool) {
	if release.Draft {
		return desktopVersion{}, "", false
	}
	if !release.Prerelease {
		if m := officialTag.FindStringSubmatch(release.TagName); m != nil {
			major, e1 := strconv.Atoi(m[1])
			minor, e2 := strconv.Atoi(m[2])
			if e1 == nil && e2 == nil {
				return desktopVersion{major: major, minor: minor}, fmt.Sprintf("%d.%d.0", major, minor), true
			}
		}
		return desktopVersion{}, "", false
	}
	m := testTag.FindStringSubmatch(release.TagName)
	if m == nil {
		return desktopVersion{}, "", false
	}
	major, e1 := strconv.Atoi(m[1])
	minor, e2 := strconv.Atoi(m[2])
	if e1 != nil || e2 != nil || major == 0 && m[3] != "" || major >= 1 && m[3] == "" {
		return desktopVersion{}, "", false
	}
	if major == 0 {
		return desktopVersion{major: 0, minor: minor}, fmt.Sprintf("0.%d.0", minor), true
	}
	betaIndex, e3 := strconv.Atoi(m[3])
	if e3 != nil {
		return desktopVersion{}, "", false
	}
	return desktopVersion{major: major, minor: minor, beta: true, betaIndex: betaIndex},
		fmt.Sprintf("%d.%d.0-beta.%d", major, minor, betaIndex), true
}

func laterDesktopVersion(a, b desktopVersion) bool {
	if a.major != b.major {
		return a.major > b.major
	}
	if a.minor != b.minor {
		return a.minor > b.minor
	}
	if a.beta != b.beta {
		return !a.beta // stable wins at the same major.minor
	}
	return a.beta && a.betaIndex > b.betaIndex
}

func pickDesktopRelease(releases []desktopRelease, acceptTests bool) (desktopRelease, string, bool) {
	var best desktopRelease
	var bestVersion desktopVersion
	var bestName string
	found := false
	for _, item := range releases {
		if item.Prerelease && !acceptTests {
			continue
		}
		version, expected, ok := parseDesktopRelease(item)
		if !ok {
			continue
		}
		if !found || laterDesktopVersion(version, bestVersion) {
			best, bestVersion, bestName, found = item, version, expected, true
		}
	}
	return best, bestName, found
}

type desktopManifestCache struct {
	sync.Mutex
	entries map[bool]struct {
		content []byte
		until   time.Time
	}
}

func desktopUpdaterHandler(client *http.Client, releasesURL string) http.HandlerFunc {
	if client == nil {
		client = &http.Client{Timeout: 7 * time.Second}
	}
	cache := &desktopManifestCache{entries: make(map[bool]struct {
		content []byte
		until   time.Time
	})}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		includeTests := r.Header.Get("X-Una-Include-Prerelease") == "1"
		cache.Lock()
		entry, cached := cache.entries[includeTests]
		cache.Unlock()
		if cached && time.Now().Before(entry.until) {
			respondDesktopManifest(w, entry.content)
			return
		}
		data, err := latestDesktopManifest(r.Context(), client, releasesURL, includeTests)
		if err != nil {
			http.Error(w, "update information temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		cache.Lock()
		cache.entries[includeTests] = struct {
			content []byte
			until   time.Time
		}{content: data, until: time.Now().Add(2 * time.Minute)}
		cache.Unlock()
		respondDesktopManifest(w, data)
	}
}

func respondDesktopManifest(w http.ResponseWriter, data []byte) {
	if len(data) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_, _ = w.Write(data)
}

func readDesktopJSON(ctx context.Context, client *http.Client, uri string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "UnaSeller-Desktop-Updater")
	req.Header.Set("Accept", "application/vnd.github+json")
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("updater source status %d", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(out); err != nil {
		return err
	}
	return nil
}

func latestDesktopManifest(ctx context.Context, client *http.Client, releasesURL string, includeTests bool) ([]byte, error) {
	var releases []desktopRelease
	if err := readDesktopJSON(ctx, client, releasesURL, &releases); err != nil {
		return nil, err
	}
	release, expectedVersion, found := pickDesktopRelease(releases, includeTests)
	if !found {
		return nil, nil // No eligible published release yet
	}
	var manifestURL string
	for _, asset := range release.Assets {
		if asset.Name == "latest.json" {
			manifestURL = asset.BrowserDownloadURL
			break
		}
	}
	// No SSRF: only fetch the manifest from the matching official GitHub tag.
	if !strings.HasPrefix(manifestURL, desktopAssetPrefix+release.TagName+"/") ||
		!strings.HasSuffix(manifestURL, "/latest.json") {
		return nil, errors.New("signed release manifest not published")
	}
	var manifest desktopManifest
	if err := readDesktopJSON(ctx, client, manifestURL, &manifest); err != nil {
		return nil, err
	}
	windows, ok := manifest.Platforms["windows-x86_64"]
	if !ok || manifest.Version != expectedVersion ||
		!strings.HasPrefix(windows.URL, desktopAssetPrefix+release.TagName+"/") ||
		!strings.HasSuffix(windows.URL, "_x64-setup.exe") || strings.TrimSpace(windows.Signature) == "" {
		return nil, errors.New("release manifest does not match published tag and signed installer")
	}
	return json.Marshal(manifest)
}

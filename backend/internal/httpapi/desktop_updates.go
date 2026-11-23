package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

const desktopReleasesURL = "https://api.github.com/repos/Fan6512/una-seller/releases?per_page=50"
const desktopAssetPrefix = "https://github.com/Fan6512/una-seller/releases/download/"

var (
	officialTag = regexp.MustCompile(`^desktop-v([1-9][0-9]*)\.(0|[1-9][0-9]*)\.0$`)
	testTag     = regexp.MustCompile(`^desktop-test-v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.0(?:-beta\.([1-9][0-9]*))?$`)
)

type desktopReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type desktopRelease struct {
	TagName    string                `json:"tag_name"`
	Draft      bool                  `json:"draft"`
	Prerelease bool                  `json:"prerelease"`
	Assets     []desktopReleaseAsset `json:"assets"`
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
	beta                    bool
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

// The releases listing is shared by both channels. A single gate also merges
// simultaneous refreshes; this endpoint has low traffic and no extra worker.
const desktopCacheTTL = 10 * time.Minute
const desktopRetryDelay = time.Minute
const desktopUpstreamTimeout = 8 * time.Second
const desktopCheckTimeout = 17 * time.Second

type desktopCachedManifest struct {
	content []byte
	until   time.Time
}

type desktopManifestCache struct {
	refresh       chan struct{}
	releases      []desktopRelease
	releasesUntil time.Time
	entries       map[bool]desktopCachedManifest
	retryAfter    map[bool]time.Time
}

func (c *desktopManifestCache) lookup(ctx context.Context, client *http.Client, releasesURL string, includeTests bool, ttl time.Duration) ([]byte, error) {
	fetchCtx, cancel := context.WithTimeout(ctx, desktopCheckTimeout)
	defer cancel()
	select {
	case c.refresh <- struct{}{}:
		defer func() { <-c.refresh }()
	case <-fetchCtx.Done():
		return nil, fetchCtx.Err()
	}
	if err := fetchCtx.Err(); err != nil {
		return nil, err
	}

	now := time.Now()
	last, found := c.entries[includeTests]
	if found && now.Before(last.until) {
		return last.content, nil
	}
	// Only a previously validated signed release manifest is safe to serve stale.
	staleOrError := func(err error) ([]byte, error) {
		if found && len(last.content) > 0 {
			return last.content, nil
		}
		return nil, err
	}
	if now.Before(c.retryAfter[includeTests]) {
		return staleOrError(errors.New("updater upstream retry backoff"))
	}
	if !now.Before(c.releasesUntil) {
		var releases []desktopRelease
		if err := readDesktopJSON(fetchCtx, client, releasesURL, &releases); err != nil {
			// A canceled client request must not poison the shared retry window.
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			c.retryAfter[includeTests] = time.Now().Add(desktopRetryDelay)
			return staleOrError(err)
		}
		c.releases = releases
		c.releasesUntil = time.Now().Add(ttl)
	}

	data, err := desktopManifestForReleases(fetchCtx, client, c.releases, includeTests)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		c.retryAfter[includeTests] = time.Now().Add(desktopRetryDelay)
		return staleOrError(err)
	}
	delete(c.retryAfter, includeTests)
	c.entries[includeTests] = desktopCachedManifest{content: data, until: time.Now().Add(ttl)}
	return data, nil
}

func desktopUpdaterHandler(client *http.Client, releasesURL string) http.HandlerFunc {
	return desktopUpdaterHandlerWithTTL(client, releasesURL, desktopCacheTTL)
}

func desktopUpdaterHandlerWithTTL(client *http.Client, releasesURL string, ttl time.Duration) http.HandlerFunc {
	if client == nil {
		client = &http.Client{Timeout: desktopUpstreamTimeout}
	}
	cache := &desktopManifestCache{
		refresh:    make(chan struct{}, 1),
		entries:    make(map[bool]desktopCachedManifest),
		retryAfter: make(map[bool]time.Time),
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		includeTests := r.Header.Get("X-Una-Include-Prerelease") == "1"
		data, err := cache.lookup(r.Context(), client, releasesURL, includeTests, ttl)
		if err != nil {
			http.Error(w, "update information temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
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

type desktopSourceStatus int

func (status desktopSourceStatus) Error() string {
	return fmt.Sprintf("updater source status %d", status)
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
		return desktopSourceStatus(response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(out); err != nil {
		return err
	}
	return nil
}

func desktopManifestForReleases(ctx context.Context, client *http.Client, releases []desktopRelease, includeTests bool) ([]byte, error) {
	candidates := slices.Clone(releases)
	for {
		release, expectedVersion, found := pickDesktopRelease(candidates, includeTests)
		if !found {
			return nil, nil
		}
		manifest, err := readDesktopManifest(ctx, client, release, expectedVersion)
		if !errors.Is(err, errDesktopManifestInvalid) {
			return manifest, err
		}
		candidates = slices.DeleteFunc(candidates, func(candidate desktopRelease) bool {
			return candidate.TagName == release.TagName
		})
	}
}

var errDesktopManifestInvalid = errors.New("release does not have a valid signed installer manifest")

func readDesktopManifest(ctx context.Context, client *http.Client, release desktopRelease, expectedVersion string) ([]byte, error) {
	prefix := desktopAssetPrefix + release.TagName + "/"
	var manifestURL string
	for _, asset := range release.Assets {
		if asset.Name == "latest.json" && asset.BrowserDownloadURL == prefix+asset.Name {
			manifestURL = asset.BrowserDownloadURL
			break
		}
	}
	if manifestURL == "" {
		return nil, errDesktopManifestInvalid
	}
	var manifest desktopManifest
	if err := readDesktopJSON(ctx, client, manifestURL, &manifest); err != nil {
		var sourceStatus desktopSourceStatus
		var syntax *json.SyntaxError
		var badType *json.UnmarshalTypeError
		if (errors.As(err, &sourceStatus) && (sourceStatus == http.StatusNotFound || sourceStatus == http.StatusGone)) ||
			errors.As(err, &syntax) || errors.As(err, &badType) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, errDesktopManifestInvalid
		}
		return nil, err
	}
	windows, ok := manifest.Platforms["windows-x86_64"]
	if !ok || manifest.Version != expectedVersion || strings.TrimSpace(windows.Signature) == "" ||
		!slices.ContainsFunc(release.Assets, func(asset desktopReleaseAsset) bool {
			return strings.HasSuffix(asset.Name, "_x64-setup.exe") && windows.URL == prefix+asset.Name &&
				asset.BrowserDownloadURL == windows.URL
		}) {
		return nil, errDesktopManifestInvalid
	}
	return json.Marshal(manifest)
}

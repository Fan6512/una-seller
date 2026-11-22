package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDesktopUpdateReleaseEligibility(t *testing.T) {
	releases := []desktopRelease{
		{TagName: "desktop-v0.1.1"}, // pre-1.0 legacy release is not an approved stable channel
		{TagName: "desktop-test-v0.3.0", Prerelease: true},
		{TagName: "desktop-test-v0.4.0", Prerelease: true, Draft: true},
		{TagName: "desktop-v1.0.0", Draft: true},
	}
	if _, _, ok := pickDesktopRelease(releases, false); ok {
		t.Fatal("stable-only should not expose drafts, legacy tags or test releases")
	}
	selected, version, ok := pickDesktopRelease(releases, true)
	if !ok || selected.TagName != "desktop-test-v0.3.0" || version != "0.3.0" {
		t.Fatalf("unexpected opted-in update: %+v %q %v", selected, version, ok)
	}
	releases[3].Draft = false
	selected, version, ok = pickDesktopRelease(releases, false)
	if !ok || selected.TagName != "desktop-v1.0.0" || version != "1.0.0" {
		t.Fatalf("stable release not selected: %+v %q", selected, version)
	}
	selected, version, ok = pickDesktopRelease(releases, true)
	if !ok || selected.TagName != "desktop-v1.0.0" || version != "1.0.0" {
		t.Fatalf("published stable should win over older test: %+v %q", selected, version)
	}
}

func TestDesktopUpdateLaterBetaAfterStable(t *testing.T) {
	releases := []desktopRelease{
		{TagName: "desktop-v1.0.0"},
		{TagName: "desktop-test-v1.1.0-beta.1", Prerelease: true},
		{TagName: "desktop-test-v1.1.0-beta.2", Prerelease: true},
		{TagName: "desktop-v1.1.0"},
	}
	selected, version, ok := pickDesktopRelease(releases[:3], true)
	if !ok || selected.TagName != "desktop-test-v1.1.0-beta.2" || version != "1.1.0-beta.2" {
		t.Fatalf("beta selection wrong: %+v %q", selected, version)
	}
	selected, version, ok = pickDesktopRelease(releases, true)
	if !ok || selected.TagName != "desktop-v1.1.0" || version != "1.1.0" {
		t.Fatalf("stable release must supersede beta of same version: %+v %q", selected, version)
	}
}

func TestDesktopUpdateHttpGateAndNoRelease(t *testing.T) {
	gh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]desktopRelease{
			{TagName: "desktop-test-v0.3.0", Prerelease: true, Draft: true},
		})
	}))
	defer gh.Close()
	handler := desktopUpdaterHandler(gh.Client(), gh.URL)
	post := httptest.NewRequest(http.MethodPost, "/api/v1/desktop/updates", nil)
	w := httptest.NewRecorder()
	handler(w, post)
	if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != "GET" {
		t.Fatalf("POST must be rejected: %d", w.Code)
	}
	for _, include := range []string{"", "1"} {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/desktop/updates", nil)
		if include != "" {
			request.Header.Set("X-Una-Include-Prerelease", include)
		}
		rec := httptest.NewRecorder()
		handler(rec, request)
		if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 ||
			rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("no release must be a 204 without a body: status=%d body=%q", rec.Code, rec.Body.String())
		}
	}
}

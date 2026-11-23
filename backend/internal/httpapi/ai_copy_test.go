package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/migrations"
)

const aiCopyFixture = `{"titles":{"en":"Glass cup","es":"Taza de vidrio","pt":"Copo de vidro"},"descriptions":{"en":"A glass cup.\n\nFor use at home.","es":"Una taza de vidrio.","pt":"Um copo de vidro."},"selling_points":{"en":["Made of glass"],"es":["De vidrio"],"pt":["De vidro"]},"missing_facts":["Capacity", "Dimensions"]}`

func mockAICompletion(w http.ResponseWriter, content string) {
	_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{
		map[string]any{"finish_reason": "stop", "message": map[string]string{"content": content}},
	}})
}

func TestAICopyValidatesStructuredOutput(t *testing.T) {
	nfc := strings.Repeat("e\u0301", 59) + "!"
	content := strings.Replace(aiCopyFixture, "Taza de vidrio", nfc, 1)
	candidate, err := parseAICopy(content)
	if err != nil || utf8.RuneCountInString(candidate.Titles.Spanish) != 60 ||
		candidate.Titles.Spanish != strings.Repeat("é", 59)+"!" ||
		candidate.Descriptions.English != "A glass cup.\n\nFor use at home." ||
		len(candidate.SellingPoints.Portuguese) != 1 || len(candidate.MissingFacts) != 2 {
		t.Fatalf("unexpected candidate: %+v, %v", candidate, err)
	}
	for name, input := range map[string]string{
		"trailing JSON":     aiCopyFixture + "{}",
		"missing facts":     strings.Replace(aiCopyFixture, `,"missing_facts":["Capacity", "Dimensions"]`, "", 1),
		"null facts":        strings.Replace(aiCopyFixture, `["Capacity", "Dimensions"]`, "null", 1),
		"blank fact":        strings.Replace(aiCopyFixture, "Capacity", "  ", 1),
		"repeated fact":     strings.Replace(aiCopyFixture, "Dimensions", "capacity", 1),
		"nonstring fact":    strings.Replace(aiCopyFixture, `"Capacity"`, `{"name":"capacity"}`, 1),
		"invalid surrogate": strings.Replace(aiCopyFixture, "Capacity", `\uD800`, 1),
		"HTML":              strings.Replace(aiCopyFixture, "Una taza de vidrio.", "<b>Una taza</b>", 1),
		"Markdown":          strings.Replace(aiCopyFixture, "Una taza de vidrio.", "**Una taza**", 1),
		"external link":     strings.Replace(aiCopyFixture, "A glass cup.", "Buy at store.example.com", 1),
		"email address":     strings.Replace(aiCopyFixture, "A glass cup.", "seller@example.com", 1),
		"phone number":      strings.Replace(aiCopyFixture, "A glass cup.", "Call +1 (555) 555-0100", 1),
		"promotion":         strings.Replace(aiCopyFixture, "A glass cup.", "Free shipping", 1),
		"duplicate key":     strings.Replace(aiCopyFixture, `"es":"Taza`, `"es":"Uno","es":"Taza`, 1),
		"wrong casing":      strings.Replace(aiCopyFixture, `"titles"`, `"Titles"`, 1),
		"unexpected field":  strings.Replace(aiCopyFixture, `"titles":`, `"additional":true,"titles":`, 1),
		"missing locale":    strings.Replace(aiCopyFixture, `,"pt":"Copo de vidro"`, "", 1),
		"no selling points": strings.Replace(aiCopyFixture, `["Made of glass"]`, "[]", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseAICopy(input); !errors.Is(err, errAICopyResponse) {
				t.Fatalf("invalid output accepted: %v", err)
			}
		})
	}
	emptyFacts := strings.Replace(aiCopyFixture, `["Capacity", "Dimensions"]`, "[]", 1)
	if result, err := parseAICopy(emptyFacts); err != nil || result.MissingFacts == nil {
		t.Fatalf("explicit empty missing facts rejected: %v", err)
	}
}

func TestAICopyProviderResponseBoundsAndRedirect(t *testing.T) {
	var redirects atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirects.Add(1)
	}))
	defer target.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer large":
			_, _ = io.WriteString(w, strings.Repeat(" ", maxAICopyResponse+1))
		case "Bearer redirect":
			http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
		case "Bearer truncated":
			_, _ = io.WriteString(w, `{"choices":[{"finish_reason":"length","message":{"content":"{}"}}]}`)
		}
	}))
	defer upstream.Close()
	for _, scenario := range []string{"large", "redirect", "truncated"} {
		service, err := NewAICopyService(upstream.URL, scenario, "mock")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.generate(t.Context(), aiCopyFacts{}); !errors.Is(err, errAICopyResponse) {
			t.Fatalf("%s response accepted: %v", scenario, err)
		}
	}
	if redirects.Load() != 0 {
		t.Fatal("followed upstream redirect")
	}
	for _, base := range []string{"http://external.example/v1", "https://user:key@host/v1", "https://host/v1?secret=value"} {
		if service, err := NewAICopyService(base, "synthetic", "mock"); service != nil || err == nil {
			t.Fatal("invalid provider configuration accepted")
		}
	}
}

func TestAICopyOwnedProductGenerationAndSaveReadback(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 25*time.Second)
	defer cancel()
	store, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.DB.Close()
	if err = migrations.Apply(ctx, store.DB); err != nil {
		t.Fatal(err)
	}
	const owner = "2a9c0000-0000-4000-8000-000000000085"
	const other = "2a9c0000-0000-4000-8000-000000000086"
	defer func() {
		cleanup := context.Background()
		_, _ = store.DB.Exec(cleanup, "DELETE FROM app.own_products WHERE owner_user_id=$1::uuid", owner)
		_, _ = store.DB.Exec(cleanup, "DELETE FROM app.source_products WHERE owner_user_id=$1::uuid", owner)
		_, _ = store.DB.Exec(cleanup, "DELETE FROM app.users WHERE id=$1::uuid", owner)
	}()
	source, _, err := store.Save(ctx, owner, collections.Input{
		Platform: "amazon", Site: "US", SourceURL: "https://www.amazon.com/dp/B000000085", Status: "complete",
		Title: collections.Field[string]{Present: true, Value: "Original cup"},
	})
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	started, cancelled := make(chan struct{}), make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/v1/chat/completions" || r.Method != http.MethodPost ||
			r.Header.Get("Authorization") != "Bearer synthetic-provider-key" {
			t.Error("invalid AI upstream request")
		}
		var request struct {
			Model          string                `json:"model"`
			ResponseFormat struct{ Type string } `json:"response_format"`
			Messages       []struct{ Role, Content string }
		}
		if json.NewDecoder(r.Body).Decode(&request) != nil || request.Model != "mock" ||
			request.ResponseFormat.Type != "json_object" || len(request.Messages) != 2 ||
			request.Messages[1].Role != "user" {
			t.Error("invalid AI request shape")
			w.WriteHeader(400)
			return
		}
		facts := request.Messages[1].Content
		if strings.Contains(facts, source.ID) || strings.Contains(facts, "PRIVATE") || strings.Contains(facts, owner) ||
			!strings.Contains(facts, "glass") || !strings.Contains(request.Messages[0].Content, "Do not invent") {
			t.Error("AI received non-allowlisted data or omitted factual inputs")
		}
		switch {
		case strings.Contains(facts, "Cancel"):
			close(started)
			<-r.Context().Done()
			close(cancelled)
		case strings.Contains(facts, "Timeout"):
			<-r.Context().Done()
		case strings.Contains(facts, "Unavailable"):
			http.Error(w, "synthetic-provider-key PRIVATE error", http.StatusTooManyRequests)
		default:
			mockAICompletion(w, strings.Replace(aiCopyFixture, "Taza de vidrio", strings.Repeat("é", 61), 1))
		}
	}))
	defer upstream.Close()
	service, err := NewAICopyService(upstream.URL+"/v1", "synthetic-provider-key", "mock")
	if err != nil {
		t.Fatal(err)
	}
	handler := HandlerWithServices(store, stubVerifier{id: owner}, nil, nil, service)
	otherHandler := HandlerWithServices(store, stubVerifier{id: other}, nil, nil, service)
	call := func(ctx context.Context, h http.Handler, method, path, jsonBody, origin, access string) *httptest.ResponseRecorder {
		r := httptest.NewRequestWithContext(ctx, method, path, strings.NewReader(jsonBody))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Origin", origin)
		r.Header.Set("Authorization", access)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	const origin = "http://tauri.localhost"
	const access = "Bearer synthetic-session"
	const genPath = "/api/v1/products/ai-copy"
	draftPath := "/api/v1/collections/" + source.ID + "/working-draft"
	facts := `"facts":{"titles":{"en":"Glass cup","es":"","pt":""},"attributes":[{"name":"Material","value":"glass"}]}`
	input := `{"source_product_id":"` + source.ID + `",` + facts + `}`
	for _, test := range []struct {
		handler                      http.Handler
		method, origin, access, body string
		want                         int
	}{
		{handler, "POST", origin, "", input, 401},
		{otherHandler, "POST", origin, access, input, 404},
		{handler, "POST", "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", access, input, 403},
		{handler, "OPTIONS", origin, "", "", 204},
		{handler, "GET", origin, access, "", 405},
		{handler, "POST", origin, access, strings.Replace(input, `"facts":`, `"notes":"PRIVATE", "facts":`, 1), 400},
		{handler, "POST", origin, access, strings.Replace(input, "Glass cup", "Glass cup shop@example.com", 1), 400},
		{handler, "POST", origin, access, strings.Replace(input, "Glass cup", "Glass cup Bearer synthetic-access-token", 1), 400},
	} {
		if w := call(ctx, test.handler, test.method, genPath, test.body, test.origin, test.access); w.Code != test.want {
			t.Fatalf("%s status=%d expected=%d", test.method, w.Code, test.want)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("unauthorized or invalid requests reached AI")
	}
	unconfigured := HandlerWithCollections(store, stubVerifier{id: owner})
	if w := call(ctx, unconfigured, "POST", genPath, input, origin, access); w.Code != 503 {
		t.Fatalf("unconfigured generation status=%d", w.Code)
	}
	generation := call(ctx, handler, "POST", genPath, input, origin, access)
	var candidate aiCopyCandidate
	if generation.Code != 200 || json.Unmarshal(generation.Body.Bytes(), &candidate) != nil ||
		utf8.RuneCountInString(candidate.Titles.Spanish) != 61 {
		t.Fatal("valid mock response or over-limit candidate missing")
	}
	if draft, err := store.GetWorkingDraft(ctx, owner, source.ID); err != nil || draft != nil {
		t.Fatalf("generate created draft fields: %v", err)
	}

	legacy := strings.Repeat("🌿", 20000)
	draftInput := collections.WorkingDraftInput{Title: "Glass cup", Description: legacy, Notes: "PRIVATE",
		ListingFields: &collections.ListingFields{Titles: candidate.Titles}}
	saveDraft := func(want int) {
		t.Helper()
		body, _ := json.Marshal(draftInput)
		if w := call(ctx, handler, "PUT", draftPath, string(body), origin, access); w.Code != want {
			t.Fatalf("draft save status=%d expected=%d", w.Code, want)
		}
	}
	saveDraft(400)
	draftInput.ListingFields.Titles.Spanish = strings.Repeat("e\u0301", 59) + "!"
	eco := strings.Repeat("🫗", 20000)
	large := collections.LocalizedDescriptions{English: eco, Spanish: eco, Portuguese: eco}
	draftInput.ListingFields.Descriptions = &large
	saveDraft(200)
	draft, err := store.GetWorkingDraft(ctx, owner, source.ID)
	if err != nil || draft == nil || draft.ListingFields.Descriptions == nil ||
		*draft.ListingFields.Descriptions != large || draft.Description != legacy ||
		draft.ListingFields.Titles.Spanish != strings.Repeat("é", 59)+"!" {
		t.Fatalf("adopted draft did not round-trip: %v", err)
	}
	// An older desktop saves listing_fields without the new optional key.
	draftInput.ExpectedRevision = draft.Revision
	draftInput.ListingFields.Descriptions = nil
	saveDraft(200)
	own, _, err := store.Promote(ctx, owner, source.ID, draft.Revision+1)
	if err != nil || own.ListingFields.Descriptions == nil || *own.ListingFields.Descriptions != large || own.Description != legacy {
		t.Fatalf("promotion dropped adopted or legacy description: %v", err)
	}
	ownInput := `{"own_product_id":"` + own.ID + `",` + facts + `}`
	if w := call(ctx, otherHandler, "POST", genPath, ownInput, origin, access); w.Code != 404 {
		t.Fatalf("foreign own product generation status=%d", w.Code)
	}
	if w := call(ctx, handler, "POST", genPath, ownInput, origin, access); w.Code != 200 {
		t.Fatalf("own product generation status=%d", w.Code)
	}
	ownPath := "/api/v1/products/" + own.ID
	if after, err := store.GetOwnProduct(ctx, owner, own.ID); err != nil || after.Revision != own.Revision ||
		after.ListingFields.Descriptions == nil || *after.ListingFields.Descriptions != large ||
		after.ListingFields.Titles != own.ListingFields.Titles {
		t.Fatalf("generation altered formal own product fields: %v", err)
	}
	edit := collections.OwnProductEdit{Title: own.Title, Description: own.Description, Notes: own.Notes,
		ListingFields: &own.ListingFields, ExpectedRevision: own.Revision}
	saveOwn := func(want int) {
		t.Helper()
		body, _ := json.Marshal(edit)
		if w := call(ctx, handler, "PUT", ownPath, string(body), origin, access); w.Code != want {
			t.Fatalf("own product save status=%d expected=%d", w.Code, want)
		}
	}
	edit.ListingFields.Titles.Portuguese = strings.Repeat("🫗", 61)
	saveOwn(400)
	edit.ListingFields.Titles.Portuguese = "Copo de vidro"
	large.Spanish = "Una taza de vidrio.\n\nPara uso en casa."
	edit.ListingFields.Descriptions = &large
	saveOwn(200)
	edit.ExpectedRevision++
	edit.ListingFields.Descriptions = nil
	saveOwn(200)
	saved, err := store.GetOwnProduct(ctx, owner, own.ID)
	if err != nil || saved.ListingFields.Descriptions == nil ||
		*saved.ListingFields.Descriptions != large || saved.Description != legacy {
		t.Fatalf("own product descriptions did not round-trip: %v", err)
	}
	// Simulate a saved title from a version that allowed regional titles up to 500 characters.
	legacyTitle := strings.Repeat("X", 61)
	if _, err := store.DB.Exec(ctx, `UPDATE app.own_products
		SET listing_fields=jsonb_set(listing_fields,'{titles,es}',to_jsonb($1::text)) WHERE id=$2::uuid`, legacyTitle, own.ID); err != nil {
		t.Fatal(err)
	}
	edit.ExpectedRevision = saved.Revision
	edit.ListingFields.Titles.Spanish = legacyTitle
	saveOwn(200)
	edit.ExpectedRevision++
	edit.ListingFields.Titles.Spanish = legacyTitle + "!"
	saveOwn(400)
	if w := call(ctx, handler, "POST", genPath, strings.Replace(input, "Glass cup", "Unavailable glass cup", 1), origin, access); w.Code != 502 || strings.Contains(w.Body.String(), "synthetic-provider-key") || strings.Contains(w.Body.String(), "PRIVATE") {
		t.Fatalf("upstream error response not safely mapped: %d", w.Code)
	}

	reqCtx, abort := context.WithCancel(ctx)
	defer abort()
	done := make(chan struct{})
	go func() {
		defer close(done)
		call(reqCtx, handler, "POST", genPath, strings.Replace(input, "Glass cup", "Cancel glass cup", 1), origin, access)
	}()
	select {
	case <-started:
		abort()
	case <-time.After(2 * time.Second):
		t.Fatal("AI cancel fixture never started")
	}
	select {
	case <-cancelled:
	case <-time.After(2 * time.Second):
		t.Fatal("request cancellation did not reach provider")
	}
	<-done
	service.client.Timeout = 80 * time.Millisecond
	if w := call(ctx, handler, "POST", genPath, strings.Replace(input, "Glass cup", "Timeout glass cup", 1), origin, access); w.Code != 504 {
		t.Fatalf("upstream timeout status=%d", w.Code)
	}
}

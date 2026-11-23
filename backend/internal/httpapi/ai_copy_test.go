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

const aiTitleFixture = `{"titles":{"en":"Glass cup","es":"Taza de vidrio","pt":"Copo de vidro"}}`
const aiEnglishFixture = `{"description":"A glass cup.\n\nFor use at home.","selling_points":["Made of glass"],"missing_facts":["Capacity", "Dimensions"]}`
const aiSavedEnglish = "Glass cup, model A12.\n\nCapacity 250 mL."
const aiTranslateFixture = `{"es":"Taza de vidrio, modelo A12.\n\nCapacidad de 250 mL.","pt":"Copo de vidro, modelo A12.\n\nCapacidade de 250 mL.","review_terms":["A12"]}`

func mockAICompletion(w http.ResponseWriter, content string) {
	_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{
		map[string]any{"finish_reason": "stop", "message": map[string]string{"content": content}},
	}})
}

func TestAICopyValidatesStructuredOutput(t *testing.T) {
	nfc := strings.Repeat("e\u0301", 59) + "!"
	content := strings.Replace(aiTitleFixture, "Taza de vidrio", nfc, 1)
	candidate, ok := parseAICopyTitles(content)
	if !ok || utf8.RuneCountInString(candidate.Titles.Spanish) != 60 ||
		candidate.Titles.Spanish != strings.Repeat("é", 59)+"!" {
		t.Fatalf("unexpected title candidate: %+v", candidate)
	}
	for name, input := range map[string]string{
		"trailing JSON":    aiTitleFixture + "{}",
		"duplicate key":    strings.Replace(aiTitleFixture, `"es":"Taza`, `"es":"Uno","es":"Taza`, 1),
		"wrong casing":     strings.Replace(aiTitleFixture, `"titles"`, `"Titles"`, 1),
		"unexpected field": strings.Replace(aiTitleFixture, `"titles":`, `"additional":true,"titles":`, 1),
		"missing locale":   strings.Replace(aiTitleFixture, `,"pt":"Copo de vidro"`, "", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseAICopyTitles(input); ok {
				t.Fatal("invalid title accepted")
			}
		})
	}
	for name, input := range map[string]string{
		"missing facts":     strings.Replace(aiEnglishFixture, `,"missing_facts":["Capacity", "Dimensions"]`, "", 1),
		"null facts":        strings.Replace(aiEnglishFixture, `["Capacity", "Dimensions"]`, "null", 1),
		"blank fact":        strings.Replace(aiEnglishFixture, "Capacity", "  ", 1),
		"repeated fact":     strings.Replace(aiEnglishFixture, "Dimensions", "capacity", 1),
		"nonstring fact":    strings.Replace(aiEnglishFixture, `"Capacity"`, `{"name":"capacity"}`, 1),
		"invalid surrogate": strings.Replace(aiEnglishFixture, "Capacity", `\uD800`, 1),
		"HTML":              strings.Replace(aiEnglishFixture, "A glass cup.", "<b>A cup</b>", 1),
		"Markdown":          strings.Replace(aiEnglishFixture, "A glass cup.", "**A cup**", 1),
		"external link":     strings.Replace(aiEnglishFixture, "A glass cup.", "Buy at store.example.com", 1),
		"email address":     strings.Replace(aiEnglishFixture, "A glass cup.", "seller@example.com", 1),
		"phone number":      strings.Replace(aiEnglishFixture, "A glass cup.", "Call +1 (555) 555-0100", 1),
		"promotion":         strings.Replace(aiEnglishFixture, "A glass cup.", "Free shipping", 1),
		"no selling points": strings.Replace(aiEnglishFixture, `["Made of glass"]`, "[]", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := parseAICopyEnglish(input); ok {
				t.Fatal("invalid English candidate accepted")
			}
		})
	}
	emptyFacts := strings.Replace(aiEnglishFixture, `["Capacity", "Dimensions"]`, "[]", 1)
	if result, ok := parseAICopyEnglish(emptyFacts); !ok || result.MissingFacts == nil ||
		result.Description != "A glass cup.\n\nFor use at home." || len(result.SellingPoints) != 1 {
		t.Fatal("valid English-only candidate rejected")
	}
	if result, ok := parseAICopyTranslation(aiTranslateFixture, aiSavedEnglish); !ok || len(result.ReviewTerms) != 1 {
		t.Fatal("faithful translation with an uncertain model term rejected")
	}
	for _, input := range []string{
		strings.Replace(aiTranslateFixture, "250 mL", "500 mL", 1),
		strings.Replace(aiTranslateFixture, "A12", "A10", 1),
		strings.Replace(aiTranslateFixture, "A12", "B12", 1),
		strings.Replace(aiTranslateFixture, "250 mL", "250 cm", 1),
		strings.Replace(aiTranslateFixture, `\n\n`, " ", 1),
		strings.Replace(aiTranslateFixture, `,"review_terms":["A12"]`, "", 1),
	} {
		if _, ok := parseAICopyTranslation(input, aiSavedEnglish); ok {
			t.Fatal("translation with changed numbers, paragraphs or missing fields accepted")
		}
	}
}

func TestAICopyLocaleNumbers(t *testing.T) {
	if matched, notes := compareAICopyNumbers("Capacity 5.5 cm", "Capacidad 5,5 cm"); !matched || len(notes) != 0 {
		t.Fatalf("decimal comma rejected or flagged: %v %v", matched, notes)
	}
	if matched, _ := compareAICopyNumbers("Capacity 250 g", "Capacidad 500 g"); matched {
		t.Fatal("changed quantity accepted")
	}
	if matched, notes := compareAICopyNumbers("Pack 1,000 pieces", "Paquete 1.000 piezas"); !matched || len(notes) != 1 {
		t.Fatalf("ambiguous grouping not flagged: %v %v", matched, notes)
	}
	source := "Capacity 5.5 cm.\\n\\nPack 1,000 pieces."
	body := `{"es":"Capacidad 5,5 cm.\\n\\nPaquete 1.000 piezas.","pt":"Capacidade 5,5 cm.\\n\\nPacote 1.000 peças.","review_terms":[]}`
	result, ok := parseAICopyTranslation(body, source)
	if !ok || len(result.ReviewTerms) != 2 {
		t.Fatalf("localized formatting should require review instead of 502: %+v %t", result, ok)
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
		if _, err := service.complete(t.Context(), aiCopyTitlesPrompt, aiCopyFacts{}); !errors.Is(err, errAICopyResponse) {
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

func TestAICopyConcurrencyAndCooldown(t *testing.T) {
	service, err := NewAICopyService("https://provider.example/v1", "test-key", "mock")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	first, ok := service.acquire("owner-1", now)
	if !ok {
		t.Fatal("first request rejected")
	}
	if _, ok := service.acquire("owner-1", now.Add(2*aiCopyInterval)); ok {
		t.Fatal("same owner entered while a request was still active")
	}
	for _, owner := range []string{"owner-2", "owner-3"} {
		release, ok := service.acquire(owner, now)
		if !ok {
			t.Fatal("global cap rejected an available slot")
		}
		defer release()
	}
	if _, ok := service.acquire("owner-4", now); ok {
		t.Fatal("global cap was exceeded")
	}
	first()
	if _, ok := service.acquire("owner-1", now.Add(aiCopyInterval-time.Millisecond)); ok {
		t.Fatal("same owner bypassed the minimum request interval")
	}
	if next, ok := service.acquire("owner-1", now.Add(aiCopyInterval)); !ok {
		t.Fatal("release or cooldown did not allow retry")
	} else {
		next()
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
	fields := collections.ListingFields{
		Titles: collections.LocalizedTitles{English: "Saved glass cup"},
		Brand:  "Original Brand", Model: "A12", SellerSKU: "CUP-A12", CategoryID: "CBT1",
		WeightG: "120", LengthCM: "5.5", BulletPoints: []string{"Glass body", "note@example.com"},
		Attributes: []collections.ListingAttribute{{Name: "Material", Value: "glass"}},
		CategoryAttributes: &[]collections.ListingCategoryAttribute{
			{CategoryID: "CBT1", ID: "CAPACITY", Value: "250", UnitID: "mL"},
			{CategoryID: "CBT2", ID: "HISTORIC", Value: "HISTORICAL_DATA"},
		},
		Variants: &[]collections.ListingVariant{{
			SellerSKU: "CUP-A12-B", Attributes: map[string]string{"Color": "Blue"},
			Price: "19.99", Currency: "CNY", ImageURLs: []string{"https://example.com/private-image"},
		}},
	}
	createdDraft, err := store.SaveWorkingDraft(ctx, owner, source.ID, collections.WorkingDraftInput{
		Title: "Saved glass cup", Notes: "PRIVATE notes", ListingFields: &fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var providerMode atomic.Int32
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
		switch providerMode.Load() {
		case 1:
			close(started)
			<-r.Context().Done()
			close(cancelled)
			return
		case 2:
			<-r.Context().Done()
			return
		case 3:
			http.Error(w, "synthetic-provider-key PRIVATE error", http.StatusTooManyRequests)
			return
		}
		facts := request.Messages[1].Content
		prompt := request.Messages[0].Content
		if prompt == aiCopyTranslationPrompt {
			var exact map[string]string
			if json.Unmarshal([]byte(facts), &exact) != nil || len(exact) != 1 || exact["english_description"] != aiSavedEnglish {
				t.Error("translation did not use the exact saved English text as its only input")
			}
			mockAICompletion(w, aiTranslateFixture)
			return
		}
		for _, banned := range []string{source.ID, owner, "PRIVATE", "HISTORICAL_DATA", "note@", "https://", "19.99", "CNY", "CBT2"} {
			if strings.Contains(facts, banned) {
				t.Error("AI received non-allowlisted data")
			}
		}
		for _, required := range []string{"Original cup", "Saved glass cup", "Original Brand", "A12", "CBT1", "CAPACITY", "250", "mL", "120", "5.5", "Glass body", "Material", "glass", "CUP-A12-B", "Blue"} {
			if !strings.Contains(facts, required) {
				t.Errorf("AI omitted saved fact %q", required)
			}
		}
		switch prompt {
		case aiCopyTitlesPrompt:
			mockAICompletion(w, strings.Replace(aiTitleFixture, "Taza de vidrio", strings.Repeat("é", 61), 1))
		case aiCopyEnglishPrompt:
			mockAICompletion(w, aiEnglishFixture)
		default:
			t.Error("unknown AI prompt")
		}
	}))
	defer upstream.Close()
	service, err := NewAICopyService(upstream.URL+"/v1", "synthetic-provider-key", "mock")
	if err != nil {
		t.Fatal(err)
	}
	allowNext := func() {
		t.Helper()
		service.mu.Lock()
		defer service.mu.Unlock()
		if service.active != 0 || service.owners[owner].active {
			t.Fatal("completed AI request retained an active slot")
		}
		state := service.owners[owner]
		state.next = time.Now().Add(-aiCopyInterval)
		service.owners[owner] = state
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
	input := `{"source_product_id":"` + source.ID + `","operation":"titles","expected_revision":1}`
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
		{handler, "POST", origin, access, strings.Replace(input, `"operation":`, `"facts":{"title":"forged"}, "operation":`, 1), 400},
		{handler, "POST", origin, access, strings.Replace(input, `"operation":`, `"english_description":"forged", "operation":`, 1), 400},
		{handler, "POST", origin, access, strings.Replace(input, `"expected_revision":1`, `"expected_revision":0`, 1), 409},
		{handler, "POST", origin, access, strings.Replace(input, "titles", "translate", 1), 422},
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
	var candidate aiCopyTitles
	if generation.Code != 200 || json.Unmarshal(generation.Body.Bytes(), &candidate) != nil ||
		utf8.RuneCountInString(candidate.Titles.Spanish) != 61 {
		t.Fatal("valid mock response or over-limit candidate missing")
	}
	if w := call(ctx, handler, "POST", genPath, strings.Replace(input, "titles", "english", 1), origin, access); w.Code != 429 {
		t.Fatalf("owner cooldown status=%d", w.Code)
	}
	allowNext()
	english := call(ctx, handler, "POST", genPath, strings.Replace(input, "titles", "english", 1), origin, access)
	var masterCandidate aiCopyEnglish
	if english.Code != 200 || json.Unmarshal(english.Body.Bytes(), &masterCandidate) != nil ||
		len(masterCandidate.MissingFacts) != 2 || strings.Contains(english.Body.String(), `"descriptions"`) {
		t.Fatal("English-only mock response missing")
	}
	if draft, err := store.GetWorkingDraft(ctx, owner, source.ID); err != nil || draft == nil ||
		draft.Revision != createdDraft.Revision || draft.ListingFields.Descriptions != nil {
		t.Fatalf("generate created or replaced official draft fields: %v", err)
	}

	legacy := strings.Repeat("🌿", 20000)
	draftInput := collections.WorkingDraftInput{Title: "Glass cup", Description: legacy, Notes: "PRIVATE",
		ListingFields: &fields, ExpectedRevision: createdDraft.Revision}
	draftInput.ListingFields.Titles.Spanish = candidate.Titles.Spanish
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
	ownInput := `{"own_product_id":"` + own.ID + `","operation":"titles","expected_revision":1}`
	if w := call(ctx, otherHandler, "POST", genPath, ownInput, origin, access); w.Code != 404 {
		t.Fatalf("foreign own product generation status=%d", w.Code)
	}
	allowNext()
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
	large.English = aiSavedEnglish
	edit.ListingFields.Descriptions = &large
	saveOwn(200)
	edit.ExpectedRevision++
	translationInput := strings.Replace(strings.Replace(ownInput, "titles", "translate", 1), `"expected_revision":1`, `"expected_revision":2`, 1)
	allowNext()
	translation := call(ctx, handler, "POST", genPath, translationInput, origin, access)
	var translated struct {
		aiCopyTranslation
		SourceDigest string `json:"source_digest"`
	}
	if translation.Code != 200 || json.Unmarshal(translation.Body.Bytes(), &translated) != nil ||
		translated.SourceDigest == "" || translated.SourceDigest != collections.EnglishDescriptionDigest(aiSavedEnglish) {
		t.Fatal("confirmed English translation missing a correct source digest")
	}
	large.Spanish = translated.Spanish
	edit.ListingFields.DescriptionTranslation = &collections.DescriptionTranslation{SpanishSourceDigest: strings.Repeat("a", 64)}
	saveOwn(400)
	edit.ListingFields.DescriptionTranslation.SpanishSourceDigest = translated.SourceDigest
	saveOwn(200)
	edit.ExpectedRevision++
	large.English = "Revised English master."
	edit.ListingFields.DescriptionTranslation = nil
	saveOwn(200)
	edit.ExpectedRevision++
	edit.ListingFields.Descriptions = nil
	saveOwn(200)
	saved, err := store.GetOwnProduct(ctx, owner, own.ID)
	if err != nil || saved.ListingFields.Descriptions == nil ||
		*saved.ListingFields.Descriptions != large || saved.Description != legacy ||
		saved.ListingFields.DescriptionTranslation == nil ||
		saved.ListingFields.DescriptionTranslation.SpanishSourceDigest != translated.SourceDigest ||
		saved.ListingFields.DescriptionTranslation.EnglishDigest == translated.SourceDigest ||
		saved.ListingFields.DescriptionTranslation.PortugueseSourceDigest != "" {
		t.Fatalf("own product descriptions, stale source or legacy data did not round-trip: %v", err)
	}
	if w := call(ctx, handler, "POST", genPath, translationInput, origin, access); w.Code != 409 {
		t.Fatalf("stale translation request status=%d", w.Code)
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
	allowNext()
	providerMode.Store(3)
	if w := call(ctx, handler, "POST", genPath, input, origin, access); w.Code != 409 {
		t.Fatalf("stale draft generation status=%d", w.Code)
	}
	input = strings.Replace(input, `"expected_revision":1`, `"expected_revision":3`, 1)
	if w := call(ctx, handler, "POST", genPath, input, origin, access); w.Code != 502 || strings.Contains(w.Body.String(), "synthetic-provider-key") || strings.Contains(w.Body.String(), "PRIVATE") {
		t.Fatalf("upstream error response not safely mapped: %d", w.Code)
	}

	reqCtx, abort := context.WithCancel(ctx)
	defer abort()
	done := make(chan struct{})
	allowNext()
	providerMode.Store(1)
	go func() {
		defer close(done)
		call(reqCtx, handler, "POST", genPath, input, origin, access)
	}()
	select {
	case <-started:
		if w := call(ctx, handler, "POST", genPath, input, origin, access); w.Code != 429 {
			t.Errorf("active owner duplicate status=%d", w.Code)
		}
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
	allowNext()
	providerMode.Store(2)
	service.client.Timeout = 80 * time.Millisecond
	if w := call(ctx, handler, "POST", genPath, input, origin, access); w.Code != 504 {
		t.Fatalf("upstream timeout status=%d", w.Code)
	}
}

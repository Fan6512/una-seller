package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func TestCategoryAttributesEndpointAccessAndRules(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing in CI")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cs, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.DB.Close()
	if err := migrations.Apply(ctx, cs.DB); err != nil {
		t.Fatal(err)
	}
	var owner, store string
	if err := cs.DB.QueryRow(ctx, "INSERT INTO app.users(id) VALUES(gen_random_uuid()) RETURNING id::text").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = cs.DB.Exec(context.Background(), "DELETE FROM app.meli_store_connections WHERE owner_user_id=$1::uuid", owner)
		_, _ = cs.DB.Exec(context.Background(), "DELETE FROM app.users WHERE id=$1::uuid", owner)
	}()
	const seller = "678921"
	const access = "synthetic_category_access"
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	cipher, err := meli.NewTokenCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := cipher.Seal(access, []byte("meli|"+owner+"|"+seller+"|access"))
	if err != nil {
		t.Fatal(err)
	}
	markets := `[{"user_id":987654,"site_id":"MLB","logistic_type":"remote"}]`
	err = cs.DB.QueryRow(ctx, `INSERT INTO app.meli_store_connections
		(owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,
		access_token_expires_at,marketplaces,marketplaces_status)
		VALUES ($1::uuid,$2,$3,$3,now()+interval '30 minutes',$4::jsonb,'verified')
		RETURNING id::text`, owner, seller, encrypted, markets).Scan(&store)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 12 * time.Second, Transport: listingTestTransport(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer "+access ||
			r.URL.Host != "api.mercadolibre.com" {
			t.Error("unexpected platform request or credentials")
		}
		var body string
		switch r.URL.Path {
		case "/categories/CBT771/attributes":
			body = `[{"id":"BRAND","name":"Brand","value_type":"string","tags":{"required":true}},
 {"id":"CAPACITY","name":"Capacity","value_type":"number_unit","tags":{"new_required":true},"allowed_units":[{"id":"l","name":"L"}]},
 {"id":"MATERIAL","name":"Material","value_type":"list","values":[{"id":"200","name":"Glass"}]},
 {"id":"COUNT","name":"Count","value_type":"number","tags":{"required":true}}]`
		case "/categories/CBT771":
			body = `{"id":"CBT771","path_from_root":[{"id":"CBT1","name":"Home"},{"id":"CBT771","name":"Jugs"}]}`
		case "/categories/CBT771/technical_specs/input":
			body = `{"groups":[{"components":[{"component":"TEXT_INPUT","attributes":[
 {"id":"BRAND","tags":["required"]},{"id":"SIZE","tags":["conditional_required"]},
 {"id":"STYLE","tags":["required"]}]}]}]}`
		case "/categories/CBT772/attributes":
			body = `[]`
		case "/categories/CBT772":
			body = `{"id":"CBT772"}`
		case "/categories/CBT772/technical_specs/input":
			body = `{"groups":[]}`
		case "/categories/CBT773/attributes":
			body = `{"not":"an attribute array"}`
		case "/categories/MLB9891/attributes":
			body = `[{"id":"BRAND","name":"Marca","value_type":"string","tags":{"required":true}},
				{"id":"COLOR","name":"Color","value_type":"list","tags":{"conditional_required":true},
				 "values":[{"id":"1","name":"Rosa"},{"id":"2","name":"Azul"}]}]`
		case "/categories/MLB9891":
			body = `{"id":"MLB9891","path_from_root":[{"id":"MLB1","name":"Acessórios"},{"id":"MLB9891","name":"Filtros"}]}`
		default:
			t.Errorf("unexpected platform path: %s", r.URL.Path)
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("{}")), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}
	svc, err := meli.NewService(cs.DB, meli.Config{ClientID: "synthetic", ClientSecret: "synthetic", RedirectURI: "https://example.invalid" + meli.CallbackPath, TokenEncryptionKey: key, HTTPClient: client})
	if err != nil {
		t.Fatal(err)
	}
	handler := HandlerWithCollections(cs, stubVerifier{id: owner}, svc)
	path := "/api/v1/stores/" + store + "/category-attributes"
	check := func(method, url, bearer string, expected int, code string) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, url, nil)
		if bearer != "" {
			req.Header.Set("Authorization", "Bearer "+bearer)
		}
		out := httptest.NewRecorder()
		handler.ServeHTTP(out, req)
		var obj map[string]any
		if err := json.Unmarshal(out.Body.Bytes(), &obj); err != nil {
			t.Fatalf("unreadable JSON: %v", err)
		}
		if out.Code != expected || (code != "" && obj["code"] != code) {
			t.Fatalf("%s: HTTP %d, response %#v, expected %d/%s", url, out.Code, obj, expected, code)
		}
		return obj
	}
	scope := "?site=MLB&logistic_type=remote&category_id=MLB9891"
	check("GET", path+scope, "", 401, "")
	check("GET", path+"?site=MLM&logistic_type=remote&category_id=MLB9891",
		"session", 400, "category_invalid")
	check("GET", path+"?site=CBT&logistic_type=&category_id=CBT9891",
		"session", 400, "category_invalid")
	check("GET", path+"?site=MLB&logistic_type=fulfillment&category_id=MLB9891",
		"session", 403, "listing_site_unverified")
	check("GET", "/api/v1/stores/00000000-0000-4000-8000-000000000001/category-attributes"+scope,
		"session", 404, "listing_store_missing")
	check("POST", path+scope, "session", 405, "")
	valid := check("GET", path+scope, "session", 200, "")
	if valid["category_id"] != "MLB9891" {
		t.Fatal("wrong category")
	}
	attrs, ok := valid["attributes"].([]any)
	if !ok || len(attrs) != 2 {
		t.Fatalf("missing platform attributes: %#v", valid)
	}
	if _, ok := valid["path"].([]any); !ok {
		t.Fatalf("missing category breadcrumb: %#v", valid)
	}
	cbt := "/api/v1/stores/cbt-category-attributes?category_id="
	check("GET", cbt+"CBT771", "", 401, "")
	check("GET", cbt+"MLB9891", "session", 400, "")
	check("GET", cbt+"CBT773", "session", 502, "")
	global := check("GET", cbt+"CBT771", "session", 200, "")
	if global["category_id"] != "CBT771" || global["technical_specs_verified"] != true ||
		global["special_requirements"] != true || len(global["technical_requirements"].([]any)) != 2 ||
		len(global["path"].([]any)) != 2 || len(global["attributes"].([]any)) != 4 {
		t.Fatalf("incomplete CBT rule response: %#v", global)
	}
	empty := check("GET", cbt+"CBT772", "session", 200, "")
	if len(empty["attributes"].([]any)) != 0 || empty["technical_specs_verified"] != true {
		t.Fatalf("empty category misinterpreted: %#v", empty)
	}
	if _, ok := global["translations_zh"]; ok {
		t.Fatal("unconfigured translator supplied translations")
	}

	translator := NewAzureTranslationService("fixture-key", "")
	translator.client.Transport = listingTestTransport(func(r *http.Request) (*http.Response, error) {
		var texts []struct{ Text string }
		if err := json.NewDecoder(r.Body).Decode(&texts); err != nil {
			t.Fatal(err)
		}
		if r.URL.Query().Get("to") != "zh-Hans" || r.Header.Get("Authorization") != "" {
			t.Fatal("invalid translation request")
		}
		labels := map[string]string{"Brand": "品牌", "Glass": "玻璃"}
		results := make([]map[string][]map[string]string, 0, len(texts))
		for _, item := range texts {
			label := labels[item.Text]
			if label == "" {
				label = item.Text
			}
			results = append(results, map[string][]map[string]string{
				"translations": {{"text": label, "to": "zh-Hans"}},
			})
		}
		data, err := json.Marshal(results)
		if err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(string(data)))}, nil
	})
	handler = HandlerWithServices(cs, stubVerifier{id: owner}, svc, translator)
	localized := check("GET", cbt+"CBT771", "session", 200, "")
	labels, ok := localized["translations_zh"].(map[string]any)
	if !ok || labels["Brand"] != "品牌" || labels["Glass"] != "玻璃" {
		t.Fatalf("missing CBT translations: %#v", localized["translations_zh"])
	}
	for _, key := range []string{"category_id", "attributes", "path"} {
		if !reflect.DeepEqual(global[key], localized[key]) {
			t.Fatalf("translation changed official %s", key)
		}
	}
	if strings.Contains(fmt.Sprint(valid, global, localized), access) {
		t.Fatal("platform token leaked to client")
	}
}

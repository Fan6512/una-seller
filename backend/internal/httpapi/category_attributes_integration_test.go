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
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/cbtcategory"
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
		case "/marketplace/domain_discovery/search":
			body = `[{"category_id":"CBT771","category_name":"Jugs"},{"category_id":"CBT772","category_name":"Other"}]`
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
 {"id":"STYLE","name":"Style","tags":["required"]}]}]}]}`
		case "/categories/CBT772/attributes":
			body = `[]`
		case "/categories/CBT772":
			body = `{"id":"CBT772"}`
		case "/categories/CBT772/technical_specs/input":
			body = `{"groups":[]}`
		case "/categories/CBT773/attributes":
			body = `{"not":"an attribute array"}`
		case "/categories/MLB9891/attributes":
			body = `[{"id":"BRAND","name":"Marca","value_type":"string","attribute_group_name":"Outros","tags":{"required":true},"values":[{"id":"3","name":"ExampleBrand"}]},
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

	dictionaryDir := t.TempDir()
	const dictionaryJSON = `{"version":{"md5":"c370aeb2af31f17b3eb45f701d422605","created":"2026-10-14T21:12:35Z"},"categories":{
	 "CBT1":{"id":"CBT1","name":"Home","path":[{"id":"CBT1","name":"Home"}],"depth":1,"has_children":true,"status":"enabled","zh":"家居"},
	 "CBT771":{"id":"CBT771","name":"Jugs","parent_id":"CBT1","path":[{"id":"CBT1","name":"Home"},{"id":"CBT771","name":"Jugs"}],"depth":2,"listing_allowed":true,"status":"enabled","zh":"水壶"},
	 "CBT772":{"id":"CBT772","name":"Other","path":[{"id":"CBT772","name":"Other"}],"depth":1,"listing_allowed":true,"status":"disabled"}}}`
	if err := os.WriteFile(filepath.Join(dictionaryDir, "active.json"), []byte(dictionaryJSON), 0600); err != nil {
		t.Fatal(err)
	}
	dictionary := cbtcategory.NewReader(dictionaryDir)
	handler = HandlerWithCategoryDictionary(cs, stubVerifier{id: owner}, svc, nil, dictionary, nil)
	offline := check("GET", cbt+"CBT771", "session", 200, "")
	if !reflect.DeepEqual(offline["path_zh"], []any{"家居", "水壶"}) || offline["translations_zh"] != nil {
		t.Fatalf("offline dictionary failed without provider: %#v", offline)
	}

	providerCalls := 0
	translator := NewAzureTranslationService("fixture-key", "")
	translator.provider.(*azureTranslationProvider).client.Transport = listingTestTransport(func(r *http.Request) (*http.Response, error) {
		providerCalls++
		var texts []struct{ Text string }
		if err := json.NewDecoder(r.Body).Decode(&texts); err != nil {
			t.Fatal(err)
		}
		source := r.URL.Query().Get("from")
		if r.URL.Query().Get("to") != "zh-Hans" || r.Header.Get("Authorization") != "" || source != "en" && source != "pt" {
			t.Fatal("invalid translation request")
		}
		labels := map[string]string{"Brand": "品牌", "Glass": "玻璃", "Style": "风格", "Marca": "品牌", "Acessórios": "配件", "Filtros": "滤芯", "Outros": "其他属性", "Rosa": "粉色"}
		results := make([]map[string][]map[string]string, 0, len(texts))
		for _, item := range texts {
			switch item.Text {
			case "Home", "Jugs":
				t.Fatal("offline CBT category path reached the live translation provider")
			case "ExampleBrand":
				t.Fatal("identifier choice reached the live translation provider")
			case "Marca", "Acessórios", "Filtros", "Outros", "Rosa":
				if source != "pt" {
					t.Fatal("attribute reached wrong language bucket")
				}
			}
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
	handler = HandlerWithCategoryDictionary(cs, stubVerifier{id: owner}, svc, translator, dictionary, nil)
	localized := check("GET", cbt+"CBT771", "session", 200, "")
	labels, ok := localized["translations_zh"].(map[string]any)
	if !ok || labels["Brand"] != "品牌" || labels["Glass"] != "玻璃" || labels["Style"] != "风格" {
		t.Fatalf("missing CBT translations: %#v", localized["translations_zh"])
	}
	published := check("GET", path+scope, "session", 200, "")
	publishedLabels, ok := published["translations_zh"].(map[string]any)
	if !ok || publishedLabels["Marca"] != "品牌" || publishedLabels["Acessórios"] != "配件" ||
		publishedLabels["Filtros"] != "滤芯" || publishedLabels["Outros"] != "其他属性" || publishedLabels["Rosa"] != "粉色" || published["path_zh"] != nil {
		t.Fatalf("site-only category should retain its official path and localize metadata: %#v", published)
	}
	if names, ok := localized["technical_requirement_names"].(map[string]any); !ok || names["STYLE"] != "Style" {
		t.Fatalf("missing official technical requirement name: %#v", localized)
	}
	for _, key := range []string{"category_id", "attributes", "path", "technical_requirements", "technical_requirement_names"} {
		if !reflect.DeepEqual(global[key], localized[key]) {
			t.Fatalf("translation changed official %s", key)
		}
	}
	for _, key := range []string{"category_id", "attributes", "path"} {
		if !reflect.DeepEqual(valid[key], published[key]) {
			t.Fatalf("translation changed published category %s", key)
		}
	}
	beforeRead := providerCalls
	predicted := check("GET", "/api/v1/categories/predict?q=Glass+jar", "session", 200, "")
	candidates, ok := predicted["items"].([]any)
	if !ok || len(candidates) != 1 {
		t.Fatalf("unavailable category should not appear in new suggestions: %#v", predicted)
	}
	candidate := candidates[0].(map[string]any)
	if candidate["category_id"] != "CBT771" || candidate["category_name"] != "Jugs" ||
		!reflect.DeepEqual(candidate["path_zh"], []any{"家居", "水壶"}) {
		t.Fatalf("prediction changed official identity or lost offline path: %#v", candidate)
	}
	if _, err := cs.DB.Exec(ctx, `INSERT INTO app.meli_published_listings(store_id,site_id,logistic_type,listing_id,details) VALUES
	  ($1::uuid,'CBT','','CBT1000', '{"id":"CBT1000","seller_id":678921,"category_id":"CBT771"}'),
	  ($1::uuid,'CBT','','CBT1001', '{"id":"CBT1001","seller_id":666666,"category_id":"CBT771"}'),
	  ($1::uuid,'MLB','remote','MLB201', '{"id":"MLB201","seller_id":987654,"category_id":"MLB9891","cbt_item_id":"CBT1000"}'),
	  ($1::uuid,'MLB','remote','MLB202', '{"id":"MLB202","seller_id":987654,"category_id":"MLB9891","cbt_item_id":"CBT1001"}'),
	  ($1::uuid,'MLB','remote','MLB203', '{"id":"MLB203","seller_id":987654,"category_id":"MLB771"}'),
	  ($1::uuid,'MLB','remote','MLB204', '{"id":"MLB204","seller_id":987654,"category_id":"MLB9891","parent_item_id":"CBT1000"}'),
	  ($1::uuid,'MLB','remote','MLB205', '{"id":"MLB205","seller_id":666666,"category_id":"MLB9891","cbt_item_id":"CBT1000"}')`, store); err != nil {
		t.Fatal(err)
	}
	list := check("GET", "/api/v1/stores/"+store+"/published-listings?site=MLB&logistic_type=remote", "session", 200, "")
	items := list["items"].([]any)
	if len(items) != 4 {
		t.Fatalf("wrong seller or subset returned: %#v", list)
	}
	for i, raw := range items {
		item := raw.(map[string]any)
		breadcrumb, matched := item["cbt_category_path"].(map[string]any)
		if matched != (i == 0 || i == 3) || matched && (item["category_id"] != "MLB9891" || !reflect.DeepEqual(breadcrumb["zh"], []any{"家居", "水壶"})) {
			t.Fatalf("wrong storefront/global linkage on listing: %#v", item)
		}
	}
	if providerCalls != beforeRead {
		t.Fatal("prediction or listing view called the live translator")
	}
	if strings.Contains(fmt.Sprint(valid, global, localized, published, predicted, list), access) {
		t.Fatal("platform token leaked to client")
	}
}

package meli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/publication"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestProductPublicationPostgresHTTP(t *testing.T) {
	for _, scenario := range []struct {
		name                  string
		mode                  publication.TitleMode
		family, partial, lost bool
	}{
		{name: "single_auto", mode: publication.PlatformAuto},
		{name: "family_auto", mode: publication.PlatformAuto, family: true},
		{name: "family_localized", mode: publication.Localized, family: true},
		{name: "partial", mode: publication.PlatformAuto, family: true, partial: true},
		{name: "lost_response", mode: publication.PlatformAuto, family: true, lost: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			s, db, owner, store := setupRefreshTest(t)
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			var parent string
			if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '2 hours'
 WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&parent); err != nil {
				t.Fatal(err)
			}
			const child = "987654"
			const category = "CBT416197"
			const red, blue = "11111111111111111111111111111111", "22222222222222222222222222222222"
			quantity, net := 2, "25.125"
			f := collections.ListingFields{
				Titles:       collections.LocalizedTitles{English: "Cotton shirt", Portuguese: "Camisa de algodão"},
				Descriptions: &collections.LocalizedDescriptions{English: "Cotton.", Portuguese: "De algodão."},
				CategoryID:   category, SellerSKU: "RED", AvailableQuantity: &quantity, TargetNetProceedsUSD: &net,
				PurchasePrice: "99.99", Attributes: []collections.ListingAttribute{{Name: "采集属性", Value: "辅助说明"}},
				CategoryAttributes: &[]collections.ListingCategoryAttribute{
					{CategoryID: category, ID: "ITEM_CONDITION", ValueID: "2230284", Value: "new"},
					{CategoryID: category, ID: "BRAND", Value: " Example "},
					{CategoryID: category, ID: "UNIT_VOLUME", Value: "010.00", UnitID: "mL"},
				},
			}
			pictures := publicationPictures{publication.SingleKey: {"photo-red"}}
			if scenario.family {
				f.VariantIdentityVersion = 1
				f.Variants = &[]collections.ListingVariant{
					{VariantID: red, SellerSKU: "RED", AvailableQuantity: &quantity, TargetNetProceedsUSD: &net,
						CategoryAttributes: &[]collections.ListingCategoryAttribute{{CategoryID: category, ID: "COLOR", ValueID: "1", Value: "wrong label"}}},
					{VariantID: blue, SellerSKU: "BLUE", AvailableQuantity: new(int), TargetNetProceedsUSD: &net,
						CategoryAttributes: &[]collections.ListingCategoryAttribute{{CategoryID: category, ID: "COLOR", ValueID: "2", Value: "blue"}}},
				}
				pictures = publicationPictures{red: {"photo-red"}, blue: {"photo-blue"}}
			}
			var product string
			fields, err := json.Marshal(f)
			if err != nil {
				t.Fatal(err)
			}
			if err = db.QueryRow(ctx, `INSERT INTO app.own_products(owner_user_id,source_platform,source_site,source_url,source_title,title,notes,listing_fields)
 VALUES($1::uuid,'amazon','US','https://example.invalid/product','source','saved','中文备注',$2::jsonb) RETURNING id::text`, owner, string(fields)).Scan(&product); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				for _, query := range []string{
					`DELETE FROM app.cbt_publication_identities i USING app.cbt_publication_attempts a WHERE i.request_id=a.request_id AND a.own_product_id=$1::uuid`,
					`DELETE FROM app.cbt_publication_attempts WHERE own_product_id=$1::uuid`,
					`DELETE FROM app.cbt_publication_inputs WHERE own_product_id=$1::uuid`,
					`DELETE FROM app.own_products WHERE id=$1::uuid`,
				} {
					if _, err := db.Exec(c, query, product); err != nil {
						t.Error(err)
					}
				}
			})
			img, err := json.Marshal(pictures)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = db.Exec(ctx, `INSERT INTO app.cbt_publication_inputs(owner_user_id,own_product_id,store_id,product_revision,site_id,
 parent_seller_id,site_seller_id,listing_type_id,pictures_by_variant,approved_until)
 VALUES($1::uuid,$2::uuid,$3::uuid,1,'MLB',$4,$5,'gold_pro',$6::jsonb,now()+interval '1 hour')`, owner, product, store, parent, child, string(img)); err != nil {
				t.Fatal(err)
			}
			box := &collections.Store{DB: db}
			p, err := box.GetOwnProduct(ctx, owner, product)
			if err != nil {
				t.Fatal(err)
			}
			targets, err := s.PublicationTargets(ctx, owner, p)
			if err != nil || len(targets) != 1 || targets[0].ListingType != "gold_pro" {
				t.Fatalf("targets=%+v err=%v", targets, err)
			}

			const globalFamily = "9007199254740993"
			const siteFamily = "9007199254740995"
			var posts atomic.Int32
			var resolved atomic.Bool
			postStarted, release := make(chan struct{}), make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
					t.Error("missing scoped OAuth token")
				}
				if r.Method == http.MethodPost {
					if posts.Add(1) != 1 {
						t.Error("duplicate official POST")
						http.Error(w, "duplicate", 500)
						return
					}
					want := publication.SinglePath
					if scenario.family {
						want = publication.FamilyPath
					}
					if r.URL.Path != want {
						t.Errorf("unexpected POST path %s", r.URL.Path)
					}
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Error(err)
					}
					var persisted []byte
					var digest, state, mode string
					var variants int
					err = db.QueryRow(r.Context(), `SELECT payload,encode(payload_sha256,'hex'),state,title_mode,
 (SELECT count(*) FROM app.cbt_publication_identities i WHERE i.request_id=a.request_id)
 FROM app.cbt_publication_attempts a WHERE request_id=$1::uuid`, r.Header.Get("X-Request-ID")).Scan(&persisted, &digest, &state, &mode, &variants)
					if err != nil || state != "submitting" || mode != string(scenario.mode) || !bytes.Equal(persisted, body) ||
						digest != (publication.Prepared{Body: body}).PayloadSHA256() || variants != len(pictures) {
						t.Errorf("intent not durable before POST: %s %s %d %v", state, mode, variants, err)
					}
					if strings.Contains(string(body), "99.99") || strings.Contains(string(body), "辅助说明") || strings.Contains(string(body), "中文备注") ||
						!strings.Contains(string(body), `"global_net_proceeds":25.125`) || !strings.Contains(string(body), `"name":"10 mL"`) ||
						!strings.Contains(string(body), `"listing_type_id":"gold_pro"`) || !strings.Contains(string(body), `"id":"photo-red"`) ||
						(strings.Contains(string(body), `"family_name":"Camisa de algodão"`) != (scenario.mode == publication.Localized)) {
						t.Errorf("unexpected wire payload: %s", body)
					}
					close(postStarted)
					select {
					case <-release:
					case <-r.Context().Done():
						return
					}
					if scenario.lost {
						conn, _, err := w.(http.Hijacker).Hijack()
						if err != nil {
							t.Error(err)
							return
						}
						_ = conn.Close()
						return
					}
					hint := func(n int, site bool) string {
						dest := ""
						if site {
							dest = fmt.Sprintf(`{"item_id":"MLB%d","site_id":"MLB","seller_id":%s,"user_product_id":"MLBU%d","logistic_type":"remote"}`, n, child, n)
						}
						return fmt.Sprintf(`{"seller_id":%s,"site_id":"CBT","siteless_user_product_id":"U%d","siteless_family_id":%s,"site_items":[%s]}`, parent, n, globalFamily, dest)
					}
					w.WriteHeader(http.StatusCreated)
					if scenario.family {
						fmt.Fprintf(w, "[%s,%s]", hint(2, !scenario.partial), hint(1, true))
						return
					}
					fmt.Fprint(w, hint(1, true))
					return
				}
				switch r.URL.Path {
				case "/users/me":
					fmt.Fprintf(w, `{"id":%s,"site_id":"CBT","status":{"sell":{"allow":true}}}`, parent)
				case "/users/" + parent:
					fmt.Fprint(w, `{"tags":["user_product_seller"]}`)
				case "/marketplace/users/" + parent:
					fmt.Fprintf(w, `{"user_id":%s,"site_id":"CBT","marketplaces":[{"user_id":%s,"site_id":"MLB","logistic_type":"remote","pricing_model":"net_proceeds","user_product":true}]}`, parent, child)
				case "/categories/" + category + "/technical_specs/input":
					fmt.Fprint(w, `{"groups":[{"components":[{"component":"TEXT","attributes":[{"id":"UNIT_VOLUME","tags":["required"]}]}]}]}`)
				case "/categories/" + category + "/attributes":
					fmt.Fprint(w, `[
 {"id":"ITEM_CONDITION","value_type":"list","tags":{"required":true},"values":[{"id":"2230284","name":"New"}]},
 {"id":"BRAND","value_type":"string","hierarchy":"PARENT_PK"},
 {"id":"COLOR","value_type":"list","hierarchy":"CHILD_PK","values":[{"id":"1","name":"Red"},{"id":"2","name":"Blue"}]},
 {"id":"UNIT_VOLUME","value_type":"number_unit","allowed_units":[{"id":"mL"}]}]`)
				case "/marketplace/users/cap":
					fmt.Fprintf(w, `[{"user_id":%s,"site_id":"MLB","logistic_type":"remote","quota":10,"total_items":1}]`, child)
				case "/user-products/U1", "/user-products/U2":
					if r.Header.Get("X-API-Version") != "2" {
						t.Error("missing UP API version")
					}
					n := r.URL.Path[len(r.URL.Path)-1:]
					sku := "RED"
					if n == "2" {
						sku = "BLUE"
					}
					fmt.Fprintf(w, `{"id":"U%s","user_id":%s,"family_id":%s,"attributes":[{"id":"SELLER_SKU","values":[{"name":%q}]}]}`, n, parent, globalFamily, sku)
				case "/marketplace/user-products/U1/mapping", "/marketplace/user-products/U2/mapping":
					n := strings.Split(r.URL.Path, "/")[3][1:]
					fmt.Fprintf(w, `[{"owner_id":%s,"site_id":"CBT","siteless_user_product_id":"U%s","site_items":[{"item_id":"MLB%s","site_id":"MLB","logistic_type":"remote"}]}]`, parent, n, n)
				case "/marketplace/items/MLB1", "/marketplace/items/MLB2":
					n := r.URL.Path[len(r.URL.Path)-1:]
					seller := child
					if scenario.partial && n == "2" && !resolved.Load() {
						seller = "8000"
					}
					fmt.Fprintf(w, `{"id":"MLB%s","seller_id":%s,"site_id":"MLB","user_logistic_type":"remote","user_product_id":"MLBU%s","family_id":%s}`, n, seller, n, siteFamily)
				case "/user-products/MLBU1", "/user-products/MLBU2":
					n := r.URL.Path[len(r.URL.Path)-1:]
					title := "Automatic translation"
					if scenario.mode == publication.Localized && resolved.Load() {
						title = "Camisa de algodão"
					}
					fmt.Fprintf(w, `{"id":"MLBU%s","user_id":%s,"family_id":%s,"family_name":%q}`, n, child, siteFamily, title)
				case "/sites/CBT/user-products-families/" + globalFamily:
					fmt.Fprintf(w, `{"family_id":%s,"user_id":%s,"user_products_ids":["U1","U2"]}`, globalFamily, parent)
				case "/sites/MLB/user-products-families/" + siteFamily:
					fmt.Fprintf(w, `{"family_id":%s,"user_id":%s,"user_products_ids":["MLBU1","MLBU2"]}`, siteFamily, child)
				default:
					t.Errorf("unexpected GET %s", r.URL.Path)
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			s.apiURL, s.client = server.URL, server.Client()
			in := PublicationConfirmation{StoreID: store, Site: "MLB", Revision: 1, TitleMode: scenario.mode}
			if _, err := s.SubmitPublication(ctx, owner, p, PublicationConfirmation{StoreID: store, Site: "MLB", Revision: 2, TitleMode: scenario.mode}); !errors.Is(err, ErrPublicationConflict) {
				t.Fatalf("stale revision: %v", err)
			}
			if _, err := s.SubmitPublication(ctx, owner, p, PublicationConfirmation{StoreID: store, Site: "MLM", Revision: 1, TitleMode: scenario.mode}); !errors.Is(err, ErrPublicationNotReady) {
				t.Fatalf("unapproved site: %v", err)
			}
			if scenario.name == "single_auto" {
				if _, err = db.Exec(ctx, `UPDATE app.own_products SET listing_fields=jsonb_set(listing_fields,'{category_attributes,2,unit_id}','"fake"') WHERE id=$1::uuid`, product); err != nil {
					t.Fatal(err)
				}
				bad, err := box.GetOwnProduct(ctx, owner, product)
				if err != nil {
					t.Fatal(err)
				}
				var invalid PublicationInputError
				if _, err = s.SubmitPublication(ctx, owner, bad, in); !errors.As(err, &invalid) {
					t.Fatalf("unverified unit: %v", err)
				}
				if _, err = db.Exec(ctx, `UPDATE app.own_products SET listing_fields=$2::jsonb WHERE id=$1::uuid`, product, string(fields)); err != nil {
					t.Fatal(err)
				}
			}
			type submission struct {
				result PublicationResult
				err    error
			}
			done := make(chan submission, 1)
			go func() { r, err := s.SubmitPublication(ctx, owner, p, in); done <- submission{r, err} }()
			select {
			case <-postStarted:
			case early := <-done:
				t.Fatalf("create never reached mock: %+v, %v", early.result, early.err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			duplicate, err := s.SubmitPublication(ctx, owner, p, in)
			close(release)
			if err != nil || duplicate.State != "submitting" || len(duplicate.Pending) != len(pictures) {
				t.Fatalf("concurrent duplicate=%+v err=%v", duplicate, err)
			}
			got := <-done
			wantMatched := len(pictures)
			if scenario.partial {
				wantMatched = 1
			}
			if scenario.lost {
				wantMatched = 0
			}
			if got.err != nil || got.result.RequestID != duplicate.RequestID ||
				len(got.result.Variants) != wantMatched || len(got.result.Pending) != len(pictures)-wantMatched {
				t.Fatalf("readback=%+v err=%v", got.result, got.err)
			}
			for _, v := range got.result.Variants {
				if (v.Listing.SKU == "RED" && (v.Listing.ItemID != "MLB1" || (scenario.family && v.Key != red))) ||
					(v.Listing.SKU == "BLUE" && (v.Listing.ItemID != "MLB2" || v.Key != blue)) ||
					v.Listing.SitelessFamilyID != globalFamily || v.Listing.SiteFamilyID != siteFamily ||
					v.TitleVerified != (scenario.mode == publication.PlatformAuto) {
					t.Fatalf("incorrect variant association %+v", v)
				}
			}
			again, err := s.SubmitPublication(ctx, owner, p, in)
			if err != nil || again.State != got.result.State || again.RequestID != duplicate.RequestID || posts.Load() != 1 {
				t.Fatalf("duplicate=%+v err=%v", again, err)
			}
			if scenario.family {
				_, err := db.Exec(ctx, `UPDATE app.own_products SET listing_fields=jsonb_set(listing_fields,'{variants}', (listing_fields->'variants')-0) WHERE id=$1::uuid`, product)
				var check *pgconn.PgError
				if !errors.As(err, &check) || check.ConstraintName != "cbt_publication_variant_identity" {
					t.Fatalf("identity deletion allowed: %v", err)
				}
			}
			resolved.Store(true)
			read, err := s.VerifyPublication(ctx, owner, product, store, "MLB")
			if err != nil || read.RequestID != duplicate.RequestID || posts.Load() != 1 {
				t.Fatalf("verify=%+v err=%v", read, err)
			}
			if scenario.lost {
				if read.State != "pending_verify" || len(read.Pending) != len(pictures) {
					t.Fatalf("unknown result was resolved without evidence: %+v", read)
				}
				return
			}
			if read.State != "success" || len(read.Variants) != len(pictures) || len(read.Pending) != 0 {
				t.Fatalf("verified=%+v", read)
			}
			for _, v := range read.Variants {
				if !v.TitleVerified {
					t.Fatalf("title readback not confirmed: %+v", v)
				}
			}
		})
	}
}

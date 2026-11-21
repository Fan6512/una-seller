package meli

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOrderSearchVerifiedOwnerScopeAndPagination(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces=
		'[{"user_id":987654,"site_id":"MLB","logistic_type":"remote"},
		  {"user_id":987654,"site_id":"MLM","logistic_type":"remote"}]'::jsonb
		WHERE id=$1::uuid`, store)
	if err != nil {
		t.Fatal(err)
	}

	var second, foreignOwner, foreignStore string
	err = db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
		(owner_user_id,seller_id,status,marketplaces,marketplaces_status)
		VALUES($1::uuid,gen_random_uuid()::text,'reauth_required',
		 '[{"user_id":987655,"site_id":"MLB","logistic_type":"remote"}]'::jsonb,
		 'verified') RETURNING id::text`, owner).Scan(&second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, _ = db.Exec(c, `DELETE FROM app.meli_store_connections WHERE id=$1::uuid`, second)
		if foreignStore != "" {
			_, _ = db.Exec(c, `DELETE FROM app.meli_store_connections WHERE id=$1::uuid`, foreignStore)
		}
		if foreignOwner != "" {
			_, _ = db.Exec(c, `DELETE FROM app.users WHERE id=$1::uuid`, foreignOwner)
		}
	})
	err = db.QueryRow(ctx, `INSERT INTO app.users(id) VALUES(gen_random_uuid()) RETURNING id::text`).Scan(&foreignOwner)
	if err != nil {
		t.Fatal(err)
	}
	err = db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
		(owner_user_id,seller_id,status,marketplaces,marketplaces_status)
		VALUES($1::uuid,gen_random_uuid()::text,'reauth_required',
		 '[{"user_id":987656,"site_id":"MLB","logistic_type":"remote"}]'::jsonb,
		 'verified') RETURNING id::text`, foreignOwner).Scan(&foreignStore)
	if err != nil {
		t.Fatal(err)
	}

	for _, row := range []struct{ store, site, seller, id, title string }{
		{store, "MLB", "987654", "10001", "Blue denim trousers"},
		{store, "MLM", "987654", "20002", "Yellow mini massager"},
		{second, "MLB", "987655", "30003", "Denim jacket"},
		{foreignStore, "MLB", "987656", "40004", "Yellow hidden product"},
	} {
		items := fmt.Sprintf(`[{"item_id":"%s-X","title":%q,"quantity":1,"unit_price":"12","currency_id":"USD"}]`, row.site, row.title)
		_, err = db.Exec(ctx, `INSERT INTO app.meli_orders
		 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
		  date_created,last_updated,order_items)
		 VALUES($1::uuid,$2,'remote',$3,$4,'paid','USD',12,now(),now(),$5::jsonb)`,
			row.store, row.site, row.seller, row.id, items)
		if err != nil {
			t.Fatal(err)
		}
	}
	base := OrderSearchFilter{Store: store, Site: "MLB", Logistic: "*", Query: "20002"}
	page, err := s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("site filter leaked: %+v %v", page, err)
	}
	base.Site = "*"
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "20002" ||
		page.Items[0].SiteID != "MLM" || page.Items[0].StoreID != store {
		t.Fatalf("cross-site Order ID: %+v %v", page, err)
	}
	base.Query = "denim"
	base.Site = "MLB"
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "10001" {
		t.Fatalf("site-limited title search: %+v %v", page, err)
	}
	base.Store = "all"
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 2 {
		t.Fatalf("all owned stores search: %+v %v", page, err)
	}
	base.Query = ""
	base.Site = "*"
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 3 {
		t.Fatalf("all owned orders excluding foreign: %+v %v", page, err)
	}
	// Server authorization is checked per row, not by trusting the desktop's store choice.
	base.Store = foreignStore
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("foreign store leaked %+v %v", page, err)
	}
	base.Store = "all"
	_, err = db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces_status='unverified'
		WHERE id=$1::uuid`, second)
	if err != nil {
		t.Fatal(err)
	}
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 2 {
		t.Fatalf("unverified mapping leaked %+v %v", page, err)
	}

	// Concurrent owners may have identical Order IDs; they remain isolated.
	base.Query = "40004"
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || len(page.Items) != 0 {
		t.Fatalf("foreign order ID leaked %+v %v", page, err)
	}

	// One global page is capped at 20 orders, with deterministic next-page offset.
	for i := 0; i < 21; i++ {
		_, err = db.Exec(ctx, `INSERT INTO app.meli_orders
		  (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,
		   total_amount,date_created,last_updated)
		  VALUES($1::uuid,'MLB','remote','987654',$2,'paid','USD',12,now(),now())`,
			store, fmt.Sprintf("5%04d", i))
		if err != nil {
			t.Fatal(err)
		}
	}
	base.Query = ""
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || !page.HasMore || len(page.Items) != 20 {
		t.Fatalf("global page 1: %+v %v", page, err)
	}
	base.Offset = 20
	page, err = s.SearchOrders(ctx, owner, base)
	if err != nil || page.HasMore || len(page.Items) != 3 {
		t.Fatalf("global page 2: %+v %v", page, err)
	}
}

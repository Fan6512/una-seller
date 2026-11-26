package meli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
)

const customerMessageID = "a35cf79864a845ca9a3bf6aee59bb4d7"

func TestCustomerNoticeScope(t *testing.T) {
	auth := AuthorizedStore{SellerID: "1234", Verified: true, Markets: []Marketplace{
		{UserID: "987654", SiteID: "MLB", LogisticType: "remote"},
		{UserID: "987654", SiteID: "MLB", LogisticType: "fulfillment"},
		{UserID: "987655", SiteID: "MLM", LogisticType: "remote"},
	}}
	for _, tc := range []struct {
		user, seller, site, want string
	}{
		{"1234", "987654", "MLB", "987654"},
		{"987654", "1234", "MLB", "987654"},
		{"987655", "987654", "MLB", ""},
		{"1234", "987654", "MLM", ""},
		{"1234", "1234", "CBT", ""},
		{"1234", "0", "MLB", ""},
	} {
		got, err := customerSeller(auth, tc.user, tc.seller, tc.site)
		if got != tc.want || (err == nil) != (tc.want != "") {
			t.Fatalf("scope=%+v got=%q err=%v", tc, got, err)
		}
	}
	auth.Markets = append(auth.Markets, Marketplace{UserID: "999", SiteID: "MLB", LogisticType: "other"})
	if _, err := customerSeller(auth, "987654", "1234", "MLB"); !errors.Is(err, notifications.ErrScope) {
		t.Fatal("notification identity disambiguated parent response")
	}
}

func TestCustomerNotificationProjectionPostgres(t *testing.T) {
	status := "UNANSWERED"
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/marketplace/questions/123":
			fmt.Fprintf(w, `{"id":123,"seller_id":987654,"site_id":"MLB","item_id":"CBT123","status":%q,
"text":"question","text_translated":"translated","date_created":"2026-11-20T00:00:00Z","from":{"id":111}}`, status)
		case "/marketplace/messages/" + customerMessageID:
			fmt.Fprintf(w, `{"messages":[{"id":%q,"site_id":"MLB","status":"available","text":"message",
"from":{"user_id":111,"name":"private"},"to":{"user_id":987654},
"message_date":{"created":"2026-11-20T00:00:00Z"},"message_resources":[{"name":"packs","id":"700"},{"name":"sellers","id":"987654"}]}]}`, customerMessageID)
		default:
			http.NotFound(w, r)
		}
	})
	q := c.job("marketplace_questions", "123", "987654")
	resource, err := c.prepare(t, q)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_customer_notifications WHERE store_id=$1`, c.store); got != 0 {
		t.Fatal("prepare wrote a projection")
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	status = "ANSWERED"
	newer, err := c.prepare(t, q)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, newer); err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_customer_notifications WHERE store_id=$1
AND topic='marketplace_questions' AND status='ANSWERED' AND body='question'`, c.store); got != 1 {
		t.Fatal("old response overwrote newer question")
	}
	msg, err := c.prepare(t, c.job("marketplace_messages", customerMessageID, "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, msg); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_customer_notifications WHERE store_id=$1
AND topic='marketplace_messages' AND related_id='700' AND direction='incoming' AND body='message'`, c.store); got != 1 {
		t.Fatal("message missing or incorrectly scoped")
	}
	status = "DELETED"
	deleted, err := c.prepare(t, q)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, deleted); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_customer_notifications WHERE store_id=$1
AND topic='marketplace_questions' AND status='DELETED' AND body=''`, c.store); got != 1 {
		t.Fatal("deleted question text retained")
	}
	if _, err := c.db.Exec(t.Context(), `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, c.store); err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, msg); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("revoked mapping accepted: %v", err)
	}
}

func TestCustomerMessageRejectsWrongParticipant(t *testing.T) {
	auth := AuthorizedStore{SellerID: "1234", Verified: true, Markets: []Marketplace{
		{UserID: "987654", SiteID: "MLB", LogisticType: "remote"},
	}}
	job := notifications.Job{Topic: "marketplace_messages", ResourceID: customerMessageID, NotificationUserID: "1234"}
	raw := json.RawMessage(fmt.Sprintf(`{"messages":[{"id":%q,"site_id":"MLB","status":"available",
"from":{"user_id":111},"to":{"user_id":222},"message_date":{"created":"2026-11-20T00:00:00Z"},
"message_resources":[{"name":"packs","id":"700"},{"name":"sellers","id":"987654"}]}]}`, customerMessageID))
	if _, err := decodeCustomerNotice(raw, job, auth, time.Now()); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("unrelated participants accepted: %v", err)
	}
}

func TestCustomerMessageAcceptsQuotedIDs(t *testing.T) {
	auth := AuthorizedStore{SellerID: "1234", Verified: true, Markets: []Marketplace{
		{UserID: "987654", SiteID: "MLB", LogisticType: "remote"},
	}}
	job := notifications.Job{Topic: "marketplace_messages", ResourceID: customerMessageID, NotificationUserID: "1234"}
	raw := json.RawMessage(fmt.Sprintf(`{"messages":[{"id":%q,"site_id":"MLB","status":"available","text":"message",
"from":{"user_id":"111"},"to":{"user_id":"987654"},"message_date":{"created":"2026-11-20T00:00:00Z"},
"message_resources":[{"name":"packs","id":"700"},{"name":"sellers","id":"987654"}]}]}`, customerMessageID))
	n, err := decodeCustomerNotice(raw, job, auth, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if n.seller != "987654" || n.related != "700" || n.direction != "incoming" {
		t.Fatalf("quoted IDs decoded incorrectly: %+v", n)
	}
}

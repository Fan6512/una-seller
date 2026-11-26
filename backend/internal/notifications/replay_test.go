package notifications

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func missedItemPage(user string, count int) []byte {
	raw := `{"messages":[`
	for i := 0; i < count; i++ {
		if i > 0 {
			raw += ","
		}
		raw += fmt.Sprintf(`{"_id":"test-%d","application_id":%q,"user_id":%q,"topic":"marketplace_items","resource":"/marketplace/resource/items/MLB%d/","received":"2026-11-20T00:00:00Z","request":{"data":"untrusted"}}`, i, testApp, user, 1000+i)
	}
	return []byte(raw + "]}")
}

func TestMissedFeedNormalization(t *testing.T) {
	events, err := decodeMissedFeedPage(missedItemPage(testChildA, 1), testApp, "marketplace_items")
	if err != nil || len(events) != 1 || events[0].ResourceID != "MLB1000" {
		t.Fatalf("item history alias: %v", err)
	}
	message := []byte(`{"messages":[{"_id":"history-id","application_id":"123","user_id":"456","topic":"marketplace_messages","resource":"aaaaaaaaaaaaaaaa","received":"2026-11-20T00:00:00Z"}]}`)
	if got, err := decodeMissedFeedPage(message, "123", "marketplace_messages"); err != nil || len(got) != 1 {
		t.Fatalf("bare message ID: %v", err)
	}
	for _, raw := range []string{
		`{}`, `{"messages":null}`,
		`{"messages":[{"_id":"x","application_id":"123","user_id":"456","topic":"marketplace_messages","resource":"https://example.invalid/aaaaaaaaaaaaaaaa","received":"2026-11-20T00:00:00Z"}]}`,
		`{"messages":[{"_id":"x","application_id":"124","user_id":"456","topic":"marketplace_messages","resource":"aaaaaaaaaaaaaaaa","received":"2026-11-20T00:00:00Z"}]}`,
	} {
		if _, err := decodeMissedFeedPage([]byte(raw), "123", "marketplace_messages"); err == nil {
			t.Fatal("invalid replay accepted")
		}
	}
}

func TestMissedFeedPageCheckpointPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	const topic = "marketplace_items"
	mustExec(t, ctx, p, `DELETE FROM app.meli_missed_feed_replay WHERE application_id=$1 AND topic=$2`, testApp, topic)
	defer mustExec(t, ctx, p, `DELETE FROM app.meli_missed_feed_replay WHERE application_id=$1 AND topic=$2`, testApp, topic)
	fetch := func(_ context.Context, gotTopic string, offset int) ([]byte, error) {
		if gotTopic != topic || offset != 0 {
			t.Fatal("unexpected starting cursor")
		}
		return missedItemPage(testChildA, MissedFeedPageSize), nil
	}
	first, err := p.ReplayMissedFeedPage(ctx, testApp, topic, fetch)
	if err != nil || first.NextOffset != 8 || first.Records != 10 || first.CycleComplete {
		t.Fatalf("first page: %+v %v", first, err)
	}
	assertCount(t, ctx, p, 10, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='missed_feed'`, testStoreA)
	// Replay contains an unknown seller: fail closed and leave checkpoint.
	_, err = p.ReplayMissedFeedPage(ctx, testApp, topic, func(_ context.Context, _ string, offset int) ([]byte, error) {
		if offset != 8 {
			t.Fatal("did not resume")
		}
		return missedItemPage("999999999", 1), nil
	})
	if !errors.Is(err, ErrScope) {
		t.Fatalf("unknown scope accepted: %v", err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_missed_feed_replay WHERE application_id=$1 AND topic=$2 AND next_offset=8 AND revision=1`, testApp, topic)
	last, err := p.ReplayMissedFeedPage(ctx, testApp, topic, func(_ context.Context, _ string, offset int) ([]byte, error) {
		if offset != 8 {
			t.Fatal("checkpoint changed on error")
		}
		return missedItemPage(testChildA, 2), nil
	})
	if err != nil || !last.CycleComplete || last.NextOffset != 0 {
		t.Fatalf("finish page: %+v %v", last, err)
	}
	assertCount(t, ctx, p, 10, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='missed_feed'`, testStoreA)
	assertCount(t, ctx, p, 10, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND generation=1`, testStoreA)
}

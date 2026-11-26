package notifications

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const MissedFeedPageSize = 10

// Replay only the official top-level envelope. request.data, URLs and headers
// are not decoded or trusted. Keep the live receiver's path rules unchanged.
func decodeMissedFeedPage(raw []byte, appID, topic string) ([]Event, error) {
	bad := errors.New("invalid missed feed page")
	prefix, supported := topicPaths[topic]
	if !supported || !idPattern.MatchString(appID) {
		return nil, bad
	}
	var page struct {
		Messages []struct {
			ID            string          `json:"_id"`
			ApplicationID json.RawMessage `json:"application_id"`
			UserID        json.RawMessage `json:"user_id"`
			Topic         string          `json:"topic"`
			Resource      string          `json:"resource"`
			Received      string          `json:"received"`
		} `json:"messages"`
	}
	if len(raw) > 2<<20 || json.Unmarshal(raw, &page) != nil || page.Messages == nil || len(page.Messages) > MissedFeedPageSize {
		return nil, bad
	}
	events := make([]Event, 0, len(page.Messages))
	for _, row := range page.Messages {
		user := id(row.UserID)
		if id(row.ApplicationID) != appID || user == "" || row.Topic != topic ||
			row.ID == "" || len(row.ID) > 256 {
			return nil, bad
		}
		// Only two separately documented/observed history aliases are allowed:
		// bare message ID, and marketplace/resource/items/ID/ (one slash).
		resource, ok := strings.CutPrefix(row.Resource, prefix)
		if !ok && topic == "marketplace_messages" {
			resource, ok = row.Resource, true
		}
		if !ok && topic == "marketplace_items" {
			if rest, matched := strings.CutPrefix(row.Resource, "/marketplace/resource/items/"); matched {
				resource, ok = strings.CutSuffix(rest, "/")
			}
		}
		received, err := time.Parse(time.RFC3339Nano, row.Received)
		if !ok || !ValidResourceID(topic, resource) || err != nil || received.IsZero() {
			return nil, bad
		}
		// Share receipt identity with realtime delivery when the official ID
		// matches exactly. If the provider uses another ID, active-job
		// coalescing and verified GET/upsert still make replay safe.
		hash := sha256.Sum256([]byte("id:" + row.ID))
		events = append(events, Event{ApplicationID: appID, UserID: user, Topic: topic, ResourceID: resource,
			Key: hex.EncodeToString(hash[:]), ReceivedAt: received.UTC()})
	}
	return events, nil
}

type ReplayProgress struct {
	Offset, NextOffset, Records int
	CycleComplete               bool
}

type MissedFeedFetch func(context.Context, string, int) ([]byte, error)

// ReplayMissedFeedPage is an operator-driven bounded page, not a scheduler.
// The credential store scopes the cursor; each event is routed by its own identity.
// No provider call holds a DB transaction. A revision fence prevents concurrent
// invocations from advancing a stale page. On any malformed/unmapped record,
// backpressure or commit failure, the page and checkpoint roll back together.
func (p PostgresInbox) ReplayMissedFeedPage(ctx context.Context, appID, topic, credentialStoreID string, fetch MissedFeedFetch) (ReplayProgress, error) {
	var result ReplayProgress
	if !idPattern.MatchString(appID) || !ValidReplayTopic(topic) || fetch == nil {
		return result, errors.New("invalid missed feed replay configuration")
	}
	var revision int64
	err := p.DB.QueryRow(ctx, `SELECT next_offset,revision FROM app.meli_missed_feed_replay
		WHERE application_id=$1 AND topic=$2 AND credential_store_id=$3::uuid`, appID, topic, credentialStoreID).Scan(&result.Offset, &revision)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return result, err
	}
	raw, err := fetch(ctx, topic, result.Offset)
	if err != nil {
		return result, err
	}
	events, err := decodeMissedFeedPage(raw, appID, topic)
	if err != nil {
		return result, err
	}
	next := 0
	complete := len(events) < MissedFeedPageSize
	if !complete {
		// Two-record overlap reduces (but cannot eliminate) offset drift.
		next = result.Offset + MissedFeedPageSize - 2
		if next > 100000 {
			return result, errors.New("missed feed offset requires operator review")
		}
	}
	err = pgx.BeginFunc(ctx, p.DB, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `SET LOCAL synchronous_commit=on`); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(queueAdmissionLock)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO app.meli_missed_feed_replay(application_id,topic,credential_store_id)
			VALUES($1,$2,$3::uuid) ON CONFLICT DO NOTHING`, appID, topic, credentialStoreID); err != nil {
			return err
		}
		var offset int
		var current int64
		if err := tx.QueryRow(ctx, `SELECT next_offset,revision FROM app.meli_missed_feed_replay
			WHERE application_id=$1 AND topic=$2 AND credential_store_id=$3::uuid FOR UPDATE`, appID, topic, credentialStoreID).Scan(&offset, &current); err != nil {
			return err
		}
		if offset != result.Offset || current != revision {
			return errors.New("missed feed checkpoint changed")
		}
		var pending int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM app.meli_notification_jobs
			WHERE application_id=$1 AND dead_at IS NULL`, appID).Scan(&pending); err != nil {
			return err
		}
		if pending+len(events) > 32 {
			return ErrCapacity
		}
		for _, event := range events {
			if err := p.enqueueInTx(ctx, tx, event, nil, nil, "missed_feed"); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `UPDATE app.meli_missed_feed_replay SET next_offset=$4,
			revision=revision+1,updated_at=now() WHERE application_id=$1 AND topic=$2 AND credential_store_id=$3::uuid`, appID, topic, credentialStoreID, next)
		return err
	})
	if err != nil {
		return result, err
	}
	result.NextOffset, result.Records, result.CycleComplete = next, len(events), complete
	return result, nil
}

func ValidReplayTopic(topic string) bool {
	_, ok := topicPaths[topic]
	return ok
}

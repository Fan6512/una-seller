package notifications

import (
	"context"
	"errors"
	"strings"
)

// Explicit opt-in only: refuse empty, unknown, duplicate or whitespace-only
// topic names. No wildcard and no implicit expansion during a release.
func ParseEnabledTopics(value string) ([]string, error) {
	var topics []string
	seen := make(map[string]bool)
	for _, part := range strings.Split(value, ",") {
		topic := strings.TrimSpace(part)
		if _, exists := topicPaths[topic]; !exists || seen[topic] {
			return nil, errors.New("invalid enabled notification topic set")
		}
		seen[topic] = true
		topics = append(topics, topic)
	}
	return topics, nil
}

// Preflight detects unapplied schema without automatically migrating or writing.
// It is not a substitute for the migration receipt and production smoke test.
func (p PostgresInbox) CheckActivationSchema(ctx context.Context) error {
	var ready bool
	err := p.DB.QueryRow(ctx, `SELECT
		to_regclass('app.meli_notification_inbox') IS NOT NULL
		AND to_regclass('app.meli_notification_jobs') IS NOT NULL
		AND to_regclass('app.meli_notification_cooldowns') IS NOT NULL
		AND to_regclass('app.meli_order_read_coordination') IS NOT NULL
		AND to_regclass('app.meli_customer_notifications') IS NOT NULL
		AND to_regclass('app.meli_claim_notifications') IS NOT NULL
		AND to_regclass('app.meli_item_notifications') IS NOT NULL
		AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='app'
		  AND table_name='meli_notification_inbox' AND column_name='source')`).Scan(&ready)
	if err != nil {
		return err
	}
	if !ready {
		return errors.New("notification schema migrations required")
	}
	return nil
}

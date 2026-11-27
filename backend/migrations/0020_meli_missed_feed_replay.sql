-- WH-06: provenance and private replay checkpoints per credential store.
-- The cursor is only advanced in the transaction that admits the page.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint
        WHERE conrelid='app.meli_notification_inbox'::regclass
          AND conname='meli_notification_inbox_wh06_source_check') THEN
        ALTER TABLE app.meli_notification_inbox
            DROP CONSTRAINT IF EXISTS meli_notification_inbox_source_check;
        ALTER TABLE app.meli_notification_inbox
            ADD CONSTRAINT meli_notification_inbox_wh06_source_check
            CHECK (source IN ('webhook','compensation','missed_feed'));
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS app.meli_missed_feed_replay (
    application_id text NOT NULL CHECK (application_id ~ '^[1-9][0-9]{0,29}$'),
    topic text NOT NULL CHECK (topic IN ('marketplace_orders','marketplace_shipments',
        'marketplace_questions','marketplace_messages','marketplace_claims','marketplace_items')),
    credential_store_id uuid NOT NULL,
    next_offset integer NOT NULL DEFAULT 0 CHECK (next_offset BETWEEN 0 AND 100000),
    revision bigint NOT NULL DEFAULT 0 CHECK (revision>=0),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY(application_id,topic,credential_store_id)
);
ALTER TABLE app.meli_missed_feed_replay ENABLE ROW LEVEL SECURITY;

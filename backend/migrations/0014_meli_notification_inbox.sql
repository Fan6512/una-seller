-- OPS-01 / WH-01: private, durable notification inbox. No raw payloads.
-- A delivery is a hint. A handler must reauthorize and GET official data.
CREATE TABLE IF NOT EXISTS app.meli_notification_inbox (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    parent_seller_id text NOT NULL,
    application_id text NOT NULL,
    notification_user_id text NOT NULL,
    topic text NOT NULL CHECK (topic IN ('marketplace_orders','marketplace_shipments')),
    resource_id text NOT NULL CHECK (resource_id ~ '^[1-9][0-9]{0,29}$'),
    event_key text NOT NULL CHECK (length(event_key)=64),
    source_received_at timestamptz NOT NULL,
    enqueued_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (store_id,owner_user_id,application_id,notification_user_id,topic,resource_id,event_key)
);
CREATE INDEX IF NOT EXISTS meli_notification_inbox_pending_idx
    ON app.meli_notification_inbox(enqueued_at,id);
ALTER TABLE app.meli_notification_inbox ENABLE ROW LEVEL SECURITY;

-- Backfill once if an isolated environment already applied the original WH-01
-- schema. Reapplying 0014 must not recreate completed jobs from dedup receipts.
DO $$
BEGIN
    IF to_regclass('app.meli_notification_jobs') IS NULL THEN
        CREATE TABLE app.meli_notification_jobs (
            id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
            store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
            owner_user_id uuid NOT NULL REFERENCES app.users(id),
            parent_seller_id text NOT NULL,
            application_id text NOT NULL,
            notification_user_id text NOT NULL,
            topic text NOT NULL CHECK (topic IN ('marketplace_orders','marketplace_shipments')),
            resource_id text NOT NULL CHECK (resource_id ~ '^[1-9][0-9]{0,29}$'),
            generation bigint NOT NULL DEFAULT 1,
            pending_since timestamptz NOT NULL DEFAULT clock_timestamp(),
            available_at timestamptz NOT NULL DEFAULT clock_timestamp(),
            attempts integer NOT NULL DEFAULT 0 CHECK (attempts BETWEEN 0 AND 8),
            lease_token uuid,
            lease_until timestamptz,
            dead_at timestamptz,
            last_failure text,
            CHECK ((lease_token IS NULL) = (lease_until IS NULL)),
            CHECK (dead_at IS NULL OR lease_token IS NULL)
        );
        INSERT INTO app.meli_notification_jobs
            (store_id, owner_user_id, parent_seller_id, application_id, notification_user_id,
             topic, resource_id, pending_since)
        SELECT DISTINCT ON (store_id, owner_user_id, parent_seller_id, application_id, topic, resource_id)
            store_id, owner_user_id, parent_seller_id, application_id, notification_user_id,
            topic, resource_id, enqueued_at
        FROM app.meli_notification_inbox
        ORDER BY store_id, owner_user_id, parent_seller_id, application_id, topic, resource_id, enqueued_at DESC, id DESC;
    END IF;
END $$;
CREATE UNIQUE INDEX IF NOT EXISTS meli_notification_jobs_active_idx
    ON app.meli_notification_jobs (store_id, owner_user_id, parent_seller_id, application_id, topic, resource_id)
    WHERE dead_at IS NULL;
CREATE INDEX IF NOT EXISTS meli_notification_jobs_claim_idx
    ON app.meli_notification_jobs (application_id, topic, available_at, pending_since)
    WHERE dead_at IS NULL;
CREATE INDEX IF NOT EXISTS meli_notification_jobs_dead_idx
    ON app.meli_notification_jobs (dead_at) WHERE dead_at IS NOT NULL;
ALTER TABLE app.meli_notification_jobs ENABLE ROW LEVEL SECURITY;

CREATE TABLE IF NOT EXISTS app.meli_notification_cooldowns (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    parent_seller_id text NOT NULL,
    application_id text NOT NULL,
    available_at timestamptz NOT NULL,
    PRIMARY KEY (store_id, owner_user_id, parent_seller_id, application_id)
);
CREATE INDEX IF NOT EXISTS meli_notification_cooldowns_expiry_idx
    ON app.meli_notification_cooldowns (available_at);
ALTER TABLE app.meli_notification_cooldowns ENABLE ROW LEVEL SECURITY;

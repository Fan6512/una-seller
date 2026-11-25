-- OPS-01 / WH-01: private, durable notification inbox. No raw payloads.
-- Delivery is a hint only; a future worker must reauthorize and GET official data.
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

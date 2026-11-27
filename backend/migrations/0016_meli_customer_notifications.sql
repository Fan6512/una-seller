-- WH-04: expand the durable inbox to two documented customer topics.
-- Keep opaque message identifiers bounded and path-safe.
DO $$
DECLARE tbl text;
BEGIN
    FOREACH tbl IN ARRAY ARRAY['meli_notification_inbox','meli_notification_jobs'] LOOP
        IF NOT EXISTS (SELECT 1 FROM pg_constraint
            WHERE conrelid=to_regclass('app.' || tbl) AND conname IN (tbl || '_wh04_resource_check',tbl || '_wh05_resource_check')) THEN
            EXECUTE format('ALTER TABLE app.%I DROP CONSTRAINT IF EXISTS %I',tbl,tbl || '_topic_check');
            EXECUTE format('ALTER TABLE app.%I DROP CONSTRAINT IF EXISTS %I',tbl,tbl || '_resource_id_check');
            EXECUTE format('ALTER TABLE app.%I ADD CONSTRAINT %I CHECK (
                (topic IN (''marketplace_orders'',''marketplace_shipments'',''marketplace_questions'') AND resource_id ~ ''^[1-9][0-9]{0,29}$'')
                OR (topic=''marketplace_messages'' AND resource_id ~ ''^[A-Za-z0-9]{16,64}$''))',tbl,tbl || '_wh04_resource_check');
        END IF;
    END LOOP;
END $$;

-- Customer module owns its projection. No raw provider JSON, attachments,
-- buyer account IDs/names, authorization tokens or official write actions.
CREATE TABLE IF NOT EXISTS app.meli_customer_notifications (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    parent_seller_id text NOT NULL,
    seller_id text NOT NULL,
    site_id text NOT NULL,
    topic text NOT NULL CHECK (topic IN ('marketplace_questions','marketplace_messages')),
    resource_id text NOT NULL,
    related_id text NOT NULL DEFAULT '',
    status text NOT NULL,
    body text NOT NULL DEFAULT '' CHECK (octet_length(body)<=16384),
    translated_body text NOT NULL DEFAULT '' CHECK (octet_length(translated_body)<=16384),
    answer text NOT NULL DEFAULT '' CHECK (octet_length(answer)<=16384),
    direction text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL,
    checked_at timestamptz NOT NULL,
    PRIMARY KEY(store_id,site_id,seller_id,topic,resource_id)
);
CREATE INDEX IF NOT EXISTS meli_customer_notifications_owner_checked_idx
    ON app.meli_customer_notifications(owner_user_id,checked_at DESC);
ALTER TABLE app.meli_customer_notifications ENABLE ROW LEVEL SECURITY;

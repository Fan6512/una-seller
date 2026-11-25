-- WH-05: expand queue checks without modifying historical migrations.
DO $$
DECLARE tbl text;
BEGIN
    FOREACH tbl IN ARRAY ARRAY['meli_notification_inbox','meli_notification_jobs'] LOOP
        IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('app.' || tbl)
            AND conname=tbl || '_wh05_resource_check') THEN
            EXECUTE format('ALTER TABLE app.%I DROP CONSTRAINT IF EXISTS %I',tbl,tbl || '_wh04_resource_check');
            EXECUTE format('ALTER TABLE app.%I ADD CONSTRAINT %I CHECK (
                (topic IN (''marketplace_orders'',''marketplace_shipments'',''marketplace_questions'',''marketplace_claims'') AND resource_id ~ ''^[1-9][0-9]{0,29}$'')
                OR (topic=''marketplace_messages'' AND resource_id ~ ''^[A-Za-z0-9]{16,64}$'')
                OR (topic=''marketplace_items'' AND resource_id ~ ''^[A-Z]{3}[1-9][0-9]{0,29}$'' AND left(resource_id,3)<>''CBT''))',tbl,tbl || '_wh05_resource_check');
        END IF;
    END LOOP;
END $$;
CREATE TABLE IF NOT EXISTS app.meli_claim_notifications (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    parent_seller_id text NOT NULL,
    seller_id text NOT NULL,
    site_id text NOT NULL,
    claim_id text NOT NULL,
    resource_type text NOT NULL,
    related_id text NOT NULL,
    status text NOT NULL,
    stage text NOT NULL,
    claim_type text NOT NULL,
    reason_code text NOT NULL,
    due_at timestamptz,
    allowed_actions text NOT NULL DEFAULT '' CHECK (length(allowed_actions)<=512),
    last_updated timestamptz NOT NULL,
    checked_at timestamptz NOT NULL,
    PRIMARY KEY(store_id,site_id,seller_id,claim_id)
);
ALTER TABLE app.meli_claim_notifications ADD COLUMN IF NOT EXISTS due_at timestamptz;
ALTER TABLE app.meli_claim_notifications ADD COLUMN IF NOT EXISTS allowed_actions text NOT NULL DEFAULT '';
CREATE TABLE IF NOT EXISTS app.meli_item_notifications (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    parent_seller_id text NOT NULL,
    seller_id text NOT NULL,
    site_id text NOT NULL,
    item_id text NOT NULL,
    status text NOT NULL,
    sub_status text NOT NULL DEFAULT '' CHECK (length(sub_status)<=160),
    price numeric(18,6) CHECK (price>=0),
    available_quantity integer CHECK (available_quantity>=0),
    last_updated timestamptz NOT NULL,
    checked_at timestamptz NOT NULL,
    PRIMARY KEY(store_id,site_id,seller_id,item_id)
);
ALTER TABLE app.meli_item_notifications ADD COLUMN IF NOT EXISTS sub_status text NOT NULL DEFAULT '';
ALTER TABLE app.meli_item_notifications ADD COLUMN IF NOT EXISTS price numeric(18,6);
CREATE INDEX IF NOT EXISTS meli_claim_notifications_owner_checked_idx ON app.meli_claim_notifications(owner_user_id,checked_at DESC);
CREATE INDEX IF NOT EXISTS meli_item_notifications_owner_checked_idx ON app.meli_item_notifications(owner_user_id,checked_at DESC);
ALTER TABLE app.meli_claim_notifications ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_item_notifications ENABLE ROW LEVEL SECURITY;

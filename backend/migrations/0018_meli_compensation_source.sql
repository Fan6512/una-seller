-- OPS-03: distinguish internal repair from actual webhook receipts.
ALTER TABLE app.meli_notification_inbox ADD COLUMN IF NOT EXISTS source
    text NOT NULL DEFAULT 'webhook' CHECK (source IN ('webhook','compensation'));
CREATE INDEX IF NOT EXISTS meli_notification_inbox_recent_resource_idx
    ON app.meli_notification_inbox
    (application_id,store_id,owner_user_id,parent_seller_id,topic,resource_id,enqueued_at DESC);

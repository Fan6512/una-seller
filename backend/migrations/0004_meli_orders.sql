-- M3-ORDERS-001. Private, service-owned CBT orders: no buyer PII or raw payloads.
CREATE TABLE IF NOT EXISTS app.meli_orders (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    order_id text NOT NULL CHECK (order_id ~ '^[0-9]{1,30}$'),
    pack_id text,
    status text NOT NULL,
    currency_id text NOT NULL,
    total_amount numeric(24,8) NOT NULL,
    date_created timestamptz NOT NULL,
    last_updated timestamptz NOT NULL,
    shipment_id text,
    order_items jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(order_items) = 'array'),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (store_id,site_id,logistic_type,order_id)
);

CREATE INDEX IF NOT EXISTS meli_orders_page_idx
ON app.meli_orders (store_id,site_id,logistic_type,date_created DESC,order_id DESC);

CREATE TABLE IF NOT EXISTS app.meli_order_syncs (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    window_from timestamptz,
    window_to timestamptz,
    next_offset integer NOT NULL DEFAULT 0 CHECK (next_offset >= 0),
    last_completed_to timestamptz,
    last_synced_at timestamptz,
    lease_id uuid,
    lease_until timestamptz,
    PRIMARY KEY (store_id,site_id,logistic_type),
    CHECK ((window_from IS NULL) = (window_to IS NULL))
);

-- Expand existing 0004 installations without discarding orders or cursors.
ALTER TABLE app.meli_order_syncs
    ADD COLUMN IF NOT EXISTS pending_order_ids text[] NOT NULL DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS pending_page_count integer NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pending_has_more boolean NOT NULL DEFAULT false;

ALTER TABLE app.meli_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_order_syncs ENABLE ROW LEVEL SECURITY;

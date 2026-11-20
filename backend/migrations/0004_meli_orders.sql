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
    PRIMARY KEY (store_id,site_id,logistic_type,seller_id,order_id)
);

-- Upgrade the earlier Draft key so overlapping IDs can coexist across child sellers.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint
        WHERE conrelid='app.meli_orders'::regclass AND conname='meli_orders_pkey'
        AND (SELECT array_agg(a.attname ORDER BY col.ord)
             FROM unnest(conkey) WITH ORDINALITY AS col(attnum,ord)
             JOIN pg_attribute a ON a.attrelid='app.meli_orders'::regclass AND a.attnum=col.attnum)
            = ARRAY['store_id','site_id','logistic_type','order_id']::name[]) THEN
        ALTER TABLE app.meli_orders DROP CONSTRAINT meli_orders_pkey,
            ADD PRIMARY KEY (store_id,site_id,logistic_type,seller_id,order_id);
    END IF;
END $$;

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
    ADD COLUMN IF NOT EXISTS pending_has_more boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS review_id uuid,
    ADD COLUMN IF NOT EXISTS review_reason text CHECK (
        (review_id IS NULL AND review_reason IS NULL) OR
        (review_id IS NOT NULL AND review_reason IS NOT NULL
            AND review_reason IN ('invalid_detail','detail_not_found') AND pending_order_ids[1] IS NOT NULL)
    );

CREATE TABLE IF NOT EXISTS app.meli_order_skips (
    review_id uuid PRIMARY KEY,
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    order_id text NOT NULL CHECK (order_id ~ '^[0-9]{1,30}$'),
    search_offset integer NOT NULL CHECK (search_offset >= 0),
    reason text NOT NULL CHECK (reason IN ('invalid_detail','detail_not_found')),
    skipped_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS meli_order_skips_scope_lookup_idx
ON app.meli_order_skips (store_id,site_id,logistic_type,seller_id,order_id,reason);

-- An explicit rewind is auditable, but does not erase orders or skip decisions.
CREATE TABLE IF NOT EXISTS app.meli_order_resets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    previous_offset integer NOT NULL CHECK (previous_offset > 0),
    window_from timestamptz NOT NULL,
    window_to timestamptz NOT NULL,
    reset_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE app.meli_orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_order_syncs ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_order_skips ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_order_resets ENABLE ROW LEVEL SECURITY;

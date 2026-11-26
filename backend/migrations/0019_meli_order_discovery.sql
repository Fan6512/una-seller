-- OPS-03: one private search checkpoint per exact current seller scope.
-- No dependency on the manual historical order sync cursor.
CREATE TABLE IF NOT EXISTS app.meli_order_discovery (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    parent_seller_id text NOT NULL,
    application_id text NOT NULL,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    completed_to timestamptz NOT NULL DEFAULT date_trunc('second',now())-interval '24 hours',
    window_from timestamptz,
    window_to timestamptz,
    next_offset integer NOT NULL DEFAULT 0 CHECK (next_offset>=0),
    pending_ids text[] NOT NULL DEFAULT '{}' CHECK (cardinality(pending_ids)<=120),
    pending_count integer NOT NULL DEFAULT 0 CHECK (pending_count BETWEEN 0 AND 3),
    pending_more boolean NOT NULL DEFAULT false,
    next_due_at timestamptz NOT NULL DEFAULT now(),
    revision bigint NOT NULL DEFAULT 0 CHECK (revision>=0),
    last_completed_at timestamptz,
    PRIMARY KEY(store_id,owner_user_id,parent_seller_id,application_id,site_id,logistic_type,seller_id),
    CHECK ((window_from IS NULL)=(window_to IS NULL)),
    CHECK (window_from IS NULL OR window_from<window_to),
    CHECK ((cardinality(pending_ids)>0)=(pending_count>0))
);
CREATE INDEX IF NOT EXISTS meli_order_discovery_due_idx
ON app.meli_order_discovery(application_id,next_due_at);
ALTER TABLE app.meli_order_discovery ENABLE ROW LEVEL SECURITY;

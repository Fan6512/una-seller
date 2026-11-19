CREATE TABLE IF NOT EXISTS app.meli_published_listings (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    listing_id text NOT NULL,
    details jsonb NOT NULL,
    PRIMARY KEY (store_id, site_id, logistic_type, listing_id)
);

CREATE TABLE IF NOT EXISTS app.meli_listing_syncs (
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    synced_at timestamptz NOT NULL,
    PRIMARY KEY (store_id, site_id, logistic_type)
);

ALTER TABLE app.meli_published_listings ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_listing_syncs ENABLE ROW LEVEL SECURITY;

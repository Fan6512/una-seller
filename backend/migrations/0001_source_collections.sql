-- M0-03 PR-A: private source-collection data, accessed only by the Go API.
-- Apply only after approval: psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f backend/migrations/0001_source_collections.sql
CREATE SCHEMA IF NOT EXISTS app;
REVOKE ALL ON SCHEMA app FROM PUBLIC;

CREATE TABLE IF NOT EXISTS app.users (
    id uuid PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS app.source_products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    source_platform text NOT NULL CHECK (source_platform IN ('amazon', 'mercadolibre')),
    source_site text NOT NULL,
    source_url text NOT NULL,
    normalized_url text NOT NULL,
    title text NOT NULL CHECK (length(title) BETWEEN 1 AND 500),
    price numeric(20,4),
    currency text,
    category text,
    description text,
    bullet_points jsonb NOT NULL DEFAULT '[]'::jsonb,
    collection_status text NOT NULL CHECK (collection_status IN ('complete', 'partial')),
    collection_warnings jsonb NOT NULL DEFAULT '[]'::jsonb,
    first_collected_at timestamptz NOT NULL DEFAULT now(),
    last_collected_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE(owner_user_id, source_platform, source_site, normalized_url)
);

CREATE TABLE IF NOT EXISTS app.source_product_variants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_product_id uuid NOT NULL REFERENCES app.source_products(id) ON DELETE CASCADE,
    variant_key text NOT NULL,
    attributes jsonb NOT NULL CHECK (jsonb_typeof(attributes) = 'object'),
    price numeric(20,4),
    currency text,
    sort_order integer NOT NULL CHECK (sort_order >= 0),
    UNIQUE(source_product_id, variant_key)
);

CREATE TABLE IF NOT EXISTS app.source_product_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_product_id uuid NOT NULL REFERENCES app.source_products(id) ON DELETE CASCADE,
    variant_id uuid REFERENCES app.source_product_variants(id) ON DELETE CASCADE,
    image_url text NOT NULL,
    image_order integer NOT NULL CHECK (image_order >= 0),
    image_type text NOT NULL CHECK (image_type IN ('main', 'additional', 'variant'))
);
CREATE INDEX IF NOT EXISTS source_products_recent_idx ON app.source_products(owner_user_id, last_collected_at DESC, id);
CREATE INDEX IF NOT EXISTS source_images_product_idx ON app.source_product_images(source_product_id, variant_id, image_order);
CREATE INDEX IF NOT EXISTS source_variants_product_idx ON app.source_product_variants(source_product_id, sort_order);

-- These are private Go-API tables, not Supabase REST tables. No anon/authenticated grants.
ALTER TABLE app.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.source_products ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.source_product_variants ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.source_product_images ENABLE ROW LEVEL SECURITY;

-- M1-A/2: a product-box draft is a standalone immutable snapshot of source facts
-- plus independently editable merchant fields. Never mutate a captured source.
CREATE TABLE IF NOT EXISTS app.own_products (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_user_id uuid NOT NULL REFERENCES app.users(id),
  source_product_id uuid UNIQUE REFERENCES app.source_products(id) ON DELETE SET NULL,
  source_platform text NOT NULL,
  source_site text NOT NULL,
  source_url text NOT NULL,
  source_title text NOT NULL,
  title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 500),
  description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 20000),
  notes text NOT NULL DEFAULT '' CHECK (char_length(notes) <= 5000),
  price numeric(20,4),
  currency text,
  category text,
  bullet_points jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(bullet_points)='array'),
  images jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(images)='array'),
  variants jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(variants)='array'),
  revision bigint NOT NULL DEFAULT 1 CHECK (revision >= 1),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS own_products_owner_recent_idx
  ON app.own_products (owner_user_id,updated_at DESC,id DESC);
ALTER TABLE app.own_products ENABLE ROW LEVEL SECURITY;
-- No anon/authenticated grants: service-owned private app schema.

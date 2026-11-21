-- M1-A/3: additive editing payload, shared by pre-approval working drafts and
-- independent product-box drafts. The existing source snapshots are immutable.
ALTER TABLE app.collection_working_drafts
  ADD COLUMN IF NOT EXISTS listing_fields jsonb NOT NULL DEFAULT '{}'::jsonb
  CHECK (jsonb_typeof(listing_fields)='object');
ALTER TABLE app.own_products
  ADD COLUMN IF NOT EXISTS listing_fields jsonb NOT NULL DEFAULT '{}'::jsonb
  CHECK (jsonb_typeof(listing_fields)='object');

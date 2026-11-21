-- M1-A/1: own working copy for a captured source; never rewrite source data.
CREATE TABLE IF NOT EXISTS app.collection_working_drafts (
    source_product_id uuid PRIMARY KEY REFERENCES app.source_products(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 500),
    description text NOT NULL DEFAULT '' CHECK (char_length(description) <= 20000),
    notes text NOT NULL DEFAULT '' CHECK (char_length(notes) <= 5000),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision >= 1),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS collection_working_drafts_owner_idx
    ON app.collection_working_drafts(owner_user_id, updated_at DESC);
ALTER TABLE app.collection_working_drafts ENABLE ROW LEVEL SECURITY;

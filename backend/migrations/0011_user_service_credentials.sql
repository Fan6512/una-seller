-- P0 SEC-KEY-001: private, owner-scoped encrypted provider keys.
-- Explicit api migrate only; no automatic production schema changes.
-- Reserve 0010 for Draft PR #66 shipment. 0011 can be applied independently.
CREATE TABLE IF NOT EXISTS app.user_service_credentials (
    owner_user_id uuid NOT NULL REFERENCES app.users(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('ai', 'translation')),
    provider text NOT NULL CHECK (char_length(provider) BETWEEN 1 AND 64),
    base_url text NOT NULL DEFAULT '' CHECK (char_length(base_url) <= 500),
    model text NOT NULL DEFAULT '' CHECK (char_length(model) <= 128),
    region text NOT NULL DEFAULT '' CHECK (char_length(region) <= 80),
    app_id text NOT NULL DEFAULT '' CHECK (char_length(app_id) <= 80),
    key_id text NOT NULL DEFAULT 'v1' CHECK (char_length(key_id) BETWEEN 1 AND 64),
    encrypted_data bytea NOT NULL CHECK (octet_length(encrypted_data) BETWEEN 32 AND 8192),
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (owner_user_id, kind)
);
-- Safe additive upgrade for any already-applied Draft table.
ALTER TABLE app.user_service_credentials
    ADD COLUMN IF NOT EXISTS app_id text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS key_id text NOT NULL DEFAULT 'v1';
ALTER TABLE app.user_service_credentials ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON app.user_service_credentials FROM PUBLIC;
-- The Go service owns these private records and verifies Supabase owner
-- on every request. No Data API access, SECURITY DEFINER, or broad policies.

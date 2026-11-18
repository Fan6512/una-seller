-- Issue #21 / M0-ML-AUTH: Mercado Libre CBT OAuth account connections.
-- Applied by backend/migrations/migrations.go; this is a private schema, never Data API exposed.
CREATE TABLE IF NOT EXISTS app.meli_oauth_states (
    state_hash bytea PRIMARY KEY CHECK (octet_length(state_hash) = 32),
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    pkce_verifier_ciphertext bytea,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS meli_oauth_states_expires_idx ON app.meli_oauth_states(expires_at);

CREATE TABLE IF NOT EXISTS app.meli_store_connections (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    -- The unique seller identity comes only from the authenticated Mercado Libre /users/me response.
    -- seller_id is the CBT PARENT / global user_id; site-specific child IDs stay in marketplaces.
    seller_id text NOT NULL UNIQUE CHECK (length(seller_id) BETWEEN 1 AND 64),
    marketplaces jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(marketplaces) = 'array'),
    marketplaces_status text NOT NULL DEFAULT 'unverified'
        CHECK (marketplaces_status IN ('verified', 'unverified')),
    marketplaces_checked_at timestamptz,
    nickname text,
    status text NOT NULL DEFAULT 'connected'
        CHECK (status IN ('connected', 'reauth_required', 'refresh_unknown', 'disconnected')),
    access_token_ciphertext bytea,
    refresh_token_ciphertext bytea,
    access_token_expires_at timestamptz,
    refresh_lease_until timestamptz,
    refresh_lock uuid,
    last_checked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (status <> 'connected' OR (
        access_token_ciphertext IS NOT NULL AND
        refresh_token_ciphertext IS NOT NULL AND
        access_token_expires_at IS NOT NULL
    ))
);
-- Also upgrade a database that applied the early Draft #23 table before these
-- CBT hierarchy fields were added. CREATE TABLE IF NOT EXISTS alone would skip them.
ALTER TABLE app.meli_store_connections
    ADD COLUMN IF NOT EXISTS marketplaces jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(marketplaces) = 'array'),
    ADD COLUMN IF NOT EXISTS marketplaces_status text NOT NULL DEFAULT 'unverified'
        CHECK (marketplaces_status IN ('verified', 'unverified')),
    ADD COLUMN IF NOT EXISTS marketplaces_checked_at timestamptz;

CREATE INDEX IF NOT EXISTS meli_store_owner_idx
    ON app.meli_store_connections(owner_user_id, updated_at DESC, id);

-- The Go service owns access to these private tables; no anon/authenticated grants or policies.
ALTER TABLE app.meli_oauth_states ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.meli_store_connections ENABLE ROW LEVEL SECURITY;

-- CBT C0: owner-scoped, sanitized publication model evidence; not publishing authorization.
-- 0010 is reserved for Draft #66. This is additive, applied only after approval.
-- Disconnect may retain historical evidence. Every consumer must join the owner
-- and store identity to app.meli_store_connections and require status = 'connected'
-- before treating evidence as current; even then it is not write authorization.
CREATE TABLE IF NOT EXISTS app.meli_cbt_c0_evidence (
    store_id uuid PRIMARY KEY REFERENCES app.meli_store_connections(id) ON DELETE CASCADE,
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    publication_mode text NOT NULL
        CHECK (publication_mode IN ('fully_managed', 'standard_up', 'legacy_candidate', 'undetermined')),
    evidence jsonb NOT NULL CHECK (jsonb_typeof(evidence) = 'object'),
    checked_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS meli_cbt_c0_evidence_owner_idx
    ON app.meli_cbt_c0_evidence(owner_user_id, checked_at DESC);
ALTER TABLE app.meli_cbt_c0_evidence ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON app.meli_cbt_c0_evidence FROM PUBLIC;
-- Supabase anon/authenticated roles have no grants on this private schema;
-- do not reference platform-only roles in generic PostgreSQL CI.

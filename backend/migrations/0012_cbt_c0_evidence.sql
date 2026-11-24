-- CBT C0: owner-scoped, sanitized publication model evidence; not publishing authorization.
-- 0010 is reserved for Draft #66. This is additive, applied only after approval.
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
REVOKE ALL ON app.meli_cbt_c0_evidence FROM PUBLIC, anon, authenticated;

-- WH-03: application-wide pacing and cooldown for manual and background reads.
-- Private service metadata only; no OAuth tokens or resource identifiers.
CREATE TABLE IF NOT EXISTS app.meli_order_read_coordination (
    application_id text PRIMARY KEY,
    next_request_at timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    next_search_at timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00',
    retry_until timestamptz NOT NULL DEFAULT '1970-01-01 00:00:00+00'
);
ALTER TABLE app.meli_order_read_coordination ENABLE ROW LEVEL SECURITY;

-- Public reference data, not an owner ledger or settlement FX. One immutable
-- first accepted quote per Beijing order creation day; no daily revaluation.
CREATE TABLE IF NOT EXISTS app.daily_usd_cny_rates (
    accounting_day date PRIMARY KEY,
    observation_day date NOT NULL,
    raw_rate numeric NOT NULL CHECK (raw_rate >= 0.1 AND raw_rate < 100 AND scale(raw_rate) <= 12),
    locked_rate numeric(3,1) NOT NULL CHECK (locked_rate = trunc(raw_rate, 1) AND locked_rate > 0),
    source text NOT NULL CHECK (source = 'frankfurter_v2_ecb'),
    fetched_at timestamptz NOT NULL DEFAULT now(),
    CHECK (observation_day < accounting_day AND observation_day >= accounting_day - 7)
);
ALTER TABLE app.daily_usd_cny_rates ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON app.daily_usd_cny_rates FROM PUBLIC;
DO $$
DECLARE role_name text;
BEGIN
    FOREACH role_name IN ARRAY ARRAY['anon','authenticated','service_role'] LOOP
        IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname=role_name) THEN
            EXECUTE format('REVOKE ALL ON app.daily_usd_cny_rates FROM %I', role_name);
        END IF;
    END LOOP;
END;
$$;

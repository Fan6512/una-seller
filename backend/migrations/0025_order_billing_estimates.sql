-- FIN-02A: independently collected billing estimates, not confirmation or settlement.
-- Extra cost is an explicit owner input independent of the net source.
ALTER TABLE app.meli_orders
    ADD COLUMN IF NOT EXISTS expected_profit_billing_estimate jsonb
    CHECK (expected_profit_billing_estimate IS NULL OR jsonb_typeof(expected_profit_billing_estimate) = 'object'),
    ADD COLUMN IF NOT EXISTS expected_profit_extra_cny numeric
    CHECK (expected_profit_extra_cny IS NULL OR (
        expected_profit_extra_cny >= 0 AND expected_profit_extra_cny < 1000000000000
        AND expected_profit_extra_cny = trunc(expected_profit_extra_cny, 2)
    ));

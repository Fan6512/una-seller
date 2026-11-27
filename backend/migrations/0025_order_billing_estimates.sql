-- FIN-02A: independently collected billing estimates, not confirmation or settlement.
-- Extra cost is an explicit owner input independent of the net source.
ALTER TABLE app.meli_orders
    ADD COLUMN IF NOT EXISTS expected_profit_billing_estimate jsonb
    CHECK (expected_profit_billing_estimate IS NULL OR jsonb_typeof(expected_profit_billing_estimate) = 'object'),
    ADD COLUMN IF NOT EXISTS expected_profit_billing_attempt jsonb
    CHECK (expected_profit_billing_attempt IS NULL OR jsonb_typeof(expected_profit_billing_attempt) = 'object'),
    ADD COLUMN IF NOT EXISTS expected_profit_extra_cny numeric
    CHECK (expected_profit_extra_cny IS NULL OR (
        expected_profit_extra_cny >= 0 AND expected_profit_extra_cny < 1000000000000
        AND expected_profit_extra_cny = trunc(expected_profit_extra_cny, 2)
    ));

-- Bound the per-order local sharing veto in both detail and dashboard reads.
CREATE INDEX IF NOT EXISTS meli_orders_profit_pack_idx ON app.meli_orders (store_id, pack_id);
CREATE INDEX IF NOT EXISTS meli_orders_profit_shipment_idx ON app.meli_orders (store_id, shipment_id);

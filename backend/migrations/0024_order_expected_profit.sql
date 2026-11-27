-- FIN-02A: optional owner-confirmed order net input, not a settled ledger.
-- Keep the existing full order key and RLS. Confirmation is bound to the
-- current financial evidence and is ignored when that evidence changes.
ALTER TABLE app.meli_orders
    ADD COLUMN IF NOT EXISTS expected_profit_confirmation jsonb
    CHECK (expected_profit_confirmation IS NULL OR jsonb_typeof(expected_profit_confirmation) = 'object');

-- A2: last successful sanitized single-order billing snapshot, never raw official JSON.
-- The existing order PK scopes the record by store/site/logistic/seller/order.
ALTER TABLE app.meli_orders
  ADD COLUMN IF NOT EXISTS billing_snapshot jsonb
  CHECK (billing_snapshot IS NULL OR jsonb_typeof(billing_snapshot) = 'object');

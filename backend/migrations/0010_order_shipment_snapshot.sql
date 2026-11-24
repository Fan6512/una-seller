-- Scoped to the existing order row and its current shipment.
ALTER TABLE app.meli_orders
  ADD COLUMN IF NOT EXISTS shipment_snapshot jsonb
  CHECK (shipment_snapshot IS NULL OR jsonb_typeof(shipment_snapshot) = 'object');

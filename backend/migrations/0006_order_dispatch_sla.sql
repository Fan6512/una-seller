-- M3-ORDER-SLA-001: official CBT dispatch deadline from /marketplace/shipments/{id}/sla.
-- Never infer this from order creation time or estimated_handling_limit.
CREATE TABLE IF NOT EXISTS app.meli_order_dispatch_sla (
    store_id uuid NOT NULL,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    order_id text NOT NULL,
    shipment_id text NOT NULL CHECK (shipment_id ~ '^[1-9][0-9]{0,29}$'),
    sla_status text NOT NULL CHECK (sla_status ~ '^[a-z_]{1,40}$'),
    expected_date timestamptz,
    checked_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (store_id,site_id,logistic_type,seller_id,order_id),
    FOREIGN KEY (store_id,site_id,logistic_type,seller_id,order_id)
      REFERENCES app.meli_orders(store_id,site_id,logistic_type,seller_id,order_id)
      ON DELETE CASCADE
);
ALTER TABLE app.meli_order_dispatch_sla ENABLE ROW LEVEL SECURITY;

-- M3-ORDER-PURCHASE-001: manual 1688 purchase linkage owned by an existing CBT order.
-- Extension auto-population is a later opt-in flow; never infer CNY costs from USD sales.
CREATE TABLE IF NOT EXISTS app.meli_order_purchases (
    store_id uuid NOT NULL,
    site_id text NOT NULL,
    logistic_type text NOT NULL,
    seller_id text NOT NULL,
    order_id text NOT NULL,
    purchase_order_id text NOT NULL CHECK (purchase_order_id ~ '^[0-9]{8,40}$'),
    cost_cny numeric(14,2) CHECK (cost_cny >= 0),
    cost_source text NOT NULL DEFAULT 'manual' CHECK (cost_source IN ('manual','extension')),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (store_id,site_id,logistic_type,seller_id,order_id),
    FOREIGN KEY (store_id,site_id,logistic_type,seller_id,order_id)
      REFERENCES app.meli_orders(store_id,site_id,logistic_type,seller_id,order_id) ON DELETE CASCADE
);
ALTER TABLE app.meli_order_purchases ENABLE ROW LEVEL SECURITY;

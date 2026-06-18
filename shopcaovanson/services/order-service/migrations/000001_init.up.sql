CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    status           VARCHAR(30) NOT NULL DEFAULT 'pending',
    total_amount     NUMERIC(15,2) NOT NULL,
    shipping_name    VARCHAR(255) NOT NULL,
    shipping_phone   VARCHAR(20) NOT NULL,
    shipping_address TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id              UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id            UUID NOT NULL,
    product_name_snapshot VARCHAR(500) NOT NULL,
    product_image_snapshot TEXT,
    unit_price            NUMERIC(15,2) NOT NULL,
    quantity              INT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

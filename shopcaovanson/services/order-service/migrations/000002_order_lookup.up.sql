CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE SEQUENCE IF NOT EXISTS order_number_seq START WITH 1000001;

ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_number VARCHAR(20);

UPDATE orders
SET order_number = 'ORD-' || LPAD(nextval('order_number_seq')::text, 10, '0')
WHERE order_number IS NULL;

ALTER TABLE orders ALTER COLUMN order_number SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
CREATE INDEX IF NOT EXISTS idx_orders_shipping_phone ON orders(shipping_phone);
CREATE INDEX IF NOT EXISTS idx_orders_shipping_name_trgm ON orders USING gin (shipping_name gin_trgm_ops);

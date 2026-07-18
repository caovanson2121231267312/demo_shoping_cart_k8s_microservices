-- Admin list: filter status + sort by created_at
CREATE INDEX IF NOT EXISTS idx_orders_status_created_at ON orders (status, created_at DESC);

-- Date-range admin filters
CREATE INDEX IF NOT EXISTS idx_orders_created_at_id ON orders (created_at DESC, id DESC);

-- ILIKE search on order_number (admin search box)
CREATE INDEX IF NOT EXISTS idx_orders_order_number_trgm ON orders USING gin (order_number gin_trgm_ops);

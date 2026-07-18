-- Admin list: ORDER BY created_at DESC
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);

-- Storefront default filter + sort
CREATE INDEX IF NOT EXISTS idx_products_active_created_at ON products (is_active, created_at DESC) WHERE is_active = TRUE;

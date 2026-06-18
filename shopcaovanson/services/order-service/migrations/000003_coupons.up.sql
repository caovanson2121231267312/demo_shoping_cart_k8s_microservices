CREATE TABLE IF NOT EXISTS coupons (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(50) NOT NULL UNIQUE,
    type            VARCHAR(20) NOT NULL CHECK (type IN ('percent', 'fixed')),
    value           NUMERIC(15,2) NOT NULL,
    min_order       NUMERIC(15,2) NOT NULL DEFAULT 0,
    max_discount    NUMERIC(15,2),
    usage_limit     INT,
    used_count      INT NOT NULL DEFAULT 0,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_active ON coupons(is_active);

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS coupon_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS subtotal_amount NUMERIC(15,2);

UPDATE orders SET subtotal_amount = total_amount WHERE subtotal_amount IS NULL;
ALTER TABLE orders ALTER COLUMN subtotal_amount SET NOT NULL;

INSERT INTO coupons (code, type, value, min_order, max_discount, usage_limit, is_active, expires_at)
VALUES
    ('WELCOME10', 'percent', 10, 200000, 500000, 10000, TRUE, NOW() + INTERVAL '1 year'),
    ('FREESHIP', 'fixed', 30000, 500000, NULL, 5000, TRUE, NOW() + INTERVAL '6 months'),
    ('SALE50K', 'fixed', 50000, 1000000, NULL, 1000, TRUE, NOW() + INTERVAL '3 months')
ON CONFLICT (code) DO NOTHING;

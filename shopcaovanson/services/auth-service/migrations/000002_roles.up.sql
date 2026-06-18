CREATE EXTENSION IF NOT EXISTS pg_trgm;

ALTER TABLE users ALTER COLUMN role TYPE VARCHAR(30);

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING gin (email gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_full_name_trgm ON users USING gin (full_name gin_trgm_ops);

UPDATE users SET role = 'super_admin' WHERE email = 'admin@shop.com' AND role = 'admin';

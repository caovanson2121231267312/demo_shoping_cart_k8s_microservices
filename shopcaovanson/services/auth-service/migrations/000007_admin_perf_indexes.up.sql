-- Admin list: ORDER BY created_at DESC (+ role filter)
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_role_created_at ON users (role, created_at DESC);

-- Dashboard: COUNT active users
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users (is_active) WHERE is_active = TRUE;

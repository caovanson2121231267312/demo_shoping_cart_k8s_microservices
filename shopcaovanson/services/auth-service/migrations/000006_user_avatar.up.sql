ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_key TEXT;

CREATE INDEX IF NOT EXISTS idx_users_avatar_key ON users (avatar_key) WHERE avatar_key IS NOT NULL;

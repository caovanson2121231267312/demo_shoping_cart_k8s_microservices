ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_reset_otp_hash VARCHAR(128),
    ADD COLUMN IF NOT EXISTS password_reset_expires_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_password_reset_otp_hash ON users(password_reset_otp_hash);

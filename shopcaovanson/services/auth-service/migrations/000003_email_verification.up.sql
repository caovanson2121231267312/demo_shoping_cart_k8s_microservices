ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS verification_token VARCHAR(128),
    ADD COLUMN IF NOT EXISTS verification_expires_at TIMESTAMPTZ;

UPDATE users SET email_verified = TRUE WHERE email_verified = FALSE AND verification_token IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_verification_token ON users(verification_token);

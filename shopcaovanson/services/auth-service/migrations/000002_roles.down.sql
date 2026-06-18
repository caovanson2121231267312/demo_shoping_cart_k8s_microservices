DROP INDEX IF EXISTS idx_users_full_name_trgm;
DROP INDEX IF EXISTS idx_users_email_trgm;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users DROP COLUMN IF EXISTS is_active;

UPDATE users SET role = 'admin' WHERE email = 'admin@shop.com' AND role = 'super_admin';

ALTER TABLE users ALTER COLUMN role TYPE VARCHAR(20);

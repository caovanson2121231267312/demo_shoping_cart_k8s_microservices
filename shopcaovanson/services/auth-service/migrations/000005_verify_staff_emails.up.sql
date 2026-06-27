UPDATE users
SET email_verified = TRUE,
    verification_token = NULL,
    verification_expires_at = NULL,
    updated_at = NOW()
WHERE role IN ('support', 'staff', 'manager', 'admin', 'super_admin')
  AND email_verified = FALSE;

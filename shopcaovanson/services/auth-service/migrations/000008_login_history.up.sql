-- Chi tiết mỗi lần đăng nhập (thành công / thất bại)
CREATE TABLE IF NOT EXISTS login_history (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    email TEXT NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    failure_reason TEXT,
    ip_address TEXT,
    user_agent TEXT,
    device TEXT,
    browser TEXT,
    os TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_login_history_created_at ON login_history (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_email_created ON login_history (email, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_user_created ON login_history (user_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_login_history_success_created ON login_history (success, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_login_history_day ON login_history ((CAST((created_at AT TIME ZONE 'Asia/Ho_Chi_Minh') AS date)));

-- Báo cáo Excel tổng hợp theo ngày (file trên MinIO)
CREATE TABLE IF NOT EXISTS login_reports (
    id UUID PRIMARY KEY,
    report_date DATE NOT NULL,
    object_key TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_size BIGINT NOT NULL DEFAULT 0,
    total_logins INT NOT NULL DEFAULT 0,
    success_count INT NOT NULL DEFAULT 0,
    failure_count INT NOT NULL DEFAULT 0,
    unique_users INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'ready',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_login_reports_date UNIQUE (report_date)
);

CREATE INDEX IF NOT EXISTS idx_login_reports_date ON login_reports (report_date DESC);

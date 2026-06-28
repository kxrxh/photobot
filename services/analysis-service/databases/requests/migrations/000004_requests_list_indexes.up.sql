CREATE INDEX IF NOT EXISTS idx_requests_user_platform_created
    ON requests (user_id, platform, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_requests_user_platform_status_created
    ON requests (user_id, platform, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_requests_old_completed_created
    ON requests (created_at)
    WHERE status = 'completed';

CREATE INDEX IF NOT EXISTS idx_requests_old_failed_created
    ON requests (created_at)
    WHERE status = 'failed';

CREATE TYPE request_status AS ENUM (
    'created', 'processing', 'waiting_for_confirmation', 'completed', 'failed'
);

CREATE TABLE requests (
    id VARCHAR PRIMARY KEY,
    user_id VARCHAR NOT NULL,
    platform VARCHAR NOT NULL DEFAULT 'telegram',
    product VARCHAR NOT NULL,
    status request_status NOT NULL,
    year VARCHAR DEFAULT date_part('year', now())::VARCHAR,
    mass_liter DOUBLE PRECISION DEFAULT NULL,
    location VARCHAR DEFAULT NULL,
    images JSONB NOT NULL,
    classification JSONB DEFAULT NULL,
    mass_1000 DOUBLE PRECISION DEFAULT NULL,
    mass DOUBLE PRECISION DEFAULT NULL,
    temp_id VARCHAR DEFAULT NULL,

    error_message VARCHAR DEFAULT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_requests_status ON requests (status, updated_at DESC);

CREATE INDEX idx_requests_user_platform_created ON requests (user_id, platform, created_at DESC);

CREATE INDEX idx_requests_user_platform_status_created ON requests (user_id, platform, status, created_at DESC);

CREATE INDEX idx_requests_old_completed_created ON requests (created_at) WHERE status = 'completed';

CREATE INDEX idx_requests_old_failed_created ON requests (created_at) WHERE status = 'failed';

CREATE TABLE outbox_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    claimed_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_outbox_messages_status_created_at ON outbox_messages (status, created_at) WHERE status = 'pending';

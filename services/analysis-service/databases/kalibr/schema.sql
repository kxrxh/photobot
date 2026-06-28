CREATE TABLE analysis_new (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL,
    source TEXT,
    product TEXT,
    bot_message TEXT,
    files_source TEXT[],
    files_output TEXT[],
    objects JSONB,
    date_time TIMESTAMPTZ,
    scale_mm_pixel DOUBLE PRECISION,
    analysis_params JSONB
);

CREATE INDEX idx_analysis_new_user_id ON analysis_new(user_id);
CREATE INDEX idx_analysis_new_date_time ON analysis_new(date_time);
CREATE INDEX idx_analysis_new_user_date ON analysis_new(user_id, date_time DESC);

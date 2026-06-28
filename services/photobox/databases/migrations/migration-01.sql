-- =====================================================
-- Functions
-- =====================================================
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

-- =====================================================
-- Catalog Proposals
-- =====================================================
CREATE TABLE catalog_proposals (
    id SERIAL PRIMARY KEY,
    status VARCHAR(32) NOT NULL DEFAULT 'draft' CHECK (
        status IN ('draft', 'submitted', 'changes_requested', 'applied', 'rejected', 'cancelled')
    ),
    request_by INTEGER NOT NULL,
    target_weed_id INTEGER NULL,
    reviewed_by INTEGER NULL,
    reviewed_at TIMESTAMP NULL,
    review_notes TEXT,
    submitted_at TIMESTAMPTZ NULL,
    applied_by INTEGER NULL,
    applied_at TIMESTAMPTZ NULL,
    applied_weed_id INTEGER NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_catalog_proposals_status ON catalog_proposals(status);

CREATE INDEX idx_catalog_proposals_request_by ON catalog_proposals(request_by);

CREATE INDEX idx_catalog_proposals_reviewed_by ON catalog_proposals(reviewed_by);

CREATE INDEX idx_catalog_proposals_created_at ON catalog_proposals(created_at DESC);

CREATE OR REPLACE FUNCTION update_catalog_proposals_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        -- Capture submit time once
        IF NEW.status = 'submitted' AND NEW.submitted_at IS NULL THEN
            NEW.submitted_at = NOW();
        END IF;

        -- Capture moderation time (request-changes/reject/apply)
        IF NEW.status IN ('changes_requested', 'rejected', 'applied') THEN
            NEW.reviewed_at = NOW();
        END IF;

        -- Capture apply time once
        IF NEW.status = 'applied' AND NEW.applied_at IS NULL THEN
            NEW.applied_at = NOW();
        END IF;
    END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER update_catalog_proposals_updated_at BEFORE
UPDATE
    ON catalog_proposals FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_catalog_proposals_reviewed_at BEFORE
UPDATE
    ON catalog_proposals FOR EACH ROW EXECUTE FUNCTION update_catalog_proposals_timestamps();

-- =====================================================
-- Pending Weeds
-- =====================================================
CREATE TABLE pending_weeds (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    latin_name VARCHAR(255),
    description TEXT,
    length REAL,
    width REAL,
    main_group VARCHAR(32) NULL,
    main_subgroup VARCHAR(32) NULL,
    subgroup VARCHAR(64) NULL,
    is_quarantine BOOLEAN NOT NULL DEFAULT FALSE,
    harmfulness TEXT NULL,
    request_id INTEGER NOT NULL REFERENCES catalog_proposals(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_pending_weeds_request_id ON pending_weeds(request_id);
CREATE UNIQUE INDEX idx_pending_weeds_request_unique ON pending_weeds(request_id);

CREATE INDEX idx_pending_weeds_name_pattern ON pending_weeds(name text_pattern_ops);

CREATE INDEX idx_pending_weeds_created_at ON pending_weeds(created_at DESC);

CREATE TRIGGER update_pending_weeds_updated_at BEFORE
UPDATE
    ON pending_weeds FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- =====================================================
-- Pending Weed Images
-- =====================================================
CREATE TABLE pending_weed_images (
    id SERIAL PRIMARY KEY,
    pending_weed_id INTEGER NOT NULL REFERENCES pending_weeds(id) ON DELETE CASCADE,
    image_key VARCHAR(255) NOT NULL UNIQUE,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_pending_weed_images_primary ON pending_weed_images(pending_weed_id)
WHERE
    is_primary = TRUE;

CREATE INDEX idx_pending_weed_images_pending_weed_id ON pending_weed_images(pending_weed_id);

-- =====================================================
-- Pending Weed Stats
-- =====================================================
CREATE TABLE pending_weed_stats (
    id SERIAL PRIMARY KEY,
    pending_weed_id INTEGER NOT NULL REFERENCES pending_weeds(id) ON DELETE CASCADE,
    w_avg REAL NOT NULL,
    w_median REAL NOT NULL,
    w_min REAL NOT NULL,
    w_max REAL NOT NULL,
    l_avg REAL NOT NULL,
    l_median REAL NOT NULL,
    l_min REAL NOT NULL,
    l_max REAL NOT NULL,
    sq_avg REAL NOT NULL,
    sq_median REAL NOT NULL,
    sq_min REAL NOT NULL,
    sq_max REAL NOT NULL,
    r_avg REAL NOT NULL,
    r_median REAL NOT NULL,
    r_min REAL NOT NULL,
    r_max REAL NOT NULL,
    g_avg REAL NOT NULL,
    g_median REAL NOT NULL,
    g_min REAL NOT NULL,
    g_max REAL NOT NULL,
    b_avg REAL NOT NULL,
    b_median REAL NOT NULL,
    b_min REAL NOT NULL,
    b_max REAL NOT NULL,
    h_avg REAL NOT NULL,
    h_median REAL NOT NULL,
    h_min REAL NOT NULL,
    h_max REAL NOT NULL,
    s_avg REAL NOT NULL,
    s_median REAL NOT NULL,
    s_min REAL NOT NULL,
    s_max REAL NOT NULL,
    v_avg REAL NOT NULL,
    v_median REAL NOT NULL,
    v_min REAL NOT NULL,
    v_max REAL NOT NULL,
    lw_avg REAL NOT NULL,
    lw_median REAL NOT NULL,
    lw_min REAL NOT NULL,
    lw_max REAL NOT NULL,
    brt_avg REAL NOT NULL,
    brt_median REAL NOT NULL,
    brt_min REAL NOT NULL,
    brt_max REAL NOT NULL,
    solid_avg REAL NOT NULL,
    solid_median REAL NOT NULL,
    solid_min REAL NOT NULL,
    solid_max REAL NOT NULL,
    sq_sqcrl_avg REAL NOT NULL,
    sq_sqcrl_median REAL NOT NULL,
    sq_sqcrl_min REAL NOT NULL,
    sq_sqcrl_max REAL NOT NULL,
    excluded_objects JSONB DEFAULT '[]' :: jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pending_weed_id)
);

CREATE INDEX idx_pending_weed_stats_pending_weed_id ON pending_weed_stats(pending_weed_id);

CREATE TRIGGER update_pending_weed_stats_updated_at BEFORE
UPDATE
    ON pending_weed_stats FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- =====================================================
-- Pending Weed Analyses
-- =====================================================
CREATE TABLE pending_weed_analyses (
    id SERIAL PRIMARY KEY,
    pending_weed_id INTEGER NOT NULL REFERENCES pending_weeds(id) ON DELETE CASCADE,
    analysis_id VARCHAR(255) NOT NULL,
    UNIQUE(pending_weed_id, analysis_id)
);

CREATE INDEX idx_pending_weed_analyses_pending_weed_id ON pending_weed_analyses(pending_weed_id);

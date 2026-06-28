-- Migration: Store request_by as Telegram ID (BIGINT) instead of user ID (INTEGER)

ALTER TABLE catalog_proposals
    ALTER COLUMN request_by TYPE BIGINT USING request_by::BIGINT;

-- Index idx_catalog_proposals_request_by remains valid (B-tree supports BIGINT)

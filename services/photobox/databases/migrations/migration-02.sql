-- Migration: Remove draft state from catalog proposals
-- New flow: submitted -> changes_requested -> applied|rejected|cancelled

-- Migrate existing draft proposals to submitted
UPDATE catalog_proposals
SET status = 'submitted',
    submitted_at = COALESCE(submitted_at, NOW())
WHERE status = 'draft';

-- Drop old CHECK constraint (PostgreSQL names it catalog_proposals_status_check)
ALTER TABLE catalog_proposals DROP CONSTRAINT IF EXISTS catalog_proposals_status_check;

-- Add new CHECK constraint without 'draft'
ALTER TABLE catalog_proposals ADD CONSTRAINT catalog_proposals_status_check
    CHECK (status IN ('submitted', 'changes_requested', 'applied', 'rejected', 'cancelled'));

-- Update default status for new rows
ALTER TABLE catalog_proposals ALTER COLUMN status SET DEFAULT 'submitted';

-- Add trigger to set submitted_at on INSERT when status is 'submitted'
CREATE OR REPLACE FUNCTION set_submitted_at_on_insert()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'submitted' AND NEW.submitted_at IS NULL THEN
        NEW.submitted_at = NOW();
    END IF;
    RETURN NEW;
END;
$$
LANGUAGE plpgsql;

CREATE TRIGGER catalog_proposals_set_submitted_at
    BEFORE INSERT ON catalog_proposals
    FOR EACH ROW
    EXECUTE FUNCTION set_submitted_at_on_insert();

-- Migration: Allow same image_key across different pending weeds; keep uniqueness per pending_weed

ALTER TABLE pending_weed_images
    DROP CONSTRAINT IF EXISTS pending_weed_images_image_key_key;

CREATE UNIQUE INDEX IF NOT EXISTS idx_pending_weed_images_pending_weed_id_image_key
    ON pending_weed_images (pending_weed_id, image_key);

-- name: SetUserActiveClassification :one
-- Set or update the active classification for a user
INSERT INTO
  user_active_classification (user_id, classification_id)
VALUES
  (sqlc.arg(user_id), sqlc.arg(classification_id))
ON CONFLICT (user_id)
DO UPDATE SET
  classification_id = EXCLUDED.classification_id,
  updated_at = NOW()
RETURNING *;

-- name: GetUserActiveClassification :one
-- Get the active classification for a user
SELECT
  *
FROM
  user_active_classification
WHERE
  user_id = sqlc.arg(user_id);

-- name: DeleteUserActiveClassification :exec
-- Remove the active classification for a user
DELETE FROM
  user_active_classification
WHERE
  user_id = sqlc.arg(user_id); 
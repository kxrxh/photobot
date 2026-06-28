-- name: MergeReassignClassifications :exec
-- Reassign classifications from one user to another during account merge
UPDATE classifications
SET created_by = sqlc.arg(to_user_id)
WHERE created_by = sqlc.arg(from_user_id);

-- name: MergeReassignMarkups :exec
-- Reassign markups from one user to another during account merge
UPDATE markups
SET created_by = sqlc.arg(to_user_id)
WHERE created_by = sqlc.arg(from_user_id);

-- name: MergeGetUserActiveClassifications :many
-- Get both users' active classification rows for conflict resolution (newer wins)
SELECT user_id, classification_id, created_at, updated_at
FROM user_active_classification
WHERE user_id IN (sqlc.arg(from_user_id), sqlc.arg(to_user_id));

-- name: MergeUserActiveClassificationDeleteByUserID :exec
-- Delete user_active_classification row by user_id
DELETE FROM user_active_classification
WHERE user_id = sqlc.arg(user_id);

-- name: MergeUserActiveClassificationReassign :exec
-- Reassign user_active_classification from one user to another (used when from_user has the winning row)
UPDATE user_active_classification
SET user_id = sqlc.arg(to_user_id),
    updated_at = NOW()
WHERE user_id = sqlc.arg(from_user_id);

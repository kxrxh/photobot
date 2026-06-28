-- name: CreateRecoveryCode :one
INSERT INTO user_recovery_codes (user_id, code_hash)
VALUES ($1, $2)
RETURNING *;

-- name: ListUnusedRecoveryCodesByUser :many
SELECT * FROM user_recovery_codes
WHERE user_id = $1 AND used_at IS NULL
ORDER BY id;

-- name: MarkRecoveryCodeUsed :exec
UPDATE user_recovery_codes SET used_at = CURRENT_TIMESTAMP WHERE id = $1;

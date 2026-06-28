-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (telegram_id, organization_name, inn, full_name, phone_number)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateUserWithMaxId :one
INSERT INTO users (max_id, organization_name, inn, full_name, phone_number)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET organization_name = $2, inn = $3, full_name = $4, phone_number = $5
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: GetUserByTelegramId :one
SELECT * FROM users
WHERE telegram_id = $1 LIMIT 1;

-- name: GetUserByMaxId :one
SELECT * FROM users
WHERE max_id = $1 LIMIT 1;

-- name: LinkUserMaxId :exec
UPDATE users
SET max_id = $2
WHERE id = $1;

-- name: LinkUserTelegramId :exec
UPDATE users
SET telegram_id = $2
WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY full_name NULLS LAST;

-- name: GetUserByLogin :one
SELECT * FROM users
WHERE LOWER(login) = LOWER($1) LIMIT 1;

-- name: CreateWebUser :one
INSERT INTO users (login, password_hash, organization_name, inn, full_name, phone_number)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: SetPasswordHash :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: SetWebCredentials :one
UPDATE users
SET login = $2, password_hash = $3
WHERE id = $1
RETURNING *;

-- name: CopyWebCredentials :exec
UPDATE users
SET login = COALESCE(users.login, $2),
    password_hash = COALESCE(users.password_hash, $3)
WHERE id = $1;

-- name: ClearWebCredentials :exec
UPDATE users SET login = NULL, password_hash = NULL WHERE id = $1;

-- name: TransferWebCredentials :exec
WITH src AS (
    SELECT login, password_hash FROM users WHERE users.id = $2
),
cleared AS (
    UPDATE users SET login = NULL, password_hash = NULL WHERE users.id = $2
)
UPDATE users AS keep
SET
    login = COALESCE(keep.login, src.login),
    password_hash = COALESCE(keep.password_hash, src.password_hash)
FROM src
WHERE keep.id = $1
    AND (src.login IS NOT NULL OR src.password_hash IS NOT NULL);

-- name: ReassignRecoveryCodes :exec
UPDATE user_recovery_codes SET user_id = $2 WHERE user_id = $1;

-- name: UpsertDevUser :one
INSERT INTO users (login, password_hash, telegram_id, full_name)
VALUES ($1, $2, $3, 'Dev User')
ON CONFLICT (login) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    telegram_id = COALESCE(users.telegram_id, EXCLUDED.telegram_id)
RETURNING *;


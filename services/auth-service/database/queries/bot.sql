-- name: CreateBot :one
INSERT INTO bots (name, token, platform)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetBotByName :one
SELECT *
FROM bots
WHERE name = $1 AND platform = 'telegram';

-- name: GetBotByNameAndPlatform :one
SELECT *
FROM bots
WHERE name = $1 AND platform = $2;

-- name: GetBot :one
SELECT id, name, created_at, updated_at, token
FROM bots
WHERE id = $1;

-- name: ListBots :many
SELECT id, name, platform, created_at, updated_at
FROM bots
ORDER BY name;

-- name: ListBotsByPlatform :many
SELECT *
FROM bots
WHERE platform = $1
ORDER BY id;

-- name: UpdateBotName :one
UPDATE bots
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateBotToken :one
UPDATE bots
SET token = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBot :exec
DELETE
FROM bots
WHERE id = $1; 
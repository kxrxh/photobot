-- name: CreateRequest :one
INSERT INTO requests (
    id, user_id, platform, product, status, images, classification, year, mass_liter, location, mass_1000, mass, temp_id
) VALUES (
    $1, $2, $3, $4, 'created', $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: GetRequest :one
SELECT * FROM requests
WHERE id = $1 LIMIT 1;

-- name: GetRequestByIDAndUserPairs :one
SELECT * FROM requests
WHERE id = $1 AND ((user_id = $2 AND platform = $3) OR (user_id = $4 AND platform = $5))
LIMIT 1;

-- name: MarkRequestAsProcessing :exec
UPDATE requests
SET status = 'processing', updated_at = now()
WHERE id = $1 AND status = 'created';

-- name: MarkRequestAsWaitingForConfirmation :exec
UPDATE requests
SET 
    status = 'waiting_for_confirmation',
    temp_id = $2,
    updated_at = now()
WHERE id = $1;

-- name: MarkRequestAsCompleted :exec
UPDATE requests
SET status = 'completed', temp_id = $2, updated_at = now()
WHERE id = $1;

-- name: MarkRequestAsCompletedIfWaiting :execrows
UPDATE requests
SET status = 'completed', temp_id = $2, updated_at = now()
WHERE id = $1 AND status = 'waiting_for_confirmation';

-- name: MarkRequestAsFailed :exec
UPDATE requests
SET 
    status = 'failed', 
    error_message = $2,
    updated_at = now()
WHERE id = $1;

-- name: MarkRequestsAsFailedByIDs :exec
UPDATE requests
SET
    status = 'failed',
    error_message = $1,
    updated_at = now()
WHERE id = ANY($2::varchar[]) AND status = 'processing';

-- name: GetStuckProcessingJobsBatch :many
SELECT id, images FROM requests
WHERE status = 'processing'
AND updated_at < (now() - INTERVAL '10 minutes')::TIMESTAMPTZ
ORDER BY updated_at ASC
LIMIT $1;

-- name: ListRequestsByUserIDAndPlatform :many
SELECT
    id, user_id, platform, product, status, year, mass_liter, location, images,
    mass_1000, mass, temp_id, error_message, created_at, updated_at
FROM requests
WHERE user_id = $1 AND platform = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListRequestsByUserIDAndPlatformAndStatus :many
SELECT
    id, user_id, platform, product, status, year, mass_liter, location, images,
    mass_1000, mass, temp_id, error_message, created_at, updated_at
FROM requests
WHERE user_id = $1 AND platform = $2 AND status = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: ListRequestsByUserPlatformPairs :many
SELECT
    id, user_id, platform, product, status, year, mass_liter, location, images,
    mass_1000, mass, temp_id, error_message, created_at, updated_at
FROM requests
WHERE (user_id = @user_id_1 AND platform = @platform_1)
   OR (@user_id_2 <> '' AND user_id = @user_id_2 AND platform = @platform_2)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListRequestsByUserPlatformPairsAndStatus :many
SELECT
    id, user_id, platform, product, status, year, mass_liter, location, images,
    mass_1000, mass, temp_id, error_message, created_at, updated_at
FROM requests
WHERE ((user_id = @user_id_1 AND platform = @platform_1)
    OR (@user_id_2 <> '' AND user_id = @user_id_2 AND platform = @platform_2))
  AND status = @status
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: DeleteRequest :exec
DELETE FROM requests
WHERE id = $1;

-- name: DeleteRequestsByIDs :exec
DELETE FROM requests
WHERE id = ANY($1::varchar[]);

-- name: GetOldRequestsBatch :many
SELECT id, images FROM requests
WHERE (
    (status = 'completed' AND created_at < (now() - INTERVAL '7 days')::TIMESTAMPTZ)
    OR
    (status IN ('failed') AND created_at < (now() - INTERVAL '1 day')::TIMESTAMPTZ)
)
ORDER BY created_at ASC
LIMIT $1;

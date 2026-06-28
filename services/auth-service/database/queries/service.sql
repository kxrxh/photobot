-- name: CreateService :one
INSERT INTO services (service_id, service_secret)
VALUES ($1, $2)
RETURNING *;

-- name: GetServiceByServiceID :one
SELECT * FROM services
WHERE service_id = $1 LIMIT 1;

-- name: ListServices :many
SELECT service_id, created_at, updated_at FROM services
ORDER BY created_at DESC;

-- name: UpdateService :one
UPDATE services
SET service_secret = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE service_id = $1
RETURNING *;

-- name: DeleteService :exec
DELETE FROM services
WHERE service_id = $1;

-- name: IsServiceExists :one
SELECT EXISTS(SELECT 1 FROM services WHERE service_id = $1) AS exists;

-- name: CreateMarkupFraction :one
INSERT INTO markup_fractions (markup_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: GetMarkupFractionsByMarkupID :many
SELECT * FROM markup_fractions
WHERE markup_id = $1
ORDER BY id ASC;

-- name: DeleteMarkupFraction :exec
DELETE FROM markup_fractions
WHERE id = $1;

-- name: BulkAddObjectsToMarkupFraction :exec
INSERT INTO markup_fraction_objects (markup_fraction_id, object_id)
SELECT $1, unnest($2::bigint[])
ON CONFLICT (markup_fraction_id, object_id) DO NOTHING;

-- name: GetObjectsByMarkupFractionIDs :many
SELECT markup_fraction_id, object_id
FROM markup_fraction_objects
WHERE markup_fraction_id = ANY(sqlc.arg(markup_fraction_ids)::uuid[])
ORDER BY markup_fraction_id, object_id;

-- name: GetMarkupFractionsByMarkupIDs :many
SELECT * FROM markup_fractions
WHERE markup_id = ANY(sqlc.arg(markup_ids)::uuid[])
ORDER BY markup_id, id ASC;

-- name: ClearMarkupFractionObjects :exec
DELETE FROM markup_fraction_objects
WHERE markup_fraction_id = $1;

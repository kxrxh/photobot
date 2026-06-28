-- name: CreateMarkup :one
INSERT INTO markups (name, created_by)
VALUES (sqlc.arg(name), sqlc.arg(created_by))
RETURNING *;

-- name: GetMarkupByID :one
SELECT * FROM markups WHERE id = sqlc.arg(id);

-- name: UpdateMarkup :one
UPDATE markups
SET name = sqlc.arg(name), updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMarkup :exec
DELETE FROM markups WHERE id = sqlc.arg(id);

-- name: GetMarkups :many
SELECT * FROM markups
WHERE (sqlc.narg(created_by)::int IS NULL OR created_by = sqlc.narg(created_by))
  AND (sqlc.narg(name)::text IS NULL OR name ILIKE '%' || sqlc.narg(name) || '%')
ORDER BY created_at DESC;

-- name: DeleteMarkupAnalyses :exec
DELETE FROM markup_analyses WHERE markup_id = sqlc.arg(markup_id);

-- name: GetMarkupAnalysesByMarkupID :many
SELECT analysis_id FROM markup_analyses WHERE markup_id = sqlc.arg(markup_id);

-- name: GetMarkupAnalysesByMarkupIDs :many
SELECT markup_id, analysis_id
FROM markup_analyses
WHERE markup_id = ANY(sqlc.arg(markup_ids)::uuid[])
ORDER BY markup_id, analysis_id;

-- name: BulkCreateMarkupAnalyses :exec
INSERT INTO markup_analyses (markup_id, analysis_id)
SELECT sqlc.arg(markup_id), unnest(sqlc.arg(analysis_ids)::bigint[])
ON CONFLICT DO NOTHING;

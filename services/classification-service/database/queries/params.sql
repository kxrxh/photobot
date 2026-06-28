-- name: BulkCreateParams :many
-- Bulk create params
INSERT INTO
  params (name, operator, value, condition_id)
SELECT
  unnest(sqlc.arg(names)::varchar[]),
  unnest(sqlc.arg(operators)::param_operator[]),
  unnest(sqlc.arg(values)::numeric[]),
  unnest(sqlc.arg(condition_ids)::uuid[])
RETURNING
  *;

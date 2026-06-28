-- name: BulkCreateConditions :many
-- Bulk create conditions
INSERT INTO
  conditions (fraction_id, name, operator, connection, order_index)
SELECT
  unnest(sqlc.arg(fraction_ids)::uuid[]),
  unnest(sqlc.arg(names)::varchar[]),
  unnest(sqlc.arg(operators)::logic_operator[]),
  unnest(sqlc.arg(connections)::logic_operator[]),
  unnest(sqlc.arg(order_indexes)::int[])
RETURNING
  *;

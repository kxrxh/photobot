-- name: RemoveAllFractionsForClassification :exec
-- Remove all fractions for a specific classification
DELETE FROM
  fractions
WHERE
  classification_id = sqlc.arg(classification_id);

-- name: BulkCreateFractions :many
-- Bulk create fractions for a classification
INSERT INTO
  fractions (name, classification_id, order_index)
SELECT
  unnest(sqlc.arg(names)::varchar[]),
  sqlc.arg(classification_id),
  unnest(sqlc.arg(order_indexes)::int[])
RETURNING
  *;

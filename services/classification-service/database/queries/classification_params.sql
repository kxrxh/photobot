-- name: CreateClassificationParam :one
INSERT INTO classification_params (name)
VALUES (sqlc.arg(name))
RETURNING *;

-- name: DeleteClassificationParamByID :exec
DELETE FROM classification_params
WHERE id = sqlc.arg(id);

-- name: GetAllClassificationParams :many
SELECT *
FROM classification_params
ORDER BY name ASC;

-- name: ClassificationParamExistsByName :one
SELECT EXISTS (
  SELECT 1 FROM classification_params WHERE name = sqlc.arg(name)
) AS exists;
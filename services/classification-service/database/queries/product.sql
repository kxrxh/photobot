-- name: CreateProduct :one
-- Create a new product
INSERT INTO
  products (name)
VALUES
  (sqlc.arg(name)) RETURNING *;

-- name: GetProductByID :one
-- Get a product by its ID
SELECT
  *
FROM
  products
WHERE
  id = sqlc.arg(id);

-- name: GetProductByName :one
-- Get a product by its name
SELECT
  *
FROM
  products
WHERE
  name = sqlc.arg(name);

-- name: GetAllProducts :many
-- Get all products
SELECT
  *
FROM
  products;

-- name: UpdateProduct :one
-- Update an existing product by ID
UPDATE
  products
SET
  name = sqlc.arg(name)
WHERE
  id = sqlc.arg(id) RETURNING *;

-- name: DeleteProduct :exec
-- Delete a product by ID
DELETE FROM
  products
WHERE
  id = sqlc.arg(id);
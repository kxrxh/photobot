-- name: GetRole :one
SELECT * FROM roles WHERE id = sqlc.arg('id') LIMIT 1;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = sqlc.arg('name') LIMIT 1;

-- name: CreateRole :one
INSERT INTO roles (name)
VALUES ($1)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: GetUserRoles :many
SELECT r.* FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = sqlc.arg('user_id');

-- name: AddUserRole :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RemoveUserRole :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: CountUsersByRole :one
SELECT COUNT(*) FROM user_roles WHERE role_id = sqlc.arg('role_id'); 
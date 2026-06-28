-- name: CreateWeedNote :one
INSERT INTO
    weed_notes (weed_id, note, created_by)
VALUES
    (
        sqlc.arg('weed_id'),
        sqlc.arg('note'),
        sqlc.arg('created_by')
    ) RETURNING *;

-- name: GetWeedNotes :many
SELECT
    *
FROM
    weed_notes
WHERE
    weed_id = sqlc.arg('weed_id');

-- name: DeleteWeedNote :exec
DELETE FROM
    weed_notes
WHERE
    id = sqlc.arg('id');

-- name: EditWeedNote :one
UPDATE
    weed_notes
SET
    note = sqlc.arg('note')
WHERE
    id = sqlc.arg('id') RETURNING *;

-- name: CheckIsAuthorized :one
SELECT
    EXISTS (
        SELECT
            1
        FROM
            weed_notes
        WHERE
            id = sqlc.arg('id')
            AND created_by = sqlc.arg('created_by')
    );
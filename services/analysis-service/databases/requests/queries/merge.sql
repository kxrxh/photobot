-- name: MergeReassignRequests :exec
UPDATE requests
SET user_id = $2
WHERE user_id = $1;

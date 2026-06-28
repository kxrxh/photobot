-- name: CreateOutboxMessage :one
INSERT INTO outbox_messages (
    topic, payload, status
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: ClaimOutboxMessages :many
WITH to_claim AS (
  SELECT id FROM outbox_messages
  WHERE status = 'pending'
  ORDER BY created_at ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
)
UPDATE outbox_messages m
SET status = 'claimed', claimed_at = now()
FROM to_claim tc
WHERE m.id = tc.id
RETURNING m.id, m.topic, m.payload, m.status, m.created_at, m.claimed_at, m.published_at;

-- name: ResetStuckClaimedOutboxMessages :exec
UPDATE outbox_messages
SET status = 'pending', claimed_at = NULL
WHERE status = 'claimed' AND claimed_at < (now() - INTERVAL '10 minutes')::TIMESTAMPTZ;

-- name: MarkOutboxMessageAsPublished :exec
UPDATE outbox_messages
SET status = 'published', published_at = now()
WHERE id = $1;

-- name: DeletePublishedOutboxMessages :exec
DELETE FROM outbox_messages
WHERE status = 'published' AND published_at < (now() - INTERVAL '1 day')::TIMESTAMPTZ;

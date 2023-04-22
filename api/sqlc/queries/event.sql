-- name: ListEvents :many
SELECT * FROM event ORDER BY created_at DESC;

-- name: EventByID :one
SELECT * FROM event WHERE event_id = $1;

-- name: EventByUserID :many
SELECT * FROM event WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateEvent :one
INSERT INTO event (user_id, event_type, metadata)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteEvent :exec
DELETE FROM event WHERE event_id = $1; 

-- name: UpdateEvent :one
UPDATE event SET user_id = $2, event_type = $3, metadata = $4
WHERE event_id = $1
RETURNING *;

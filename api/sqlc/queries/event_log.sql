-- name: ListEventLogs :many
SELECT * FROM event_log ORDER BY created_at DESC;

-- name: EventLogByID :one
SELECT * FROM event_log WHERE event_id = $1;

-- name: EventLogByUserID :many
SELECT * FROM event_log WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateEventLog :one
INSERT INTO event_log (user_id, event_type, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteEventLog :exec
DELETE FROM event_log WHERE event_id = $1; 

-- name: UpdateEventLog :one
UPDATE event_log SET user_id = $2, event_type = $3, ip_address = $4, user_agent = $5, metadata = $6
WHERE event_id = $1
RETURNING *;

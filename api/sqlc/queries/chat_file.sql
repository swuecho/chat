-- name: CreateChatFile :one
INSERT INTO chat_file (name, data, user_id, chat_session_uuid)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListChatFiles :many
SELECT id, name, data, created_at, user_id, chat_session_uuid
FROM chat_file
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteChatFile :one
DELETE FROM chat_file
WHERE id = $1
RETURNING *;

-- name: CreateChatFile :one
INSERT INTO chat_file (name, data, user_id, chat_session_uuid)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListChatFilesBySessionUUID :many
SELECT id, name
FROM chat_file
WHERE user_id = $1 and chat_session_uuid = $2
ORDER BY created_at DESC;

-- name: ListChatFilesWithContentBySessionUUID :many
SELECT *
FROM chat_file
WHERE user_id = $1 and chat_session_uuid = $2
ORDER BY created_at DESC;


-- name: GetChatFileByID :one
SELECT id, name, data, created_at, user_id, chat_session_uuid
FROM chat_file
WHERE id = $1;

-- name: DeleteChatFile :one
DELETE FROM chat_file
WHERE id = $1
RETURNING *;

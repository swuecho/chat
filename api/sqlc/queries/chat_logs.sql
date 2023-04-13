-- name: ListChatLogs :many
SELECT * FROM chat_logs ORDER BY id;

-- name: ChatLogByID :one
SELECT * FROM chat_logs WHERE id = $1;

-- name: CreateChatLog :one
INSERT INTO chat_logs (session, question, answer)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateChatLog :one
UPDATE chat_logs SET session = $2, question = $3, answer = $4
WHERE id = $1
RETURNING *;

-- name: DeleteChatLog :exec
DELETE FROM chat_logs WHERE id = $1;
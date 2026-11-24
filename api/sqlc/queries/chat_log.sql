-- name: CreateChatLog :one
INSERT INTO chat_logs (session, question, answer)
VALUES ($1, $2, $3)
RETURNING *;

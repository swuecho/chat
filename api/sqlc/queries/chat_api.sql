-- name: ListChatAPIs :many
SELECT * FROM chat_api ORDER BY id;

-- name: ChatAPIByID :one
SELECT * FROM chat_api WHERE id = $1;

-- name: CreateChatAPI :one
INSERT INTO chat_api (name, url, auth_header, auth_key)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateChatAPI :one
UPDATE chat_api SET name = $2, url = $3, auth_header = $4, auth_key = $5
WHERE id = $1
RETURNING *;

-- name: DeleteChatAPI :exec
DELETE FROM chat_api WHERE id = $1;
-- name: ListChatAPIs :many
SELECT * FROM chat_api ORDER BY id;

-- name: ChatAPIByID :one
SELECT * FROM chat_api WHERE id = $1;

-- name: CreateChatAPI :one
INSERT INTO chat_api (name, label, is_default, url, api_auth_header, api_auth_key)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateChatAPI :one
UPDATE chat_api SET name = $2, label = $3, is_default = $4, url = $5, api_auth_header = $6, api_auth_key = $7
WHERE id = $1
RETURNING *;

-- name: DeleteChatAPI :exec
DELETE FROM chat_api WHERE id = $1;

-- name: GetDefaultChatAPI :one
SELECT * FROM chat_api WHERE is_default = true;
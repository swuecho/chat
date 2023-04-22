-- name: ListChatModels :many
SELECT * FROM chat_model ORDER BY id;

-- name: ListSystemChatModels :many
SELECT * FROM chat_model
where user_id in (select id from auth_user where is_superuser = true)
ORDER BY id;

-- name: ChatModelByID :one
SELECT * FROM chat_model WHERE id = $1;

-- name: ChatModelByName :one
SELECT * FROM chat_model WHERE name = $1;

-- name: CreateChatModel :one
INSERT INTO chat_model (name, label, is_default, url, api_auth_header, api_auth_key, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateChatModel :one
UPDATE chat_model SET name = $2, label = $3, is_default = $4, url = $5, api_auth_header = $6, api_auth_key = $7
WHERE id = $1 and user_id = $8
RETURNING *;

-- name: UpdateChatModelKey :one
UPDATE chat_model SET api_auth_key = $2
WHERE id = $1
RETURNING *;

-- name: DeleteChatModel :exec
DELETE FROM chat_model WHERE id = $1 and user_id = $2;

-- name: GetDefaultChatModel :one
SELECT * FROM chat_model WHERE is_default = true
and user_id in (select id from auth_user where is_superuser = true)
;
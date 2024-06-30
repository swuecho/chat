-- name: ListChatModels :many
SELECT * FROM chat_model ORDER BY order_number;

-- name: ListSystemChatModels :many
SELECT * FROM chat_model
where user_id in (select id from auth_user where is_superuser = true)
ORDER BY order_number, id desc;

-- name: ChatModelByID :one
SELECT * FROM chat_model WHERE id = $1;

-- name: ChatModelByName :one
SELECT * FROM chat_model WHERE name = $1;

-- name: CreateChatModel :one
INSERT INTO chat_model (name, label, is_default, url, api_auth_header, api_auth_key, user_id, enable_per_mode_ratelimit, max_token, default_token, order_number, http_time_out )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateChatModel :one
UPDATE chat_model SET name = $2, label = $3, is_default = $4, url = $5, api_auth_header = $6, api_auth_key = $7, enable_per_mode_ratelimit = $9,
max_token = $10, default_token = $11, order_number = $12, http_time_out = $13. is_enable = $14
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
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
INSERT INTO chat_model (name, label, is_default, url, api_auth_header, api_auth_key, user_id, enable_per_mode_ratelimit, max_token, default_token, order_number, http_time_out, api_type )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: UpdateChatModel :one
UPDATE chat_model SET name = $2, label = $3, is_default = $4, url = $5, api_auth_header = $6, api_auth_key = $7, enable_per_mode_ratelimit = $9,
max_token = $10, default_token = $11, order_number = $12, http_time_out = $13, is_enable = $14, api_type = $15
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
ORDER BY order_number, id
LIMIT 1;

-- name: GetTitleChatModel :one
SELECT * FROM chat_model
WHERE is_title_model = true AND is_enable = true
  AND user_id IN (SELECT id FROM auth_user WHERE is_superuser = true)
LIMIT 1;

-- name: SetTitleChatModel :one
WITH updated AS (
    UPDATE chat_model old_model
    SET is_title_model = false
    WHERE old_model.is_title_model = true
      AND old_model.id <> sqlc.arg(model_id)
      AND EXISTS (
          SELECT 1 FROM chat_model selected_model
          WHERE selected_model.id = sqlc.arg(model_id)
            AND selected_model.user_id = sqlc.arg(user_id)
            AND selected_model.is_enable = true
            AND selected_model.api_type = 'gemini'
            AND selected_model.user_id IN (SELECT id FROM auth_user WHERE is_superuser = true)
      )
    RETURNING old_model.id
)
UPDATE chat_model target
SET is_title_model = true
WHERE target.id = sqlc.arg(model_id)
  AND target.user_id = sqlc.arg(user_id)
  AND target.is_enable = true
  AND target.api_type = 'gemini'
  AND target.user_id IN (SELECT id FROM auth_user WHERE is_superuser = true)
  AND (SELECT count(*) FROM updated) >= 0
RETURNING target.*;

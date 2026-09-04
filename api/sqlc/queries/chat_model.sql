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
INSERT INTO chat_model (name, label, is_default, url, api_auth_header, api_auth_key, user_id, enable_per_mode_ratelimit, order_number, http_time_out, api_type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateChatModel :one
UPDATE chat_model SET name = sqlc.arg(name), label = sqlc.arg(label), is_default = sqlc.arg(is_default),
url = sqlc.arg(url), api_auth_header = sqlc.arg(api_auth_header), api_auth_key = sqlc.arg(api_auth_key),
enable_per_mode_ratelimit = sqlc.arg(enable_per_mode_ratelimit),
max_token = COALESCE(sqlc.narg(max_token)::integer, max_token), default_token = COALESCE(sqlc.narg(default_token)::integer, default_token),
order_number = sqlc.arg(order_number), http_time_out = sqlc.arg(http_time_out), is_enable = sqlc.arg(is_enable), api_type = sqlc.arg(api_type)
WHERE id = sqlc.arg(id) and user_id = sqlc.arg(user_id)
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
            AND selected_model.user_id IN (SELECT id FROM auth_user WHERE is_superuser = true)
      )
    RETURNING old_model.id
)
UPDATE chat_model target
SET is_title_model = true
WHERE target.id = sqlc.arg(model_id)
  AND target.user_id = sqlc.arg(user_id)
  AND target.is_enable = true
  AND target.user_id IN (SELECT id FROM auth_user WHERE is_superuser = true)
  AND (SELECT count(*) FROM updated) >= 0
RETURNING target.*;

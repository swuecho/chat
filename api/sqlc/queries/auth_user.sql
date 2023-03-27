-- name: GetAllAuthUsers :many
SELECT * FROM auth_user ORDER BY id;

-- name: ListAuthUsers :many
SELECT * FROM auth_user ORDER BY id LIMIT $1 OFFSET $2;

-- name: GetAuthUserByID :one
SELECT * FROM auth_user WHERE id = $1;


-- name: GetAuthUserByEmail :one
SELECT * FROM auth_user WHERE email = $1;

-- name: CreateAuthUser :one
INSERT INTO auth_user (email, "password", first_name, last_name, username, is_staff, is_superuser)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateAuthUser :one
UPDATE auth_user SET "password" = $2, is_superuser = $3, username = $4, first_name = $5, last_name = $6, email = $7, last_login = now() 
WHERE id = $1
RETURNING *;

-- name: DeleteAuthUser :exec
DELETE FROM auth_user WHERE email = $1;

-- name: GetUserByEmail :one
SELECT * FROM auth_user WHERE email = $1;

-- name: UpdateUserPassword :exec
UPDATE auth_user SET "password" = $2 WHERE email = $1;

-- name: GetTotalActiveUserCount :one
SELECT COUNT(*) FROM auth_user WHERE is_active = true;


-- name: UpdateAuthUserRateLimitByEmail :one
INSERT INTO auth_user_management (user_id, rate_limit, created_at, updated_at)
VALUES ((SELECT id FROM auth_user WHERE email = $1), $2, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET rate_limit = $2, updated_at = NOW()
RETURNING rate_limit;

-- name: GetUserStats :many
SELECT 
    auth_user.email AS user_email,
    COALESCE(user_stats.total_messages, 0) AS total_chat_messages,
    COALESCE(user_stats.total_messages_3_days, 0) AS total_chat_messages_3_days,
    COALESCE(auth_user_management.rate_limit, 0) AS rate_limit
FROM auth_user
LEFT JOIN (
    SELECT chat_message_stats.user_id, 
           SUM(total_messages) AS total_messages, 
           SUM(CASE WHEN created_at >= NOW() - INTERVAL '3 days' THEN total_messages ELSE 0 END) AS total_messages_3_days
    FROM (
        SELECT user_id, COUNT(*) AS total_messages, MAX(created_at) AS created_at
        FROM chat_message
        GROUP BY user_id, chat_session_uuid
    ) AS chat_message_stats
    GROUP BY chat_message_stats.user_id
) AS user_stats ON auth_user.id = user_stats.user_id
LEFT JOIN auth_user_management ON auth_user.id = auth_user_management.user_id
ORDER BY total_chat_messages DESC, auth_user.id DESC
OFFSET $2
LIMIT $1;
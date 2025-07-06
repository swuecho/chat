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
UPDATE auth_user SET first_name = $2, last_name= $3, last_login = now() 
WHERE id = $1
RETURNING first_name, last_name, email;

-- name: UpdateAuthUserByEmail :one
UPDATE auth_user SET first_name = $2, last_name= $3, last_login = now() 
WHERE email = $1
RETURNING first_name, last_name, email;

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
    auth_user.first_name,
    auth_user.last_name,
    auth_user.email AS user_email,
    COALESCE(user_stats.total_messages, 0) AS total_chat_messages,
    COALESCE(user_stats.total_token_count, 0) AS total_token_count,
    COALESCE(user_stats.total_messages_3_days, 0) AS total_chat_messages_3_days,
    COALESCE(user_stats.total_token_count_3_days, 0) AS total_token_count_3_days,
    COALESCE(auth_user_management.rate_limit, @default_rate_limit::INTEGER) AS rate_limit
FROM auth_user
LEFT JOIN (
    SELECT chat_message_stats.user_id, 
           SUM(total_messages) AS total_messages, 
           SUM(total_token_count) AS total_token_count,
           SUM(CASE WHEN created_at >= NOW() - INTERVAL '3 days' THEN total_messages ELSE 0 END) AS total_messages_3_days,
           SUM(CASE WHEN created_at >= NOW() - INTERVAL '3 days' THEN total_token_count ELSE 0 END) AS total_token_count_3_days
    FROM (
        SELECT user_id, COUNT(*) AS total_messages, SUM(token_count) as total_token_count, MAX(created_at) AS created_at
        FROM chat_message
        GROUP BY user_id, chat_session_uuid
    ) AS chat_message_stats
    GROUP BY chat_message_stats.user_id
) AS user_stats ON auth_user.id = user_stats.user_id
LEFT JOIN auth_user_management ON auth_user.id = auth_user_management.user_id
ORDER BY total_chat_messages DESC, auth_user.id DESC
OFFSET $2
LIMIT $1;

-- name: GetUserAnalysisByEmail :one
SELECT 
    auth_user.first_name,
    auth_user.last_name,
    auth_user.email AS user_email,
    COALESCE(user_stats.total_messages, 0) AS total_messages,
    COALESCE(user_stats.total_token_count, 0) AS total_tokens,
    COALESCE(user_stats.total_sessions, 0) AS total_sessions,
    COALESCE(user_stats.total_messages_3_days, 0) AS messages_3_days,
    COALESCE(user_stats.total_token_count_3_days, 0) AS tokens_3_days,
    COALESCE(auth_user_management.rate_limit, @default_rate_limit::INTEGER) AS rate_limit
FROM auth_user
LEFT JOIN (
    SELECT 
        stats.user_id, 
        SUM(stats.total_messages) AS total_messages, 
        SUM(stats.total_token_count) AS total_token_count,
        COUNT(DISTINCT stats.chat_session_uuid) AS total_sessions,
        SUM(CASE WHEN stats.created_at >= NOW() - INTERVAL '3 days' THEN stats.total_messages ELSE 0 END) AS total_messages_3_days,
        SUM(CASE WHEN stats.created_at >= NOW() - INTERVAL '3 days' THEN stats.total_token_count ELSE 0 END) AS total_token_count_3_days
    FROM (
        SELECT user_id, chat_session_uuid, COUNT(*) AS total_messages, SUM(token_count) as total_token_count, MAX(created_at) AS created_at
        FROM chat_message
        WHERE is_deleted = false
        GROUP BY user_id, chat_session_uuid
    ) AS stats
    GROUP BY stats.user_id
) AS user_stats ON auth_user.id = user_stats.user_id
LEFT JOIN auth_user_management ON auth_user.id = auth_user_management.user_id
WHERE auth_user.email = $1;

-- name: GetUserModelUsageByEmail :many
SELECT 
    COALESCE(cm.model, 'unknown') AS model,
    COUNT(*) AS message_count,
    COALESCE(SUM(cm.token_count), 0) AS token_count,
    MAX(cm.created_at)::timestamp AS last_used
FROM chat_message cm
INNER JOIN auth_user au ON cm.user_id = au.id
WHERE au.email = $1 
    AND cm.is_deleted = false 
    AND cm.role = 'assistant'
    AND cm.model IS NOT NULL 
    AND cm.model != ''
GROUP BY cm.model
ORDER BY message_count DESC;

-- name: GetUserRecentActivityByEmail :many
SELECT 
    DATE(cm.created_at) AS activity_date,
    COUNT(*) AS messages,
    COALESCE(SUM(cm.token_count), 0) AS tokens,
    COUNT(DISTINCT cm.chat_session_uuid) AS sessions
FROM chat_message cm
INNER JOIN auth_user au ON cm.user_id = au.id
WHERE au.email = $1 
    AND cm.is_deleted = false 
    AND cm.created_at >= NOW() - INTERVAL '30 days'
GROUP BY DATE(cm.created_at)
ORDER BY activity_date DESC
LIMIT 30;

-- name: GetUserSessionHistoryByEmail :many
SELECT 
    cs.uuid AS session_id,
    cs.model,
    COALESCE(COUNT(cm.id), 0) AS message_count,
    COALESCE(SUM(cm.token_count), 0) AS token_count,
    COALESCE(MIN(cm.created_at), cs.created_at)::timestamp AS created_at,
    COALESCE(MAX(cm.created_at), cs.updated_at)::timestamp AS updated_at
FROM chat_session cs
INNER JOIN auth_user au ON cs.user_id = au.id
LEFT JOIN chat_message cm ON cs.uuid = cm.chat_session_uuid AND cm.is_deleted = false
WHERE au.email = $1 AND cs.active = true
GROUP BY cs.uuid, cs.model, cs.created_at, cs.updated_at
ORDER BY cs.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserSessionHistoryCountByEmail :one
SELECT COUNT(DISTINCT cs.uuid) AS total_sessions
FROM chat_session cs
INNER JOIN auth_user au ON cs.user_id = au.id
WHERE au.email = $1 AND cs.active = true;
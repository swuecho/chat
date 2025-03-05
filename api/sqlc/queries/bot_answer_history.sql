-- Bot Answer History Queries --

-- name: CreateBotAnswerHistory :one
INSERT INTO bot_answer_history (
    bot_uuid,
    user_id,
    prompt,
    answer,
    model,
    tokens_used
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetBotAnswerHistoryByID :one
SELECT 
    bah.id,
    bah.bot_uuid,
    bah.user_id,
    bah.prompt,
    bah.answer,
    bah.model,
    bah.tokens_used,
    bah.created_at,
    bah.updated_at,
    au.username AS user_username,
    au.email AS user_email
FROM bot_answer_history bah
JOIN auth_user au ON bah.user_id = au.id
WHERE bah.id = $1;

-- name: GetBotAnswerHistoryByBotUUID :many
SELECT 
    bah.id,
    bah.bot_uuid,
    bah.user_id,
    bah.prompt,
    bah.answer,
    bah.model,
    bah.tokens_used,
    bah.created_at,
    bah.updated_at,
    au.username AS user_username,
    au.email AS user_email
FROM bot_answer_history bah
JOIN auth_user au ON bah.user_id = au.id
WHERE bah.bot_uuid = $1
ORDER BY bah.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetBotAnswerHistoryByUserID :many
SELECT 
    bah.id,
    bah.bot_uuid,
    bah.user_id,
    bah.prompt,
    bah.answer,
    bah.model,
    bah.tokens_used,
    bah.created_at,
    bah.updated_at,
    au.username AS user_username,
    au.email AS user_email
FROM bot_answer_history bah
JOIN auth_user au ON bah.user_id = au.id
WHERE bah.user_id = $1
ORDER BY bah.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateBotAnswerHistory :one
UPDATE bot_answer_history
SET
    answer = $2,
    tokens_used = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteBotAnswerHistory :exec
DELETE FROM bot_answer_history WHERE id = $1;

-- name: GetBotAnswerHistoryCountByBotUUID :one
SELECT COUNT(*) FROM bot_answer_history WHERE bot_uuid = $1;

-- name: GetBotAnswerHistoryCountByUserID :one
SELECT COUNT(*) FROM bot_answer_history WHERE user_id = $1;

-- name: GetLatestBotAnswerHistoryByBotUUID :many
SELECT 
    bah.id,
    bah.bot_uuid,
    bah.user_id,
    bah.prompt,
    bah.answer,
    bah.model,
    bah.tokens_used,
    bah.created_at,
    bah.updated_at,
    au.username AS user_username,
    au.email AS user_email
FROM bot_answer_history bah
JOIN auth_user au ON bah.user_id = au.id
WHERE bah.bot_uuid = $1
ORDER BY bah.created_at DESC
LIMIT $2;

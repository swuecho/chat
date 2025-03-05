-- name: CreateChatComment :one
INSERT INTO chat_comment (
    uuid,
    chat_session_uuid,
    chat_message_uuid, 
    content,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $5
) RETURNING *;

-- name: GetCommentsBySessionUUID :many
SELECT 
    cc.uuid,
    cc.chat_message_uuid,
    cc.content,
    cc.created_at,
    au.username AS author_username,
    au.email AS author_email
FROM chat_comment cc
JOIN auth_user au ON cc.created_by = au.id
WHERE cc.chat_session_uuid = $1
ORDER BY cc.created_at DESC;

-- name: GetCommentsByMessageUUID :many
SELECT 
    cc.uuid,
    cc.content,
    cc.created_at,
    au.username AS author_username,
    au.email AS author_email
FROM chat_comment cc
JOIN auth_user au ON cc.created_by = au.id
WHERE cc.chat_message_uuid = $1
ORDER BY cc.created_at DESC;


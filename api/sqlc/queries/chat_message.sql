-- name: GetAllChatMessages :many
SELECT * FROM chat_message ORDER BY id;

-- name: GetChatMessagesBySessionUUID :many
SELECT cm.*
FROM chat_message cm
INNER JOIN chat_session cs ON cm.chat_session_uuid = cs.uuid
WHERE cs.active = true and cs.uuid = $1 
ORDER BY cm.id 
OFFSET $2
LIMIT $3;


-- name: GetChatMessageBySessionUUID :one
SELECT cm.*
FROM chat_message cm
INNER JOIN chat_session cs ON cm.chat_session_uuid = cs.uuid
WHERE cs.active = true and cs.uuid = $1 
ORDER BY cm.id 
OFFSET $2
LIMIT $1;


-- name: GetChatMessageByID :one
SELECT * FROM chat_message WHERE id = $1;


-- name: CreateChatMessage :one
INSERT INTO chat_message (chat_session_uuid, uuid, role, content, score, user_id, created_by, updated_by, raw)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateChatMessage :one
UPDATE chat_message SET role = $2, content = $3, score = $4, user_id = $5, updated_by = $6, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteChatMessage :exec
DELETE FROM chat_message WHERE id = $1;

-- by uuid

-- name: GetChatMessageByUUID :one
SELECT * FROM chat_message WHERE uuid = $1;

-- name: CreateChatMessageByUUID :one
INSERT INTO chat_message (uuid, chat_session_uuid, role, content, score, user_id, created_by, updated_by, raw)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateChatMessageByUUID :one
UPDATE chat_message SET content = $2, updated_at = now() 
WHERE uuid = $1
RETURNING *;

-- name: DeleteChatMessageByUUID :exec
DELETE FROM chat_message WHERE uuid = $1;


-- name: HasChatMessagePermission :one
SELECT COUNT(*) > 0 as has_permission
FROM chat_message cm
INNER JOIN chat_session cs ON cm.chat_session_uuid = cs.uuid
INNER JOIN auth_user au ON cs.user_id = au.id
WHERE cm.id = $1 AND (cs.user_id = $2 OR au.is_superuser) and cs.active = true;


-- name: GetLatestMessagesBySessionUUID :many
SELECT *
FROM chat_message
Where chat_session_uuid in 
(
    SELECT chat_session_uuid 
    FROM chat_message
    WHERE chat_message.chat_session_uuid = $1
    ORDER BY created_at DESC
    LIMIT $2 
)
ORDER BY created_at ;


-- name: GetFirstMessageBySessionUUID :one
SELECT *
FROM chat_message
WHERE chat_session_uuid = $1
ORDER BY created_at 
LIMIT 1;

-- name: GetLastNChatMessages :many
SELECT *
FROM chat_message
WHERE chat_message.id in (
    SELECT id 
    FROM chat_message cm
    WHERE cm.chat_session_uuid = $3 
            AND cm.id < (SELECT id FROM chat_message WHERE chat_message.uuid = $1)
    ORDER BY cm.created_at DESC
    LIMIT $2
) 
ORDER BY created_at;


-- name: UpdateChatMessageContent :exec
UPDATE chat_message
SET content = $2, updated_at = now()
WHERE uuid = $1 ;


-- name: DeleteChatMessagesBySesionUUID :exec
DELETE FROM chat_message
WHERE chat_session_uuid = $1;


-- name: GetChatMessagesCount :one
-- Get total chat message count for user in last 10 minutes
SELECT COUNT(*)
FROM chat_message
WHERE user_id = $1
AND created_at >= NOW() - INTERVAL '10 minutes';

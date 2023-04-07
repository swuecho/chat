-- name: GetAllChatPrompts :many
SELECT * FROM chat_prompt 
WHERE is_deleted = false
ORDER BY id;

-- name: GetChatPromptByID :one
SELECT * FROM chat_prompt
WHERE is_deleted = false and  id = $1;

-- name: CreateChatPrompt :one
INSERT INTO chat_prompt (uuid, chat_session_uuid, role, content, token_count, user_id, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateChatPrompt :one
UPDATE chat_prompt SET chat_session_uuid = $2, role = $3, content = $4, score = $5, user_id = $6, updated_at = now(), updated_by = $7
WHERE id = $1
RETURNING *;

-- name: UpdateChatPromptByUUID :one
UPDATE chat_prompt SET content = $2, token_count = $3, updated_at = now()
WHERE uuid = $1 and is_deleted = false
RETURNING *;

-- name: DeleteChatPrompt :exec
UPDATE chat_prompt 
SET is_deleted = true, updated_at = now()
WHERE id = $1;

-- name: GetChatPromptsByUserID :many
SELECT *
FROM chat_prompt 
WHERE user_id = $1 and is_deleted = false
ORDER BY id;

-- name: GetChatPromptsBysession_uuid :many
SELECT *
FROM chat_prompt 
WHERE chat_session_uuid = $1 and is_deleted = false
ORDER BY id;


-- name: GetChatPromptsBySessionUUID :many
SELECT *
FROM chat_prompt 
WHERE chat_session_uuid = $1 and is_deleted = false
ORDER BY id;

-- name: GetOneChatPromptBySessionUUID :one
SELECT *
FROM chat_prompt 
WHERE chat_session_uuid = $1 and is_deleted = false
ORDER BY id
LIMIT 1;




-- name: HasChatPromptPermission :one
SELECT COUNT(*) > 0 as has_permission
FROM chat_prompt cp
INNER JOIN auth_user au ON cp.user_id = au.id
WHERE cp.id = $1 AND (cp.user_id = $2 OR au.is_superuser) AND cp.is_deleted = false;


-- name: DeleteChatPromptByUUID :exec
UPDATE chat_prompt
SET is_deleted = true, updated_at = now()
WHERE uuid = $1;
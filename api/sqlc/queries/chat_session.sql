-- name: GetAllChatSessions :many
SELECT * FROM chat_session 
where active = true
ORDER BY id;

-- name: CreateChatSession :one
INSERT INTO chat_session (user_id, topic, max_length, uuid)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateChatSession :one
UPDATE chat_session SET user_id = $2, topic = $3, updated_at = now(), active = $4
WHERE id = $1
RETURNING *;

-- name: DeleteChatSession :exec
DELETE FROM chat_session 
WHERE id = $1;

-- name: GetChatSessionByID :one
SELECT * FROM chat_session WHERE id = $1;

-- name: GetChatSessionByUUID :one
SELECT * FROM chat_session 
WHERE active = true and uuid = $1
order by updated_at;

-- name: CreateChatSessionByUUID :one
INSERT INTO chat_session (user_id, uuid, topic, created_at, active,  max_length)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateChatSessionByUUID :one
UPDATE chat_session SET user_id = $2, topic = $3, updated_at = now()
WHERE uuid = $1
RETURNING *;

-- name: CreateOrUpdateChatSessionByUUID :one
INSERT INTO chat_session(uuid, user_id, topic, keep_length, max_length, temperature, model, max_tokens, top_p, debug)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (uuid) 
DO UPDATE SET
keep_length = EXCLUDED.keep_length,
max_length = EXCLUDED.max_length,
debug = EXCLUDED.debug,
max_tokens = EXCLUDED.max_tokens,
temperature = EXCLUDED.temperature, 
top_p = EXCLUDED.top_p,
model = EXCLUDED.model,
topic = CASE WHEN chat_session.topic IS NULL THEN EXCLUDED.topic ELSE chat_session.topic END,
updated_at = now()
returning *;

-- name: UpdateChatSessionTopicByUUID :one
INSERT INTO chat_session(uuid, user_id, topic)
VALUES ($1, $2, $3)
ON CONFLICT (uuid) 
DO UPDATE SET
topic = EXCLUDED.topic, 
updated_at = now()
returning *;

-- name: DeleteChatSessionByUUID :exec
update chat_session set active = false
WHERE uuid = $1
returning *;

-- name: GetChatSessionsByUserID :many
SELECT cs.*
FROM chat_session cs
WHERE cs.user_id = $1 and cs.active = true
ORDER BY cs.id;

-- name: HasChatSessionPermission :one
SELECT COUNT(*) > 0 as has_permission
FROM chat_session cs
INNER JOIN auth_user au ON cs.user_id = au.id
WHERE cs.id = $1 AND (cs.user_id = $2 OR au.is_superuser);


-- name: UpdateSessionMaxLength :one
UPDATE chat_session
SET max_length = $2,
    updated_at = now()
WHERE uuid = $1
RETURNING *;

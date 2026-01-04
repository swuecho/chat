-- name: GetAllChatSessions :many
SELECT * FROM chat_session 
where active = true
ORDER BY id;

-- name: CreateChatSession :one
INSERT INTO chat_session (user_id, topic, max_length, uuid, model)
VALUES ($1, $2, $3, $4, $5)
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

-- name: GetChatSessionByUUIDWithInActive :one
SELECT * FROM chat_session 
WHERE uuid = $1
order by updated_at;

-- name: CreateChatSessionByUUID :one
INSERT INTO chat_session (user_id, uuid, topic, created_at, active,  max_length, model)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateChatSessionByUUID :one
UPDATE chat_session SET user_id = $2, topic = $3, updated_at = now()
WHERE uuid = $1
RETURNING *;

-- name: CreateOrUpdateChatSessionByUUID :one
INSERT INTO chat_session(uuid, user_id, topic, max_length, temperature, model, max_tokens, top_p, n, debug, summarize_mode, code_runner_enabled, workspace_id, explore_mode)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (uuid) 
DO UPDATE SET
max_length = EXCLUDED.max_length, 
debug = EXCLUDED.debug,
max_tokens = EXCLUDED.max_tokens,
temperature = EXCLUDED.temperature, 
top_p = EXCLUDED.top_p,
n= EXCLUDED.n,
model = EXCLUDED.model,
summarize_mode = EXCLUDED.summarize_mode,
code_runner_enabled = EXCLUDED.code_runner_enabled,
workspace_id = CASE WHEN EXCLUDED.workspace_id IS NOT NULL THEN EXCLUDED.workspace_id ELSE chat_session.workspace_id END,
topic = CASE WHEN chat_session.topic IS NULL THEN EXCLUDED.topic ELSE chat_session.topic END,
explore_mode = EXCLUDED.explore_mode,
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
LEFT JOIN (
    SELECT chat_session_uuid, MAX(created_at) AS latest_message_time
    FROM chat_message
    GROUP BY chat_session_uuid
) cm ON cs.uuid = cm.chat_session_uuid
WHERE cs.user_id = $1 AND cs.active = true
ORDER BY 
    cm.latest_message_time DESC,
    cs.id DESC;


-- SELECT cs.*
-- FROM chat_session cs
-- WHERE cs.user_id = $1 and cs.active = true
-- ORDER BY cs.updated_at DESC;

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

-- name: CreateChatSessionInWorkspace :one
INSERT INTO chat_session (user_id, uuid, topic, created_at, active, max_length, model, workspace_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: UpdateSessionWorkspace :one
UPDATE chat_session 
SET workspace_id = $2, updated_at = now()
WHERE uuid = $1
RETURNING *;

-- name: GetSessionsByWorkspaceID :many
SELECT cs.*
FROM chat_session cs
LEFT JOIN (
    SELECT chat_session_uuid, MAX(created_at) AS latest_message_time
    FROM chat_message
    GROUP BY chat_session_uuid
) cm ON cs.uuid = cm.chat_session_uuid
WHERE cs.workspace_id = $1 AND cs.active = true
ORDER BY 
    cm.latest_message_time DESC,
    cs.id DESC;

-- name: GetSessionsGroupedByWorkspace :many
SELECT 
    cs.*,
    w.uuid as workspace_uuid,
    w.name as workspace_name,
    w.color as workspace_color,
    w.icon as workspace_icon
FROM chat_session cs
LEFT JOIN chat_workspace w ON cs.workspace_id = w.id
LEFT JOIN (
    SELECT chat_session_uuid, MAX(created_at) AS latest_message_time
    FROM chat_message
    GROUP BY chat_session_uuid
) cm ON cs.uuid = cm.chat_session_uuid
WHERE cs.user_id = $1 AND cs.active = true
ORDER BY 
    w.order_position ASC,
    cm.latest_message_time DESC,
    cs.id DESC;

-- name: MigrateSessionsToDefaultWorkspace :exec
UPDATE chat_session 
SET workspace_id = $2
WHERE user_id = $1 AND workspace_id IS NULL;

-- name: GetSessionsWithoutWorkspace :many
SELECT * FROM chat_session 
WHERE user_id = $1 AND workspace_id IS NULL AND active = true;

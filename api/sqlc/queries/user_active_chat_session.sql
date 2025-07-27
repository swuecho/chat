-- Simplified unified queries for active sessions

-- name: UpsertUserActiveSession :one
INSERT INTO user_active_chat_session (user_id, workspace_id, chat_session_uuid)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, COALESCE(workspace_id, -1))
DO UPDATE SET 
    chat_session_uuid = EXCLUDED.chat_session_uuid,
    updated_at = now()
RETURNING *;

-- name: GetUserActiveSession :one
SELECT * FROM user_active_chat_session 
WHERE user_id = $1 AND (
    (workspace_id IS NULL AND $2::int IS NULL) OR 
    (workspace_id = $2)
);

-- name: GetAllUserActiveSessions :many
SELECT * FROM user_active_chat_session
WHERE user_id = $1
ORDER BY workspace_id NULLS FIRST, updated_at DESC;

-- name: DeleteUserActiveSession :exec
DELETE FROM user_active_chat_session
WHERE user_id = $1 AND (
    (workspace_id IS NULL AND $2::int IS NULL) OR 
    (workspace_id = $2)
);

-- name: DeleteUserActiveSessionBySession :exec
DELETE FROM user_active_chat_session
WHERE user_id = $1 AND chat_session_uuid = $2;


-- name: ListUserActiveChatSessions :many
SELECT * FROM user_active_chat_session ORDER BY id;

-- name: GetUserActiveChatSession :one
SELECT * FROM user_active_chat_session WHERE user_id = $1;

-- name: CreateUserActiveChatSession :one
INSERT INTO user_active_chat_session (user_id, chat_session_uuid)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateUserActiveChatSession :one
UPDATE user_active_chat_session SET chat_session_uuid = @chat_session_uuid, updated_at = now()
WHERE user_id = @user_id
RETURNING *;

-- name: DeleteUserActiveChatSession :exec
DELETE FROM user_active_chat_session WHERE user_id = $1;


-- name: CreateOrUpdateUserActiveChatSession :one
INSERT INTO user_active_chat_session(user_id, chat_session_uuid)
VALUES ($1, $2)
ON CONFLICT (user_id) 
DO UPDATE SET
chat_session_uuid = EXCLUDED.chat_session_uuid,
updated_at = now()
returning *;
-- name: ListUserChatModelPrivileges :many
SELECT * FROM user_chat_model_privilege ORDER BY id;

-- name: ListUserChatModelPrivilegesRateLimit :many
SELECT ucmp.id, au.email as user_email, CONCAT_WS('',au.last_name, au.first_name) as full_name, cm.name chat_model_name, ucmp.rate_limit  
FROM user_chat_model_privilege ucmp 
INNER JOIN chat_model cm ON cm.id = ucmp.chat_model_id
INNER JOIN auth_user au ON au.id = ucmp.user_id
ORDER by au.last_login DESC;
-- TODO add ratelimit
-- LIMIT 1000

-- name: ListUserChatModelPrivilegesByUserID :many
SELECT * FROM user_chat_model_privilege 
WHERE user_id = $1
ORDER BY id;

-- name: UserChatModelPrivilegeByID :one
SELECT * FROM user_chat_model_privilege WHERE id = $1;

-- name: CreateUserChatModelPrivilege :one
INSERT INTO user_chat_model_privilege (user_id, chat_model_id, rate_limit, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUserChatModelPrivilege :one
UPDATE user_chat_model_privilege SET rate_limit = $2, updated_at = now(), updated_by = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUserChatModelPrivilege :exec
DELETE FROM user_chat_model_privilege WHERE id = $1;

-- name: UserChatModelPrivilegeByUserAndModelID :one
SELECT * FROM user_chat_model_privilege WHERE user_id = $1 AND chat_model_id = $2;

-- name: RateLimiteByUserAndSessionUUID :one
SELECT ucmp.rate_limit, cm.name AS chat_model_name
FROM user_chat_model_privilege ucmp
JOIN chat_session cs ON cs.user_id = ucmp.user_id
JOIN chat_model cm ON (cm.id = ucmp.chat_model_id AND cs.model = cm.name and cm.enable_per_mode_ratelimit = true)
WHERE cs.uuid = $1
  AND ucmp.user_id = $2;
  -- AND cs.model = cm.name 

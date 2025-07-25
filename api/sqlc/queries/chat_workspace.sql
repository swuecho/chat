-- name: CreateWorkspace :one
INSERT INTO chat_workspace (uuid, user_id, name, description, color, icon, is_default, order_position)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetWorkspaceByUUID :one
SELECT * FROM chat_workspace 
WHERE uuid = $1;

-- name: GetWorkspacesByUserID :many
SELECT * FROM chat_workspace 
WHERE user_id = $1
ORDER BY order_position ASC, created_at ASC;

-- name: UpdateWorkspace :one
UPDATE chat_workspace 
SET name = $2, description = $3, color = $4, icon = $5, updated_at = now()
WHERE uuid = $1
RETURNING *;

-- name: UpdateWorkspaceOrder :one
UPDATE chat_workspace 
SET order_position = $2, updated_at = now()
WHERE uuid = $1
RETURNING *;

-- name: DeleteWorkspace :exec
DELETE FROM chat_workspace 
WHERE uuid = $1;

-- name: GetDefaultWorkspaceByUserID :one
SELECT * FROM chat_workspace 
WHERE user_id = $1 AND is_default = true
LIMIT 1;

-- name: SetDefaultWorkspace :one
UPDATE chat_workspace 
SET is_default = $2, updated_at = now()
WHERE uuid = $1
RETURNING *;

-- name: GetWorkspaceWithSessionCount :many
SELECT 
    w.*,
    COUNT(cs.id) as session_count
FROM chat_workspace w
LEFT JOIN chat_session cs ON w.id = cs.workspace_id AND cs.active = true
WHERE w.user_id = $1
GROUP BY w.id
ORDER BY w.order_position ASC, w.created_at ASC;

-- name: CreateDefaultWorkspace :one
INSERT INTO chat_workspace (uuid, user_id, name, description, color, icon, is_default, order_position)
VALUES ($1, $2, 'General', 'Default workspace for all conversations', '#6366f1', 'folder', true, 0)
RETURNING *;

-- name: HasWorkspacePermission :one
SELECT COUNT(*) > 0 as has_permission
FROM chat_workspace w
INNER JOIN auth_user au ON w.user_id = au.id
WHERE w.uuid = $1 AND (w.user_id = $2 OR au.is_superuser);
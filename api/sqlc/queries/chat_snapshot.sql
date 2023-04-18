-- name: ListChatSnapshots :many
SELECT * FROM chat_snapshot ORDER BY id;

-- name: ChatSnapshotByID :one
SELECT * FROM chat_snapshot WHERE id = $1;

-- name: CreateChatSnapshot :one
INSERT INTO chat_snapshot (uuid, user_id, title, summary, tags, conversation )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateChatSnapshot :one
UPDATE chat_snapshot
SET uuid = $2, user_id = $3, title = $4, summary = $5, tags = $6, conversation = $7, created_at = $8
WHERE id = $1
RETURNING *;

-- name: DeleteChatSnapshot :exec
DELETE FROM chat_snapshot WHERE id = $1;

-- name: ChatSnapshotByUUID :one
SELECT * FROM chat_snapshot WHERE uuid = $1;


-- name: ChatSnapshotMetaByUserID :many
SELECT uuid, title, summary, tags, created_at
FROM chat_snapshot WHERE user_id = $1
order by id desc;
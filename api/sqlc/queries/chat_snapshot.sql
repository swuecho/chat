-- name: ListChatSnapshots :many
SELECT * FROM chat_snapshot ORDER BY id;

-- name: ChatSnapshotByID :one
SELECT * FROM chat_snapshot WHERE id = $1;

-- name: CreateChatSnapshot :one
INSERT INTO chat_snapshot (uuid, user_id, title, model, summary, tags, conversation ,session, text )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;


-- name: CreateChatBot :one
INSERT INTO chat_snapshot (uuid, user_id, typ, title, model, summary, tags, conversation ,session, text )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateChatSnapshot :one
UPDATE chat_snapshot
SET uuid = $2, user_id = $3, title = $4, summary = $5, tags = $6, conversation = $7, created_at = $8
WHERE id = $1
RETURNING *;


-- name: DeleteChatSnapshot :one
DELETE FROM chat_snapshot WHERE uuid = $1
and user_id = $2
RETURNING *;

-- name: ChatSnapshotByUUID :one
SELECT * FROM chat_snapshot WHERE uuid = $1;

-- name: ChatSnapshotByUserIdAndUuid :one
SELECT * FROM chat_snapshot WHERE user_id = $1 AND uuid = $2;

-- name: ChatSnapshotMetaByUserID :many
SELECT uuid, title, summary, tags, created_at, typ 
FROM chat_snapshot WHERE user_id = $1
order by created_at desc;

-- name: UpdateChatSnapshotMetaByUUID :exec
UPDATE chat_snapshot
SET title = $2, summary = $3
WHERE uuid = $1 and user_id = $4;

-- name: ChatSnapshotSearch :many
SELECT uuid, title, ts_rank(search_vector, websearch_to_tsquery(@search), 1) as rank
FROM chat_snapshot
WHERE search_vector @@ websearch_to_tsquery(@search) AND user_id = $1
ORDER BY rank DESC
LIMIT 20;
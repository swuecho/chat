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
FROM chat_snapshot WHERE user_id = $1 and typ = $2
order by created_at desc
LIMIT $3 OFFSET $4;

-- name: UpdateChatSnapshotMetaByUUID :exec
UPDATE chat_snapshot
SET title = $2, summary = $3
WHERE uuid = $1 and user_id = $4;

-- name: UpdateChatBotModel :one
UPDATE chat_snapshot
SET model = sqlc.arg(input_model),
    session = jsonb_set(session, '{model}', to_jsonb(sqlc.arg(input_model)::text), true)
WHERE chat_snapshot.uuid = sqlc.arg(uuid)
  AND chat_snapshot.user_id = sqlc.arg(bot_user_id)
  AND chat_snapshot.typ = 'chatbot'
  AND EXISTS (
    SELECT 1
    FROM chat_model
    WHERE chat_model.name = sqlc.arg(input_model)
      AND chat_model.is_enable = true
      AND chat_model.user_id IN (SELECT auth_user.id FROM auth_user WHERE auth_user.is_superuser = true)
  )
RETURNING *;

-- name: UpdateChatBotSettings :one
UPDATE chat_snapshot
SET title = sqlc.arg(input_title),
    summary = sqlc.arg(input_summary),
    model = sqlc.arg(input_model),
    session = jsonb_set(session, '{model}', to_jsonb(sqlc.arg(input_model)::text), true)
WHERE chat_snapshot.uuid = sqlc.arg(uuid)
  AND chat_snapshot.user_id = sqlc.arg(bot_user_id)
  AND chat_snapshot.typ = 'chatbot'
  AND EXISTS (
    SELECT 1
    FROM chat_model
    WHERE chat_model.name = sqlc.arg(input_model)
      AND chat_model.is_enable = true
      AND chat_model.user_id IN (SELECT auth_user.id FROM auth_user WHERE auth_user.is_superuser = true)
  )
RETURNING *;

-- name: ChatSnapshotCountByUserIDAndType :one
SELECT COUNT(*)
FROM chat_snapshot
WHERE user_id = $1 AND ($2::text = '' OR typ = $2);

-- name: ChatSnapshotSearch :many
SELECT uuid, title, ts_rank(search_vector, websearch_to_tsquery(@search), 1) as rank
FROM chat_snapshot
WHERE search_vector @@ websearch_to_tsquery(@search) AND user_id = $1
ORDER BY rank DESC
LIMIT 20;

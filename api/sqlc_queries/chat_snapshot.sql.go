// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: chat_snapshot.sql

package sqlc_queries

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const chatSnapshotByID = `-- name: ChatSnapshotByID :one
SELECT id, uuid, user_id, title, summary, model, tags, session, conversation, created_at, text, search_vector FROM chat_snapshot WHERE id = $1
`

func (q *Queries) ChatSnapshotByID(ctx context.Context, id int32) (ChatSnapshot, error) {
	row := q.db.QueryRow(ctx, chatSnapshotByID, id)
	var i ChatSnapshot
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.UserID,
		&i.Title,
		&i.Summary,
		&i.Model,
		&i.Tags,
		&i.Session,
		&i.Conversation,
		&i.CreatedAt,
		&i.Text,
		&i.SearchVector,
	)
	return i, err
}

const chatSnapshotByUUID = `-- name: ChatSnapshotByUUID :one
SELECT id, uuid, user_id, title, summary, model, tags, session, conversation, created_at, text, search_vector FROM chat_snapshot WHERE uuid = $1
`

func (q *Queries) ChatSnapshotByUUID(ctx context.Context, uuid string) (ChatSnapshot, error) {
	row := q.db.QueryRow(ctx, chatSnapshotByUUID, uuid)
	var i ChatSnapshot
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.UserID,
		&i.Title,
		&i.Summary,
		&i.Model,
		&i.Tags,
		&i.Session,
		&i.Conversation,
		&i.CreatedAt,
		&i.Text,
		&i.SearchVector,
	)
	return i, err
}

const chatSnapshotMetaByUserID = `-- name: ChatSnapshotMetaByUserID :many
SELECT uuid, title, summary, tags, created_at
FROM chat_snapshot WHERE user_id = $1
order by created_at desc
`

type ChatSnapshotMetaByUserIDRow struct {
	Uuid      string           `json:"uuid"`
	Title     string           `json:"title"`
	Summary   string           `json:"summary"`
	Tags      []byte           `json:"tags"`
	CreatedAt pgtype.Timestamp `json:"createdAt"`
}

func (q *Queries) ChatSnapshotMetaByUserID(ctx context.Context, userID int32) ([]ChatSnapshotMetaByUserIDRow, error) {
	rows, err := q.db.Query(ctx, chatSnapshotMetaByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatSnapshotMetaByUserIDRow
	for rows.Next() {
		var i ChatSnapshotMetaByUserIDRow
		if err := rows.Scan(
			&i.Uuid,
			&i.Title,
			&i.Summary,
			&i.Tags,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const chatSnapshotSearch = `-- name: ChatSnapshotSearch :many
SELECT uuid, title, ts_rank(search_vector, websearch_to_tsquery($2), 1) as rank
FROM chat_snapshot
WHERE search_vector @@ websearch_to_tsquery($2) AND user_id = $1
ORDER BY rank DESC
LIMIT 20
`

type ChatSnapshotSearchParams struct {
	UserID int32  `json:"userID"`
	Search string `json:"search"`
}

type ChatSnapshotSearchRow struct {
	Uuid  string  `json:"uuid"`
	Title string  `json:"title"`
	Rank  float32 `json:"rank"`
}

func (q *Queries) ChatSnapshotSearch(ctx context.Context, arg ChatSnapshotSearchParams) ([]ChatSnapshotSearchRow, error) {
	rows, err := q.db.Query(ctx, chatSnapshotSearch, arg.UserID, arg.Search)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatSnapshotSearchRow
	for rows.Next() {
		var i ChatSnapshotSearchRow
		if err := rows.Scan(&i.Uuid, &i.Title, &i.Rank); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const createChatSnapshot = `-- name: CreateChatSnapshot :one
INSERT INTO chat_snapshot (uuid, user_id, title, model, summary, tags, conversation ,session, text )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, uuid, user_id, title, summary, model, tags, session, conversation, created_at, text, search_vector
`

type CreateChatSnapshotParams struct {
	Uuid         string `json:"uuid"`
	UserID       int32  `json:"userID"`
	Title        string `json:"title"`
	Model        string `json:"model"`
	Summary      string `json:"summary"`
	Tags         []byte `json:"tags"`
	Conversation []byte `json:"conversation"`
	Session      []byte `json:"session"`
	Text         string `json:"text"`
}

func (q *Queries) CreateChatSnapshot(ctx context.Context, arg CreateChatSnapshotParams) (ChatSnapshot, error) {
	row := q.db.QueryRow(ctx, createChatSnapshot,
		arg.Uuid,
		arg.UserID,
		arg.Title,
		arg.Model,
		arg.Summary,
		arg.Tags,
		arg.Conversation,
		arg.Session,
		arg.Text,
	)
	var i ChatSnapshot
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.UserID,
		&i.Title,
		&i.Summary,
		&i.Model,
		&i.Tags,
		&i.Session,
		&i.Conversation,
		&i.CreatedAt,
		&i.Text,
		&i.SearchVector,
	)
	return i, err
}

const deleteChatSnapshot = `-- name: DeleteChatSnapshot :one
DELETE FROM chat_snapshot WHERE uuid = $1
and user_id = $2
RETURNING id, uuid, user_id, title, summary, model, tags, session, conversation, created_at, text, search_vector
`

type DeleteChatSnapshotParams struct {
	Uuid   string `json:"uuid"`
	UserID int32  `json:"userID"`
}

func (q *Queries) DeleteChatSnapshot(ctx context.Context, arg DeleteChatSnapshotParams) (ChatSnapshot, error) {
	row := q.db.QueryRow(ctx, deleteChatSnapshot, arg.Uuid, arg.UserID)
	var i ChatSnapshot
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.UserID,
		&i.Title,
		&i.Summary,
		&i.Model,
		&i.Tags,
		&i.Session,
		&i.Conversation,
		&i.CreatedAt,
		&i.Text,
		&i.SearchVector,
	)
	return i, err
}

const listChatSnapshots = `-- name: ListChatSnapshots :many
SELECT id, uuid, user_id, title, summary, model, tags, session, conversation, created_at, text, search_vector FROM chat_snapshot ORDER BY id
`

func (q *Queries) ListChatSnapshots(ctx context.Context) ([]ChatSnapshot, error) {
	rows, err := q.db.Query(ctx, listChatSnapshots)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatSnapshot
	for rows.Next() {
		var i ChatSnapshot
		if err := rows.Scan(
			&i.ID,
			&i.Uuid,
			&i.UserID,
			&i.Title,
			&i.Summary,
			&i.Model,
			&i.Tags,
			&i.Session,
			&i.Conversation,
			&i.CreatedAt,
			&i.Text,
			&i.SearchVector,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateChatSnapshot = `-- name: UpdateChatSnapshot :one
UPDATE chat_snapshot
SET uuid = $2, user_id = $3, title = $4, summary = $5, tags = $6, conversation = $7, created_at = $8
WHERE id = $1
RETURNING id, uuid, user_id, title, summary, model, tags, session, conversation, created_at, text, search_vector
`

type UpdateChatSnapshotParams struct {
	ID           int32            `json:"id"`
	Uuid         string           `json:"uuid"`
	UserID       int32            `json:"userID"`
	Title        string           `json:"title"`
	Summary      string           `json:"summary"`
	Tags         []byte           `json:"tags"`
	Conversation []byte           `json:"conversation"`
	CreatedAt    pgtype.Timestamp `json:"createdAt"`
}

func (q *Queries) UpdateChatSnapshot(ctx context.Context, arg UpdateChatSnapshotParams) (ChatSnapshot, error) {
	row := q.db.QueryRow(ctx, updateChatSnapshot,
		arg.ID,
		arg.Uuid,
		arg.UserID,
		arg.Title,
		arg.Summary,
		arg.Tags,
		arg.Conversation,
		arg.CreatedAt,
	)
	var i ChatSnapshot
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.UserID,
		&i.Title,
		&i.Summary,
		&i.Model,
		&i.Tags,
		&i.Session,
		&i.Conversation,
		&i.CreatedAt,
		&i.Text,
		&i.SearchVector,
	)
	return i, err
}

const updateChatSnapshotMetaByUUID = `-- name: UpdateChatSnapshotMetaByUUID :exec
UPDATE chat_snapshot
SET title = $2, summary = $3
WHERE uuid = $1 and user_id = $4
`

type UpdateChatSnapshotMetaByUUIDParams struct {
	Uuid    string `json:"uuid"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	UserID  int32  `json:"userID"`
}

func (q *Queries) UpdateChatSnapshotMetaByUUID(ctx context.Context, arg UpdateChatSnapshotMetaByUUIDParams) error {
	_, err := q.db.Exec(ctx, updateChatSnapshotMetaByUUID,
		arg.Uuid,
		arg.Title,
		arg.Summary,
		arg.UserID,
	)
	return err
}

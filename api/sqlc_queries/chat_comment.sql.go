// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: chat_comment.sql

package sqlc_queries

import (
	"context"
	"time"
)

const createChatComment = `-- name: CreateChatComment :one
INSERT INTO chat_comment (
    uuid,
    chat_session_uuid,
    chat_message_uuid, 
    content,
    created_by,
    updated_by
) VALUES (
    $1, $2, $3, $4, $5, $5
) RETURNING id, uuid, chat_session_uuid, chat_message_uuid, content, created_at, updated_at, created_by, updated_by
`

type CreateChatCommentParams struct {
	Uuid            string `json:"uuid"`
	ChatSessionUuid string `json:"chatSessionUuid"`
	ChatMessageUuid string `json:"chatMessageUuid"`
	Content         string `json:"content"`
	CreatedBy       int32  `json:"createdBy"`
}

func (q *Queries) CreateChatComment(ctx context.Context, arg CreateChatCommentParams) (ChatComment, error) {
	row := q.db.QueryRowContext(ctx, createChatComment,
		arg.Uuid,
		arg.ChatSessionUuid,
		arg.ChatMessageUuid,
		arg.Content,
		arg.CreatedBy,
	)
	var i ChatComment
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.ChatMessageUuid,
		&i.Content,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
	)
	return i, err
}

const getCommentsByMessageUUID = `-- name: GetCommentsByMessageUUID :many
SELECT 
    cc.uuid,
    cc.content,
    cc.created_at,
    au.username AS author_username,
    au.email AS author_email
FROM chat_comment cc
JOIN auth_user au ON cc.created_by = au.id
WHERE cc.chat_message_uuid = $1
ORDER BY cc.created_at DESC
`

type GetCommentsByMessageUUIDRow struct {
	Uuid           string    `json:"uuid"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	AuthorUsername string    `json:"authorUsername"`
	AuthorEmail    string    `json:"authorEmail"`
}

func (q *Queries) GetCommentsByMessageUUID(ctx context.Context, chatMessageUuid string) ([]GetCommentsByMessageUUIDRow, error) {
	rows, err := q.db.QueryContext(ctx, getCommentsByMessageUUID, chatMessageUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetCommentsByMessageUUIDRow
	for rows.Next() {
		var i GetCommentsByMessageUUIDRow
		if err := rows.Scan(
			&i.Uuid,
			&i.Content,
			&i.CreatedAt,
			&i.AuthorUsername,
			&i.AuthorEmail,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCommentsBySessionUUID = `-- name: GetCommentsBySessionUUID :many
SELECT 
    cc.uuid,
    cc.chat_message_uuid,
    cc.content,
    cc.created_at,
    au.username AS author_username,
    au.email AS author_email
FROM chat_comment cc
JOIN auth_user au ON cc.created_by = au.id
WHERE cc.chat_session_uuid = $1
ORDER BY cc.created_at DESC
`

type GetCommentsBySessionUUIDRow struct {
	Uuid            string    `json:"uuid"`
	ChatMessageUuid string    `json:"chatMessageUuid"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"createdAt"`
	AuthorUsername  string    `json:"authorUsername"`
	AuthorEmail     string    `json:"authorEmail"`
}

func (q *Queries) GetCommentsBySessionUUID(ctx context.Context, chatSessionUuid string) ([]GetCommentsBySessionUUIDRow, error) {
	rows, err := q.db.QueryContext(ctx, getCommentsBySessionUUID, chatSessionUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetCommentsBySessionUUIDRow
	for rows.Next() {
		var i GetCommentsBySessionUUIDRow
		if err := rows.Scan(
			&i.Uuid,
			&i.ChatMessageUuid,
			&i.Content,
			&i.CreatedAt,
			&i.AuthorUsername,
			&i.AuthorEmail,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

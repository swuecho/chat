// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: chat_file.sql

package sqlc_queries

import (
	"context"
	"time"
)

const createChatFile = `-- name: CreateChatFile :one
INSERT INTO chat_file (name, data, user_id, chat_session_uuid, mime_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, name, data, created_at, user_id, chat_session_uuid, mime_type
`

type CreateChatFileParams struct {
	Name            string `json:"name"`
	Data            []byte `json:"data"`
	UserID          int32  `json:"userID"`
	ChatSessionUuid string `json:"chatSessionUuid"`
	MimeType        string `json:"mimeType"`
}

func (q *Queries) CreateChatFile(ctx context.Context, arg CreateChatFileParams) (ChatFile, error) {
	row := q.db.QueryRowContext(ctx, createChatFile,
		arg.Name,
		arg.Data,
		arg.UserID,
		arg.ChatSessionUuid,
		arg.MimeType,
	)
	var i ChatFile
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Data,
		&i.CreatedAt,
		&i.UserID,
		&i.ChatSessionUuid,
		&i.MimeType,
	)
	return i, err
}

const deleteChatFile = `-- name: DeleteChatFile :one
DELETE FROM chat_file
WHERE id = $1
RETURNING id, name, data, created_at, user_id, chat_session_uuid, mime_type
`

func (q *Queries) DeleteChatFile(ctx context.Context, id int32) (ChatFile, error) {
	row := q.db.QueryRowContext(ctx, deleteChatFile, id)
	var i ChatFile
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Data,
		&i.CreatedAt,
		&i.UserID,
		&i.ChatSessionUuid,
		&i.MimeType,
	)
	return i, err
}

const getChatFileByID = `-- name: GetChatFileByID :one
SELECT id, name, data, created_at, user_id, chat_session_uuid
FROM chat_file
WHERE id = $1
`

type GetChatFileByIDRow struct {
	ID              int32     `json:"id"`
	Name            string    `json:"name"`
	Data            []byte    `json:"data"`
	CreatedAt       time.Time `json:"createdAt"`
	UserID          int32     `json:"userID"`
	ChatSessionUuid string    `json:"chatSessionUuid"`
}

func (q *Queries) GetChatFileByID(ctx context.Context, id int32) (GetChatFileByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getChatFileByID, id)
	var i GetChatFileByIDRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Data,
		&i.CreatedAt,
		&i.UserID,
		&i.ChatSessionUuid,
	)
	return i, err
}

const listChatFilesBySessionUUID = `-- name: ListChatFilesBySessionUUID :many
SELECT id, name
FROM chat_file
WHERE user_id = $1 and chat_session_uuid = $2
ORDER BY created_at
`

type ListChatFilesBySessionUUIDParams struct {
	UserID          int32  `json:"userID"`
	ChatSessionUuid string `json:"chatSessionUuid"`
}

type ListChatFilesBySessionUUIDRow struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (q *Queries) ListChatFilesBySessionUUID(ctx context.Context, arg ListChatFilesBySessionUUIDParams) ([]ListChatFilesBySessionUUIDRow, error) {
	rows, err := q.db.QueryContext(ctx, listChatFilesBySessionUUID, arg.UserID, arg.ChatSessionUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListChatFilesBySessionUUIDRow
	for rows.Next() {
		var i ListChatFilesBySessionUUIDRow
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
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

const listChatFilesWithContentBySessionUUID = `-- name: ListChatFilesWithContentBySessionUUID :many
SELECT id, name, data, created_at, user_id, chat_session_uuid, mime_type
FROM chat_file
WHERE chat_session_uuid = $1
ORDER BY created_at
`

func (q *Queries) ListChatFilesWithContentBySessionUUID(ctx context.Context, chatSessionUuid string) ([]ChatFile, error) {
	rows, err := q.db.QueryContext(ctx, listChatFilesWithContentBySessionUUID, chatSessionUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatFile
	for rows.Next() {
		var i ChatFile
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Data,
			&i.CreatedAt,
			&i.UserID,
			&i.ChatSessionUuid,
			&i.MimeType,
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

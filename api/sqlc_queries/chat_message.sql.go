// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.17.2
// source: chat_message.sql

package sqlc_queries

import (
	"context"
	"encoding/json"
)

const createChatMessage = `-- name: CreateChatMessage :one
INSERT INTO chat_message (chat_session_uuid, uuid, role, content, token_count, score, user_id, created_by, updated_by, raw)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw
`

type CreateChatMessageParams struct {
	ChatSessionUuid string
	Uuid            string
	Role            string
	Content         string
	TokenCount      int32
	Score           float64
	UserID          int32
	CreatedBy       int32
	UpdatedBy       int32
	Raw             json.RawMessage
}

func (q *Queries) CreateChatMessage(ctx context.Context, arg CreateChatMessageParams) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, createChatMessage,
		arg.ChatSessionUuid,
		arg.Uuid,
		arg.Role,
		arg.Content,
		arg.TokenCount,
		arg.Score,
		arg.UserID,
		arg.CreatedBy,
		arg.UpdatedBy,
		arg.Raw,
	)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const deleteChatMessage = `-- name: DeleteChatMessage :exec
UPDATE chat_message set is_deleted = true, updated_at = now()
WHERE id = $1
`

func (q *Queries) DeleteChatMessage(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteChatMessage, id)
	return err
}

const deleteChatMessageByUUID = `-- name: DeleteChatMessageByUUID :exec
UPDATE chat_message SET is_deleted = true, updated_at = now()
WHERE uuid = $1
`

func (q *Queries) DeleteChatMessageByUUID(ctx context.Context, uuid string) error {
	_, err := q.db.ExecContext(ctx, deleteChatMessageByUUID, uuid)
	return err
}

const deleteChatMessagesBySesionUUID = `-- name: DeleteChatMessagesBySesionUUID :exec
UPDATE chat_message 
SET is_deleted = true, updated_at = now()
WHERE id in (SELECT id from chat_message WHERE is_deleted = false and chat_message.chat_session_uuid = $1 ORDER BY id OFFSET $2)
`

type DeleteChatMessagesBySesionUUIDParams struct {
	ChatSessionUuid string
	Offset          int32
}

func (q *Queries) DeleteChatMessagesBySesionUUID(ctx context.Context, arg DeleteChatMessagesBySesionUUIDParams) error {
	_, err := q.db.ExecContext(ctx, deleteChatMessagesBySesionUUID, arg.ChatSessionUuid, arg.Offset)
	return err
}

const getAllChatMessages = `-- name: GetAllChatMessages :many
SELECT id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw FROM chat_message 
WHERE is_deleted = false
ORDER BY id
`

func (q *Queries) GetAllChatMessages(ctx context.Context) ([]ChatMessage, error) {
	rows, err := q.db.QueryContext(ctx, getAllChatMessages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatMessage
	for rows.Next() {
		var i ChatMessage
		if err := rows.Scan(
			&i.ID,
			&i.Uuid,
			&i.ChatSessionUuid,
			&i.Role,
			&i.Content,
			&i.Score,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.IsDeleted,
			&i.TokenCount,
			&i.Raw,
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

const getChatMessageByID = `-- name: GetChatMessageByID :one
SELECT id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw FROM chat_message 
WHERE is_deleted = false and id = $1
`

func (q *Queries) GetChatMessageByID(ctx context.Context, id int32) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, getChatMessageByID, id)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const getChatMessageBySessionUUID = `-- name: GetChatMessageBySessionUUID :one
SELECT cm.id, cm.uuid, cm.chat_session_uuid, cm.role, cm.content, cm.score, cm.user_id, cm.created_at, cm.updated_at, cm.created_by, cm.updated_by, cm.is_deleted, cm.token_count, cm.raw
FROM chat_message cm
INNER JOIN chat_session cs ON cm.chat_session_uuid = cs.uuid
WHERE cm.is_deleted = false and cs.active = true and cs.uuid = $1 
ORDER BY cm.id 
OFFSET $2
LIMIT $1
`

type GetChatMessageBySessionUUIDParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) GetChatMessageBySessionUUID(ctx context.Context, arg GetChatMessageBySessionUUIDParams) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, getChatMessageBySessionUUID, arg.Limit, arg.Offset)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const getChatMessageByUUID = `-- name: GetChatMessageByUUID :one

SELECT id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw FROM chat_message 
WHERE is_deleted = false and uuid = $1
`

// -- UUID ----
func (q *Queries) GetChatMessageByUUID(ctx context.Context, uuid string) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, getChatMessageByUUID, uuid)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const getChatMessagesBySessionUUID = `-- name: GetChatMessagesBySessionUUID :many
SELECT cm.id, cm.uuid, cm.chat_session_uuid, cm.role, cm.content, cm.score, cm.user_id, cm.created_at, cm.updated_at, cm.created_by, cm.updated_by, cm.is_deleted, cm.token_count, cm.raw
FROM chat_message cm
INNER JOIN chat_session cs ON cm.chat_session_uuid = cs.uuid
WHERE cm.is_deleted = false and cs.active = true and cs.uuid = $1  
ORDER BY cm.id 
OFFSET $2
LIMIT $3
`

type GetChatMessagesBySessionUUIDParams struct {
	Uuid   string
	Offset int32
	Limit  int32
}

func (q *Queries) GetChatMessagesBySessionUUID(ctx context.Context, arg GetChatMessagesBySessionUUIDParams) ([]ChatMessage, error) {
	rows, err := q.db.QueryContext(ctx, getChatMessagesBySessionUUID, arg.Uuid, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatMessage
	for rows.Next() {
		var i ChatMessage
		if err := rows.Scan(
			&i.ID,
			&i.Uuid,
			&i.ChatSessionUuid,
			&i.Role,
			&i.Content,
			&i.Score,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.IsDeleted,
			&i.TokenCount,
			&i.Raw,
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

const getChatMessagesCount = `-- name: GetChatMessagesCount :one
SELECT COUNT(*)
FROM chat_message
WHERE user_id = $1
AND created_at >= NOW() - INTERVAL '10 minutes'
`

// Get total chat message count for user in last 10 minutes
func (q *Queries) GetChatMessagesCount(ctx context.Context, userID int32) (int64, error) {
	row := q.db.QueryRowContext(ctx, getChatMessagesCount, userID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getFirstMessageBySessionUUID = `-- name: GetFirstMessageBySessionUUID :one
SELECT id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw
FROM chat_message
WHERE chat_session_uuid = $1 and is_deleted = false
ORDER BY created_at 
LIMIT 1
`

func (q *Queries) GetFirstMessageBySessionUUID(ctx context.Context, chatSessionUuid string) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, getFirstMessageBySessionUUID, chatSessionUuid)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const getLastNChatMessages = `-- name: GetLastNChatMessages :many
SELECT id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw
FROM chat_message
WHERE chat_message.id in (
    SELECT id 
    FROM chat_message cm
    WHERE cm.chat_session_uuid = $3 
            AND cm.id < (SELECT id FROM chat_message WHERE chat_message.uuid = $1)
            AND cm.is_deleted = false
    ORDER BY cm.created_at DESC
    LIMIT $2
) 
ORDER BY created_at
`

type GetLastNChatMessagesParams struct {
	Uuid            string
	Limit           int32
	ChatSessionUuid string
}

func (q *Queries) GetLastNChatMessages(ctx context.Context, arg GetLastNChatMessagesParams) ([]ChatMessage, error) {
	rows, err := q.db.QueryContext(ctx, getLastNChatMessages, arg.Uuid, arg.Limit, arg.ChatSessionUuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatMessage
	for rows.Next() {
		var i ChatMessage
		if err := rows.Scan(
			&i.ID,
			&i.Uuid,
			&i.ChatSessionUuid,
			&i.Role,
			&i.Content,
			&i.Score,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.IsDeleted,
			&i.TokenCount,
			&i.Raw,
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

const getLatestMessagesBySessionUUID = `-- name: GetLatestMessagesBySessionUUID :many
SELECT id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw
FROM chat_message
Where chat_message.id in 
(
    SELECT chat_message.id
    FROM chat_message
    WHERE chat_message.chat_session_uuid = $1 and chat_message.is_deleted = false
    ORDER BY created_at DESC
    LIMIT $2
)
ORDER BY created_at
`

type GetLatestMessagesBySessionUUIDParams struct {
	ChatSessionUuid string
	Limit           int32
}

func (q *Queries) GetLatestMessagesBySessionUUID(ctx context.Context, arg GetLatestMessagesBySessionUUIDParams) ([]ChatMessage, error) {
	rows, err := q.db.QueryContext(ctx, getLatestMessagesBySessionUUID, arg.ChatSessionUuid, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChatMessage
	for rows.Next() {
		var i ChatMessage
		if err := rows.Scan(
			&i.ID,
			&i.Uuid,
			&i.ChatSessionUuid,
			&i.Role,
			&i.Content,
			&i.Score,
			&i.UserID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.IsDeleted,
			&i.TokenCount,
			&i.Raw,
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

const hasChatMessagePermission = `-- name: HasChatMessagePermission :one
SELECT COUNT(*) > 0 as has_permission
FROM chat_message cm
INNER JOIN chat_session cs ON cm.chat_session_uuid = cs.uuid
INNER JOIN auth_user au ON cs.user_id = au.id
WHERE cm.is_deleted = false and  cm.id = $1 AND (cs.user_id = $2 OR au.is_superuser) and cs.active = true
`

type HasChatMessagePermissionParams struct {
	ID     int32
	UserID int32
}

func (q *Queries) HasChatMessagePermission(ctx context.Context, arg HasChatMessagePermissionParams) (bool, error) {
	row := q.db.QueryRowContext(ctx, hasChatMessagePermission, arg.ID, arg.UserID)
	var has_permission bool
	err := row.Scan(&has_permission)
	return has_permission, err
}

const updateChatMessage = `-- name: UpdateChatMessage :one
UPDATE chat_message SET role = $2, content = $3, score = $4, user_id = $5, updated_by = $6, updated_at = now()
WHERE id = $1
RETURNING id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw
`

type UpdateChatMessageParams struct {
	ID        int32
	Role      string
	Content   string
	Score     float64
	UserID    int32
	UpdatedBy int32
}

func (q *Queries) UpdateChatMessage(ctx context.Context, arg UpdateChatMessageParams) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, updateChatMessage,
		arg.ID,
		arg.Role,
		arg.Content,
		arg.Score,
		arg.UserID,
		arg.UpdatedBy,
	)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const updateChatMessageByUUID = `-- name: UpdateChatMessageByUUID :one
UPDATE chat_message SET content = $2, token_count = $3,  updated_at = now() 
WHERE uuid = $1
RETURNING id, uuid, chat_session_uuid, role, content, score, user_id, created_at, updated_at, created_by, updated_by, is_deleted, token_count, raw
`

type UpdateChatMessageByUUIDParams struct {
	Uuid       string
	Content    string
	TokenCount int32
}

func (q *Queries) UpdateChatMessageByUUID(ctx context.Context, arg UpdateChatMessageByUUIDParams) (ChatMessage, error) {
	row := q.db.QueryRowContext(ctx, updateChatMessageByUUID, arg.Uuid, arg.Content, arg.TokenCount)
	var i ChatMessage
	err := row.Scan(
		&i.ID,
		&i.Uuid,
		&i.ChatSessionUuid,
		&i.Role,
		&i.Content,
		&i.Score,
		&i.UserID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CreatedBy,
		&i.UpdatedBy,
		&i.IsDeleted,
		&i.TokenCount,
		&i.Raw,
	)
	return i, err
}

const updateChatMessageContent = `-- name: UpdateChatMessageContent :exec
UPDATE chat_message
SET content = $2, updated_at = now(), token_count = $3
WHERE uuid = $1
`

type UpdateChatMessageContentParams struct {
	Uuid       string
	Content    string
	TokenCount int32
}

func (q *Queries) UpdateChatMessageContent(ctx context.Context, arg UpdateChatMessageContentParams) error {
	_, err := q.db.ExecContext(ctx, updateChatMessageContent, arg.Uuid, arg.Content, arg.TokenCount)
	return err
}

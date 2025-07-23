package repository

import (
	"context"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type chatMessageRepository struct {
	queries *sqlc_queries.Queries
}

func NewChatMessageRepository(q *sqlc_queries.Queries) ChatMessageRepository {
	return &chatMessageRepository{queries: q}
}

func (r *chatMessageRepository) GetBySessionUUID(ctx context.Context, sessionUUID string, limit int32) ([]sqlc_queries.ChatMessage, error) {
	return r.queries.GetChatMessagesBySessionUUID(ctx, sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid:   sessionUUID,
		Offset: 0,
		Limit:  limit,
	})
}

func (r *chatMessageRepository) Create(ctx context.Context, params sqlc_queries.CreateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	return r.queries.CreateChatMessage(ctx, params)
}

func (r *chatMessageRepository) GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatMessage, error) {
	return r.queries.GetChatMessageByUUID(ctx, uuid)
}

func (r *chatMessageRepository) GetByID(ctx context.Context, id int32) (sqlc_queries.ChatMessage, error) {
	return r.queries.GetChatMessageByID(ctx, id)
}

func (r *chatMessageRepository) GetLatestBySessionUUID(ctx context.Context, sessionUUID string, limit int32) ([]sqlc_queries.ChatMessage, error) {
	return r.queries.GetLatestMessagesBySessionUUID(ctx, sqlc_queries.GetLatestMessagesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		Limit:           limit,
	})
}

func (r *chatMessageRepository) GetFirstBySessionUUID(ctx context.Context, sessionUUID string) (sqlc_queries.ChatMessage, error) {
	return r.queries.GetFirstMessageBySessionUUID(ctx, sessionUUID)
}

func (r *chatMessageRepository) GetAll(ctx context.Context) ([]sqlc_queries.ChatMessage, error) {
	return r.queries.GetAllChatMessages(ctx)
}

func (r *chatMessageRepository) Update(ctx context.Context, params sqlc_queries.UpdateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	return r.queries.UpdateChatMessage(ctx, params)
}

func (r *chatMessageRepository) UpdateByUUID(ctx context.Context, params sqlc_queries.UpdateChatMessageByUUIDParams) (sqlc_queries.ChatMessage, error) {
	return r.queries.UpdateChatMessageByUUID(ctx, params)
}

func (r *chatMessageRepository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteChatMessage(ctx, id)
}

func (r *chatMessageRepository) DeleteByUUID(ctx context.Context, uuid string) error {
	return r.queries.DeleteChatMessageByUUID(ctx, uuid)
}
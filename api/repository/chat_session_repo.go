package repository

import (
	"context"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type chatSessionRepository struct {
	queries *sqlc_queries.Queries
}

func NewChatSessionRepository(q *sqlc_queries.Queries) ChatSessionRepository {
	return &chatSessionRepository{queries: q}
}

func (r *chatSessionRepository) GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatSession, error) {
	return r.queries.GetChatSessionByUUID(ctx, uuid)
}

func (r *chatSessionRepository) GetByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatSession, error) {
	return r.queries.GetChatSessionsByUserID(ctx, userID)
}

func (r *chatSessionRepository) Create(ctx context.Context, params sqlc_queries.CreateChatSessionParams) (sqlc_queries.ChatSession, error) {
	return r.queries.CreateChatSession(ctx, params)
}

func (r *chatSessionRepository) Update(ctx context.Context, params sqlc_queries.UpdateChatSessionParams) (sqlc_queries.ChatSession, error) {
	return r.queries.UpdateChatSession(ctx, params)
}

func (r *chatSessionRepository) UpdateTopicByUUID(ctx context.Context, params sqlc_queries.UpdateChatSessionTopicByUUIDParams) (sqlc_queries.ChatSession, error) {
	return r.queries.UpdateChatSessionTopicByUUID(ctx, params)
}

func (r *chatSessionRepository) Delete(ctx context.Context, uuid string) error {
	return r.queries.DeleteChatSessionByUUID(ctx, uuid)
}
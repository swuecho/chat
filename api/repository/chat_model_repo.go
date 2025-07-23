package repository

import (
	"context"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type chatModelRepository struct {
	queries *sqlc_queries.Queries
}

func NewChatModelRepository(q *sqlc_queries.Queries) ChatModelRepository {
	return &chatModelRepository{queries: q}
}

func (r *chatModelRepository) GetByName(ctx context.Context, name string) (sqlc_queries.ChatModel, error) {
	return r.queries.ChatModelByName(ctx, name)
}

func (r *chatModelRepository) GetAll(ctx context.Context) ([]sqlc_queries.ChatModel, error) {
	return r.queries.ListChatModels(ctx)
}

func (r *chatModelRepository) GetByID(ctx context.Context, id int32) (sqlc_queries.ChatModel, error) {
	return r.queries.ChatModelByID(ctx, id)
}

func (r *chatModelRepository) GetSystemModels(ctx context.Context) ([]sqlc_queries.ChatModel, error) {
	return r.queries.ListSystemChatModels(ctx)
}

func (r *chatModelRepository) GetModelUsageStats(ctx context.Context, timePeriod string) ([]sqlc_queries.GetLatestUsageTimeOfModelRow, error) {
	return r.queries.GetLatestUsageTimeOfModel(ctx, timePeriod)
}

func (r *chatModelRepository) GetDefaultModel(ctx context.Context) (sqlc_queries.ChatModel, error) {
	return r.queries.GetDefaultChatModel(ctx)
}

func (r *chatModelRepository) Create(ctx context.Context, params sqlc_queries.CreateChatModelParams) (sqlc_queries.ChatModel, error) {
	return r.queries.CreateChatModel(ctx, params)
}

func (r *chatModelRepository) Update(ctx context.Context, params sqlc_queries.UpdateChatModelParams) (sqlc_queries.ChatModel, error) {
	return r.queries.UpdateChatModel(ctx, params)
}

func (r *chatModelRepository) Delete(ctx context.Context, id int32) error {
	// The DeleteChatModel requires both ID and UserID, but we don't have userID in this interface
	// For now, we'll use a placeholder userID - this should be enhanced with proper authorization
	params := sqlc_queries.DeleteChatModelParams{
		ID:     id,
		UserID: 0, // This should be passed from the service layer
	}
	return r.queries.DeleteChatModel(ctx, params)
}
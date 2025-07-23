package repository

import (
	"context"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type chatPromptRepository struct {
	queries *sqlc_queries.Queries
}

func NewChatPromptRepository(q *sqlc_queries.Queries) ChatPromptRepository {
	return &chatPromptRepository{queries: q}
}

func (r *chatPromptRepository) GetBySessionUUID(ctx context.Context, sessionUUID string) ([]sqlc_queries.ChatPrompt, error) {
	return r.queries.GetChatPromptsBySessionUUID(ctx, sessionUUID)
}

func (r *chatPromptRepository) GetByID(ctx context.Context, id int32) (sqlc_queries.ChatPrompt, error) {
	return r.queries.GetChatPromptByID(ctx, id)
}

func (r *chatPromptRepository) GetAll(ctx context.Context) ([]sqlc_queries.ChatPrompt, error) {
	return r.queries.GetAllChatPrompts(ctx)
}

func (r *chatPromptRepository) Create(ctx context.Context, params sqlc_queries.CreateChatPromptParams) (sqlc_queries.ChatPrompt, error) {
	return r.queries.CreateChatPrompt(ctx, params)
}

func (r *chatPromptRepository) Update(ctx context.Context, params sqlc_queries.UpdateChatPromptParams) (sqlc_queries.ChatPrompt, error) {
	return r.queries.UpdateChatPrompt(ctx, params)
}

func (r *chatPromptRepository) Delete(ctx context.Context, id int32) error {
	return r.queries.DeleteChatPrompt(ctx, id)
}
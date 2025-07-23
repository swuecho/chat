package repository

import (
	"context"
	
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Start with minimal core repositories to demonstrate the pattern

// ChatSessionRepository handles chat session database operations
type ChatSessionRepository interface {
	GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatSession, error)
	GetByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatSession, error)
	Create(ctx context.Context, params sqlc_queries.CreateChatSessionParams) (sqlc_queries.ChatSession, error)
	Update(ctx context.Context, params sqlc_queries.UpdateChatSessionParams) (sqlc_queries.ChatSession, error)
	UpdateTopicByUUID(ctx context.Context, params sqlc_queries.UpdateChatSessionTopicByUUIDParams) (sqlc_queries.ChatSession, error)
	Delete(ctx context.Context, uuid string) error
}

// ChatMessageRepository handles chat message database operations
type ChatMessageRepository interface {
	GetBySessionUUID(ctx context.Context, sessionUUID string, limit int32) ([]sqlc_queries.ChatMessage, error)
	GetByID(ctx context.Context, id int32) (sqlc_queries.ChatMessage, error)
	GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatMessage, error)
	GetLatestBySessionUUID(ctx context.Context, sessionUUID string, limit int32) ([]sqlc_queries.ChatMessage, error)
	GetFirstBySessionUUID(ctx context.Context, sessionUUID string) (sqlc_queries.ChatMessage, error)
	GetAll(ctx context.Context) ([]sqlc_queries.ChatMessage, error)
	Create(ctx context.Context, params sqlc_queries.CreateChatMessageParams) (sqlc_queries.ChatMessage, error)
	Update(ctx context.Context, params sqlc_queries.UpdateChatMessageParams) (sqlc_queries.ChatMessage, error)
	UpdateByUUID(ctx context.Context, params sqlc_queries.UpdateChatMessageByUUIDParams) (sqlc_queries.ChatMessage, error)
	Delete(ctx context.Context, id int32) error
	DeleteByUUID(ctx context.Context, uuid string) error
}

// ChatModelRepository handles chat model database operations
type ChatModelRepository interface {
	GetByName(ctx context.Context, name string) (sqlc_queries.ChatModel, error)
	GetByID(ctx context.Context, id int32) (sqlc_queries.ChatModel, error)
	GetAll(ctx context.Context) ([]sqlc_queries.ChatModel, error)
	GetSystemModels(ctx context.Context) ([]sqlc_queries.ChatModel, error)
	GetModelUsageStats(ctx context.Context, timePeriod string) ([]sqlc_queries.GetLatestUsageTimeOfModelRow, error)
	GetDefaultModel(ctx context.Context) (sqlc_queries.ChatModel, error)
	Create(ctx context.Context, params sqlc_queries.CreateChatModelParams) (sqlc_queries.ChatModel, error)
	Update(ctx context.Context, params sqlc_queries.UpdateChatModelParams) (sqlc_queries.ChatModel, error)
	Delete(ctx context.Context, id int32) error
}

// ChatPromptRepository handles chat prompt database operations
type ChatPromptRepository interface {
	GetBySessionUUID(ctx context.Context, sessionUUID string) ([]sqlc_queries.ChatPrompt, error)
	GetByID(ctx context.Context, id int32) (sqlc_queries.ChatPrompt, error)
	GetAll(ctx context.Context) ([]sqlc_queries.ChatPrompt, error)
	Create(ctx context.Context, params sqlc_queries.CreateChatPromptParams) (sqlc_queries.ChatPrompt, error)
	Update(ctx context.Context, params sqlc_queries.UpdateChatPromptParams) (sqlc_queries.ChatPrompt, error)
	Delete(ctx context.Context, id int32) error
}

// CoreRepositoryManager aggregates core repositories  
type CoreRepositoryManager interface {
	ChatSession() ChatSessionRepository
	ChatMessage() ChatMessageRepository
	ChatModel() ChatModelRepository
	ChatPrompt() ChatPromptRepository
}
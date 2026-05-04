// Package provider defines LLM provider interfaces and shared configuration.
//
// Providers implement the ChatModel interface to support different LLM backends:
// OpenAI, Claude, Gemini, Ollama, and custom API-compatible models.
//
// The Handler interface decouples providers from the HTTP layer,
// allowing them to be tested independently.
package provider

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// StreamChunk represents a single chunk in a streaming LLM response.
type StreamChunk struct {
	ID          string            // answer ID
	Content     string            // delta text content
	Done        bool              // true for the terminal chunk
	FinalAnswer *models.LLMAnswer // set on Done (nil on error)
	Err         error             // non-nil if a stream error occurred
}

// ChatModel is the interface all LLM providers must implement.
// Stream returns a channel of StreamChunk and an optional immediate error.
// The channel is closed when streaming completes or fails.
type ChatModel interface {
	Stream(ctx context.Context, session sqlc_queries.ChatSession,
		messages []models.Message, chatUuid string,
		regenerate bool, stream bool) (<-chan StreamChunk, error)
}

// Config holds global configuration needed by providers.
type Config struct {
	OpenAIKey    string
	OpenAIProxy  string
	RateLimiter  *rate.Limiter
	DefaultLimit int
}

// Handler provides request-scoped dependencies that providers need.
type Handler interface {
	Queries() *sqlc_queries.Queries
	CheckModelAccess(ctx context.Context, chatSessionUuid, model string, userID int32) error
	Config() Config
}

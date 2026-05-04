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
	"net/http"

	"golang.org/x/time/rate"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatModel is the interface all LLM providers must implement.
type ChatModel interface {
	Stream(ctx context.Context, w http.ResponseWriter, session sqlc_queries.ChatSession,
		messages []models.Message, chatUuid string,
		regenerate bool, stream bool) (*models.LLMAnswer, error)
}

// Config holds global configuration needed by providers.
type Config struct {
	OpenAIKey     string
	OpenAIProxy   string
	RateLimiter   *rate.Limiter
	DefaultLimit  int
}

// Handler provides request-scoped dependencies that providers need.
type Handler interface {
	Queries() *sqlc_queries.Queries
	CheckModelAccess(w http.ResponseWriter, chatSessionUuid, model string, userID int32) bool
	Config() Config
}

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
)

// Session contains only the chat settings needed by an LLM provider. It is
// intentionally independent of the database representation of a chat session.
type Session struct {
	UUID        string
	UserID      int32
	Model       string
	MaxTokens   int32
	Temperature float64
	TopP        float64
	N           int32
	Debug       bool
}

// ModelConfig describes an upstream model endpoint without exposing a
// persistence record to provider implementations.
type ModelConfig struct {
	Name                    string
	URL                     string
	APIAuthHeader           string
	APIAuthKey              string
	APIType                 string
	EnablePerModelRateLimit bool
}

// File contains the attachment data providers may add to an LLM request.
type File struct {
	Name     string
	Data     []byte
	MIMEType string
}

// Request is a fully resolved LLM invocation. Providers do not perform
// persistence lookups; the application supplies all required configuration.
type Request struct {
	Session    Session
	Model      ModelConfig
	Files      []File
	Messages   []models.Message
	ChatUUID   string
	Regenerate bool
	Stream     bool
	NewID      func() string
}

// StreamChunk represents a single chunk in a streaming LLM response.
type StreamChunk struct {
	ID          string            // answer ID
	Content     string            // delta text content
	Done        bool              // true for the terminal chunk
	FinalAnswer *models.LLMAnswer // set on Done (nil on error)
	Err         error             // non-nil if a stream error occurred
}

// emitChunk prevents a provider goroutine from being stranded when its request
// is canceled or its consumer stops reading.
func emitChunk(ctx context.Context, ch chan<- StreamChunk, chunk StreamChunk) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- chunk:
		return true
	}
}

// ChatModel is the interface all LLM providers must implement.
// Stream returns a channel of StreamChunk and an optional immediate error.
// The channel is closed when streaming completes or fails.
type ChatModel interface {
	Stream(context.Context, Request) (<-chan StreamChunk, error)
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
	CheckModelAccess(ctx context.Context, chatSessionUuid, model string, userID int32) error
	Config() Config
}

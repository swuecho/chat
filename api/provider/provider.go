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
	"time"

	"golang.org/x/time/rate"

	"github.com/swuecho/chat_backend/domain"
)

// DefaultOutputTokens is sent to APIs that require an output limit when the
// caller sets none. Every current Claude model accepts 64K output tokens.
const DefaultOutputTokens int32 = 64000

// OutputTokens returns the caller's limit, or DefaultOutputTokens when unset.
func (s Session) OutputTokens() int32 {
	if s.MaxTokens > 0 {
		return s.MaxTokens
	}
	return DefaultOutputTokens
}

// Session contains only the chat settings needed by an LLM provider. It is
// intentionally independent of the database representation of a chat session.
type Session struct {
	UUID   string
	UserID int32
	Model  string
	// MaxTokens caps output only for internal callers that want short replies
	// (titles, suggestions). Zero means "no app-level limit": providers that
	// require a value use DefaultOutputTokens, the rest omit it.
	MaxTokens   int32
	Temperature float64
	TopP        float64
	N           int32
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
	Messages   []domain.Message
	ChatUUID   string
	Regenerate bool
	Stream     bool
	NewID      func() string
}

// systemMessageOnlyMessage is returned when a request has no user input
// beyond the system prompt.
const systemMessageOnlyMessage = "Usage error, system message input, not user input"

// completionRequestTimeout bounds non-streaming completion requests.
const completionRequestTimeout = 5 * time.Minute

// StreamChunk represents a single chunk in a streaming LLM response.
type StreamChunk struct {
	ID          string            // answer ID
	Content     string            // delta text content
	Done        bool              // true for the terminal chunk
	FinalAnswer *domain.LLMAnswer // set on Done (nil on error)
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

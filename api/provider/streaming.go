// Package provider — Streaming response helpers shared by all providers.
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// --- Streaming infrastructure ---

// StreamingResponse represents a common streaming response structure.
type StreamingResponse struct {
	AnswerID string
	Content  string
	IsFinal  bool
}

// FlushResponse sends a streaming response to the client.
func FlushResponse(w http.ResponseWriter, flusher http.Flusher, response StreamingResponse) error {
	if response.Content == "" && !response.IsFinal {
		return nil
	}
	streamResponse := openai.ChatCompletionStreamResponse{
		ID: response.AnswerID,
		Choices: []openai.ChatCompletionStreamChoice{
			{Index: 0, Delta: openai.ChatCompletionStreamChoiceDelta{Content: response.Content}},
		},
	}
	data, err := json.Marshal(streamResponse)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()
	return nil
}

// SetupSSEStream configures the response writer for Server-Sent Events.
// Delegates to pkg/util.SetupSSE.
var SetupSSEStream = util.SetupSSE

// --- Text buffer for streaming ---

// TextBuffer accumulates streaming text chunks across multiple choices.
type TextBuffer struct {
	builders []strings.Builder
	prefix   string
	suffix   string
}

// NewTextBuffer creates a new TextBuffer for n parallel choices.
func NewTextBuffer(n int32, prefix, suffix string) *TextBuffer {
	return &TextBuffer{
		builders: make([]strings.Builder, n),
		prefix:   prefix,
		suffix:   suffix,
	}
}

// AppendByIndex adds text to the buffer at the given index.
func (tb *TextBuffer) AppendByIndex(index int, text string) {
	if index >= 0 && index < len(tb.builders) {
		tb.builders[index].WriteString(text)
	}
}

// String joins all buffers with the given separator.
func (tb *TextBuffer) String(separator string) string {
	var result strings.Builder
	n := len(tb.builders)
	for i, builder := range tb.builders {
		if n > 1 {
			result.WriteString(fmt.Sprintf("\n%d\n---\n", i+1))
		}
		result.WriteString(tb.prefix)
		result.WriteString(builder.String())
		result.WriteString(tb.suffix)
		if i < len(tb.builders)-1 {
			result.WriteString(separator)
		}
	}
	return result.String()
}

// --- Shared helpers ---

// SetStreamingHeaders sets common headers for upstream streaming requests.
func SetStreamingHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
}

// ShouldFlushContent determines when to flush content based on common rules.
func ShouldFlushContent(content string, lastFlushLength int, isSmallContent bool) bool {
	return strings.Contains(content, "\n") ||
		(isSmallContent && len(content) < 200) ||
		(len(content)-lastFlushLength) >= 500
}

// GetChatModel retrieves a chat model by name from the database.
func GetChatModel(ctx context.Context, q *sqlc_queries.Queries, modelName string) (*sqlc_queries.ChatModel, error) {
	chatModel, err := q.ChatModelByName(ctx, modelName)
	if err != nil {
		return nil, dto.ErrResourceNotFound("chat model: " + modelName)
	}
	return &chatModel, nil
}

// GetChatFiles retrieves chat files for a session.
func GetChatFiles(ctx context.Context, q *sqlc_queries.Queries, sessionUUID string) ([]sqlc_queries.ChatFile, error) {
	chatFiles, err := q.ListChatFilesWithContentBySessionUUID(ctx, sessionUUID)
	if err != nil {
		return nil, dto.ErrInternalUnexpected.WithMessage("Failed to get chat files").WithDebugInfo(err.Error())
	}
	return chatFiles, nil
}

// buildStreamResponse creates a simple streaming response struct.
func buildStreamResponse(answerID, content string) openai.ChatCompletionStreamResponse {
	return openai.ChatCompletionStreamResponse{
		ID: answerID,
		Choices: []openai.ChatCompletionStreamChoice{
			{Index: 0, Delta: openai.ChatCompletionStreamChoiceDelta{Content: content}},
		},
	}
}

// firstN returns the first n runes of s.
// FirstN returns the first n characters of s.
func FirstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

// --- Utility functions ---

// NewUUID generates a new UUID v7 string.
var NewUUID = util.NewUUID

// generateAnswerID creates an answer ID or reuses chatUuid in regenerate mode.
func generateAnswerID(chatUuid string, regenerate bool) string {
	if regenerate {
		return chatUuid
	}
	return NewUUID()
}

// GetTokenCount returns the number of tokens in the given content.
var GetTokenCount = util.TokenCount

// GetPerWordStreamLimit returns the per-word stream limit from env or default.
var GetPerWordStreamLimit = util.PerWordStreamLimit

// Package provider — Streaming response helpers shared by all providers.
package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/pkg/util"
)

// --- Streaming infrastructure ---

type AnswerEventType string

const (
	AnswerEventStarted        AnswerEventType = "started"
	AnswerEventDelta          AnswerEventType = "delta"
	AnswerEventReasoningDelta AnswerEventType = "reasoning_delta"
	AnswerEventSuggested      AnswerEventType = "suggested_questions"
	AnswerEventCompleted      AnswerEventType = "completed"
	AnswerEventFailed         AnswerEventType = "failed"
	AnswerEventCanceled       AnswerEventType = "canceled"
)

// AnswerEvent is the typed protocol shared by streaming adapters. Completed,
// failed, and canceled are terminal events.
type AnswerEvent struct {
	Type               AnswerEventType `json:"type"`
	AnswerID           string          `json:"answerId,omitempty"`
	Delta              string          `json:"delta,omitempty"`
	SuggestedQuestions []string        `json:"suggestedQuestions,omitempty"`
	Persisted          bool            `json:"persisted"`
	Code               string          `json:"code,omitempty"`
	Message            string          `json:"message,omitempty"`
}

// FlushAnswerEvent emits a typed SSE event. A completed event must only be sent
// after the corresponding database mutation has succeeded.
func FlushAnswerEvent(w http.ResponseWriter, event AnswerEvent) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func IsTerminalAnswerEvent(event AnswerEvent) bool {
	return event.Type == AnswerEventCompleted || event.Type == AnswerEventFailed || event.Type == AnswerEventCanceled
}

// AnswerEventWriter enforces the terminal-event protocol for one response.
// It is safe to share among workflow branches.
type AnswerEventWriter struct {
	w        http.ResponseWriter
	mu       sync.Mutex
	terminal bool
}

func NewAnswerEventWriter(w http.ResponseWriter) *AnswerEventWriter {
	return &AnswerEventWriter{w: w}
}

func (s *AnswerEventWriter) Emit(event AnswerEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.terminal {
		return fmt.Errorf("answer stream already terminated")
	}
	if event.Type == AnswerEventCompleted && !event.Persisted {
		return fmt.Errorf("completed answer event must be persisted")
	}
	if err := FlushAnswerEvent(s.w, event); err != nil {
		return err
	}
	if IsTerminalAnswerEvent(event) {
		s.terminal = true
	}
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

var newUUID = util.NewUUID

// generateAnswerID creates an answer ID or reuses chatUuid in regenerate mode.
func generateAnswerID(chatUuid string, regenerate bool, generator ...func() string) string {
	if regenerate {
		return chatUuid
	}
	if len(generator) > 0 && generator[0] != nil {
		return generator[0]()
	}
	return newUUID()
}

// GetTokenCount returns the number of tokens in the given content.
var GetTokenCount = util.TokenCount

// GetPerWordStreamLimit returns the per-word stream limit from env or default.
var GetPerWordStreamLimit = util.PerWordStreamLimit

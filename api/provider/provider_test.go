package provider

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/swuecho/chat_backend/dto"
	"golang.org/x/time/rate"
)

func TestConfig(t *testing.T) {
	cfg := Config{
		OpenAIKey:   "key",
		OpenAIProxy: "proxy",
		RateLimiter: rate.NewLimiter(1, 1),
	}
	if cfg.OpenAIKey != "key" {
		t.Error("expected key")
	}
}

func TestStreamingResponse(t *testing.T) {
	w := httptest.NewRecorder()
	flusher, err := SetupSSEStream(w)
	if err != nil {
		t.Fatal(err)
	}
	if flusher == nil {
		t.Error("expected flusher")
	}
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Error("expected text/event-stream content type")
	}
}

func TestFlushStreamEvent(t *testing.T) {
	w := httptest.NewRecorder()
	err := FlushStreamEvent(w, "completed", StreamEvent{
		Type:      "completed",
		AnswerID:  "answer-1",
		Persisted: true,
	})
	if err != nil {
		t.Fatalf("FlushStreamEvent() error = %v", err)
	}

	body := w.Body.String()
	if !strings.HasPrefix(body, "event: completed\ndata: ") {
		t.Fatalf("unexpected SSE frame: %q", body)
	}

	data := strings.TrimSuffix(strings.TrimPrefix(body, "event: completed\ndata: "), "\n\n")
	var event StreamEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		t.Fatalf("invalid event JSON: %v", err)
	}
	if event.AnswerID != "answer-1" || !event.Persisted {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestAnswerEventWriterAllowsExactlyOneValidTerminalEvent(t *testing.T) {
	w := httptest.NewRecorder()
	writer := NewAnswerEventWriter(w)
	if err := writer.Emit(AnswerEvent{Type: AnswerEventCompleted, AnswerID: "a", Persisted: true}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Emit(AnswerEvent{Type: AnswerEventFailed}); err == nil {
		t.Fatal("expected second terminal event to be rejected")
	}

	invalid := NewAnswerEventWriter(httptest.NewRecorder())
	if err := invalid.Emit(AnswerEvent{Type: AnswerEventCompleted}); err == nil {
		t.Fatal("expected non-persisted completion to be rejected")
	}
}

func TestNewUUID(t *testing.T) {
	id1 := newUUID()
	id2 := newUUID()
	if id1 == "" || id2 == "" {
		t.Error("expected non-empty UUIDs")
	}
	if id1 == id2 {
		t.Error("expected unique UUIDs")
	}
}

func TestGenerateAnswerIDUsesInjectedGenerator(t *testing.T) {
	got := generateAnswerID("", false, func() string { return "fixed-answer-id" })
	if got != "fixed-answer-id" {
		t.Fatalf("got %q", got)
	}
	if got := generateAnswerID("existing", true, func() string { return "unused" }); got != "existing" {
		t.Fatalf("regeneration changed answer ID: %q", got)
	}
}

func TestGetTokenCount(t *testing.T) {
	count, err := GetTokenCount("hello world")
	if err != nil {
		t.Fatal(err)
	}
	if count <= 0 {
		t.Errorf("expected positive token count, got %d", count)
	}
}

func TestFirstN(t *testing.T) {
	if got := FirstN("hello world", 5); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	if got := FirstN("hi", 10); got != "hi" {
		t.Errorf("expected 'hi', got %q", got)
	}
}

func TestGetPerWordStreamLimit(t *testing.T) {
	limit := GetPerWordStreamLimit()
	if limit <= 0 {
		t.Errorf("expected positive limit, got %d", limit)
	}
}

func TestTextBuffer(t *testing.T) {
	tb := NewTextBuffer(1, "", "")
	tb.AppendByIndex(0, "hello")
	if got := tb.String(""); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestGetModelBaseURL(t *testing.T) {
	url, err := GetModelBaseURL("https://api.openai.com/v1/chat/completions")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://api.openai.com/v1" {
		t.Errorf("expected 'https://api.openai.com/v1', got %q", url)
	}
}

func TestBuildStreamResponse(t *testing.T) {
	resp := buildStreamResponse("id-123", "hello")
	if resp.ID != "id-123" {
		t.Errorf("expected 'id-123', got %q", resp.ID)
	}
	if len(resp.Choices) != 1 {
		t.Fatal("expected 1 choice")
	}
	if resp.Choices[0].Delta.Content != "hello" {
		t.Errorf("expected 'hello', got %q", resp.Choices[0].Delta.Content)
	}
}

// Ensure dto is referenced (used by provider code).
var _ = dto.ErrInternalUnexpected
var _ = context.Background

package provider

import (
	"context"
	"net/http/httptest"
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

func TestNewUUID(t *testing.T) {
	id1 := NewUUID()
	id2 := NewUUID()
	if id1 == "" || id2 == "" {
		t.Error("expected non-empty UUIDs")
	}
	if id1 == id2 {
		t.Error("expected unique UUIDs")
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

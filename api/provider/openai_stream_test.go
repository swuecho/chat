package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestDoChatStreamHandlesReasoningAndChoicelessChunks(t *testing.T) {
	lines := []string{
		`{"id":"chatcmpl-abc","choices":[{"index":0,"delta":{"role":"assistant","reasoning_content":"think "}}]}`,
		`{"id":"chatcmpl-abc","choices":[{"index":0,"delta":{"reasoning_content":"hard"}}]}`,
		`{"id":"chatcmpl-abc","choices":[{"index":0,"delta":{"content":"Hello"}}]}`,
		`{"id":"chatcmpl-abc","choices":[],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
		`{"id":"chatcmpl-abc","choices":[{"index":0,"delta":{"content":" world"}}]}`,
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		for _, line := range lines {
			fmt.Fprintf(w, "data: %s\n\n", line)
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	config := openai.DefaultConfig("test-key")
	config.BaseURL = server.URL
	client := openai.NewClientWithConfig(config)
	req := openai.ChatCompletionRequest{Model: "deepseek-reasoner", Stream: true}

	ch := make(chan StreamChunk, 16)
	go func() {
		defer close(ch)
		doChatStream(context.Background(), ch, client, req, 1, "session", false, func() string { return "id" }, server.URL, server.URL)
	}()

	var streamed strings.Builder
	var final StreamChunk
	for chunk := range ch {
		if chunk.Err != nil {
			t.Fatalf("unexpected stream error: %v", chunk.Err)
		}
		if chunk.Done {
			final = chunk
			continue
		}
		streamed.WriteString(chunk.Content)
	}

	if got, want := streamed.String(), "<think>think hard</think>Hello world"; got != want {
		t.Fatalf("streamed content = %q, want %q", got, want)
	}
	if final.FinalAnswer == nil {
		t.Fatal("expected a final answer")
	}
	if final.FinalAnswer.Answer != "Hello world" {
		t.Fatalf("final answer = %q, want %q", final.FinalAnswer.Answer, "Hello world")
	}
	if !strings.Contains(final.FinalAnswer.ReasoningContent, "think hard") {
		t.Fatalf("reasoning content = %q, want it to contain %q", final.FinalAnswer.ReasoningContent, "think hard")
	}
}

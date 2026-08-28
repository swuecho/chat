package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCompletionURL(t *testing.T) {
	tests := map[string]string{
		"https://api.openai.com/v1":                  "https://api.openai.com/v1/chat/completions",
		"https://api.openai.com/v1/":                 "https://api.openai.com/v1/chat/completions",
		"https://api.openai.com/v1/chat/completions": "https://api.openai.com/v1/chat/completions",
		"https://example.com/custom-endpoint":        "https://example.com/custom-endpoint",
	}
	for input, want := range tests {
		if got := completionURL(input); got != want {
			t.Errorf("completionURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestOpenAIErrorShape(t *testing.T) {
	w := httptest.NewRecorder()
	openAIError(w, http.StatusUnauthorized, "Invalid API key", "invalid_request_error", "invalid_api_key")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "invalid_api_key" || body.Error.Type != "invalid_request_error" || body.Error.Message == "" {
		t.Fatalf("unexpected error: %+v", body.Error)
	}
}

func TestGatewayHTTPClientProxy(t *testing.T) {
	h := NewGatewayHandler(nil, "http://127.0.0.1:7890")
	client, err := h.httpClient(time.Second)
	if err != nil {
		t.Fatal(err)
	}
	transport := client.Transport.(*http.Transport)
	request := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
	proxy, err := transport.Proxy(request)
	if err != nil {
		t.Fatal(err)
	}
	if proxy == nil || proxy.String() != "http://127.0.0.1:7890" {
		t.Fatalf("proxy = %v", proxy)
	}
}

func TestCopyEndToEndHeaders(t *testing.T) {
	source := http.Header{
		"Accept":         {"application/json"},
		"OpenAI-Project": {"project-1"},
		"Connection":     {"keep-alive, X-Internal"},
		"X-Internal":     {"do-not-forward"},
	}
	destination := make(http.Header)
	copyEndToEndHeaders(destination, source)
	if destination.Get("Accept") != "application/json" || destination.Get("OpenAI-Project") != "project-1" {
		t.Fatalf("end-to-end headers were not copied: %v", destination)
	}
	if destination.Get("Connection") != "" || destination.Get("X-Internal") != "" {
		t.Fatalf("hop-by-hop headers were copied: %v", destination)
	}
}

func TestBodyObservationIsBoundedButHashesFullBody(t *testing.T) {
	body := []byte("0123456789")
	observation := observeBytes(body, 4)
	if string(observation.sample) != "0123" || !observation.truncated() || observation.byteCount != 10 {
		t.Fatalf("unexpected observation: sample=%q truncated=%v bytes=%d", observation.sample, observation.truncated(), observation.byteCount)
	}
	want := sha256.Sum256(body)
	if observation.digest() != hex.EncodeToString(want[:]) {
		t.Fatalf("hash did not cover the complete body")
	}
}

func TestClassifyRequestDoesNotRetainContent(t *testing.T) {
	classification := classifyRequest([]byte(`{"model":"test","messages":[{"role":"user","content":"private text"}],"tools":[]}`))
	if bytes.Contains(classification, []byte("private text")) {
		t.Fatal("classification retained message content")
	}
	if !bytes.Contains(classification, []byte(`"message_count":1`)) || !bytes.Contains(classification, []byte(`"has_tools":true`)) {
		t.Fatalf("unexpected classification: %s", classification)
	}
}

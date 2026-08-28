package handler

import (
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

// Package handler provides HTTP request handlers for the chat API.
//
// Handlers are organized by domain:
//   - Auth: authentication, user management
//   - Chat: streaming, completion, session/message/prompt CRUD
//   - Workspace: workspace and session organization
//   - Model: LLM model management and privileges
//   - Snapshot: conversation snapshots and bots
//   - File: file upload and download
//   - Admin: user stats, rate limits
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/pkoukk/tiktoken-go"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/middleware"
)

// NewUUID generates a new UUID v7 string.
func NewUUID() string {
	uuidv7, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString()
	}
	return uuidv7.String()
}

// getTokenCount returns the estimated token count for a text string.
func getTokenCount(content string) (int, error) {
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, err
	}
	return len(tke.Encode(content, nil, nil)), nil
}

// firstNWords returns the first n words of a string.
func firstNWords(s string, n int) string {
	if s == "" {
		return ""
	}
	words := splitWords(s)
	if len(words) <= n {
		return s
	}
	result := ""
	for i, w := range words[:n] {
		if i > 0 {
			result += " "
		}
		result += w
	}
	return result
}

// splitWords splits a string by whitespace.
func splitWords(s string) []string {
	var words []string
	word := ""
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// setSSEHeader sets headers for Server-Sent Events.
func setSSEHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Del("Content-Length")
	w.Header().Del("Content-Encoding")
}

// setupSSEStream prepares an SSE stream and returns a Flusher.
func setupSSEStream(w http.ResponseWriter) (http.Flusher, error) {
	setSSEHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New(dto.ErrorStreamUnsupported)
	}
	return flusher, nil
}

// getPerWordStreamLimit returns the per-word streaming limit from env.
func getPerWordStreamLimit() int {
	val := os.Getenv("PER_WORD_STREAM_LIMIT")
	if val == "" {
		return 200
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 200
	}
	return n
}

// getPaginationParams extracts limit and offset from query parameters.
func getPaginationParams(r *http.Request) (limit int32, offset int32) {
	limit = 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if l, err := strconv.ParseInt(v, 10, 32); err == nil {
			limit = int32(l)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if o, err := strconv.ParseInt(v, 10, 32); err == nil {
			offset = int32(o)
		}
	}
	return
}

// getLimitParam extracts an optional limit parameter with a default.
func getLimitParam(r *http.Request, defaultLimit int32) int32 {
	v := r.URL.Query().Get("limit")
	if v == "" {
		return defaultLimit
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return defaultLimit
	}
	return int32(n)
}

// DecodeJSON decodes a JSON request body into the target.
func DecodeJSON(r *http.Request, target any) error {
	return json.NewDecoder(r.Body).Decode(target)
}

// getUserID extracts the authenticated user ID from the request context.
func getUserID(ctx context.Context) (int32, error) {
	userIdValue := ctx.Value(middleware.UserContextKey)
	if userIdValue == nil {
		return 0, fmt.Errorf("no user Id in context")
	}
	userIDStr, _ := userIdValue.(string)
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %s", userIDStr)
	}
	return int32(userIDInt), nil
}

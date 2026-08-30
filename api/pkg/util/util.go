// Package util provides shared utility functions used across the application.
package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

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

// TokenCount returns the estimated token count for a text string.
func TokenCount(content string) (int, error) {
	tke, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return 0, err
	}
	return len(tke.Encode(content, nil, nil)), nil
}

// FirstNWords returns the first n words of a string.
func FirstNWords(s string, n int) string {
	if s == "" {
		return ""
	}
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ")
}

// SetupSSE configures the response writer for Server-Sent Events
// and returns a Flusher for streaming.
func SetupSSE(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Del("Content-Length")
	w.Header().Del("Content-Encoding")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New(dto.ErrorStreamUnsupported)
	}
	return flusher, nil
}

// PerWordStreamLimit returns the per-word streaming limit from the environment.
func PerWordStreamLimit() int {
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

// PaginationParams extracts limit and offset from HTTP query parameters.
func PaginationParams(r *http.Request) (limit int32, offset int32) {
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

// LimitParam extracts an optional limit parameter with a default value.
func LimitParam(r *http.Request, defaultLimit int32) int32 {
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

// Validator is implemented by request types that can validate themselves after
// successful JSON decoding.
type Validator interface {
	Validate() error
}

// DecodeJSON strictly decodes exactly one JSON object into target. Unknown
// fields and trailing JSON values are rejected. If target implements Validator,
// validation runs only after decoding succeeds.
func DecodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body must not be empty")
		}
		return fmt.Errorf("decode request body: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request body must contain exactly one JSON value")
		}
		return fmt.Errorf("request body must contain exactly one JSON value: %w", err)
	}

	if validator, ok := target.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validate request body: %w", err)
		}
	}

	return nil
}

// UserID extracts the authenticated user ID from the request context.
func UserID(ctx context.Context) (int32, error) {
	val := ctx.Value(middleware.UserContextKey)
	if val == nil {
		return 0, fmt.Errorf("no user ID in context")
	}
	userIDStr, ok := val.(string)
	if !ok {
		return 0, fmt.Errorf("user ID in context is not a string")
	}
	userIDInt, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID: %s", userIDStr)
	}
	return int32(userIDInt), nil
}

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
	"github.com/swuecho/chat_backend/requestctx"
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

// PaginationParams extracts and validates bounded limit and offset parameters.
func PaginationParams(r *http.Request) (limit int32, offset int32, err error) {
	limit = 100
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, parseErr := strconv.ParseInt(v, 10, 32)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("limit must be an integer")
		}
		limit = int32(parsed)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		parsed, parseErr := strconv.ParseInt(v, 10, 32)
		if parseErr != nil {
			return 0, 0, fmt.Errorf("offset must be an integer")
		}
		offset = int32(parsed)
	}
	if limit < 1 || limit > 500 {
		return 0, 0, fmt.Errorf("limit must be between 1 and 500")
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("offset must not be negative")
	}
	return limit, offset, nil
}

// LimitParam extracts and validates an optional bounded limit parameter.
func LimitParam(r *http.Request, defaultLimit int32) (int32, error) {
	v := r.URL.Query().Get("limit")
	if v == "" {
		if defaultLimit < 1 || defaultLimit > 500 {
			return 0, fmt.Errorf("default limit must be between 1 and 500")
		}
		return defaultLimit, nil
	}
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("limit must be an integer")
	}
	if n < 1 || n > 500 {
		return 0, fmt.Errorf("limit must be between 1 and 500")
	}
	return int32(n), nil
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
	if userID, err := requestctx.UserID(ctx); err == nil {
		return userID, nil
	}
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

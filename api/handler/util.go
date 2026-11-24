// Package handler provides HTTP request handlers for the chat API.
package handler

import (
	"context"
	"net/http"

	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/pkg/util"
)

// Re-exported from pkg/util for convenient use within the handler package.

var (
	NewUUID            = util.NewUUID
	SetupSSE           = util.SetupSSE
	PerWordStreamLimit = util.PerWordStreamLimit
	PaginationParams   = util.PaginationParams
	LimitParam         = util.LimitParam
	DecodeJSON         = util.DecodeJSON
)

func endpoint(handler httpx.HandlerFunc) http.HandlerFunc { return httpx.Adapt(handler) }

// streamEndpoint is deliberately separate from endpoint: after SSE headers or
// bytes are committed, failures must be sent as stream events, not JSON errors.
func streamEndpoint(handler httpx.StreamHandlerFunc) http.HandlerFunc {
	return httpx.AdaptStream(handler)
}
func respondJSON(w http.ResponseWriter, status int, value any) error {
	return httpx.JSON(w, status, value)
}
func noContent(w http.ResponseWriter) error { return httpx.NoContent(w) }
func respondStatus(w http.ResponseWriter, status int) error {
	return httpx.Status(w, status)
}
func authenticatedUserID(r *http.Request) (int32, error) {
	principal, err := httpx.Principal(r)
	if err == nil {
		return principal.UserID, nil
	}
	// Compatibility for package tests and callers that still construct the
	// legacy context directly. Production auth middleware always sets Principal.
	userID, legacyErr := getUserID(r.Context())
	if legacyErr != nil {
		return 0, err
	}
	return userID, nil
}
func positiveInt32Param(r *http.Request, name string) (int32, error) {
	return httpx.Int32Param(r, name)
}
func positiveInt64Param(r *http.Request, name string) (int64, error) {
	return httpx.Int64Param(r, name)
}
func pageResponse[T any](items []T, total int64, page httpx.Page) httpx.PageResponse[T] {
	return httpx.PageResponse[T]{Items: items, Total: total, Limit: page.Limit, Offset: page.Offset}
}

func getTokenCount(content string) (int, error)                  { return util.TokenCount(content) }
func getUserID(ctx context.Context) (int32, error)               { return util.UserID(ctx) }
func setupSSEStream(w http.ResponseWriter) (http.Flusher, error) { return util.SetupSSE(w) }

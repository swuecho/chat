// Package handler provides HTTP request handlers for the chat API.
package handler

import (
	"context"
	"net/http"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/validation"
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
func firstNWords(s string, n int) string                         { return util.FirstNWords(s, n) }
func getUserID(ctx context.Context) (int32, error)               { return util.UserID(ctx) }
func setSSEHeader(w http.ResponseWriter)                         { _, _ = util.SetupSSE(w) }
func setupSSEStream(w http.ResponseWriter) (http.Flusher, error) { return util.SetupSSE(w) }
func getPerWordStreamLimit() int                                 { return util.PerWordStreamLimit() }
func getPaginationParams(r *http.Request) (int32, int32, error)  { return util.PaginationParams(r) }
func getLimitParam(r *http.Request, d int32) (int32, error)      { return util.LimitParam(r, d) }

func validateUUIDParam(w http.ResponseWriter, field, value string) bool {
	if err := validation.UUID(field, value, true); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput(err.Error()))
		return false
	}
	return true
}

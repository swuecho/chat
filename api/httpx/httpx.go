// Package httpx provides the HTTP transport boundary shared by handlers.
package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/requestctx"
	"github.com/swuecho/chat_backend/validation"
)

// HandlerFunc returns failures to one outer HTTP error boundary. It is for
// ordinary responses; handlers that commit an SSE stream use the stream API.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func Adapt(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			Error(w, r, err)
		}
	}
}

// JSON marshals before committing headers so encoding failures can still be
// handled by the outer error boundary.
func JSON(w http.ResponseWriter, status int, value any) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode JSON response: %w", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("write JSON response: %w", err)
	}
	return nil
}

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func Status(w http.ResponseWriter, status int) error {
	w.WriteHeader(status)
	return nil
}

func Error(w http.ResponseWriter, r *http.Request, err error) {
	apiErr := dto.ToAPIError(err)
	if apiErr.DebugInfo != "" {
		slog.Error("api request failed", "request_id", requestctx.RequestID(r.Context()),
			"method", r.Method, "path", r.URL.Path, "code", apiErr.Code,
			"detail", apiErr.Detail, "error", apiErr.DebugInfo)
	}
	response := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
	}{Code: apiErr.Code, Message: apiErr.Message, Detail: apiErr.Detail}
	if writeErr := JSON(w, apiErr.HTTPCode, response); writeErr != nil {
		slog.Error("failed to write API error", "request_id", requestctx.RequestID(r.Context()), "error", writeErr)
	}
}

func DecodeJSON(r *http.Request, target any) error {
	if err := util.DecodeJSON(r, target); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error())
	}
	return nil
}

func Principal(r *http.Request) (requestctx.Principal, error) {
	principal, err := requestctx.PrincipalFrom(r.Context())
	if err != nil {
		return requestctx.Principal{}, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error())
	}
	return principal, nil
}

func UUIDParam(r *http.Request, name string) (string, error) {
	value := mux.Vars(r)[name]
	if err := validation.UUID(name, value, true); err != nil {
		return "", domain.Invalid(err.Error())
	}
	return value, nil
}

func Int32Param(r *http.Request, name string) (int32, error) {
	value := mux.Vars(r)[name]
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed <= 0 {
		return 0, domain.Invalid(name + " must be a positive integer")
	}
	return int32(parsed), nil
}

func Int64Param(r *http.Request, name string) (int64, error) {
	value := mux.Vars(r)[name]
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, domain.Invalid(name + " must be a positive integer")
	}
	return parsed, nil
}

type Page struct {
	Limit  int32
	Offset int32
}

func ParsePage(r *http.Request) (Page, error) {
	limit, offset, err := util.PaginationParams(r)
	if err != nil {
		return Page{}, domain.Invalid(err.Error())
	}
	return Page{Limit: limit, Offset: offset}, nil
}

func ParseLimit(r *http.Request, defaultLimit int32) (int32, error) {
	limit, err := util.LimitParam(r, defaultLimit)
	if err != nil {
		return 0, domain.Invalid(err.Error())
	}
	return limit, nil
}

// PageResponse is the standard offset-pagination response envelope.
type PageResponse[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func Invalid(detail string) error {
	return domain.Invalid(detail)
}

func IsResponseWriteError(err error) bool {
	return err != nil && (errors.Is(err, http.ErrHandlerTimeout) || errors.Is(err, http.ErrAbortHandler))
}

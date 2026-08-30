// Package dto provides API error types and catalog.
package dto

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/swuecho/chat_backend/domain"
)

// APIError represents a standardized error response for the API.
type APIError struct {
	HTTPCode  int    `json:"-"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Detail    string `json:"detail,omitempty"`
	DebugInfo string `json:"-"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("[%s] %s %s", e.Code, e.Message, e.Detail)
}

// WithDetail adds detail to an APIError.
func (e APIError) WithDetail(detail string) APIError {
	e.Detail = detail
	return e
}

// WithDebugInfo adds debug info to an APIError.
func (e APIError) WithDebugInfo(debugInfo string) APIError {
	e.DebugInfo = debugInfo
	return e
}

// WithMessage sets the message of an APIError.
func (e APIError) WithMessage(message string) APIError {
	e.Message = message
	return e
}

// Error code prefixes by domain
const (
	ErrAuth       = "AUTH"
	ErrValidation = "VALD"
	ErrResource   = "RES"
	ErrDatabase   = "DB"
	ErrExternal   = "EXT"
	ErrInternal   = "INTN"
	ErrModel      = "MODEL"
)

// Pre-defined API errors
var (
	// Auth errors
	ErrAuthInvalidCredentials = APIError{
		HTTPCode: http.StatusUnauthorized,
		Code:     ErrAuth + "_001",
		Message:  "Invalid credentials",
	}
	ErrAuthExpiredToken = APIError{
		HTTPCode: http.StatusUnauthorized,
		Code:     ErrAuth + "_002",
		Message:  "Token has expired",
	}
	ErrAuthAdminRequired = APIError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrAuth + "_003",
		Message:  "Admin privileges required",
	}
	ErrAuthInvalidEmailOrPassword = APIError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrAuth + "_004",
		Message:  "invalid email or password",
	}
	ErrAuthAccessDenied = APIError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrAuth + "_005",
		Message:  "Access denied",
	}

	// Resource errors
	ErrResourceNotFoundGeneric = APIError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrResource + "_001",
		Message:  "Resource not found",
	}
	ErrResourceAlreadyExistsGeneric = APIError{
		HTTPCode: http.StatusConflict,
		Code:     ErrResource + "_002",
		Message:  "Resource already exists",
	}
	ErrTooManyRequests = APIError{
		HTTPCode: http.StatusTooManyRequests,
		Code:     ErrResource + "_003",
		Message:  "Rate limit exceeded",
		Detail:   "Too many requests in the given time period",
	}
	ErrChatSessionNotFound = APIError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrResource + "_004",
		Message:  "Chat session not found",
	}
	ErrChatFileNotFound = APIError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrResource + "_005",
		Message:  "Chat file not found",
	}
	ErrChatModelNotFound = APIError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrResource + "_006",
		Message:  "Chat model not found",
	}
	ErrChatMessageNotFound = APIError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrResource + "_007",
		Message:  "Chat message not found",
	}

	// Validation errors
	ErrValidationInvalidInputGeneric = APIError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrValidation + "_001",
		Message:  "Invalid input",
	}
	ErrChatFileTooLarge = APIError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrValidation + "_002",
		Message:  "File too large",
	}
	ErrChatFileInvalidType = APIError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrValidation + "_003",
		Message:  "Invalid file type",
	}
	ErrChatSessionInvalid = APIError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrValidation + "_004",
		Message:  "Invalid chat session",
	}

	// Database errors
	ErrDatabaseQuery = APIError{
		HTTPCode:  http.StatusInternalServerError,
		Code:      ErrDatabase + "_001",
		Message:   "Database query failed",
		DebugInfo: "Database operation failed - check logs for details",
	}
	ErrDatabaseConnection = APIError{
		HTTPCode:  http.StatusServiceUnavailable,
		Code:      ErrDatabase + "_002",
		Message:   "Database connection failed",
		DebugInfo: "Could not connect to database - check connection settings",
	}
	ErrDatabaseForeignKey = APIError{
		HTTPCode:  http.StatusBadRequest,
		Code:      ErrDatabase + "_003",
		Message:   "Referenced resource does not exist",
		DebugInfo: "Foreign key violation",
	}

	// External service errors
	ErrExternalTimeout = APIError{
		HTTPCode: http.StatusGatewayTimeout,
		Code:     ErrExternal + "_001",
		Message:  "External service timed out",
	}
	ErrExternalUnavailable = APIError{
		HTTPCode: http.StatusServiceUnavailable,
		Code:     ErrExternal + "_002",
		Message:  "External service unavailable",
	}

	// Internal errors
	ErrInternalUnexpected = APIError{
		HTTPCode:  http.StatusInternalServerError,
		Code:      ErrInternal + "_001",
		Message:   "An unexpected error occurred",
		DebugInfo: "Unexpected internal error - check logs for stack trace",
	}
	ErrChatStreamFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrInternal + "_004",
		Message:  "Failed to stream chat response",
	}
	ErrChatRequestFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrInternal + "_005",
		Message:  "Failed to make chat request",
	}

	// Model errors
	ErrSystemMessageError = APIError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrModel + "_001",
		Message:  "Usage error, system message input, not user input",
	}
	ErrClaudeStreamFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_002",
		Message:  "Failed to stream Claude response",
	}
	ErrClaudeRequestFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_003",
		Message:  "Failed to make Claude request",
	}
	ErrClaudeInvalidResponse = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_004",
		Message:  "Invalid response from Claude API",
	}
	ErrClaudeResponseFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_005",
		Message:  "Failed to stream Claude response",
	}

	// Deprecated: use ErrClaudeResponseFailed instead.
	ErrClaudeResponseFaild = ErrClaudeResponseFailed
	ErrOpenAIStreamFailed  = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_006",
		Message:  "Failed to stream OpenAI response",
	}
	ErrOpenAIRequestFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_007",
		Message:  "Failed to make OpenAI request",
	}
	ErrOpenAIInvalidResponse = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_008",
		Message:  "Invalid response from OpenAI API",
	}
	ErrOpenAIConfigFailed = APIError{
		HTTPCode: http.StatusInternalServerError,
		Code:     ErrModel + "_009",
		Message:  "Failed to configure OpenAI client",
	}
)

// --- Helper constructors ---

func ErrResourceNotFound(resource string) APIError {
	err := ErrResourceNotFoundGeneric
	err.Message = resource + " not found"
	return err
}

func ErrResourceAlreadyExists(resource string) APIError {
	err := ErrResourceAlreadyExistsGeneric
	err.Message = resource + " already exists"
	return err
}

func ErrValidationInvalidInput(detail string) APIError {
	err := ErrValidationInvalidInputGeneric
	err.Detail = detail
	return err
}

// CreateAPIError creates a consistent API error with optional detail and debug info.
func CreateAPIError(baseErr APIError, detail, debugInfo string) APIError {
	apiErr := baseErr
	if detail != "" {
		apiErr.Detail = detail
	}
	if debugInfo != "" {
		apiErr.DebugInfo = debugInfo
	}
	return apiErr
}

// --- HTTP response helpers ---

// RespondWithAPIError is the HTTP error boundary. It maps transport-neutral
// application errors to the stable public API error contract.
func RespondWithAPIError(w http.ResponseWriter, err error) {
	apiErr := ToAPIError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.HTTPCode)

	response := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
	}{
		Code:    apiErr.Code,
		Message: apiErr.Message,
		Detail:  apiErr.Detail,
	}

	if apiErr.DebugInfo != "" {
		slog.Error("api error", "code", apiErr.Code, "message", apiErr.Message, "detail", apiErr.Detail, "debug", apiErr.DebugInfo)
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Info("Failed to write error response", "error", err)
	}
}

// ToAPIError maps errors from the application and infrastructure layers to HTTP.
func ToAPIError(err error) APIError {
	if err == nil {
		return ErrInternalUnexpected
	}

	if failure, ok := domain.AsProviderFailure(err); ok {
		var apiErr APIError
		switch failure.Kind {
		case domain.ProviderFailureInvalidRequest:
			apiErr = ErrValidationInvalidInput("Provider rejected the request")
		case domain.ProviderFailureAuthentication:
			apiErr = ErrAuthInvalidCredentials.WithMessage("Provider authentication failed")
		case domain.ProviderFailurePermission:
			apiErr = ErrAuthAccessDenied.WithMessage("Provider denied access")
		case domain.ProviderFailureRateLimited, domain.ProviderFailureUnavailable:
			apiErr = ErrExternalUnavailable.WithMessage("Provider is temporarily unavailable")
		case domain.ProviderFailureTimeout:
			apiErr = WrapError(context.DeadlineExceeded, "provider request")
		case domain.ProviderFailureCanceled:
			apiErr = WrapError(context.Canceled, "provider request")
		default:
			apiErr = ErrInternalUnexpected.WithMessage("Provider request failed")
		}
		apiErr.DebugInfo = failure.Error()
		return apiErr
	}

	if appErr, ok := domain.AsError(err); ok {
		var apiErr APIError
		switch appErr.Kind {
		case domain.KindInvalid:
			apiErr = ErrValidationInvalidInput(appErr.Message)
		case domain.KindUnauthorized:
			apiErr = ErrAuthInvalidCredentials.WithMessage(appErr.Message)
		case domain.KindForbidden:
			apiErr = ErrAuthAccessDenied.WithMessage(appErr.Message)
		case domain.KindNotFound:
			switch appErr.Resource {
			case "Chat file":
				apiErr = ErrChatFileNotFound
			case "Chat model":
				apiErr = ErrChatModelNotFound
			case "Chat message":
				apiErr = ErrChatMessageNotFound
			case "Chat session":
				apiErr = ErrChatSessionNotFound
			default:
				apiErr = ErrResourceNotFound(appErr.Resource)
			}
			apiErr.Message = appErr.Message
		case domain.KindConflict:
			apiErr = ErrResourceAlreadyExists(appErr.Resource)
			apiErr.Message = appErr.Message
		case domain.KindUnavailable:
			apiErr = ErrExternalUnavailable.WithMessage(appErr.Message)
		default:
			apiErr = ErrInternalUnexpected.WithMessage(appErr.Message)
		}
		if appErr.Err != nil {
			apiErr.DebugInfo = appErr.Err.Error()
		}
		return apiErr
	}

	var apiErr APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return WrapError(err, "")
}

// RespondWithJSON writes a JSON response with the given status code.
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

// --- Database error mapping ---

func MapDatabaseError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrResourceNotFound("Record")
	}

	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "connection reset by peer") {
		dbErr := ErrDatabaseConnection
		dbErr.DebugInfo = err.Error()
		return dbErr
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrResourceAlreadyExists("Record")
		case "23503":
			dbErr := ErrDatabaseForeignKey
			dbErr.DebugInfo = fmt.Sprintf("Foreign key violation: %s", pgErr.Detail)
			return dbErr
		case "42P01":
			dbErr := ErrDatabaseQuery
			dbErr.Message = "Database schema error"
			dbErr.DebugInfo = fmt.Sprintf("Table does not exist: %s", pgErr.Detail)
			return dbErr
		case "42703":
			dbErr := ErrDatabaseQuery
			dbErr.Message = "Database schema error"
			dbErr.DebugInfo = fmt.Sprintf("Column does not exist: %s", pgErr.Detail)
			return dbErr
		case "53300":
			dbErr := ErrDatabaseConnection
			dbErr.Message = "Database connection limit reached"
			dbErr.DebugInfo = pgErr.Detail
			return dbErr
		}
	}

	slog.Info("Unhandled database error", "error", err)

	dbErr := ErrDatabaseQuery
	dbErr.DebugInfo = err.Error()
	return dbErr
}

// --- Error wrapping ---

// WrapError converts a standard error into an APIError.
func WrapError(err error, detail string) APIError {
	var apiErr APIError

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return APIError{
			HTTPCode:  http.StatusGatewayTimeout,
			Code:      ErrInternal + "_002",
			Message:   "Request timed out",
			Detail:    detail,
			DebugInfo: "Context deadline exceeded",
		}
	case errors.Is(err, context.Canceled):
		return APIError{
			HTTPCode:  http.StatusRequestTimeout,
			Code:      ErrInternal + "_003",
			Message:   "Request was canceled",
			Detail:    detail,
			DebugInfo: "Context was canceled",
		}
	}

	switch e := err.(type) {
	case APIError:
		apiErr = e
		if detail != "" {
			if apiErr.Detail != "" {
				apiErr.Detail = fmt.Sprintf("%s: %s", detail, apiErr.Detail)
			} else {
				apiErr.Detail = detail
			}
		}
	case *APIError:
		apiErr = *e
		if detail != "" {
			if apiErr.Detail != "" {
				apiErr.Detail = fmt.Sprintf("%s: %s", detail, apiErr.Detail)
			} else {
				apiErr.Detail = detail
			}
		}
	default:
		apiErr = ErrInternalUnexpected
		apiErr.Detail = detail
		apiErr.DebugInfo = err.Error()
	}

	return apiErr
}

// IsErrorCode checks if an error is an APIError with the specified code.
func IsErrorCode(err error, code string) bool {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.Code == code
	}
	return false
}

// --- Error catalog (for documentation) ---

var ErrorCatalog = map[string]APIError{
	ErrAuthInvalidCredentials.Code:        ErrAuthInvalidCredentials,
	ErrAuthExpiredToken.Code:              ErrAuthExpiredToken,
	ErrAuthAdminRequired.Code:             ErrAuthAdminRequired,
	ErrAuthInvalidEmailOrPassword.Code:    ErrAuthInvalidEmailOrPassword,
	ErrAuthAccessDenied.Code:              ErrAuthAccessDenied,
	ErrResourceNotFoundGeneric.Code:       ErrResourceNotFoundGeneric,
	ErrResourceAlreadyExistsGeneric.Code:  ErrResourceAlreadyExistsGeneric,
	ErrTooManyRequests.Code:               ErrTooManyRequests,
	ErrChatSessionNotFound.Code:           ErrChatSessionNotFound,
	ErrChatFileNotFound.Code:              ErrChatFileNotFound,
	ErrChatModelNotFound.Code:             ErrChatModelNotFound,
	ErrChatMessageNotFound.Code:           ErrChatMessageNotFound,
	ErrValidationInvalidInputGeneric.Code: ErrValidationInvalidInputGeneric,
	ErrChatFileTooLarge.Code:              ErrChatFileTooLarge,
	ErrChatFileInvalidType.Code:           ErrChatFileInvalidType,
	ErrChatSessionInvalid.Code:            ErrChatSessionInvalid,
	ErrDatabaseQuery.Code:                 ErrDatabaseQuery,
	ErrDatabaseConnection.Code:            ErrDatabaseConnection,
	ErrDatabaseForeignKey.Code:            ErrDatabaseForeignKey,
	ErrExternalTimeout.Code:               ErrExternalTimeout,
	ErrExternalUnavailable.Code:           ErrExternalUnavailable,
	ErrInternalUnexpected.Code:            ErrInternalUnexpected,
	ErrChatStreamFailed.Code:              ErrChatStreamFailed,
	ErrChatRequestFailed.Code:             ErrChatRequestFailed,
	ErrSystemMessageError.Code:            ErrSystemMessageError,
	ErrClaudeStreamFailed.Code:            ErrClaudeStreamFailed,
	ErrClaudeRequestFailed.Code:           ErrClaudeRequestFailed,
	ErrClaudeInvalidResponse.Code:         ErrClaudeInvalidResponse,
	ErrClaudeResponseFailed.Code:          ErrClaudeResponseFailed,
	ErrOpenAIStreamFailed.Code:            ErrOpenAIStreamFailed,
	ErrOpenAIRequestFailed.Code:           ErrOpenAIRequestFailed,
	ErrOpenAIInvalidResponse.Code:         ErrOpenAIInvalidResponse,
	ErrOpenAIConfigFailed.Code:            ErrOpenAIConfigFailed,
}

// ErrorCatalogHandler serves the error catalog as JSON.
func ErrorCatalogHandler(w http.ResponseWriter, r *http.Request) {
	type ErrorDoc struct {
		Code     string `json:"code"`
		HTTPCode int    `json:"http_code"`
		Message  string `json:"message"`
	}

	docs := make([]ErrorDoc, 0, len(ErrorCatalog))
	for code, info := range ErrorCatalog {
		docs = append(docs, ErrorDoc{
			Code:     code,
			HTTPCode: info.HTTPCode,
			Message:  info.Message,
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Code < docs[j].Code
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(docs)
}

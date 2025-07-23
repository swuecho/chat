package errors

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgconn"
)

// BusinessError represents a business logic error that can be safely exposed to clients
type BusinessError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Detail     string `json:"detail,omitempty"`
	HTTPStatus int    `json:"-"`
}

func (e BusinessError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
}

// ValidationError represents input validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// MultiValidationError represents multiple validation errors
type MultiValidationError struct {
	Errors []ValidationError `json:"errors"`
}

func (e MultiValidationError) Error() string {
	return fmt.Sprintf("validation failed: %d errors", len(e.Errors))
}

// Error constants for business errors
var (
	// Authentication errors
	ErrUnauthorized = BusinessError{
		Code:       "AUTH_001",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}
	
	ErrForbidden = BusinessError{
		Code:       "AUTH_002", 
		Message:    "Access denied",
		HTTPStatus: http.StatusForbidden,
	}
	
	ErrInvalidCredentials = BusinessError{
		Code:       "AUTH_003",
		Message:    "Invalid username or password",
		HTTPStatus: http.StatusUnauthorized,
	}
	
	// Resource errors
	ErrResourceNotFound = BusinessError{
		Code:       "RES_001",
		Message:    "Resource not found",
		HTTPStatus: http.StatusNotFound,
	}
	
	ErrResourceConflict = BusinessError{
		Code:       "RES_002",
		Message:    "Resource already exists",
		HTTPStatus: http.StatusConflict,
	}
	
	// Validation errors
	ErrInvalidInput = BusinessError{
		Code:       "VAL_001",
		Message:    "Invalid input provided",
		HTTPStatus: http.StatusBadRequest,
	}
	
	ErrMissingField = BusinessError{
		Code:       "VAL_002",
		Message:    "Required field missing",
		HTTPStatus: http.StatusBadRequest,
	}
	
	// Rate limiting errors
	ErrRateLimitExceeded = BusinessError{
		Code:       "RATE_001",
		Message:    "Rate limit exceeded",
		HTTPStatus: http.StatusTooManyRequests,
	}
	
	// Model/LLM errors
	ErrModelNotFound = BusinessError{
		Code:       "MODEL_001",
		Message:    "Model not found or unavailable",
		HTTPStatus: http.StatusNotFound,
	}
	
	ErrModelRequestFailed = BusinessError{
		Code:       "MODEL_002",
		Message:    "Model request failed",
		HTTPStatus: http.StatusServiceUnavailable,
	}
	
	// Internal errors
	ErrInternalServer = BusinessError{
		Code:       "INT_001",
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// Builder functions for common error patterns
func NotFound(resource string) BusinessError {
	return ErrResourceNotFound.WithDetail(fmt.Sprintf("%s not found", resource))
}

func Unauthorized(detail string) BusinessError {
	return ErrUnauthorized.WithDetail(detail)
}

func ValidationFailed(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

func InvalidInput(detail string) BusinessError {
	return ErrInvalidInput.WithDetail(detail)
}

// Builder methods for BusinessError
func (e BusinessError) WithDetail(detail string) BusinessError {
	e.Detail = detail
	return e
}

func (e BusinessError) WithMessage(message string) BusinessError {
	e.Message = message
	return e
}

// Database error mapping
func FromDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	
	switch {
	case err == sql.ErrNoRows:
		return ErrResourceNotFound
	case isDuplicateKeyError(err):
		return ErrResourceConflict.WithDetail("Duplicate entry")
	case isForeignKeyError(err):
		return ErrInvalidInput.WithDetail("Invalid reference")
	default:
		// Don't expose database errors to clients
		return ErrInternalServer
	}
}

// Context error handling
func FromContextError(err error) error {
	if err == nil {
		return nil
	}
	
	switch err {
	case context.Canceled:
		return BusinessError{
			Code:       "CTX_001",
			Message:    "Request canceled",
			HTTPStatus: http.StatusRequestTimeout,
		}
	case context.DeadlineExceeded:
		return BusinessError{
			Code:       "CTX_002", 
			Message:    "Request timeout",
			HTTPStatus: http.StatusRequestTimeout,
		}
	default:
		return ErrInternalServer
	}
}

// HTTP Response helpers
type ErrorResponse struct {
	Error interface{} `json:"error"`
}

func WriteErrorResponse(w http.ResponseWriter, err error) {
	var response ErrorResponse
	var status int
	
	switch e := err.(type) {
	case BusinessError:
		status = e.HTTPStatus
		response.Error = e
	case MultiValidationError:
		status = http.StatusBadRequest
		response.Error = e
	case ValidationError:
		status = http.StatusBadRequest
		response.Error = MultiValidationError{Errors: []ValidationError{e}}
	default:
		// Unknown error - don't expose details
		status = http.StatusInternalServerError
		response.Error = ErrInternalServer
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Helper functions for database error detection
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return err != nil && errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return err != nil && errors.As(err, &pgErr) && pgErr.Code == "23503"
}
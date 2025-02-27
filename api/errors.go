package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/jackc/pgconn"
)

type APIError struct {
	HTTPCode  int    `json:"-"`                // HTTP status code (not exposed in response)
	Code      string `json:"code"`             // Application-specific error code
	Message   string `json:"message"`          // Human-readable message
	Detail    string `json:"detail,omitempty"` // Optional error details
	DebugInfo string `json:"-"`                // Internal debugging info (not exposed)
}

func (e APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Define error code prefixes by domain
const (
	ErrAuth       = "AUTH" // Authentication/Authorization errors
	ErrValidation = "VALD" // Validation errors
	ErrResource   = "RES"  // Resource-related errors
	ErrDatabase   = "DB"   // Database errors
	ErrExternal   = "EXT"  // External service errors
	ErrInternal   = "INTN" // Internal application errors
)

// Define external service errors
var (
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
)

// Define all API errors
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

	// Validation errors
	ErrValidationInvalidInputGeneric = APIError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrValidation + "_001",
		Message:  "Invalid input",
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

	// Internal errors
	ErrInternalUnexpected = APIError{
		HTTPCode:  http.StatusInternalServerError,
		Code:      ErrInternal + "_001",
		Message:   "An unexpected error occurred",
		DebugInfo: "Unexpected internal error - check logs for stack trace",
	}
)

// Helper functions to create specific errors with dynamic content
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

func RespondWithAPIError(w http.ResponseWriter, err APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPCode)

	response := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
	}{
		Code:    err.Code,
		Message: err.Message,
		Detail:  err.Detail,
	}

	// Log error with debug info if available
	if err.DebugInfo != "" {
		log.Printf("Error [%s]: %s - %s", err.Code, err.Message, err.DebugInfo)
	}

	json.NewEncoder(w).Encode(response)
}

// NewAPIError creates a new APIError with the given parameters
func NewAPIError(httpCode int, code, message string) APIError {
	return APIError{
		HTTPCode: httpCode,
		Code:     code,
		Message:  message,
	}
}

// WithDetail adds detail to an APIError
func (e APIError) WithDetail(detail string) APIError {
	e.Detail = detail
	return e
}

// WithDebugInfo adds debug info to an APIError
func (e APIError) WithDebugInfo(debugInfo string) APIError {
	e.DebugInfo = debugInfo
	return e
}

func MapDatabaseError(err error) error {
	// Map common database errors to appropriate application errors
	if errors.Is(err, sql.ErrNoRows) {
		return ErrResourceNotFound("Record")
	}

	// Check for connection errors
	if strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "no such host") ||
		strings.Contains(err.Error(), "connection reset by peer") {
		dbErr := ErrDatabaseConnection
		dbErr.DebugInfo = err.Error()
		return dbErr
	}

	// Check for other specific database errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			return ErrResourceAlreadyExists("Record")
		case "23503": // Foreign key violation
			dbErr := ErrDatabaseForeignKey
			dbErr.DebugInfo = fmt.Sprintf("Foreign key violation: %s", pgErr.Detail)
			return dbErr
		case "42P01": // Undefined table
			dbErr := ErrDatabaseQuery
			dbErr.Message = "Database schema error"
			dbErr.DebugInfo = fmt.Sprintf("Table does not exist: %s", pgErr.Detail)
			return dbErr
		case "42703": // Undefined column
			dbErr := ErrDatabaseQuery
			dbErr.Message = "Database schema error"
			dbErr.DebugInfo = fmt.Sprintf("Column does not exist: %s", pgErr.Detail)
			return dbErr
		case "53300": // Too many connections
			dbErr := ErrDatabaseConnection
			dbErr.Message = "Database connection limit reached"
			dbErr.DebugInfo = pgErr.Detail
			return dbErr
		}
	}

	// Log the unhandled database error
	log.Printf("Unhandled database error: %v", err)

	// Return generic database error
	dbErr := ErrDatabaseQuery
	dbErr.DebugInfo = err.Error()
	return dbErr
}

// ErrorCatalog holds all error codes for documentation purposes
var ErrorCatalog = map[string]APIError{
	// Auth errors
	ErrAuthInvalidCredentials.Code: ErrAuthInvalidCredentials,
	ErrAuthExpiredToken.Code:       ErrAuthExpiredToken,
	ErrAuthAdminRequired.Code:      ErrAuthAdminRequired,

	// Resource errors
	ErrResourceNotFoundGeneric.Code:      ErrResourceNotFoundGeneric,
	ErrResourceAlreadyExistsGeneric.Code: ErrResourceAlreadyExistsGeneric,
	ErrTooManyRequests.Code:              ErrTooManyRequests,

	// Validation errors
	ErrValidationInvalidInputGeneric.Code: ErrValidationInvalidInputGeneric,

	// Database errors
	ErrDatabaseQuery.Code:      ErrDatabaseQuery,
	ErrDatabaseConnection.Code: ErrDatabaseConnection,
	ErrDatabaseForeignKey.Code: ErrDatabaseForeignKey,

	// External service errors
	ErrExternalTimeout.Code:     ErrExternalTimeout,
	ErrExternalUnavailable.Code: ErrExternalUnavailable,

	// External service errors
	ErrExternalTimeout.Code:     ErrExternalTimeout,
	ErrExternalUnavailable.Code: ErrExternalUnavailable,

	// Internal errors
	ErrInternalUnexpected.Code: ErrInternalUnexpected,
	ErrInternal + "_002":       {HTTPCode: http.StatusGatewayTimeout, Code: ErrInternal + "_002", Message: "Request timed out"},
	ErrInternal + "_003":       {HTTPCode: http.StatusRequestTimeout, Code: ErrInternal + "_003", Message: "Request was canceled"},
}

func WrapError(err error, detail string) APIError {
	var apiErr APIError

	// Check for context errors first
	if errors.Is(err, context.DeadlineExceeded) {
		apiErr = APIError{
			HTTPCode:  http.StatusGatewayTimeout,
			Code:      ErrInternal + "_002",
			Message:   "Request timed out",
			Detail:    detail,
			DebugInfo: "Context deadline exceeded",
		}
		return apiErr
	}

	if errors.Is(err, context.Canceled) {
		apiErr = APIError{
			HTTPCode:  http.StatusRequestTimeout,
			Code:      ErrInternal + "_003",
			Message:   "Request was canceled",
			Detail:    detail,
			DebugInfo: "Context was canceled",
		}
		return apiErr
	}

	switch e := err.(type) {
	case APIError:
		// Clone the API error and add detail
		apiErr = e
		if detail != "" {
			if apiErr.Detail != "" {
				apiErr.Detail = fmt.Sprintf("%s: %s", detail, apiErr.Detail)
			} else {
				apiErr.Detail = detail
			}
		}
	default:
		// Create a new internal error
		apiErr = ErrInternalUnexpected
		apiErr.Detail = detail
		apiErr.DebugInfo = err.Error()
	}

	return apiErr
}

// IsErrorCode checks if an error is an APIError with the specified code
func IsErrorCode(err error, code string) bool {
	if apiErr, ok := err.(APIError); ok {
		return apiErr.Code == code
	}
	return false
}

// Add a handler to serve the error catalog
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

	// Sort by error code
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Code < docs[j].Code
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(docs)
}

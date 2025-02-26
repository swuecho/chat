package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"

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

// Define specific error codes
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

	// Resource errors
	ErrResourceNotFound = func(resource string) APIError {
		return APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     ErrResource + "_001",
			Message:  resource + " not found",
		}
	}
	ErrResourceAlreadyExists = func(resource string) APIError {
		return APIError{
			HTTPCode: http.StatusConflict,
			Code:     ErrResource + "_002",
			Message:  resource + " already exists",
		}
	}

	// Validation errors
	ErrValidationInvalidInput = func(detail string) APIError {
		return APIError{
			HTTPCode: http.StatusBadRequest,
			Code:     ErrValidation + "_001",
			Message:  "Invalid input",
			Detail:   detail,
		}
	}

	// Database errors
	ErrDatabaseQuery = APIError{
		HTTPCode:  http.StatusInternalServerError,
		Code:      ErrDatabase + "_001",
		Message:   "Database query failed",
		DebugInfo: "Database operation failed - check logs for details",
	}

	// Internal errors
	ErrInternalUnexpected = APIError{
		HTTPCode:  http.StatusInternalServerError,
		Code:      ErrInternal + "_001",
		Message:   "An unexpected error occurred",
		DebugInfo: "Unexpected internal error - check logs for stack trace",
	}
)

func RespondWithAPIError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var apiErr APIError
	switch e := err.(type) {
	case APIError:
		apiErr = e
	default:
		// Log unexpected errors
		log.Printf("Unexpected error: %v", err)
		apiErr = ErrInternalUnexpected
	}

	// Set status code from the error
	w.WriteHeader(apiErr.HTTPCode)

	// Create response object (don't expose debug info)
	response := struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail,omitempty"`
	}{
		Code:    apiErr.Code,
		Message: apiErr.Message,
		Detail:  apiErr.Detail,
	}

	// Log error with debug info if available
	if apiErr.DebugInfo != "" {
		log.Printf("Error [%s]: %s - %s", apiErr.Code, apiErr.Message, apiErr.DebugInfo)
	}

	json.NewEncoder(w).Encode(response)
}

func MapDatabaseError(err error) error {
	// Map common database errors to appropriate application errors
	if errors.Is(err, sql.ErrNoRows) {
		return ErrResourceNotFound("Record")
	}

	// Check for other specific database errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique violation
			return ErrResourceAlreadyExists("Record")
		case "23503": // Foreign key violation
			return APIError{
				HTTPCode:  http.StatusBadRequest,
				Code:      ErrDatabase + "_003",
				Message:   "Referenced resource does not exist",
				DebugInfo: fmt.Sprintf("Foreign key violation: %s", pgErr.Detail),
			}
		}
	}

	// Log the unhandled database error
	log.Printf("Unhandled database error: %v", err)

	// Return generic database error
	dbErr := ErrDatabaseQuery
	dbErr.DebugInfo = err.Error()
	return dbErr
}

var errorResourceNotFound = ErrResourceNotFound("resource")
var errorResourceAlreadyExists = ErrResourceAlreadyExists("resource")

// ErrorCatalog holds all error codes for documentation purposes
var ErrorCatalog = map[string]struct {
	HTTPCode int
	Message  string
	Example  string
}{
	"AUTH_001": {
		HTTPCode: http.StatusUnauthorized,
		Message:  "Invalid credentials",
		Example:  "Username or password is incorrect",
	},
	"AUTH_002": {
		HTTPCode: http.StatusUnauthorized,
		Message:  "Token has expired",
		Example:  "Your session has expired, please log in again",
	},
	errorResourceNotFound.Code: {
		HTTPCode: errorResourceNotFound.HTTPCode,
		Message:  errorResourceNotFound.Message,
		Example:  "The user with ID 123 could not be found",
	},
	errorResourceAlreadyExists.Code: {
		HTTPCode: errorResourceAlreadyExists.HTTPCode,
		Message:  errorResourceAlreadyExists.Message,
		Example:  "A user with that email address already exists",
	},
}

func WrapError(err error, detail string) APIError {
	var apiErr APIError

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

// Add a handler to serve the error catalog
func ErrorCatalogHandler(w http.ResponseWriter, r *http.Request) {
	type ErrorDoc struct {
		Code     string `json:"code"`
		HTTPCode int    `json:"http_code"`
		Message  string `json:"message"`
		Example  string `json:"example"`
	}

	docs := make([]ErrorDoc, 0, len(ErrorCatalog))
	for code, info := range ErrorCatalog {
		docs = append(docs, ErrorDoc{
			Code:     code,
			HTTPCode: info.HTTPCode,
			Message:  info.Message,
			Example:  info.Example,
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

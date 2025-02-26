package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
			HTTPCode: http.StatusNotFound,
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

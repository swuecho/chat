package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
)

// Common validation patterns
var (
	validationEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	validationUuidRegex  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// ValidationConfig defines validation rules for API endpoints
type ValidationConfig struct {
	MaxBodySize     int64                      // Maximum request body size
	RequiredFields  []string                   // Required JSON fields
	FieldValidators map[string]FieldValidator  // Custom field validators
	AllowedMethods  []string                   // Allowed HTTP methods
	SkipBodyBuffer  bool                       // Skip body buffering for large requests (disables field validation)
}

// FieldValidator defines a validation function for a specific field
type FieldValidator func(value interface{}) error

// Common field validators
func ValidateEmail(value interface{}) error {
	email, ok := value.(string)
	if !ok {
		return ErrValidationInvalidInput("email must be a string")
	}
	if !validationEmailRegex.MatchString(email) {
		return ErrValidationInvalidInput("invalid email format")
	}
	if len(email) > 254 {
		return ErrValidationInvalidInput("email too long")
	}
	return nil
}

func ValidateUUID(value interface{}) error {
	uuid, ok := value.(string)
	if !ok {
		return ErrValidationInvalidInput("UUID must be a string")
	}
	if !validationUuidRegex.MatchString(uuid) {
		return ErrValidationInvalidInput("invalid UUID format")
	}
	return nil
}

func ValidateStringLength(min, max int) FieldValidator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return ErrValidationInvalidInput("value must be a string")
		}
		if !utf8.ValidString(str) {
			return ErrValidationInvalidInput("invalid UTF-8 string")
		}
		if len(str) < min {
			return ErrValidationInvalidInput("string too short")
		}
		if len(str) > max {
			return ErrValidationInvalidInput("string too long")
		}
		return nil
	}
}

func ValidateNonEmpty(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ErrValidationInvalidInput("value must be a string")
	}
	if strings.TrimSpace(str) == "" {
		return ErrValidationInvalidInput("value cannot be empty")
	}
	return nil
}

// ValidationMiddleware creates a validation middleware with the given config
func ValidationMiddleware(config ValidationConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timing for performance monitoring
			start := time.Now()
			
			// Validate HTTP method
			if len(config.AllowedMethods) > 0 {
				methodAllowed := false
				for _, method := range config.AllowedMethods {
					if r.Method == method {
						methodAllowed = true
						break
					}
				}
				if !methodAllowed {
					log.WithFields(log.Fields{
						"method": r.Method,
						"path":   r.URL.Path,
						"ip":     r.RemoteAddr,
					}).Warn("Method not allowed")
					RespondWithAPIError(w, APIError{
						HTTPCode: http.StatusMethodNotAllowed,
						Code:     ErrValidation + "_100",
						Message:  "Method not allowed",
					})
					return
				}
			}

			// Skip validation for GET requests without body
			if r.Method == "GET" || r.Method == "DELETE" {
				next.ServeHTTP(w, r)
				return
			}

			// Validate content type for requests with body
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") && r.ContentLength > 0 {
				log.WithFields(log.Fields{
					"content_type": contentType,
					"path":         r.URL.Path,
					"ip":           r.RemoteAddr,
				}).Warn("Invalid content type")
				RespondWithAPIError(w, APIError{
					HTTPCode: http.StatusUnsupportedMediaType,
					Code:     ErrValidation + "_101",
					Message:  "Content-Type must be application/json",
				})
				return
			}

			// Check content length
			if config.MaxBodySize > 0 && r.ContentLength > config.MaxBodySize {
				log.WithFields(log.Fields{
					"content_length": r.ContentLength,
					"max_size":       config.MaxBodySize,
					"path":           r.URL.Path,
					"ip":             r.RemoteAddr,
				}).Warn("Request body too large")
				RespondWithAPIError(w, APIError{
					HTTPCode: http.StatusRequestEntityTooLarge,
					Code:     ErrValidation + "_102",
					Message:  "Request body too large",
				})
				return
			}

			// For requests without body, skip JSON validation
			if r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Skip body buffering for large requests or when explicitly configured
			if config.SkipBodyBuffer {
				// Just validate content length and skip field validation
				if config.MaxBodySize > 0 && r.ContentLength > config.MaxBodySize {
					log.WithFields(log.Fields{
						"content_length": r.ContentLength,
						"max_size":       config.MaxBodySize,
						"path":           r.URL.Path,
						"ip":             r.RemoteAddr,
					}).Warn("Request body exceeds size limit")
					RespondWithAPIError(w, APIError{
						HTTPCode: http.StatusRequestEntityTooLarge,
						Code:     ErrValidation + "_102",
						Message:  "Request body too large",
					})
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Create a limited reader to prevent memory exhaustion
			var limitedReader io.Reader = r.Body
			if config.MaxBodySize > 0 {
				limitedReader = io.LimitReader(r.Body, config.MaxBodySize+1)
			}
			
			// Read with streaming approach using a buffer
			var bodyBuffer bytes.Buffer
			written, err := io.CopyN(&bodyBuffer, limitedReader, config.MaxBodySize+1)
			r.Body.Close()
			
			if err != nil && err != io.EOF {
				log.WithError(err).WithFields(log.Fields{
					"path": r.URL.Path,
					"ip":   r.RemoteAddr,
				}).Error("Failed to read request body")
				RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to read request body"))
				return
			}

			// Check if body exceeds limit
			if config.MaxBodySize > 0 && written > config.MaxBodySize {
				log.WithFields(log.Fields{
					"body_size": written,
					"max_size":  config.MaxBodySize,
					"path":      r.URL.Path,
					"ip":        r.RemoteAddr,
				}).Warn("Request body exceeds size limit")
				RespondWithAPIError(w, APIError{
					HTTPCode: http.StatusRequestEntityTooLarge,
					Code:     ErrValidation + "_102",
					Message:  "Request body too large",
				})
				return
			}
			
			body := bodyBuffer.Bytes()

			// Parse JSON body
			var jsonData map[string]interface{}
			if len(body) > 0 {
				if err := json.Unmarshal(body, &jsonData); err != nil {
					log.WithError(err).WithFields(log.Fields{
						"path": r.URL.Path,
						"ip":   r.RemoteAddr,
					}).Warn("Invalid JSON in request body")
					RespondWithAPIError(w, ErrValidationInvalidInput("Invalid JSON format").WithDebugInfo(err.Error()))
					return
				}

				// Validate required fields
				for _, field := range config.RequiredFields {
					if _, exists := jsonData[field]; !exists {
						log.WithFields(log.Fields{
							"missing_field": field,
							"path":          r.URL.Path,
							"ip":            r.RemoteAddr,
						}).Warn("Missing required field")
						RespondWithAPIError(w, ErrValidationInvalidInput("Missing required field: "+field))
						return
					}
				}

				// Run field validators
				for fieldName, validator := range config.FieldValidators {
					if value, exists := jsonData[fieldName]; exists {
						if err := validator(value); err != nil {
							log.WithError(err).WithFields(log.Fields{
								"field": fieldName,
								"path":  r.URL.Path,
								"ip":    r.RemoteAddr,
							}).Warn("Field validation failed")
							RespondWithAPIError(w, WrapError(err, "Validation failed for field: "+fieldName))
							return
						}
					}
				}
			}

			// Restore body for next handler
			r.Body = io.NopCloser(bytes.NewReader(body))

			// Add validation context
			ctx := context.WithValue(r.Context(), "validation_duration", time.Since(start))
			r = r.WithContext(ctx)

			log.WithFields(log.Fields{
				"path":     r.URL.Path,
				"method":   r.Method,
				"duration": time.Since(start),
				"ip":       r.RemoteAddr,
			}).Debug("Request validation completed")

			next.ServeHTTP(w, r)
		})
	}
}

// Predefined validation configs for common endpoints
var (
	AuthValidationConfig = ValidationConfig{
		MaxBodySize:    1024 * 10, // 10KB
		RequiredFields: []string{"email", "password"},
		FieldValidators: map[string]FieldValidator{
			"email":    ValidateEmail,
			"password": ValidateStringLength(8, 128),
		},
		AllowedMethods: []string{"POST"},
	}

	ChatValidationConfig = ValidationConfig{
		MaxBodySize:    1024 * 100, // 100KB
		RequiredFields: []string{"prompt"},
		FieldValidators: map[string]FieldValidator{
			"prompt":       ValidateStringLength(1, 10000),
			"session_uuid": ValidateUUID,
			"chat_uuid":    ValidateUUID,
		},
		AllowedMethods: []string{"POST"},
	}

	FileUploadValidationConfig = ValidationConfig{
		MaxBodySize:    32 * 1024 * 1024, // 32MB
		AllowedMethods: []string{"POST", "GET", "DELETE"},
		SkipBodyBuffer: true, // Skip buffering for large file uploads
	}

	GeneralValidationConfig = ValidationConfig{
		MaxBodySize:    1024 * 50, // 50KB
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
	}
)
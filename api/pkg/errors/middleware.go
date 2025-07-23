package errors

import (
	"log"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and converts them to 500 errors
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\nStack trace:\n%s", err, debug.Stack())
				WriteErrorResponse(w, ErrInternalServer)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ErrorHandlingMiddleware provides centralized error handling
func ErrorHandlingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom ResponseWriter that can capture errors
		ew := &errorResponseWriter{ResponseWriter: w}
		next.ServeHTTP(ew, r)
		
		// If there was an error captured, handle it
		if ew.error != nil {
			WriteErrorResponse(w, ew.error)
		}
	})
}

// errorResponseWriter wraps http.ResponseWriter to capture errors
type errorResponseWriter struct {
	http.ResponseWriter
	error error
}

func (ew *errorResponseWriter) WriteError(err error) {
	ew.error = err
}

// HandlerFunc represents a handler that can return an error
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Handle converts an error-returning handler to a standard http.HandlerFunc
func Handle(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			WriteErrorResponse(w, err)
		}
	}
}

// Logger provides structured error logging
type Logger interface {
	LogError(err error, context map[string]interface{})
	LogWarning(message string, context map[string]interface{})
	LogInfo(message string, context map[string]interface{})
}

type defaultLogger struct{}

func (l defaultLogger) LogError(err error, context map[string]interface{}) {
	log.Printf("ERROR: %v, Context: %+v", err, context)
}

func (l defaultLogger) LogWarning(message string, context map[string]interface{}) {
	log.Printf("WARNING: %s, Context: %+v", message, context)
}

func (l defaultLogger) LogInfo(message string, context map[string]interface{}) {
	log.Printf("INFO: %s, Context: %+v", message, context)
}

var defaultLoggerInstance = defaultLogger{}

// LogError logs an error with context
func LogError(err error, context map[string]interface{}) {
	defaultLoggerInstance.LogError(err, context)
}

// LogWarning logs a warning with context
func LogWarning(message string, context map[string]interface{}) {
	defaultLoggerInstance.LogWarning(message, context)
}

// LogInfo logs an info message with context  
func LogInfo(message string, context map[string]interface{}) {
	defaultLoggerInstance.LogInfo(message, context)
}
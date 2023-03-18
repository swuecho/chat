package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func LoggingMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming request
		logger.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
		}).Info("Request received")

		// Create a new ResponseWriter to capture status code
		recorder := newLoggingResponseWriter(w)

		// Call the next handler in the chain
		next.ServeHTTP(recorder, r)

		// Log the outgoing response
		logger.WithFields(log.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
			"remote": r.RemoteAddr,
			"status": recorder.statusCode,
			"size":   recorder.size,
		}).Info("Response sent")
	})
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, 0}
}

// loggingResponseWriter wraps an http.ResponseWriter, allowing us to log
// the response status code and size.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lrw.ResponseWriter.Write(b)
	lrw.size += size
	return size, err
}

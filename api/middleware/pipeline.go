package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type Middleware func(http.Handler) http.Handler

// Chain composes middleware in the same order in which requests execute.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

type responseMetricsWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *responseMetricsWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseMetricsWriter) Write(payload []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(payload)
	w.bytes += n
	return n, err
}

// Unwrap allows http.ResponseController to retain Flusher, Hijacker, Pusher,
// and full-duplex support provided by the underlying writer.
func (w *responseMetricsWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		wrapped := &responseMetricsWriter{ResponseWriter: w}
		next.ServeHTTP(wrapped, r)
		status := wrapped.status
		if status == 0 {
			status = http.StatusOK
		}
		slog.Info("http request",
			"request_id", GetRequestID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"response_bytes", wrapped.bytes,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

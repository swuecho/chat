package middleware

import (
	"net/http"
	"sync"
	"time"
)

// LastRequestTracker tracks the last request time for idle detection.
type LastRequestTracker struct {
	mu          sync.RWMutex
	lastRequest time.Time
}

// NewLastRequestTracker creates a new tracker.
func NewLastRequestTracker() *LastRequestTracker {
	return &LastRequestTracker{lastRequest: time.Now()}
}

// Update updates the last request time to now.
func (t *LastRequestTracker) Update() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastRequest = time.Now()
}

// Since returns the duration since the last request.
func (t *LastRequestTracker) Since() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Since(t.lastRequest)
}

// UpdateLastRequestTime is a middleware that updates the last request time.
func UpdateLastRequestTime(tracker *LastRequestTracker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracker.Update()
			next.ServeHTTP(w, r)
		})
	}
}

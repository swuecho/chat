package main

import (
	"net/http"
	"time"
)

// Middleware to update lastRequest time
func UpdateLastRequestTime(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Update lastRequest time
		lastRequest = time.Now()

		// Call next middleware/handler
		next.ServeHTTP(w, r)
	})
}

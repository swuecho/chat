package main

import (
	"net/http"
	"time"

	"github.com/swuecho/chat_backend/middleware"
)

var requestTracker = middleware.NewLastRequestTracker()

// UpdateLastRequestTime middleware for Fly.io idle detection.
func UpdateLastRequestTime(next http.Handler) http.Handler {
	return middleware.UpdateLastRequestTime(requestTracker)(next)
}

// lastRequest is kept for backward compatibility with existing code.
var lastRequest time.Time

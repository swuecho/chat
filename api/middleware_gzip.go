package main

import (
	"net/http"

	"github.com/swuecho/chat_backend/middleware"
)

// makeGzipHandler delegates to the middleware package.
func makeGzipHandler(next http.Handler) http.Handler {
	return middleware.MakeGzipHandler(next)
}

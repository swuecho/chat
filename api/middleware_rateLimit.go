package main

import (
	"net/http"

	"github.com/swuecho/chat_backend/middleware"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// RateLimitByUserID delegates to the middleware package.
func RateLimitByUserID(q *sqlc_queries.Queries) func(http.Handler) http.Handler {
	return middleware.RateLimitByUserID(q, appConfig.OPENAI.RATELIMIT)
}

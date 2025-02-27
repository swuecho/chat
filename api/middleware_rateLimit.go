package main

import (
	"fmt"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"net/http"
)

// This function returns a middleware that limits requests from each user by their ID.
func RateLimitByUserID(q *sqlc_queries.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Get the user ID from the request, e.g. from a JWT token.
			if r.URL.Path == "/chat" || r.URL.Path == "/chat_stream" {
				ctx := r.Context()
				userIDInt, err := getUserID(ctx)
				// role := ctx.Value(roleContextKey).(string)

				if err != nil {
					apiErr := ErrAuthInvalidCredentials
					apiErr.Detail = "User identification required for rate limiting"
					apiErr.DebugInfo = err.Error()
					RespondWithAPIError(w, apiErr)
					return
				}
				messageCount, err := q.GetChatMessagesCount(r.Context(), int32(userIDInt))
				if err != nil {
					apiErr := ErrInternalUnexpected
					apiErr.Detail = "Could not get message count for rate limiting"
					apiErr.DebugInfo = err.Error()
					RespondWithAPIError(w, apiErr)
					return
				}
				maxRate, err := q.GetRateLimit(r.Context(), int32(userIDInt))
				if err != nil {
					maxRate = int32(appConfig.OPENAI.RATELIMIT)
				}

				if messageCount >= int64(maxRate) {
					apiErr := ErrTooManyRequests
					apiErr.Detail = fmt.Sprintf("Rate limit exceeded: messageCount=%d, maxRate=%d", messageCount, maxRate)
					RespondWithAPIError(w, apiErr)
					return
				}
			}
			// Call the next handler.
			next.ServeHTTP(w, r)
		})
	}
}

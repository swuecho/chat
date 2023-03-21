package main

import (
	"net/http"
	"strconv"

	"github.com/swuecho/chatgpt_backend/sqlc_queries"
)

// This function returns a middleware that limits requests from each user by their ID.
func RateLimitByUserID(q *sqlc_queries.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the user ID from the request, e.g. from a JWT token.
			ctx := r.Context()
			userIDStr := ctx.Value(userContextKey).(string)
			userIDInt, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", http.StatusBadRequest)
				return
			}
			messageCount, err := q.GetChatMessagesCount(r.Context(), int32(userIDInt))
			if err != nil {
				http.Error(w, "Error: Could not get message count.", http.StatusInternalServerError)
				return
			}

			// Check if the request exceeds the rate limit.
			if messageCount > 100 {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			// Call the next handler.
			next.ServeHTTP(w, r)
		})
	}
}

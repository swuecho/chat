package main

import (
	"net/http"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
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
					RespondWithErrorMessage(w, http.StatusUnauthorized, "no user", err)
					return
				}
				messageCount, err := q.GetChatMessagesCount(r.Context(), int32(userIDInt))
				if err != nil {
					http.Error(w, eris.Wrap(err, "error: Could not get message count. ").Error(), http.StatusInternalServerError)
					return
				}
				maxRate, err := q.GetRateLimit(r.Context(), int32(userIDInt))
				if err != nil {
					maxRate = int32(appConfig.OPENAI.RATELIMIT)
				}

				if messageCount >= int64(maxRate) {
					RespondWithErrorMessage(w, http.StatusTooManyRequests, "error.rateLimit", map[string]interface{}{
						"messageCount": messageCount,
						"maxRate":      maxRate,
					})
					return
				}
			}
			// Call the next handler.
			next.ServeHTTP(w, r)
		})
	}
}

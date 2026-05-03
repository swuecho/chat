package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// RateLimitByUserID returns a middleware that limits requests per user.
func RateLimitByUserID(q *sqlc_queries.Queries, defaultLimit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if !strings.HasSuffix(path, "/chat") && !strings.HasSuffix(path, "/chat_stream") && !strings.HasSuffix(path, "/chatbot") {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			userIDInt, err := GetUserID(ctx)
			if err != nil {
				apiErr := dto.ErrAuthInvalidCredentials
				apiErr.Detail = "User identification required for rate limiting"
				apiErr.DebugInfo = err.Error()
				dto.RespondWithAPIError(w, apiErr)
				return
			}

			messageCount, err := q.GetChatMessagesCount(r.Context(), userIDInt)
			if err != nil {
				apiErr := dto.ErrInternalUnexpected
				apiErr.Detail = "Could not get message count for rate limiting"
				apiErr.DebugInfo = err.Error()
				dto.RespondWithAPIError(w, apiErr)
				return
			}

			maxRate, err := q.GetRateLimit(r.Context(), userIDInt)
			if err != nil {
				maxRate = int32(defaultLimit)
			}

			if messageCount >= int64(maxRate) {
				apiErr := dto.ErrTooManyRequests
				apiErr.Detail = fmt.Sprintf("Rate limit exceeded: messageCount=%d, maxRate=%d", messageCount, maxRate)
				dto.RespondWithAPIError(w, apiErr)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

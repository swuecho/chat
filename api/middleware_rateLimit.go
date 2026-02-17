package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// GinRateLimitByUserID - Gin middleware for rate limiting by user ID
func GinRateLimitByUserID(q *sqlc_queries.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Only apply rate limiting to chat endpoints
		if strings.HasSuffix(path, "/chat") || strings.HasSuffix(path, "/chat_stream") || strings.HasSuffix(path, "/chatbot") {
			userIDInt, err := GetUserID(c)
			if err != nil {
				apiErr := ErrAuthInvalidCredentials
				apiErr.Detail = "User identification required for rate limiting"
				apiErr.DebugInfo = err.Error()
				apiErr.GinResponse(c)
				c.Abort()
				return
			}

			messageCount, err := q.GetChatMessagesCount(c.Request.Context(), userIDInt)
			if err != nil {
				apiErr := ErrInternalUnexpected
				apiErr.Detail = "Could not get message count for rate limiting"
				apiErr.DebugInfo = err.Error()
				apiErr.GinResponse(c)
				c.Abort()
				return
			}

			maxRate, err := q.GetRateLimit(c.Request.Context(), userIDInt)
			if err != nil {
				maxRate = int32(appConfig.OPENAI.RATELIMIT)
			}

			if messageCount >= int64(maxRate) {
				apiErr := ErrTooManyRequests
				apiErr.Detail = fmt.Sprintf("Rate limit exceeded: messageCount=%d, maxRate=%d", messageCount, maxRate)
				apiErr.GinResponse(c)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

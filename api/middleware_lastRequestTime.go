package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinUpdateLastRequestTime - Gin middleware to update last request time
func GinUpdateLastRequestTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Update lastRequest time
		lastRequest = time.Now()

		c.Next()
	}
}

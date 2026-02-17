package main

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Services holds all service dependencies for the router
type Services struct {
	SQLC        *sqlc_queries.Queries
	SecretSvc   *JWTSecretService
	AuthUserSvc *AuthUserService
}

// SetupRouter creates and configures the Gin router
func SetupRouter(services *Services, jwtSecret sqlc_queries.JwtSecret) *gin.Engine {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS configuration
	r.Use(cors.New(getCORSConfig()))

	// Gzip compression
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// API routes
	api := r.Group("/api")

	// Public routes (no authentication required)
	setupPublicRoutes(api, services)

	// User routes (authenticated)
	userGroup := api.Group("")
	userGroup.Use(GinUserAuthMiddleware(jwtSecret))
	userGroup.Use(GinRateLimitByUserID(services.SQLC))
	setupUserRoutes(userGroup, services)

	// Admin routes (authenticated + admin role required)
	adminGroup := api.Group("/admin")
	adminGroup.Use(GinAdminAuthMiddleware(jwtSecret))
	adminGroup.Use(GinRateLimitByUserID(services.SQLC))
	setupAdminRoutes(adminGroup, services)

	// Fly.io specific middleware for idle shutdown
	if os.Getenv("FLY_APP_NAME") != "" {
		r.Use(GinUpdateLastRequestTime())
	}

	// Note: Static file serving is handled in main.go via setupGinStaticFiles
	// to properly serve from embedded FS

	return r
}

// getCORSConfig returns CORS configuration
func getCORSConfig() cors.Config {
	defaultOrigins := []string{"http://localhost:9002", "http://localhost:3000"}
	allowedOrigins := append([]string{}, defaultOrigins...)
	restrictToConfigured := false

	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		parts := strings.Split(corsOrigins, ",")
		allowedOrigins = allowedOrigins[:0]
		for _, origin := range parts {
			trimmed := strings.TrimSpace(origin)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
		restrictToConfigured = true
	}

	return cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if len(allowedOrigins) == 0 {
				return true
			}
			for _, allowed := range allowedOrigins {
				if allowed == "*" {
					return true
				}
				if strings.EqualFold(origin, allowed) {
					return true
				}
			}
			if !restrictToConfigured {
				if strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") ||
					strings.HasPrefix(origin, "https://localhost:") ||
					strings.HasPrefix(origin, "https://127.0.0.1:") {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Cache-Control", "Connection", "Pragma", "Accept", "Accept-Language", "Origin", "Referer"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// setupPublicRoutes configures routes that don't require authentication
func setupPublicRoutes(api *gin.RouterGroup, services *Services) {
	// Auth routes (login/signup)
	authHandler := NewAuthUserHandler(services.SQLC)
	authHandler.GinRegisterPublic(api)

	// TTS endpoint
	api.GET("/tts", gin.WrapF(handleTTSRequest))

	// Error catalog
	api.GET("/errors", gin.WrapF(ErrorCatalogHandler))

	// Refresh token endpoint
	api.POST("/refresh", authHandler.GinRefreshToken)
}

// setupUserRoutes configures routes that require user authentication
func setupUserRoutes(user *gin.RouterGroup, services *Services) {
	// Chat models
	chatModelHandler := NewChatModelHandler(services.SQLC)
	chatModelHandler.GinRegister(user)

	// Auth user routes
	authHandler := NewAuthUserHandler(services.SQLC)
	authHandler.GinRegister(user)

	// Chat sessions
	chatSessionHandler := NewChatSessionHandler(services.SQLC)
	chatSessionHandler.GinRegister(user)

	// Active session
	activeSessionHandler := NewUserActiveChatSessionHandler(services.SQLC)
	activeSessionHandler.GinRegister(user)

	// Workspaces
	chatWorkspaceHandler := NewChatWorkspaceHandler(services.SQLC)
	chatWorkspaceHandler.GinRegister(user)

	// Messages
	chatMessageHandler := NewChatMessageHandler(services.SQLC)
	chatMessageHandler.GinRegister(user)

	// Snapshots
	chatSnapshotHandler := NewChatSnapshotHandler(services.SQLC)
	chatSnapshotHandler.GinRegister(user)

	// Chat main handler
	chatHandler := NewChatHandler(services.SQLC)
	chatHandler.GinRegister(user)

	// Prompts
	promptHandler := NewChatPromptHandler(services.SQLC)
	promptHandler.GinRegister(user)

	// Model privileges
	userModelPrivilegeHandler := NewUserChatModelPrivilegeHandler(services.SQLC)
	userModelPrivilegeHandler.GinRegister(user)

	// Files
	chatFileHandler := NewChatFileHandler(services.SQLC)
	chatFileHandler.GinRegister(user)

	// Comments
	chatCommentHandler := NewChatCommentHandler(services.SQLC)
	chatCommentHandler.GinRegister(user)

	// Bot answer history
	botAnswerHistoryHandler := NewBotAnswerHistoryHandler(services.SQLC)
	botAnswerHistoryHandler.GinRegister(user)
}

// setupAdminRoutes configures routes that require admin authentication
func setupAdminRoutes(admin *gin.RouterGroup, services *Services) {
	adminHandler := NewAdminHandler(services.AuthUserSvc)
	adminHandler.GinRegisterRoutes(admin)
}

// PrintRoutes prints all registered routes (for debugging)
func PrintRoutes(r *gin.Engine) {
	for _, route := range r.Routes() {
		println(route.Method, route.Path)
	}
}

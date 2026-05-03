package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/swuecho/chat_backend/config"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/middleware"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/static"
	"golang.org/x/time/rate"
)

//go:embed sqlc/schema.sql
var schemaBytes []byte

var appConfig config.AppConfig
var jwtSecretAndAud sqlc_queries.JwtSecret
var openAIRateLimiter *rate.Limiter

func main() {
	// Rate limiter: 3000 requests per minute, burst 500
	openAIRateLimiter = rate.NewLimiter(rate.Every(time.Minute/3000), 500)

	// Load configuration from environment variables
	appConfig = config.Load(log.Default())

	log.Printf("%+v", appConfig)

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	var connStr string
	if dbURL == "" {
		pg := appConfig.PG
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			pg.HOST, pg.PORT, pg.USER, pg.PASS, pg.DB)
	} else {
		connStr = dbURL
	}

	pgdb, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pgdb.Close()

	// Run schema
	if _, err := pgdb.Exec(string(schemaBytes)); err != nil {
		log.Fatal("Failed to execute SQL schema: ", err)
	}
	fmt.Println("SQL statements executed successfully")

	// Router setup
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()
	sqlc_q := sqlc_queries.New(pgdb)

	// JWT secret
	secretService := NewJWTSecretService(sqlc_q)
	jwtSecretAndAud, err = secretService.GetOrCreateJwtSecret(context.Background(), "chat")
	if err != nil {
		log.Fatal(err)
	}
	middleware.SetJWTSecret(jwtSecretAndAud.Secret)

	// Subrouters
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	userRouter := apiRouter.NewRoute().Subrouter()

	// Auth middleware
	adminRouter.Use(middleware.AdminAuthMiddleware)
	userRouter.Use(middleware.UserAuthMiddleware)

	// Rate limiting
	adminRouter.Use(middleware.RateLimitByUserID(sqlc_q, appConfig.OPENAI.RATELIMIT))
	userRouter.Use(middleware.RateLimitByUserID(sqlc_q, appConfig.OPENAI.RATELIMIT))

	// --- Route registration ---
	// Chat models
	ChatModelHandler := NewChatModelHandler(sqlc_q)
	ChatModelHandler.Register(userRouter)

	// Auth
	userHandler := NewAuthUserHandler(sqlc_q)
	userHandler.Register(userRouter)
	userHandler.RegisterPublicRoutes(apiRouter)

	// Admin
	authUserService := NewAuthUserService(sqlc_q)
	adminHandler := NewAdminHandler(authUserService)
	adminHandler.RegisterRoutes(adminRouter)

	// Prompts
	promptHandler := NewChatPromptHandler(sqlc_q)
	promptHandler.Register(userRouter)

	// Sessions (register before workspaces to avoid route shadowing)
	chatSessionHandler := NewChatSessionHandler(sqlc_q)
	chatSessionHandler.Register(userRouter)

	// Active sessions
	activeSessionHandler := NewUserActiveChatSessionHandler(sqlc_q)
	activeSessionHandler.Register(userRouter)

	// Workspaces
	chatWorkspaceHandler := NewChatWorkspaceHandler(sqlc_q)
	chatWorkspaceHandler.Register(userRouter)

	// Messages
	chatMessageHandler := NewChatMessageHandler(sqlc_q)
	chatMessageHandler.Register(userRouter)

	// Snapshots
	chatSnapshotHandler := NewChatSnapshotHandler(sqlc_q)
	chatSnapshotHandler.Register(userRouter)

	// Chat stream
	chatHandler := NewChatHandler(sqlc_q)
	chatHandler.Register(userRouter)

	// Model privileges
	user_model_privilege_handler := NewUserChatModelPrivilegeHandler(sqlc_q)
	user_model_privilege_handler.Register(userRouter)

	// Files
	chatFileHandler := NewChatFileHandler(sqlc_q)
	chatFileHandler.Register(userRouter)

	// Comments
	chatCommentHandler := NewChatCommentHandler(sqlc_q)
	chatCommentHandler.Register(userRouter)

	// Bot history
	botAnswerHistoryHandler := NewBotAnswerHistoryHandler(sqlc_q)
	botAnswerHistoryHandler.Register(userRouter)

	// Public endpoints
	apiRouter.HandleFunc("/tts", handleTTSRequest)
	apiRouter.HandleFunc("/errors", dto.ErrorCatalogHandler)

	// Static files
	fs := http.FileServer(http.FS(static.StaticFiles))
	cacheHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			w.Header().Set("Cache-Control", "max-age=31536000")
		} else if r.URL.Path == "" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		fs.ServeHTTP(w, r)
	})
	router.PathPrefix("/").Handler(middleware.MakeGzipHandler(cacheHandler))

	// Fly.io idle monitor
	requestTracker := middleware.NewLastRequestTracker()
	if os.Getenv("FLY_APP_NAME") != "" {
		router.Use(middleware.UpdateLastRequestTime(requestTracker))
		go func() {
			restartInterval := os.Getenv("FLY_RESTART_INTERVAL_IF_IDLE")
			if restartInterval == "" {
				restartInterval = "30m"
			}
			duration, err := time.ParseDuration(restartInterval)
			if err != nil {
				return
			}
			for {
				time.Sleep(1 * time.Minute)
				if requestTracker.Since() > duration {
					fmt.Printf("No activity for %s. Exiting.\n", restartInterval)
					os.Exit(0)
				}
			}
		}()
	}

	// CORS
	defaultOrigins := []string{"http://localhost:9002", "http://localhost:3000"}
	allowedOrigins := append([]string{}, defaultOrigins...)
	restrictToConfigured := false
	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		allowedOrigins = allowedOrigins[:0]
		for _, origin := range strings.Split(corsOrigins, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
		restrictToConfigured = true
	}

	originValidator := func(origin string) bool {
		if len(allowedOrigins) == 0 {
			return true
		}
		for _, allowed := range allowedOrigins {
			if allowed == "*" || strings.EqualFold(origin, allowed) {
				return true
			}
		}
		if !restrictToConfigured {
			if strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:") {
				return true
			}
			if strings.HasPrefix(origin, "https://localhost:") || strings.HasPrefix(origin, "https://127.0.0.1:") {
				return true
			}
		}
		return false
	}

	corsRouter := handlers.CORS(
		handlers.AllowedOriginValidator(originValidator),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Cache-Control", "Connection", "Pragma", "Accept", "Accept-Language", "Origin", "Referer"}),
		handlers.AllowCredentials(),
	)(router)

	// Log routes
	router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err1 := route.GetPathTemplate()
		met, err2 := route.GetMethods()
		fmt.Println(tpl, err1, met, err2)
		return nil
	})

	// Start server
	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", corsRouter)
	if err != nil {
		log.Fatal(err)
	}
}

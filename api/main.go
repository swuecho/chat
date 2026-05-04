package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/swuecho/chat_backend/config"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/handler"
	"github.com/swuecho/chat_backend/middleware"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/static"
	"github.com/swuecho/chat_backend/svc"
	"golang.org/x/time/rate"
)

//go:embed sqlc/schema.sql
var schemaBytes []byte

// server holds all application dependencies, avoiding package-level globals.
type server struct {
	cfg            config.AppConfig
	db             *sql.DB
	q              *sqlc_queries.Queries
	jwtSecret      sqlc_queries.JwtSecret
	rateLimiter    *rate.Limiter
	requestTracker *middleware.LastRequestTracker
}

func main() {
	if err := run(); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// --- Configuration ---
	cfg := config.Load()

	// --- Database ---
	pgdb, err := openDB(cfg)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer pgdb.Close()

	// Run schema migrations
	if _, err := pgdb.Exec(string(schemaBytes)); err != nil {
		return fmt.Errorf("schema migration: %w", err)
	}
	slog.Info("schema migration complete")

	// --- Build server ---
	srv := &server{
		cfg:            cfg,
		db:             pgdb,
		q:              sqlc_queries.New(pgdb),
		rateLimiter:    rate.NewLimiter(rate.Every(time.Minute/3000), 500),
		requestTracker: middleware.NewLastRequestTracker(),
	}

	// JWT secret
	secretSvc := svc.NewJWTSecretService(srv.q)
	jwtSecretAndAud, err := secretSvc.GetOrCreateJwtSecret(context.Background(), "chat")
	if err != nil {
		return fmt.Errorf("jwt secret: %w", err)
	}
	srv.jwtSecret = jwtSecretAndAud

	// --- Router ---
	router, rawRouter := srv.buildRouter()

	// --- Fly.io idle monitor (before wrapping in CORS) ---
	if os.Getenv("FLY_APP_NAME") != "" {
		rawRouter.Use(middleware.UpdateLastRequestTime(srv.requestTracker))
		go srv.idleMonitor()
	}

	// --- HTTP Server ---
	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
		WriteTimeout:      10 * time.Minute, // Long write timeout for streaming
		IdleTimeout:       120 * time.Second,
	}

	// --- Graceful shutdown ---
	idleConnsClosed := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		slog.Info("shutting down", "signal", sig.String())

		// Drain the rate limiter
		srv.rateLimiter.SetLimit(0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(ctx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
		close(idleConnsClosed)
	}()

	slog.Info("server starting", "addr", ":8080")
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}

	<-idleConnsClosed
	slog.Info("server stopped")
	return nil
}

// openDB creates a database connection from the given config.
func openDB(cfg config.AppConfig) (*sql.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	var connStr string
	if dbURL == "" {
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			cfg.PG.HOST, cfg.PG.PORT, cfg.PG.USER, cfg.PG.PASS, cfg.PG.DB)
	} else {
		connStr = dbURL
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

// buildRouter constructs the HTTP router and returns both the CORS-wrapped handler
// and the raw mux router (for applying middleware after construction).
func (s *server) buildRouter() (http.Handler, *mux.Router) {
	router := mux.NewRouter()
	apiRouter := router.PathPrefix("/api").Subrouter()

	// --- Global middleware ---
	router.Use(middleware.RecoveryMiddleware)
	router.Use(middleware.RequestIDMiddleware)
	router.Use(middleware.BodyLimitMiddleware)

	// --- Health check (public, before auth) ---
	apiRouter.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)

	// --- Subrouters ---
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	userRouter := apiRouter.NewRoute().Subrouter()

	// Auth middleware
	adminRouter.Use(middleware.AdminAuthMiddleware(s.jwtSecret.Secret))
	userRouter.Use(middleware.UserAuthMiddleware(s.jwtSecret.Secret))

	// Rate limiting
	rateLimitMW := middleware.RateLimitByUserID(s.q, s.cfg.OPENAI.RATELIMIT)
	adminRouter.Use(rateLimitMW)
	userRouter.Use(rateLimitMW)

	// --- Route registration ---
	s.registerRoutes(apiRouter, adminRouter, userRouter)

	// --- Static files ---
	fs := http.FileServer(http.FS(static.StaticFiles))
	router.PathPrefix("/").Handler(middleware.MakeGzipHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/static/") {
				w.Header().Set("Cache-Control", "max-age=31536000")
			} else if r.URL.Path == "" {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}
			fs.ServeHTTP(w, r)
		}),
	))

	// --- CORS ---
	return s.corsMiddleware(router), router
}

// registerRoutes wires all HTTP handlers.
func (s *server) registerRoutes(apiRouter, adminRouter, userRouter *mux.Router) {
	q := s.q
	rateLimit := int32(s.cfg.OPENAI.RATELIMIT)
	openAIKey := s.cfg.OPENAI.API_KEY
	openAIProxy := s.cfg.OPENAI.PROXY_URL
	jwtSecret := s.jwtSecret.Secret
	jwtAudience := s.jwtSecret.Audience

	// Public
	apiRouter.HandleFunc("/tts", handler.HandleTTSRequest)
	apiRouter.HandleFunc("/errors", dto.ErrorCatalogHandler)

	// Chat models
	handler.NewChatModelHandler(q).Register(userRouter)

	// Auth
	authHandler := handler.NewAuthUserHandler(q, jwtSecret, jwtAudience, rateLimit)
	authHandler.Register(userRouter)
	authHandler.RegisterPublicRoutes(apiRouter)

	// Admin
	handler.NewAdminHandler(svc.NewAuthUserService(q, jwtSecret, rateLimit), rateLimit).RegisterRoutes(adminRouter)

	// Prompts
	handler.NewChatPromptHandler(q).Register(userRouter)

	// Sessions
	handler.NewChatSessionHandler(q).Register(userRouter)

	// Active sessions
	handler.NewUserActiveChatSessionHandler(q).Register(userRouter)

	// Workspaces
	handler.NewChatWorkspaceHandler(q).Register(userRouter)

	// Messages
	handler.NewChatMessageHandler(q, openAIKey, openAIProxy).Register(userRouter)

	// Snapshots
	handler.NewChatSnapshotHandler(q).Register(userRouter)

	// Chat stream
	handler.NewChatHandler(q, s.rateLimiter, openAIKey, openAIProxy).Register(userRouter)

	// Model privileges
	handler.NewUserChatModelPrivilegeHandler(q).Register(userRouter)

	// Files
	handler.NewChatFileHandler(q).Register(userRouter)

	// Comments
	handler.NewChatCommentHandler(q).Register(userRouter)

	// Bot history
	handler.NewBotAnswerHistoryHandler(q).Register(userRouter)
}

// healthCheck returns server health status.
func (s *server) healthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	healthy := true
	checks := map[string]string{}

	if err := s.db.PingContext(ctx); err != nil {
		healthy = false
		checks["database"] = "unhealthy: " + err.Error()
	} else {
		checks["database"] = "healthy"
	}

	checks["version"] = "1.0.0"

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	dto.RespondWithJSON(w, status, map[string]interface{}{
		"status": statusToText(healthy),
		"checks": checks,
	})
}

func statusToText(healthy bool) string {
	if healthy {
		return "ok"
	}
	return "degraded"
}

// corsMiddleware configures CORS for the router.
func (s *server) corsMiddleware(router *mux.Router) http.Handler {
	allowedOrigins := []string{"http://localhost:9002", "http://localhost:3000"}
	restrict := false

	if corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); corsOrigins != "" {
		allowedOrigins = nil
		for _, origin := range strings.Split(corsOrigins, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
		restrict = true
	}

	return handlers.CORS(
		handlers.AllowedOriginValidator(func(origin string) bool {
			for _, allowed := range allowedOrigins {
				if allowed == "*" || strings.EqualFold(origin, allowed) {
					return true
				}
			}
			if !restrict {
				return strings.HasPrefix(origin, "http://localhost:") ||
					strings.HasPrefix(origin, "http://127.0.0.1:") ||
					strings.HasPrefix(origin, "https://localhost:") ||
					strings.HasPrefix(origin, "https://127.0.0.1:")
			}
			return false
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Cache-Control", "Connection", "Pragma", "Accept", "Accept-Language", "Origin", "Referer", "X-Request-Id"}),
		handlers.AllowCredentials(),
	)(router)
}

// idleMonitor periodically checks for inactivity and exits on Fly.io.
func (s *server) idleMonitor() {
	interval := os.Getenv("FLY_RESTART_INTERVAL_IF_IDLE")
	if interval == "" {
		interval = "30m"
	}
	duration, err := time.ParseDuration(interval)
	if err != nil {
		slog.Warn("invalid FLY_RESTART_INTERVAL_IF_IDLE, disabling idle monitor", "value", interval)
		return
	}
	for {
		time.Sleep(1 * time.Minute)
		if s.requestTracker.Since() > duration {
			slog.Info("idle timeout reached, exiting", "duration", duration)
			os.Exit(0)
		}
	}
}

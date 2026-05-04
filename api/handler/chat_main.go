package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"log/slog"
	"github.com/gorilla/mux"
	"golang.org/x/time/rate"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

// ChatHandler handles chat completion and streaming requests.
type ChatHandler struct {
	service         *svc.ChatService
	sessionSvc      *svc.ChatSessionService
	chatfileService *svc.ChatFileService
	requestCtx      context.Context
	rateLimiter     *rate.Limiter
}

const sessionTitleGenerationTimeout = 30 * time.Second

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(sqlc_q *sqlc_queries.Queries, rateLimiter *rate.Limiter) *ChatHandler {
	return &ChatHandler{
		service:         svc.NewChatService(sqlc_q),
		sessionSvc:      svc.NewChatSessionService(sqlc_q),
		chatfileService: svc.NewChatFileService(sqlc_q),
		requestCtx:      context.Background(),
		rateLimiter:     rateLimiter,
	}
}

// Register registers chat routes on the given router.
func (h *ChatHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_stream", h.ChatCompletionHandler).Methods(http.MethodPost)
	router.HandleFunc("/chatbot", h.ChatBotCompletionHandler).Methods(http.MethodPost)
	router.HandleFunc("/chat_instructions", h.GetChatInstructions).Methods(http.MethodGet)
}

// --- provider.Handler implementation ---

func (h *ChatHandler) RequestContext() context.Context { return h.requestCtx }
func (h *ChatHandler) Queries() *sqlc_queries.Queries  { return h.service.Q() }
func (h *ChatHandler) Config() provider.Config {
	return provider.Config{
		OpenAIKey:   svc.Cfg.OpenAIKey,
		OpenAIProxy: svc.Cfg.OpenAIProxy,
		RateLimiter: h.rateLimiter,
	}
}

// GetRequestContext is a legacy alias for RequestContext.
func (h *ChatHandler) GetRequestContext() context.Context {
	return h.requestCtx
}

// GetChatInstructions returns artifact instruction text.
func (h *ChatHandler) GetChatInstructions(w http.ResponseWriter, r *http.Request) {
	artifactInstruction, err := svc.LoadArtifactInstruction()
	if err != nil {
		slog.Warn("Failed to load artifact instruction: %v", err)
		artifactInstruction = ""
	}
	json.NewEncoder(w).Encode(dto.ChatInstructionResponse{
		ArtifactInstruction: artifactInstruction,
	})
}

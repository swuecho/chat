package handler

import (
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
	rateLimiter     *rate.Limiter
	openAIKey       string
	openAIProxy     string
}

const sessionTitleGenerationTimeout = 30 * time.Second

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(sqlc_q *sqlc_queries.Queries, rateLimiter *rate.Limiter, openAIKey, openAIProxy string) *ChatHandler {
	return &ChatHandler{
		service:         svc.NewChatService(sqlc_q, openAIKey, openAIProxy),
		sessionSvc:      svc.NewChatSessionService(sqlc_q),
		chatfileService: svc.NewChatFileService(sqlc_q),
		rateLimiter:     rateLimiter,
		openAIKey:       openAIKey,
		openAIProxy:     openAIProxy,
	}
}

// Register registers chat routes on the given router.
func (h *ChatHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_stream", h.ChatCompletionHandler).Methods(http.MethodPost)
	router.HandleFunc("/chatbot", h.ChatBotCompletionHandler).Methods(http.MethodPost)
	router.HandleFunc("/chat_instructions", h.GetChatInstructions).Methods(http.MethodGet)
}

// --- provider.Handler implementation ---

func (h *ChatHandler) Queries() *sqlc_queries.Queries { return h.service.Q() }
func (h *ChatHandler) Config() provider.Config {
	return provider.Config{
		OpenAIKey:   h.openAIKey,
		OpenAIProxy: h.openAIProxy,
		RateLimiter: h.rateLimiter,
	}
}

// GetChatInstructions returns artifact instruction text.
func (h *ChatHandler) GetChatInstructions(w http.ResponseWriter, r *http.Request) {
	artifactInstruction, err := svc.LoadArtifactInstruction()
	if err != nil {
		slog.Warn("Failed to load artifact instruction", "error", err)
		artifactInstruction = ""
	}
	json.NewEncoder(w).Encode(dto.ChatInstructionResponse{
		ArtifactInstruction: artifactInstruction,
	})
}

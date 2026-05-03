package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatHandler handles chat completion and streaming requests.
type ChatHandler struct {
	service         *ChatService
	sessionSvc      *ChatSessionService
	chatfileService *ChatFileService
	requestCtx      context.Context
}

const sessionTitleGenerationTimeout = 30 * time.Second

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(sqlc_q *sqlc_queries.Queries) *ChatHandler {
	return &ChatHandler{
		service:         NewChatService(sqlc_q),
		sessionSvc:      NewChatSessionService(sqlc_q),
		chatfileService: NewChatFileService(sqlc_q),
		requestCtx:      context.Background(),
	}
}

// Register registers chat routes on the given router.
func (h *ChatHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_stream", h.ChatCompletionHandler).Methods(http.MethodPost)
	router.HandleFunc("/chatbot", h.ChatBotCompletionHandler).Methods(http.MethodPost)
	router.HandleFunc("/chat_instructions", h.GetChatInstructions).Methods(http.MethodGet)
}

// GetRequestContext returns the current request context for streaming operations.
func (h *ChatHandler) GetRequestContext() context.Context {
	return h.requestCtx
}

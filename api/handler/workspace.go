package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

// ChatWorkspaceHandler handles HTTP requests for workspace management.
type ChatWorkspaceHandler struct {
	wsService      *svc.ChatWorkspaceService
	sessionService *svc.ChatSessionService
	activeSession  *svc.UserActiveChatSessionService
}

// NewChatWorkspaceHandler creates a new ChatWorkspaceHandler with all required services.
func NewChatWorkspaceHandler(q *sqlc_queries.Queries) *ChatWorkspaceHandler {
	return &ChatWorkspaceHandler{
		wsService:      svc.NewChatWorkspaceService(q),
		sessionService: svc.NewChatSessionService(q),
		activeSession:  svc.NewUserActiveChatSessionService(q),
	}
}

// Register registers workspace routes on the given router.
func (h *ChatWorkspaceHandler) Register(router *mux.Router) {
	router.HandleFunc("/workspaces", h.getWorkspacesByUserID).Methods(http.MethodGet)
	router.HandleFunc("/workspaces", h.createWorkspace).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/{uuid}", h.getWorkspaceByUUID).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{uuid}", h.updateWorkspace).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}", h.deleteWorkspace).Methods(http.MethodDelete)
	router.HandleFunc("/workspaces/{uuid}/reorder", h.updateWorkspaceOrder).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}/set-default", h.setDefaultWorkspace).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}/sessions", h.createSessionInWorkspace).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/{uuid}/sessions", h.getSessionsByWorkspace).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/default", h.ensureDefaultWorkspace).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/auto-migrate", h.autoMigrateLegacySessions).Methods(http.MethodPost)
}

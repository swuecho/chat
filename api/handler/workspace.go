package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/svc"
)

// ChatWorkspaceHandler handles HTTP requests for workspace management.
type ChatWorkspaceHandler struct {
	wsService *svc.ChatWorkspaceService
}

// NewChatWorkspaceHandler creates a new ChatWorkspaceHandler with all required services.
func NewChatWorkspaceHandler(wsService *svc.ChatWorkspaceService) *ChatWorkspaceHandler {
	return &ChatWorkspaceHandler{wsService: wsService}
}

// Register registers workspace routes on the given router.
func (h *ChatWorkspaceHandler) Register(router *mux.Router) {
	router.HandleFunc("/workspaces", endpoint(h.getWorkspacesByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/workspaces", endpoint(h.createWorkspace)).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/{uuid}", endpoint(h.getWorkspaceByUUID)).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{uuid}", endpoint(h.updateWorkspace)).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}", endpoint(h.deleteWorkspace)).Methods(http.MethodDelete)
	router.HandleFunc("/workspaces/{uuid}/reorder", endpoint(h.updateWorkspaceOrder)).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}/set-default", endpoint(h.setDefaultWorkspace)).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}/sessions", endpoint(h.createSessionInWorkspace)).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/{uuid}/sessions", endpoint(h.getSessionsByWorkspace)).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/default", endpoint(h.ensureDefaultWorkspace)).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/auto-migrate", endpoint(h.autoMigrateLegacySessions)).Methods(http.MethodPost)
}

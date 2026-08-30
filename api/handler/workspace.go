package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
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
func (h *ChatWorkspaceHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	apicontract.RegisterJSON(router, registry, workspaceOperation(http.MethodGet, "/workspaces",
		"listWorkspaces", "List the authenticated user's workspaces", http.StatusOK), h.getWorkspacesByUserID)
	apicontract.RegisterJSON(router, registry, workspaceOperation(http.MethodPost, "/workspaces",
		"createWorkspace", "Create a workspace", http.StatusCreated), h.createWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodGet, "/workspaces/{uuid}",
		"getWorkspace", "Get a workspace", http.StatusOK), h.getWorkspaceByUUID)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodPut, "/workspaces/{uuid}",
		"updateWorkspace", "Update a workspace", http.StatusOK), h.updateWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodDelete, "/workspaces/{uuid}",
		"deleteWorkspace", "Delete a workspace", http.StatusOK), h.deleteWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodPut, "/workspaces/{uuid}/reorder",
		"updateWorkspaceOrder", "Update a workspace order position", http.StatusOK), h.updateWorkspaceOrder)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodPut, "/workspaces/{uuid}/set-default",
		"setDefaultWorkspace", "Set the default workspace", http.StatusOK), h.setDefaultWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodPost, "/workspaces/{uuid}/sessions",
		"createWorkspaceSession", "Create a chat session in a workspace", http.StatusCreated), h.createSessionInWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperationWithUUID(http.MethodGet, "/workspaces/{uuid}/sessions",
		"listWorkspaceSessions", "List chat sessions in a workspace", http.StatusOK), h.getSessionsByWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperation(http.MethodPost, "/workspaces/default",
		"ensureDefaultWorkspace", "Ensure the default workspace exists", http.StatusOK), h.ensureDefaultWorkspace)
	apicontract.RegisterJSON(router, registry, workspaceOperation(http.MethodPost, "/workspaces/auto-migrate",
		"autoMigrateLegacySessions", "Migrate legacy sessions into the default workspace", http.StatusOK), h.autoMigrateLegacySessions)
}

func workspaceOperation(method, path, operationID, summary string, status int) apicontract.Operation {
	return apicontract.Operation{Method: method, Path: path, OperationID: operationID, Summary: summary,
		Tags: []string{"workspaces"}, SuccessStatus: status, Security: apicontract.BearerAuth()}
}

func workspaceOperationWithUUID(method, path, operationID, summary string, status int) apicontract.Operation {
	op := workspaceOperation(method, path, operationID, summary, status)
	op.Parameters = []apicontract.Parameter{apicontract.UUIDPathParameter("uuid")}
	return op
}

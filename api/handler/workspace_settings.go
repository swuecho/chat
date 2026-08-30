package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

func (h *ChatWorkspaceHandler) updateWorkspaceOrder(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req dto.UpdateWorkspaceOrderRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	workspace, err := h.wsService.UpdateWorkspaceOrder(ctx, svc.UpdateWorkspaceOrderCommand{WorkspaceUUID: workspaceUUID, UserID: userID, OrderPosition: req.OrderPosition})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update workspace order")
	}
	return respondJSON(w, http.StatusOK, workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) setDefaultWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	workspace, err := h.wsService.SetWorkspaceAsDefaultForUser(ctx, svc.SetDefaultWorkspaceCommand{UserID: userID, WorkspaceUUID: workspaceUUID})
	if err != nil {
		return dto.WrapError(err, "Failed to set default workspace")
	}
	return respondJSON(w, http.StatusOK, workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) ensureDefaultWorkspace(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	workspace, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to ensure default workspace")
	}
	return respondJSON(w, http.StatusOK, workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) autoMigrateLegacySessions(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	result, err := h.wsService.MigrateLegacyWorkspaceSessions(ctx, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to migrate legacy sessions")
	}

	response := migrationHTTPResponse{HasLegacySessions: result.HasLegacySessions, MigratedSessions: result.MigratedCount}

	if result.HasLegacySessions {
		ws := result.DefaultWorkspace
		workspace := workspaceToResponse(ws.UUID, ws.Name, ws.Description, ws.Color, ws.Icon, ws.IsDefault, ws.OrderPosition, 0, ws.CreatedAt, ws.UpdatedAt)
		response.DefaultWorkspace = &workspace
	}

	return respondJSON(w, http.StatusOK, response)
}

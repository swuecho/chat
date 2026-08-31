package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

func (h *ChatWorkspaceHandler) updateWorkspaceOrder(r *http.Request, req updateWorkspaceOrderRequest) (dto.WorkspaceResponse, error) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return dto.WorkspaceResponse{}, err
	}
	workspace, err := h.wsService.UpdateWorkspaceOrder(ctx, svc.UpdateWorkspaceOrderCommand{WorkspaceUUID: workspaceUUID, UserID: userID, OrderPosition: req.OrderPosition})
	if err != nil {
		return dto.WorkspaceResponse{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to update workspace order")
	}
	return workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt), nil
}

func (h *ChatWorkspaceHandler) setDefaultWorkspace(r *http.Request, _ apicontract.NoBody) (dto.WorkspaceResponse, error) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return dto.WorkspaceResponse{}, err
	}

	workspace, err := h.wsService.SetWorkspaceAsDefaultForUser(ctx, svc.SetDefaultWorkspaceCommand{UserID: userID, WorkspaceUUID: workspaceUUID})
	if err != nil {
		return dto.WorkspaceResponse{}, dto.WrapError(err, "Failed to set default workspace")
	}
	return workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt), nil
}

func (h *ChatWorkspaceHandler) ensureDefaultWorkspace(r *http.Request, _ apicontract.NoBody) (dto.WorkspaceResponse, error) {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return dto.WorkspaceResponse{}, err
	}

	workspace, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		return dto.WorkspaceResponse{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to ensure default workspace")
	}
	return workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt), nil
}

func (h *ChatWorkspaceHandler) autoMigrateLegacySessions(r *http.Request, _ apicontract.NoBody) (migrationHTTPResponse, error) {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return migrationHTTPResponse{}, err
	}

	result, err := h.wsService.MigrateLegacyWorkspaceSessions(ctx, userID)
	if err != nil {
		return migrationHTTPResponse{}, dto.WrapError(dto.MapDatabaseError(err), "Failed to migrate legacy sessions")
	}

	response := migrationHTTPResponse{HasLegacySessions: result.HasLegacySessions, MigratedSessions: result.MigratedCount}

	if result.HasLegacySessions {
		ws := result.DefaultWorkspace
		workspace := workspaceToResponse(ws.UUID, ws.Name, ws.Description, ws.Color, ws.Icon, ws.IsDefault, ws.OrderPosition, 0, ws.CreatedAt, ws.UpdatedAt)
		response.DefaultWorkspace = &workspace
	}

	return response, nil
}

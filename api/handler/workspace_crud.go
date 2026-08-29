package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

func (h *ChatWorkspaceHandler) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	if req.Name == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Workspace name is required"))
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	workspace, err := h.wsService.CreateWorkspace(ctx, svc.CreateWorkspaceInput{
		Uuid: uuid.New().String(), UserID: userID,
		Name: req.Name, Description: req.Description,
		Color: req.Color, Icon: req.Icon,
		IsDefault: false, OrderPosition: 0,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create workspace"))
		return
	}

	if req.IsDefault {
		workspace, err = h.wsService.SetWorkspaceAsDefaultForUser(ctx, svc.SetDefaultWorkspaceCommand{UserID: userID, WorkspaceUUID: workspace.UUID})
		if err != nil {
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to set default workspace"))
			return
		}
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) getWorkspaceByUUID(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]
	slog.Info("getWorkspaceByUUID called", "uuid", workspaceUUID)

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if !h.checkPermission(w, ctx, workspaceUUID, userID) {
		return
	}

	workspace, err := h.wsService.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := dto.ErrResourceNotFound("Workspace")
			apiErr.Message = "Workspace not found with UUID: " + workspaceUUID
			dto.RespondWithAPIError(w, apiErr)
			return
		}
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) getWorkspacesByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	workspaces, err := h.wsService.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspaces"))
		return
	}

	responses := make([]dto.WorkspaceResponse, 0, len(workspaces))
	for _, ws := range workspaces {
		responses = append(responses, workspaceToResponse(ws.UUID, ws.Name, ws.Description, ws.Color, ws.Icon, ws.IsDefault, ws.OrderPosition, ws.SessionCount, ws.CreatedAt, ws.UpdatedAt))
	}
	json.NewEncoder(w).Encode(responses)
}

func (h *ChatWorkspaceHandler) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req dto.UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	workspace, err := h.wsService.UpdateWorkspace(ctx, svc.UpdateWorkspaceInput{
		Uuid: workspaceUUID, UserID: userID, Name: req.Name, Description: req.Description,
		Color: req.Color, Icon: req.Icon,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	if err := h.wsService.DeleteWorkspace(ctx, svc.DeleteWorkspaceCommand{WorkspaceUUID: workspaceUUID, UserID: userID}); err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete workspace"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Workspace deleted successfully"})
}

// --- Helpers ---

func (h *ChatWorkspaceHandler) checkPermission(w http.ResponseWriter, ctx context.Context, workspaceUUID string, userID int32) bool {
	ok, err := h.wsService.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to check workspace permission"))
		return false
	}
	if !ok {
		apiErr := dto.ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		dto.RespondWithAPIError(w, apiErr)
		return false
	}
	return true
}

func workspaceToResponse(uuid, name, description, color, icon string, isDefault bool, orderPosition int32, sessionCount int64, createdAt, updatedAt time.Time) dto.WorkspaceResponse {
	return dto.WorkspaceResponse{
		Uuid: uuid, Name: name, Description: description,
		Color: color, Icon: icon, IsDefault: isDefault,
		OrderPosition: orderPosition, SessionCount: sessionCount,
		CreatedAt: createdAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: updatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

func (h *ChatWorkspaceHandler) createWorkspace(w http.ResponseWriter, r *http.Request) error {
	var req dto.CreateWorkspaceRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}
	if req.Name == "" {
		return dto.ErrValidationInvalidInput("Workspace name is required")
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	workspace, err := h.wsService.CreateWorkspace(ctx, svc.CreateWorkspaceInput{
		Uuid: uuid.New().String(), UserID: userID,
		Name: req.Name, Description: req.Description,
		Color: req.Color, Icon: req.Icon,
		IsDefault: false, OrderPosition: 0,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to create workspace")
	}

	if req.IsDefault {
		workspace, err = h.wsService.SetWorkspaceAsDefaultForUser(ctx, svc.SetDefaultWorkspaceCommand{UserID: userID, WorkspaceUUID: workspace.UUID})
		if err != nil {
			return dto.WrapError(dto.MapDatabaseError(err), "Failed to set default workspace")
		}
	}

	return respondJSON(w, http.StatusCreated, workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) getWorkspaceByUUID(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	if err := h.checkPermission(ctx, workspaceUUID, userID); err != nil {
		return err
	}

	workspace, err := h.wsService.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := dto.ErrResourceNotFound("Workspace")
			apiErr.Message = "Workspace not found with UUID: " + workspaceUUID
			return apiErr
		}
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspace")
	}
	return respondJSON(w, http.StatusOK, workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) getWorkspacesByUserID(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	workspaces, err := h.wsService.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspaces")
	}

	responses := make([]dto.WorkspaceResponse, 0, len(workspaces))
	for _, ws := range workspaces {
		responses = append(responses, workspaceToResponse(ws.UUID, ws.Name, ws.Description, ws.Color, ws.Icon, ws.IsDefault, ws.OrderPosition, ws.SessionCount, ws.CreatedAt, ws.UpdatedAt))
	}
	return respondJSON(w, http.StatusOK, responses)
}

func (h *ChatWorkspaceHandler) updateWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req dto.UpdateWorkspaceRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	workspace, err := h.wsService.UpdateWorkspace(ctx, svc.UpdateWorkspaceInput{
		Uuid: workspaceUUID, UserID: userID, Name: req.Name, Description: req.Description,
		Color: req.Color, Icon: req.Icon,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update workspace")
	}
	return respondJSON(w, http.StatusOK, workspaceToResponse(workspace.UUID, workspace.Name, workspace.Description, workspace.Color, workspace.Icon, workspace.IsDefault, workspace.OrderPosition, 0, workspace.CreatedAt, workspace.UpdatedAt))
}

func (h *ChatWorkspaceHandler) deleteWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	if err := h.wsService.DeleteWorkspace(ctx, svc.DeleteWorkspaceCommand{WorkspaceUUID: workspaceUUID, UserID: userID}); err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to delete workspace")
	}
	return respondJSON(w, http.StatusOK, messageHTTPStatusResponse{Message: "Workspace deleted successfully"})
}

// --- Helpers ---

func (h *ChatWorkspaceHandler) checkPermission(ctx context.Context, workspaceUUID string, userID int32) error {
	ok, err := h.wsService.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to check workspace permission")
	}
	if !ok {
		apiErr := dto.ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		return apiErr
	}
	return nil
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

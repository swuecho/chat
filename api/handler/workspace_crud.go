package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
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

	workspace, err := h.wsService.CreateWorkspace(ctx, sqlc_queries.CreateWorkspaceParams{
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
		workspace, err = h.wsService.SetWorkspaceAsDefaultForUser(ctx, userID, workspace.Uuid)
		if err != nil {
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to set default workspace"))
			return
		}
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

func (h *ChatWorkspaceHandler) getWorkspaceByUUID(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]
	log.Printf("getWorkspaceByUUID called with UUID=%s", workspaceUUID)

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

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
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
		responses = append(responses, workspaceRowToResponse(ws))
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
	if !h.checkPermission(w, ctx, workspaceUUID, userID) {
		return
	}

	workspace, err := h.wsService.UpdateWorkspace(ctx, sqlc_queries.UpdateWorkspaceParams{
		Uuid: workspaceUUID, Name: req.Name, Description: req.Description,
		Color: req.Color, Icon: req.Icon,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

func (h *ChatWorkspaceHandler) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

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
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspace"))
		return
	}
	if workspace.IsDefault {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Cannot delete default workspace"))
		return
	}

	if err := h.wsService.DeleteWorkspace(ctx, workspaceUUID); err != nil {
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

func workspaceToResponse(ws sqlc_queries.ChatWorkspace) dto.WorkspaceResponse {
	return dto.WorkspaceResponse{
		Uuid: ws.Uuid, Name: ws.Name, Description: ws.Description,
		Color: ws.Color, Icon: ws.Icon,
		IsDefault: ws.IsDefault, OrderPosition: ws.OrderPosition,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func workspaceRowToResponse(ws sqlc_queries.GetWorkspaceWithSessionCountRow) dto.WorkspaceResponse {
	return dto.WorkspaceResponse{
		Uuid: ws.Uuid, Name: ws.Name, Description: ws.Description,
		Color: ws.Color, Icon: ws.Icon,
		IsDefault: ws.IsDefault, OrderPosition: ws.OrderPosition, SessionCount: ws.SessionCount,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

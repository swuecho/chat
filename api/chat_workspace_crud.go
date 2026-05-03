// Package main — Workspace CRUD handlers: create, read, update, delete.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// createWorkspace creates a new workspace.
func (h *ChatWorkspaceHandler) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	if req.Name == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Workspace name is required"))
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
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	workspace, err := h.wsService.CreateWorkspace(ctx, sqlc_queries.CreateWorkspaceParams{
		Uuid: uuid.New().String(), UserID: userID,
		Name: req.Name, Description: req.Description,
		Color: req.Color, Icon: req.Icon,
		IsDefault: false, OrderPosition: 0,
	})
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to create workspace"))
		return
	}

	if req.IsDefault {
		workspace, err = h.wsService.SetWorkspaceAsDefaultForUser(ctx, userID, workspace.Uuid)
		if err != nil {
			RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to set default workspace"))
			return
		}
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

// getWorkspaceByUUID returns a workspace by its UUID.
func (h *ChatWorkspaceHandler) getWorkspaceByUUID(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]
	log.Printf("getWorkspaceByUUID called with UUID=%s", workspaceUUID)

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if !h.checkPermission(w, ctx, workspaceUUID, userID) {
		return
	}

	workspace, err := h.wsService.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Workspace")
			apiErr.Message = "Workspace not found with UUID: " + workspaceUUID
			RespondWithAPIError(w, apiErr)
			return
		}
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

// getWorkspacesByUserID returns all workspaces for the authenticated user.
func (h *ChatWorkspaceHandler) getWorkspacesByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	workspaces, err := h.wsService.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get workspaces"))
		return
	}

	responses := make([]WorkspaceResponse, 0, len(workspaces))
	for _, ws := range workspaces {
		responses = append(responses, workspaceRowToResponse(ws))
	}
	json.NewEncoder(w).Encode(responses)
}

// updateWorkspace updates an existing workspace.
func (h *ChatWorkspaceHandler) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req UpdateWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
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
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

// deleteWorkspace deletes a workspace.
func (h *ChatWorkspaceHandler) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	if !h.checkPermission(w, ctx, workspaceUUID, userID) {
		return
	}

	workspace, err := h.wsService.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get workspace"))
		return
	}
	if workspace.IsDefault {
		RespondWithAPIError(w, ErrValidationInvalidInput("Cannot delete default workspace"))
		return
	}

	if err := h.wsService.DeleteWorkspace(ctx, workspaceUUID); err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to delete workspace"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Workspace deleted successfully"})
}

// --- Helpers ---

// checkPermission verifies the user can access the workspace.
func (h *ChatWorkspaceHandler) checkPermission(w http.ResponseWriter, ctx context.Context, workspaceUUID string, userID int32) bool {
	ok, err := h.wsService.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to check workspace permission"))
		return false
	}
	if !ok {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return false
	}
	return true
}

// workspaceToResponse converts a ChatWorkspace to a WorkspaceResponse.
func workspaceToResponse(ws sqlc_queries.ChatWorkspace) WorkspaceResponse {
	return WorkspaceResponse{
		Uuid: ws.Uuid, Name: ws.Name, Description: ws.Description,
		Color: ws.Color, Icon: ws.Icon,
		IsDefault: ws.IsDefault, OrderPosition: ws.OrderPosition,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// workspaceRowToResponse converts a GetWorkspaceWithSessionCountRow to a WorkspaceResponse.
func workspaceRowToResponse(ws sqlc_queries.GetWorkspaceWithSessionCountRow) WorkspaceResponse {
	return WorkspaceResponse{
		Uuid: ws.Uuid, Name: ws.Name, Description: ws.Description,
		Color: ws.Color, Icon: ws.Icon,
		IsDefault: ws.IsDefault, OrderPosition: ws.OrderPosition, SessionCount: ws.SessionCount,
		CreatedAt: ws.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: ws.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

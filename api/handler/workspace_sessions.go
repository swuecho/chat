package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

func (h *ChatWorkspaceHandler) createSessionInWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]
	if !validateUUIDParam(w, "uuid", workspaceUUID) {
		return
	}

	var req dto.CreateSessionInWorkspaceRequest
	if err := DecodeJSON(r, &req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	result, err := h.wsService.CreateWorkspaceSession(ctx, svc.CreateWorkspaceSessionCommand{
		UserID: userID, WorkspaceUUID: workspaceUUID, Topic: req.Topic,
		Model: req.Model, DefaultSystemPrompt: req.DefaultSystemPrompt,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create session in workspace"))
		return
	}
	session := result.Session

	json.NewEncoder(w).Encode(workspaceSessionCreatedHTTPResponse{UUID: session.UUID, Topic: session.Topic,
		Model: session.Model, ArtifactEnabled: session.ArtifactEnabled, WorkspaceUUID: result.WorkspaceUUID,
		CreatedAt: session.CreatedAt.Format("2006-01-02T15:04:05Z")})
}

func (h *ChatWorkspaceHandler) getSessionsByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]
	if !validateUUIDParam(w, "uuid", workspaceUUID) {
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

	workspace, err := h.wsService.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspace"))
		return
	}

	sessions, err := h.wsService.GetSessionsByWorkspaceID(ctx, workspace.ID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get sessions"))
		return
	}

	responses := make([]workspaceSessionHTTPResponse, 0, len(sessions))
	for _, s := range sessions {
		responses = append(responses, workspaceSessionHTTPResponse{UUID: s.UUID, Title: s.Topic, Model: s.Model,
			WorkspaceUUID: workspaceUUID, MaxLength: s.MaxLength, Temperature: s.Temperature,
			MaxTokens: s.MaxTokens, TopP: s.TopP, N: s.N, Debug: s.Debug, SummarizeMode: s.SummarizeMode,
			ExploreMode: s.ExploreMode, ArtifactEnabled: s.ArtifactEnabled,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"), UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z")})
	}

	json.NewEncoder(w).Encode(responses)
}

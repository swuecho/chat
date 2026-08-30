package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

func (h *ChatWorkspaceHandler) createSessionInWorkspace(w http.ResponseWriter, r *http.Request) error {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req dto.CreateSessionInWorkspaceRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}

	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	result, err := h.wsService.CreateWorkspaceSession(ctx, svc.CreateWorkspaceSessionCommand{
		UserID: userID, WorkspaceUUID: workspaceUUID, Topic: req.Topic,
		Model: req.Model, DefaultSystemPrompt: req.DefaultSystemPrompt,
	})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to create session in workspace")
	}
	session := result.Session

	return respondJSON(w, http.StatusCreated, workspaceSessionCreatedHTTPResponse{UUID: session.UUID, Topic: session.Topic,
		Model: session.Model, ArtifactEnabled: session.ArtifactEnabled, WorkspaceUUID: result.WorkspaceUUID,
		CreatedAt: session.CreatedAt.Format("2006-01-02T15:04:05Z")})
}

func (h *ChatWorkspaceHandler) getSessionsByWorkspace(w http.ResponseWriter, r *http.Request) error {
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
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get workspace")
	}

	sessions, err := h.wsService.GetSessionsByWorkspaceID(ctx, workspace.ID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get sessions")
	}

	responses := make([]workspaceSessionHTTPResponse, 0, len(sessions))
	for _, s := range sessions {
		responses = append(responses, workspaceSessionHTTPResponse{UUID: s.UUID, Title: s.Topic, Model: s.Model,
			WorkspaceUUID: workspaceUUID, MaxLength: s.MaxLength, Temperature: s.Temperature,
			MaxTokens: s.MaxTokens, TopP: s.TopP, N: s.N, Debug: s.Debug, SummarizeMode: s.SummarizeMode,
			ExploreMode: s.ExploreMode, ArtifactEnabled: s.ArtifactEnabled,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"), UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z")})
	}

	return respondJSON(w, http.StatusOK, responses)
}

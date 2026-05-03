// Package main — Workspace session handlers.
package main

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// createSessionInWorkspace creates a new chat session inside a specific workspace.
func (h *ChatWorkspaceHandler) createSessionInWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req CreateSessionInWorkspaceRequest
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

	workspace, err := h.wsService.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get workspace"))
		return
	}

	// Delegate to service for session creation
	session, err := h.wsService.CreateSessionInWorkspace(ctx, userID, workspace.ID, req.Topic, req.Model, req.DefaultSystemPrompt)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to create session in workspace"))
		return
	}

	// Ensure system prompt via session service
	if _, err := h.sessionService.EnsureDefaultSystemPrompt(ctx, session.Uuid, userID, req.DefaultSystemPrompt); err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to create default system prompt"))
		return
	}

	// Set as active session
	if _, err := h.activeSession.UpsertActiveSession(ctx, userID, &workspace.ID, session.Uuid); err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to set active session"))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"uuid":            session.Uuid,
		"topic":           session.Topic,
		"model":           session.Model,
		"artifactEnabled": session.ArtifactEnabled,
		"workspaceUuid":   workspaceUUID,
		"createdAt":       session.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// getSessionsByWorkspace returns all sessions in a specific workspace.
func (h *ChatWorkspaceHandler) getSessionsByWorkspace(w http.ResponseWriter, r *http.Request) {
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

	sessions, err := h.wsService.GetSessionsByWorkspaceID(ctx, workspace.ID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get sessions"))
		return
	}

	responses := make([]map[string]interface{}, 0, len(sessions))
	for _, s := range sessions {
		responses = append(responses, sessionToMap(s, workspaceUUID))
	}

	json.NewEncoder(w).Encode(responses)
}

// sessionToMap converts a ChatSession to a response map.
func sessionToMap(s sqlc_queries.ChatSession, workspaceUUID string) map[string]interface{} {
	return map[string]interface{}{
		"uuid":            s.Uuid,
		"title":           s.Topic,
		"isEdit":          false,
		"model":           s.Model,
		"workspaceUuid":   workspaceUUID,
		"maxLength":       s.MaxLength,
		"temperature":     s.Temperature,
		"maxTokens":       s.MaxTokens,
		"topP":            s.TopP,
		"n":               s.N,
		"debug":           s.Debug,
		"summarizeMode":   s.SummarizeMode,
		"exploreMode":     s.ExploreMode,
		"artifactEnabled": s.ArtifactEnabled,
		"createdAt":       s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updatedAt":       s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

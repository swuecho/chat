package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func (h *ChatWorkspaceHandler) updateWorkspaceOrder(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req dto.UpdateWorkspaceOrderRequest
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

	workspace, err := h.wsService.UpdateWorkspaceOrder(ctx, sqlc_queries.UpdateWorkspaceOrderParams{
		Uuid: workspaceUUID, OrderPosition: req.OrderPosition,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update workspace order"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

func (h *ChatWorkspaceHandler) setDefaultWorkspace(w http.ResponseWriter, r *http.Request) {
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

	workspace, err := h.wsService.SetWorkspaceAsDefaultForUser(ctx, userID, workspaceUUID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to set default workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

func (h *ChatWorkspaceHandler) ensureDefaultWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	workspace, err := h.wsService.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to ensure default workspace"))
		return
	}

	json.NewEncoder(w).Encode(workspaceToResponse(workspace))
}

func (h *ChatWorkspaceHandler) autoMigrateLegacySessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	result, err := h.wsService.AutoMigrateLegacySessions(ctx, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to migrate legacy sessions"))
		return
	}

	response := map[string]interface{}{
		"hasLegacySessions": result.HasLegacySessions,
		"migratedSessions":  result.MigratedCount,
	}

	if result.HasLegacySessions {
		if err := h.wsService.MigrateLegacyActiveSessions(ctx, userID, result.DefaultWorkspace.ID); err != nil {
			slog.Warn("failed to migrate legacy active sessions", "error", err)
		}
		response["defaultWorkspace"] = workspaceToResponse(result.DefaultWorkspace)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

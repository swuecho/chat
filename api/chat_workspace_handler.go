package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatWorkspaceHandler struct {
	service *ChatWorkspaceService
}

func NewChatWorkspaceHandler(sqlc_q *sqlc_queries.Queries) *ChatWorkspaceHandler {
	workspaceService := NewChatWorkspaceService(sqlc_q)
	return &ChatWorkspaceHandler{
		service: workspaceService,
	}
}

func (h *ChatWorkspaceHandler) Register(router *mux.Router) {
	router.HandleFunc("/workspaces", h.getWorkspacesByUserID).Methods(http.MethodGet)
	router.HandleFunc("/workspaces", h.createWorkspace).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/{uuid}", h.getWorkspaceByUUID).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/{uuid}", h.updateWorkspace).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}", h.deleteWorkspace).Methods(http.MethodDelete)
	router.HandleFunc("/workspaces/{uuid}/reorder", h.updateWorkspaceOrder).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}/set-default", h.setDefaultWorkspace).Methods(http.MethodPut)
	router.HandleFunc("/workspaces/{uuid}/sessions", h.createSessionInWorkspace).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/{uuid}/sessions", h.getSessionsByWorkspace).Methods(http.MethodGet)
	router.HandleFunc("/workspaces/default", h.ensureDefaultWorkspace).Methods(http.MethodPost)
	router.HandleFunc("/workspaces/auto-migrate", h.autoMigrateLegacySessions).Methods(http.MethodPost)
}


type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	IsDefault   bool   `json:"isDefault"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

type UpdateWorkspaceOrderRequest struct {
	OrderPosition int32 `json:"orderPosition"`
}

type WorkspaceResponse struct {
	Uuid          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	Icon          string `json:"icon"`
	IsDefault     bool   `json:"isDefault"`
	OrderPosition int32  `json:"orderPosition"`
	SessionCount  int64  `json:"sessionCount,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

// createWorkspace creates a new workspace
func (h *ChatWorkspaceHandler) createWorkspace(w http.ResponseWriter, r *http.Request) {
	var req CreateWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Validate required fields
	if req.Name == "" {
		apiErr := ErrValidationInvalidInput("Workspace name is required")
		RespondWithAPIError(w, apiErr)
		return
	}

	// Default values
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	workspaceUUID := uuid.New().String()
	params := sqlc_queries.CreateWorkspaceParams{
		Uuid:          workspaceUUID,
		UserID:        userID,
		Name:          req.Name,
		Description:   req.Description,
		Color:         req.Color,
		Icon:          req.Icon,
		IsDefault:     req.IsDefault,
		OrderPosition: 0, // Will be updated if needed
	}

	workspace, err := h.service.CreateWorkspace(ctx, params)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to create workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	response := WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	json.NewEncoder(w).Encode(response)
}

// getWorkspaceByUUID returns a workspace by its UUID
func (h *ChatWorkspaceHandler) getWorkspaceByUUID(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]
	log.Printf("🔍 DEBUG: getWorkspaceByUUID called with UUID=%s", workspaceUUID)

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	log.Printf("🔍 DEBUG: getWorkspaceByUUID userID=%d", userID)

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Workspace")
			apiErr.Message = "Workspace not found with UUID: " + workspaceUUID
			RespondWithAPIError(w, apiErr)
			return
		}
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	response := WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	json.NewEncoder(w).Encode(response)
}

// getWorkspacesByUserID returns all workspaces for the authenticated user
func (h *ChatWorkspaceHandler) getWorkspacesByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	workspaces, err := h.service.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspaces")
		RespondWithAPIError(w, apiErr)
		return
	}

	responses := make([]WorkspaceResponse, 0)
	for _, workspace := range workspaces {
		response := WorkspaceResponse{
			Uuid:          workspace.Uuid,
			Name:          workspace.Name,
			Description:   workspace.Description,
			Color:         workspace.Color,
			Icon:          workspace.Icon,
			IsDefault:     workspace.IsDefault,
			OrderPosition: workspace.OrderPosition,
			SessionCount:  workspace.SessionCount,
			CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		responses = append(responses, response)
	}

	json.NewEncoder(w).Encode(responses)
}

// updateWorkspace updates an existing workspace
func (h *ChatWorkspaceHandler) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req UpdateWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	params := sqlc_queries.UpdateWorkspaceParams{
		Uuid:        workspaceUUID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	workspace, err := h.service.UpdateWorkspace(ctx, params)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to update workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	response := WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	json.NewEncoder(w).Encode(response)
}

// updateWorkspaceOrder updates the order position of a workspace
func (h *ChatWorkspaceHandler) updateWorkspaceOrder(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req UpdateWorkspaceOrderRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	params := sqlc_queries.UpdateWorkspaceOrderParams{
		Uuid:          workspaceUUID,
		OrderPosition: req.OrderPosition,
	}

	workspace, err := h.service.UpdateWorkspaceOrder(ctx, params)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to update workspace order")
		RespondWithAPIError(w, apiErr)
		return
	}

	response := WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	json.NewEncoder(w).Encode(response)
}

// deleteWorkspace deletes a workspace
func (h *ChatWorkspaceHandler) deleteWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	// Get workspace to check if it's default
	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	// Prevent deletion of default workspace
	if workspace.IsDefault {
		apiErr := ErrValidationInvalidInput("Cannot delete default workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	err = h.service.DeleteWorkspace(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to delete workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Workspace deleted successfully"})
}

// setDefaultWorkspace sets a workspace as the default
func (h *ChatWorkspaceHandler) setDefaultWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	// First, unset all default workspaces for this user
	workspaces, err := h.service.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspaces")
		RespondWithAPIError(w, apiErr)
		return
	}

	for _, ws := range workspaces {
		if ws.IsDefault {
			_, err = h.service.SetDefaultWorkspace(ctx, sqlc_queries.SetDefaultWorkspaceParams{
				Uuid:      ws.Uuid,
				IsDefault: false,
			})
			if err != nil {
				apiErr := WrapError(MapDatabaseError(err), "Failed to unset default workspace")
				RespondWithAPIError(w, apiErr)
				return
			}
		}
	}

	// Set the new default workspace
	workspace, err := h.service.SetDefaultWorkspace(ctx, sqlc_queries.SetDefaultWorkspaceParams{
		Uuid:      workspaceUUID,
		IsDefault: true,
	})
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to set default workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	response := WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	json.NewEncoder(w).Encode(response)
}

// ensureDefaultWorkspace ensures the user has a default workspace
func (h *ChatWorkspaceHandler) ensureDefaultWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	workspace, err := h.service.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to ensure default workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	response := WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	json.NewEncoder(w).Encode(response)
}

type CreateSessionInWorkspaceRequest struct {
	Topic string `json:"topic"`
	Model string `json:"model"`
}

// createSessionInWorkspace creates a new session in a specific workspace
func (h *ChatWorkspaceHandler) createSessionInWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	var req CreateSessionInWorkspaceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Check workspace permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	// Get workspace
	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	// Default model if not provided
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}

	// Create session
	sessionUUID := uuid.New().String()
	sessionService := NewChatSessionService(h.service.q)
	activeSessionService := NewUserActiveChatSessionService(h.service.q)
	
	sessionParams := sqlc_queries.CreateChatSessionInWorkspaceParams{
		UserID:      userID,
		Uuid:        sessionUUID,
		Topic:       req.Topic,
		CreatedAt:   time.Now(),
		Active:      true,
		MaxLength:   10,
		Model:       req.Model,
		WorkspaceID: sql.NullInt32{Int32: workspace.ID, Valid: true},
	}

	session, err := sessionService.q.CreateChatSessionInWorkspace(ctx, sessionParams)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to create session in workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	// Set as active session (use unified approach)
	_, err = activeSessionService.UpsertActiveSession(ctx, userID, &workspace.ID, sessionUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to set active session")
		RespondWithAPIError(w, apiErr)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"uuid":         session.Uuid,
		"topic":        session.Topic,
		"model":        session.Model,
		"workspaceUuid": workspaceUUID,
		"createdAt":    session.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// getSessionsByWorkspace returns all sessions in a specific workspace
func (h *ChatWorkspaceHandler) getSessionsByWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceUUID := mux.Vars(r)["uuid"]

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Check workspace permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIError(w, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIError(w, apiErr)
		return
	}

	// Get workspace
	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIError(w, apiErr)
		return
	}

	// Get sessions in workspace
	sessionService := NewChatSessionService(h.service.q)
	sessions, err := sessionService.q.GetSessionsByWorkspaceID(ctx, sql.NullInt32{Int32: workspace.ID, Valid: true})
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get sessions")
		RespondWithAPIError(w, apiErr)
		return
	}

	sessionResponses := make([]map[string]interface{}, 0)
	for _, session := range sessions {
		sessionResponse := map[string]interface{}{
			"uuid":         session.Uuid,
			"title":        session.Topic,  // Use "title" to match the original API
			"isEdit":       false,
			"model":        session.Model,
			"workspaceUuid": workspaceUUID,
			"createdAt":    session.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":    session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		sessionResponses = append(sessionResponses, sessionResponse)
	}

	json.NewEncoder(w).Encode(sessionResponses)
}


// autoMigrateLegacySessions automatically detects and migrates legacy sessions without workspace_id
func (h *ChatWorkspaceHandler) autoMigrateLegacySessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	// Check if user has any legacy sessions (sessions without workspace_id)
	sessionService := NewChatSessionService(h.service.q)
	legacySessions, err := sessionService.q.GetSessionsWithoutWorkspace(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to check for legacy sessions"))
		return
	}

	response := map[string]interface{}{
		"hasLegacySessions": len(legacySessions) > 0,
		"migratedSessions": 0,
	}

	// If no legacy sessions, return early
	if len(legacySessions) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Ensure default workspace exists
	defaultWorkspace, err := h.service.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to ensure default workspace"))
		return
	}

	// Migrate all legacy sessions to default workspace
	err = h.service.MigrateSessionsToDefaultWorkspace(ctx, userID, defaultWorkspace.ID)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to migrate legacy sessions"))
		return
	}

	// Also migrate any legacy active sessions
	activeSessionService := NewUserActiveChatSessionService(h.service.q)
	legacyActiveSessions, err := activeSessionService.q.GetAllUserActiveSessions(ctx, userID)
	if err == nil {
		for _, activeSession := range legacyActiveSessions {
			if !activeSession.WorkspaceID.Valid {
				// This is a legacy global active session, migrate it to default workspace
				_, _ = activeSessionService.UpsertActiveSession(ctx, userID, &defaultWorkspace.ID, activeSession.ChatSessionUuid)
				// Delete the old global active session by setting workspace to NULL and deleting
				_ = activeSessionService.DeleteActiveSession(ctx, userID, nil)
			}
		}
	}

	response["migratedSessions"] = len(legacySessions)
	response["defaultWorkspace"] = WorkspaceResponse{
		Uuid:          defaultWorkspace.Uuid,
		Name:          defaultWorkspace.Name,
		Description:   defaultWorkspace.Description,
		Color:         defaultWorkspace.Color,
		Icon:          defaultWorkspace.Icon,
		IsDefault:     defaultWorkspace.IsDefault,
		OrderPosition: defaultWorkspace.OrderPosition,
		CreatedAt:     defaultWorkspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     defaultWorkspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

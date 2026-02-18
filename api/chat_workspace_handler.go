package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *ChatWorkspaceHandler) Register(router *gin.RouterGroup) {
	router.GET("/workspaces", h.getWorkspacesByUserID)
	router.POST("/workspaces", h.createWorkspace)
	router.GET("/workspaces/:uuid", h.getWorkspaceByUUID)
	router.PUT("/workspaces/:uuid", h.updateWorkspace)
	router.DELETE("/workspaces/:uuid", h.deleteWorkspace)
	router.PUT("/workspaces/:uuid/reorder", h.updateWorkspaceOrder)
	router.PUT("/workspaces/:uuid/set-default", h.setDefaultWorkspace)
	router.POST("/workspaces/:uuid/sessions", h.createSessionInWorkspace)
	router.GET("/workspaces/:uuid/sessions", h.getSessionsByWorkspace)
	router.POST("/workspaces/default", h.ensureDefaultWorkspace)
	router.POST("/workspaces/auto-migrate", h.autoMigrateLegacySessions)
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
func (h *ChatWorkspaceHandler) createWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Validate required fields
	if req.Name == "" {
		apiErr := ErrValidationInvalidInput("Workspace name is required")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Default values
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
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
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, response)
}

// getWorkspaceByUUID returns a workspace by its UUID
func (h *ChatWorkspaceHandler) getWorkspaceByUUID(c *gin.Context) {
	workspaceUUID := c.Param("uuid")
	log.Printf("🔍 DEBUG: getWorkspaceByUUID called with UUID=%s", workspaceUUID)

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	log.Printf("🔍 DEBUG: getWorkspaceByUUID userID=%d", userID)

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Workspace")
			apiErr.Message = "Workspace not found with UUID: " + workspaceUUID
			RespondWithAPIErrorGin(c, apiErr)
			return
		}
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, response)
}

// getWorkspacesByUserID returns all workspaces for the authenticated user
func (h *ChatWorkspaceHandler) getWorkspacesByUserID(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	workspaces, err := h.service.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspaces")
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, responses)
}

// updateWorkspace updates an existing workspace
func (h *ChatWorkspaceHandler) updateWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	var req UpdateWorkspaceRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
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
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, response)
}

// updateWorkspaceOrder updates the order position of a workspace
func (h *ChatWorkspaceHandler) updateWorkspaceOrder(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	var req UpdateWorkspaceOrderRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	params := sqlc_queries.UpdateWorkspaceOrderParams{
		Uuid:          workspaceUUID,
		OrderPosition: req.OrderPosition,
	}

	workspace, err := h.service.UpdateWorkspaceOrder(ctx, params)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to update workspace order")
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, response)
}

// deleteWorkspace deletes a workspace
func (h *ChatWorkspaceHandler) deleteWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Get workspace to check if it's default
	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Prevent deletion of default workspace
	if workspace.IsDefault {
		apiErr := ErrValidationInvalidInput("Cannot delete default workspace")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	err = h.service.DeleteWorkspace(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to delete workspace")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "Workspace deleted successfully"})
}

// setDefaultWorkspace sets a workspace as the default
func (h *ChatWorkspaceHandler) setDefaultWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Check permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// First, unset all default workspaces for this user
	workspaces, err := h.service.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspaces")
		RespondWithAPIErrorGin(c, apiErr)
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
				RespondWithAPIErrorGin(c, apiErr)
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
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, response)
}

// ensureDefaultWorkspace ensures the user has a default workspace
func (h *ChatWorkspaceHandler) ensureDefaultWorkspace(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	workspace, err := h.service.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to ensure default workspace")
		RespondWithAPIErrorGin(c, apiErr)
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

	c.JSON(200, response)
}

type CreateSessionInWorkspaceRequest struct {
	Topic string `json:"topic"`
	Model string `json:"model"`
}

// createSessionInWorkspace creates a new session in a specific workspace
func (h *ChatWorkspaceHandler) createSessionInWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	var req CreateSessionInWorkspaceRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Check workspace permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Get workspace
	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIErrorGin(c, apiErr)
		return
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
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Set as active session (use unified approach)
	_, err = activeSessionService.UpsertActiveSession(ctx, userID, &workspace.ID, sessionUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to set active session")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	c.JSON(200, map[string]interface{}{
		"uuid":          session.Uuid,
		"topic":         session.Topic,
		"model":         session.Model,
		"codeRunnerEnabled": session.CodeRunnerEnabled,
		"artifactEnabled": session.ArtifactEnabled,
		"workspaceUuid": workspaceUUID,
		"createdAt":     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// getSessionsByWorkspace returns all sessions in a specific workspace
func (h *ChatWorkspaceHandler) getSessionsByWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Check workspace permission
	hasPermission, err := h.service.HasWorkspacePermission(ctx, workspaceUUID, userID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to check workspace permission")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Get workspace
	workspace, err := h.service.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get workspace")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Get sessions in workspace
	sessionService := NewChatSessionService(h.service.q)
	sessions, err := sessionService.q.GetSessionsByWorkspaceID(ctx, sql.NullInt32{Int32: workspace.ID, Valid: true})
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to get sessions")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	sessionResponses := make([]map[string]interface{}, 0)
	for _, session := range sessions {
		sessionResponse := map[string]interface{}{
			"uuid":          session.Uuid,
			"title":         session.Topic, // Use "title" to match the original API
			"isEdit":        false,
			"model":         session.Model,
			"workspaceUuid": workspaceUUID,
			"maxLength":     session.MaxLength,
			"temperature":   session.Temperature,
			"maxTokens":     session.MaxTokens,
			"topP":          session.TopP,
			"n":             session.N,
			"debug":         session.Debug,
			"summarizeMode": session.SummarizeMode,
			"exploreMode":   session.ExploreMode,
			"codeRunnerEnabled": session.CodeRunnerEnabled,
			"artifactEnabled": session.ArtifactEnabled,
			"createdAt":     session.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":     session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		sessionResponses = append(sessionResponses, sessionResponse)
	}

	c.JSON(200, sessionResponses)
}

// autoMigrateLegacySessions automatically detects and migrates legacy sessions without workspace_id
func (h *ChatWorkspaceHandler) autoMigrateLegacySessions(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	// Check if user has any legacy sessions (sessions without workspace_id)
	sessionService := NewChatSessionService(h.service.q)
	legacySessions, err := sessionService.q.GetSessionsWithoutWorkspace(ctx, userID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to check for legacy sessions"))
		return
	}

	response := map[string]interface{}{
		"hasLegacySessions": len(legacySessions) > 0,
		"migratedSessions":  0,
	}

	// If no legacy sessions, return early
	if len(legacySessions) == 0 {
		c.JSON(http.StatusOK, response)
		return
	}

	// Ensure default workspace exists
	defaultWorkspace, err := h.service.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to ensure default workspace"))
		return
	}

	// Migrate all legacy sessions to default workspace
	err = h.service.MigrateSessionsToDefaultWorkspace(ctx, userID, defaultWorkspace.ID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to migrate legacy sessions"))
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

	c.JSON(http.StatusOK, response)
}

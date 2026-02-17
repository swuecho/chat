package main

import (
	"database/sql"
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

// GinRegister registers routes with Gin router
func (h *ChatWorkspaceHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/workspaces", h.GinGetWorkspacesByUserID)
	rg.POST("/workspaces", h.GinCreateWorkspace)
	rg.GET("/workspaces/:uuid", h.GinGetWorkspaceByUUID)
	rg.PUT("/workspaces/:uuid", h.GinUpdateWorkspace)
	rg.DELETE("/workspaces/:uuid", h.GinDeleteWorkspace)
	rg.PUT("/workspaces/:uuid/reorder", h.GinUpdateWorkspaceOrder)
	rg.PUT("/workspaces/:uuid/set-default", h.GinSetDefaultWorkspace)
	rg.POST("/workspaces/:uuid/sessions", h.GinCreateSessionInWorkspace)
	rg.GET("/workspaces/:uuid/sessions", h.GinGetSessionsByWorkspace)
	rg.POST("/workspaces/default", h.GinEnsureDefaultWorkspace)
	rg.POST("/workspaces/auto-migrate", h.GinAutoMigrateLegacySessions)
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

type CreateSessionInWorkspaceRequest struct {
	Topic string `json:"topic"`
	Model string `json:"model"`
}

// =============================================================================
// Gin Handlers
// =============================================================================

func (h *ChatWorkspaceHandler) GinCreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	if req.Name == "" {
		ErrValidationInvalidInput("Workspace name is required").GinResponse(c)
		return
	}

	if req.Color == "" {
		req.Color = "#6366f1"
	}
	if req.Icon == "" {
		req.Icon = "folder"
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		OrderPosition: 0,
	}

	workspace, err := h.service.CreateWorkspace(c.Request.Context(), params)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to create workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinGetWorkspaceByUUID(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	workspace, err := h.service.GetWorkspaceByUUID(c.Request.Context(), workspaceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Workspace")
			apiErr.Message = "Workspace not found with UUID: " + workspaceUUID
			apiErr.GinResponse(c)
			return
		}
		WrapError(MapDatabaseError(err), "Failed to get workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinGetWorkspacesByUserID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	workspaces, err := h.service.GetWorkspaceWithSessionCount(c.Request.Context(), userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get workspaces").GinResponse(c)
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

	c.JSON(http.StatusOK, responses)
}

func (h *ChatWorkspaceHandler) GinUpdateWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	var req UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	params := sqlc_queries.UpdateWorkspaceParams{
		Uuid:        workspaceUUID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
		Icon:        req.Icon,
	}

	workspace, err := h.service.UpdateWorkspace(c.Request.Context(), params)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinUpdateWorkspaceOrder(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	var req UpdateWorkspaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	params := sqlc_queries.UpdateWorkspaceOrderParams{
		Uuid:          workspaceUUID,
		OrderPosition: req.OrderPosition,
	}

	workspace, err := h.service.UpdateWorkspaceOrder(c.Request.Context(), params)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update workspace order").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinDeleteWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	workspace, err := h.service.GetWorkspaceByUUID(c.Request.Context(), workspaceUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get workspace").GinResponse(c)
		return
	}

	if workspace.IsDefault {
		ErrValidationInvalidInput("Cannot delete default workspace").GinResponse(c)
		return
	}

	err = h.service.DeleteWorkspace(c.Request.Context(), workspaceUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to delete workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "Workspace deleted successfully"})
}

func (h *ChatWorkspaceHandler) GinSetDefaultWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	workspaces, err := h.service.GetWorkspacesByUserID(c.Request.Context(), userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get workspaces").GinResponse(c)
		return
	}

	for _, ws := range workspaces {
		if ws.IsDefault {
			_, err = h.service.SetDefaultWorkspace(c.Request.Context(), sqlc_queries.SetDefaultWorkspaceParams{
				Uuid:      ws.Uuid,
				IsDefault: false,
			})
			if err != nil {
				WrapError(MapDatabaseError(err), "Failed to unset default workspace").GinResponse(c)
				return
			}
		}
	}

	workspace, err := h.service.SetDefaultWorkspace(c.Request.Context(), sqlc_queries.SetDefaultWorkspaceParams{
		Uuid:      workspaceUUID,
		IsDefault: true,
	})
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to set default workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinEnsureDefaultWorkspace(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	workspace, err := h.service.EnsureDefaultWorkspaceExists(c.Request.Context(), userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to ensure default workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, WorkspaceResponse{
		Uuid:          workspace.Uuid,
		Name:          workspace.Name,
		Description:   workspace.Description,
		Color:         workspace.Color,
		Icon:          workspace.Icon,
		IsDefault:     workspace.IsDefault,
		OrderPosition: workspace.OrderPosition,
		CreatedAt:     workspace.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     workspace.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinCreateSessionInWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	var req CreateSessionInWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	workspace, err := h.service.GetWorkspaceByUUID(c.Request.Context(), workspaceUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get workspace").GinResponse(c)
		return
	}

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

	session, err := sessionService.q.CreateChatSessionInWorkspace(c.Request.Context(), sessionParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to create session in workspace").GinResponse(c)
		return
	}

	_, err = activeSessionService.UpsertActiveSession(c.Request.Context(), userID, &workspace.ID, sessionUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to set active session").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"uuid":              session.Uuid,
		"topic":             session.Topic,
		"model":             session.Model,
		"codeRunnerEnabled": session.CodeRunnerEnabled,
		"artifactEnabled":   session.ArtifactEnabled,
		"workspaceUuid":     workspaceUUID,
		"createdAt":         session.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

func (h *ChatWorkspaceHandler) GinGetSessionsByWorkspace(c *gin.Context) {
	workspaceUUID := c.Param("uuid")

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	hasPermission, err := h.service.HasWorkspacePermission(c.Request.Context(), workspaceUUID, userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check workspace permission").GinResponse(c)
		return
	}
	if !hasPermission {
		apiErr := ErrAuthAccessDenied
		apiErr.Message = "Access denied to workspace"
		apiErr.GinResponse(c)
		return
	}

	workspace, err := h.service.GetWorkspaceByUUID(c.Request.Context(), workspaceUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get workspace").GinResponse(c)
		return
	}

	sessionService := NewChatSessionService(h.service.q)
	sessions, err := sessionService.q.GetSessionsByWorkspaceID(c.Request.Context(), sql.NullInt32{Int32: workspace.ID, Valid: true})
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get sessions").GinResponse(c)
		return
	}

	sessionResponses := make([]map[string]interface{}, 0)
	for _, session := range sessions {
		sessionResponse := map[string]interface{}{
			"uuid":              session.Uuid,
			"title":             session.Topic,
			"isEdit":            false,
			"model":             session.Model,
			"workspaceUuid":     workspaceUUID,
			"maxLength":         session.MaxLength,
			"temperature":       session.Temperature,
			"maxTokens":         session.MaxTokens,
			"topP":              session.TopP,
			"n":                 session.N,
			"debug":             session.Debug,
			"summarizeMode":     session.SummarizeMode,
			"exploreMode":       session.ExploreMode,
			"codeRunnerEnabled": session.CodeRunnerEnabled,
			"artifactEnabled":   session.ArtifactEnabled,
			"createdAt":         session.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":         session.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
		sessionResponses = append(sessionResponses, sessionResponse)
	}

	c.JSON(http.StatusOK, sessionResponses)
}

func (h *ChatWorkspaceHandler) GinAutoMigrateLegacySessions(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	sessionService := NewChatSessionService(h.service.q)
	legacySessions, err := sessionService.q.GetSessionsWithoutWorkspace(c.Request.Context(), userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to check for legacy sessions").GinResponse(c)
		return
	}

	response := map[string]interface{}{
		"hasLegacySessions": len(legacySessions) > 0,
		"migratedSessions":  0,
	}

	if len(legacySessions) == 0 {
		c.JSON(http.StatusOK, response)
		return
	}

	defaultWorkspace, err := h.service.EnsureDefaultWorkspaceExists(c.Request.Context(), userID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to ensure default workspace").GinResponse(c)
		return
	}

	err = h.service.MigrateSessionsToDefaultWorkspace(c.Request.Context(), userID, defaultWorkspace.ID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to migrate legacy sessions").GinResponse(c)
		return
	}

	activeSessionService := NewUserActiveChatSessionService(h.service.q)
	legacyActiveSessions, err := activeSessionService.q.GetAllUserActiveSessions(c.Request.Context(), userID)
	if err == nil {
		for _, activeSession := range legacyActiveSessions {
			if !activeSession.WorkspaceID.Valid {
				_, _ = activeSessionService.UpsertActiveSession(c.Request.Context(), userID, &defaultWorkspace.ID, activeSession.ChatSessionUuid)
				_ = activeSessionService.DeleteActiveSession(c.Request.Context(), userID, nil)
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

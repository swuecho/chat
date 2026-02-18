package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	sqlc "github.com/swuecho/chat_backend/sqlc_queries"
)

// UserActiveChatSessionHandler handles requests related to active chat sessions
type UserActiveChatSessionHandler struct {
	service *UserActiveChatSessionService
}

// NewUserActiveChatSessionHandler creates a new handler instance
func NewUserActiveChatSessionHandler(sqlc_q *sqlc.Queries) *UserActiveChatSessionHandler {
	activeSessionService := NewUserActiveChatSessionService(sqlc_q)
	return &UserActiveChatSessionHandler{
		service: activeSessionService,
	}
}

// Register sets up the handler routes
func (h *UserActiveChatSessionHandler) Register(router *gin.RouterGroup) {
	router.GET("/uuid/user_active_chat_session", h.GetUserActiveChatSessionHandler)
	router.PUT("/uuid/user_active_chat_session", h.CreateOrUpdateUserActiveChatSessionHandler)

	// Per-workspace active session endpoints
	// Note: More specific routes must come before parameterized routes to avoid shadowing
	router.GET("/workspaces/active-sessions", h.GetAllWorkspaceActiveSessionsHandler)
	router.GET("/workspaces/:workspaceUuid/active-session", h.GetWorkspaceActiveSessionHandler)
	router.PUT("/workspaces/:workspaceUuid/active-session", h.SetWorkspaceActiveSessionHandler)
}

// GetUserActiveChatSessionHandler handles GET requests to get a session by user_id
func (h *UserActiveChatSessionHandler) GetUserActiveChatSessionHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	log.Printf("Getting active chat session for user %d", userID)

	// Get session from service (use unified approach for global session)
	session, err := h.service.GetActiveSession(c.Request.Context(), userID, nil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID)))
		} else {
			RespondWithAPIErrorGin(c, WrapError(err, "failed to get active chat session"))
		}
		return
	}

	c.JSON(200, session)
}

// UUID validation regex
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// CreateOrUpdateUserActiveChatSessionHandler handles PUT requests to create/update a session
func (h *UserActiveChatSessionHandler) CreateOrUpdateUserActiveChatSessionHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Parse request body
	var reqBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	// Validate session UUID format
	if !uuidRegex.MatchString(reqBody.ChatSessionUuid) {
		RespondWithAPIErrorGin(c, ErrChatSessionInvalid.WithMessage(
			fmt.Sprintf("invalid session UUID format: %s", reqBody.ChatSessionUuid)))
		return
	}

	log.Printf("Creating/updating active chat session for user %d", userID)

	// Create/update session (use unified approach for global session)
	session, err := h.service.UpsertActiveSession(c.Request.Context(), userID, nil, reqBody.ChatSessionUuid)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to create or update active chat session"))
		return
	}

	c.JSON(200, session)
}

// Per-workspace active session handlers

// GetWorkspaceActiveSessionHandler gets the active session for a specific workspace
func (h *UserActiveChatSessionHandler) GetWorkspaceActiveSessionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	workspaceUuid := c.Param("workspaceUuid")

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Get workspace to get its ID
	workspaceService := NewChatWorkspaceService(h.service.q)
	workspace, err := workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	// Get workspace active session
	session, err := h.service.GetActiveSession(ctx, userID, &workspace.ID)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrResourceNotFound("Active Session").WithMessage("no active session for workspace"))
		return
	}

	c.JSON(200, map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

// SetWorkspaceActiveSessionHandler sets the active session for a specific workspace
func (h *UserActiveChatSessionHandler) SetWorkspaceActiveSessionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	workspaceUuid := c.Param("workspaceUuid")

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Parse request body
	var requestBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("failed to parse request body"))
		return
	}

	// Validate session UUID format
	if !uuidRegex.MatchString(requestBody.ChatSessionUuid) {
		RespondWithAPIErrorGin(c, ErrChatSessionInvalid.WithMessage("invalid session UUID format"))
		return
	}

	// Get workspace to get its ID
	workspaceService := NewChatWorkspaceService(h.service.q)
	workspace, err := workspaceService.GetWorkspaceByUUID(ctx, workspaceUuid)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrResourceNotFound("Workspace").WithMessage("workspace not found"))
		return
	}

	// Set workspace active session
	session, err := h.service.UpsertActiveSession(ctx, userID, &workspace.ID, requestBody.ChatSessionUuid)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to set workspace active session"))
		return
	}

	c.JSON(200, map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

// GetAllWorkspaceActiveSessionsHandler gets all workspace active sessions for a user
func (h *UserActiveChatSessionHandler) GetAllWorkspaceActiveSessionsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Get and validate user ID
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID"))
		return
	}

	// Get all workspace active sessions
	sessions, err := h.service.GetAllActiveSessions(ctx, userID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to get workspace active sessions"))
		return
	}

	// Convert to response format with workspace UUIDs
	workspaceService := NewChatWorkspaceService(h.service.q)
	workspaces, err := workspaceService.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "failed to get workspaces"))
		return
	}

	// Create a map for workspace ID to UUID lookup
	workspaceMap := make(map[int32]string)
	for _, workspace := range workspaces {
		workspaceMap[workspace.ID] = workspace.Uuid
	}

	// Build response
	var response []map[string]interface{}
	for _, session := range sessions {
		if session.WorkspaceID.Valid {
			if workspaceUuid, exists := workspaceMap[session.WorkspaceID.Int32]; exists {
				response = append(response, map[string]interface{}{
					"workspaceUuid":   workspaceUuid,
					"chatSessionUuid": session.ChatSessionUuid,
					"updatedAt":       session.UpdatedAt,
				})
			}
		}
	}

	c.JSON(200, response)
}

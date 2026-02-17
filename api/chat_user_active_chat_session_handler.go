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

// GinRegister registers routes with Gin router
func (h *UserActiveChatSessionHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/uuid/user_active_chat_session", h.GinGetUserActiveChatSessionHandler)
	rg.PUT("/uuid/user_active_chat_session", h.GinCreateOrUpdateUserActiveChatSessionHandler)
	rg.GET("/workspaces/active-sessions", h.GinGetAllWorkspaceActiveSessionsHandler)
	rg.GET("/workspaces/:workspaceUuid/active-session", h.GinGetWorkspaceActiveSessionHandler)
	rg.PUT("/workspaces/:workspaceUuid/active-session", h.GinSetWorkspaceActiveSessionHandler)
}

// UUID validation regex
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (h *UserActiveChatSessionHandler) GinGetUserActiveChatSessionHandler(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	log.Printf("Getting active chat session for user %d", userID)

	session, err := h.service.GetActiveSession(c.Request.Context(), userID, nil)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrChatSessionNotFound.WithMessage(fmt.Sprintf("no active session for user %d", userID)).GinResponse(c)
		} else {
			WrapError(err, "failed to get active chat session").GinResponse(c)
		}
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *UserActiveChatSessionHandler) GinCreateOrUpdateUserActiveChatSessionHandler(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	var reqBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		ErrValidationInvalidInput("failed to parse request body").GinResponse(c)
		return
	}

	if !uuidRegex.MatchString(reqBody.ChatSessionUuid) {
		ErrChatSessionInvalid.WithMessage(
			fmt.Sprintf("invalid session UUID format: %s", reqBody.ChatSessionUuid)).GinResponse(c)
		return
	}

	log.Printf("Creating/updating active chat session for user %d", userID)

	session, err := h.service.UpsertActiveSession(c.Request.Context(), userID, nil, reqBody.ChatSessionUuid)
	if err != nil {
		WrapError(err, "failed to create or update active chat session").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *UserActiveChatSessionHandler) GinGetWorkspaceActiveSessionHandler(c *gin.Context) {
	workspaceUuid := c.Param("workspaceUuid")

	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	workspaceService := NewChatWorkspaceService(h.service.q)
	workspace, err := workspaceService.GetWorkspaceByUUID(c.Request.Context(), workspaceUuid)
	if err != nil {
		ErrResourceNotFound("Workspace").WithMessage("workspace not found").GinResponse(c)
		return
	}

	session, err := h.service.GetActiveSession(c.Request.Context(), userID, &workspace.ID)
	if err != nil {
		ErrResourceNotFound("Active Session").WithMessage("no active session for workspace").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

func (h *UserActiveChatSessionHandler) GinSetWorkspaceActiveSessionHandler(c *gin.Context) {
	workspaceUuid := c.Param("workspaceUuid")

	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	var requestBody struct {
		ChatSessionUuid string `json:"chatSessionUuid"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		ErrValidationInvalidInput("failed to parse request body").GinResponse(c)
		return
	}

	if !uuidRegex.MatchString(requestBody.ChatSessionUuid) {
		ErrChatSessionInvalid.WithMessage("invalid session UUID format").GinResponse(c)
		return
	}

	workspaceService := NewChatWorkspaceService(h.service.q)
	workspace, err := workspaceService.GetWorkspaceByUUID(c.Request.Context(), workspaceUuid)
	if err != nil {
		ErrResourceNotFound("Workspace").WithMessage("workspace not found").GinResponse(c)
		return
	}

	session, err := h.service.UpsertActiveSession(c.Request.Context(), userID, &workspace.ID, requestBody.ChatSessionUuid)
	if err != nil {
		WrapError(err, "failed to set workspace active session").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"chatSessionUuid": session.ChatSessionUuid,
		"workspaceUuid":   workspaceUuid,
		"updatedAt":       session.UpdatedAt,
	})
}

func (h *UserActiveChatSessionHandler) GinGetAllWorkspaceActiveSessionsHandler(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("missing or invalid user ID").GinResponse(c)
		return
	}

	sessions, err := h.service.GetAllActiveSessions(c.Request.Context(), userID)
	if err != nil {
		WrapError(err, "failed to get workspace active sessions").GinResponse(c)
		return
	}

	workspaceService := NewChatWorkspaceService(h.service.q)
	workspaces, err := workspaceService.GetWorkspacesByUserID(c.Request.Context(), userID)
	if err != nil {
		WrapError(err, "failed to get workspaces").GinResponse(c)
		return
	}

	workspaceMap := make(map[int32]string)
	for _, workspace := range workspaces {
		workspaceMap[workspace.ID] = workspace.Uuid
	}

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

	c.JSON(http.StatusOK, response)
}

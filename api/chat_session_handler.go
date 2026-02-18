package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatSessionHandler struct {
	service *ChatSessionService
}

func NewChatSessionHandler(sqlc_q *sqlc_queries.Queries) *ChatSessionHandler {
	// create a new ChatSessionService instance
	chatSessionService := NewChatSessionService(sqlc_q)
	return &ChatSessionHandler{
		service: chatSessionService,
	}
}

func (h *ChatSessionHandler) Register(router *gin.RouterGroup) {
	router.GET("/chat_sessions/user", h.getSimpleChatSessionsByUserID)

	router.PUT("/uuid/chat_sessions/max_length/:uuid", h.updateSessionMaxLength)
	router.PUT("/uuid/chat_sessions/topic/:uuid", h.updateChatSessionTopicByUUID)
	router.GET("/uuid/chat_sessions/:uuid", h.getChatSessionByUUID)
	router.PUT("/uuid/chat_sessions/:uuid", h.createOrUpdateChatSessionByUUID)
	router.DELETE("/uuid/chat_sessions/:uuid", h.deleteChatSessionByUUID)
	router.POST("/uuid/chat_sessions", h.createChatSessionByUUID)
	router.POST("/uuid/chat_session_from_snapshot/:uuid", h.createChatSessionFromSnapshot)
}

// getChatSessionByUUID returns a chat session by its UUID
func (h *ChatSessionHandler) getChatSessionByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	session, err := h.service.GetChatSessionByUUID(c.Request.Context(), uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Chat session")
			apiErr.Message = "Session not found with UUID: " + uuid
			RespondWithAPIErrorGin(c, apiErr)
			return
		} else {
			apiErr := WrapError(MapDatabaseError(err), "Failed to get chat session")
			RespondWithAPIErrorGin(c, apiErr)
			return
		}
	}

	session_resp := &ChatSessionResponse{
		Uuid:      session.Uuid,
		Topic:     session.Topic,
		MaxLength: session.MaxLength,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		CodeRunnerEnabled: session.CodeRunnerEnabled,
		ArtifactEnabled: session.ArtifactEnabled,
	}
	c.JSON(200, session_resp)
}

// createChatSessionByUUID creates a chat session by its UUID (idempotent)
func (h *ChatSessionHandler) createChatSessionByUUID(c *gin.Context) {
	var sessionParams sqlc_queries.CreateChatSessionParams
	err := c.ShouldBindJSON(&sessionParams)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	ctx := c.Request.Context()
	userIDInt, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Get or create default workspace for the user
	workspaceService := NewChatWorkspaceService(h.service.q)
	defaultWorkspace, err := workspaceService.EnsureDefaultWorkspaceExists(ctx, userIDInt)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// Use CreateOrUpdateChatSessionByUUID for idempotent session creation
	createOrUpdateParams := sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid:          sessionParams.Uuid,
		UserID:        userIDInt,
		Topic:         sessionParams.Topic,
		MaxLength:     DefaultMaxLength,
		Temperature:   DefaultTemperature,
		Model:         sessionParams.Model,
		MaxTokens:     DefaultMaxTokens,
		TopP:          DefaultTopP,
		N:             DefaultN,
		Debug:         false,
		SummarizeMode: false,
		ExploreMode:   false,
		ArtifactEnabled: false,
		WorkspaceID:   sql.NullInt32{Int32: defaultWorkspace.ID, Valid: true},
	}

	session, err := h.service.CreateOrUpdateChatSessionByUUID(c.Request.Context(), createOrUpdateParams)
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to create or update chat session")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	// set active chat session when creating a new chat session (use unified approach)
	_, err = h.service.q.UpsertUserActiveSession(c.Request.Context(),
		sqlc_queries.UpsertUserActiveSessionParams{
			UserID:          session.UserID,
			WorkspaceID:     sql.NullInt32{Valid: false},
			ChatSessionUuid: session.Uuid,
		})
	if err != nil {
		apiErr := WrapError(MapDatabaseError(err), "Failed to update or create active user session record")
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, session)
}

type UpdateChatSessionRequest struct {
	Uuid          string  `json:"uuid"`
	Topic         string  `json:"topic"`
	MaxLength     int32   `json:"maxLength"`
	Temperature   float64 `json:"temperature"`
	Model         string  `json:"model"`
	TopP          float64 `json:"topP"`
	N             int32   `json:"n"`
	MaxTokens     int32   `json:"maxTokens"`
	Debug         bool    `json:"debug"`
	SummarizeMode bool    `json:"summarizeMode"`
	CodeRunnerEnabled bool `json:"codeRunnerEnabled"`
	ArtifactEnabled bool `json:"artifactEnabled"`
	ExploreMode   bool    `json:"exploreMode"`
	WorkspaceUUID string  `json:"workspaceUuid,omitempty"`
}

// UpdateChatSessionByUUID updates a chat session by its UUID
func (h *ChatSessionHandler) createOrUpdateChatSessionByUUID(c *gin.Context) {
	var sessionReq UpdateChatSessionRequest
	err := c.ShouldBindJSON(&sessionReq)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = DefaultMaxLength
	}

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	var sessionParams sqlc_queries.CreateOrUpdateChatSessionByUUIDParams

	sessionParams.MaxLength = sessionReq.MaxLength
	sessionParams.Topic = sessionReq.Topic
	sessionParams.Uuid = sessionReq.Uuid
	sessionParams.UserID = userID
	sessionParams.Temperature = sessionReq.Temperature
	sessionParams.Model = sessionReq.Model
	sessionParams.TopP = sessionReq.TopP
	sessionParams.N = sessionReq.N
	sessionParams.MaxTokens = sessionReq.MaxTokens
	sessionParams.Debug = sessionReq.Debug
	sessionParams.SummarizeMode = sessionReq.SummarizeMode
	sessionParams.CodeRunnerEnabled = sessionReq.CodeRunnerEnabled
	sessionParams.ArtifactEnabled = sessionReq.ArtifactEnabled
	sessionParams.ExploreMode = sessionReq.ExploreMode

	// Handle workspace
	if sessionReq.WorkspaceUUID != "" {
		workspaceService := NewChatWorkspaceService(h.service.q)
		workspace, err := workspaceService.GetWorkspaceByUUID(ctx, sessionReq.WorkspaceUUID)
		if err != nil {
			apiErr := WrapError(MapDatabaseError(err), "Invalid workspace UUID")
			RespondWithAPIErrorGin(c, apiErr)
			return
		}
		sessionParams.WorkspaceID = sql.NullInt32{Int32: workspace.ID, Valid: true}
	} else {
		// Ensure default workspace exists
		workspaceService := NewChatWorkspaceService(h.service.q)
		defaultWorkspace, err := workspaceService.EnsureDefaultWorkspaceExists(ctx, userID)
		if err != nil {
			apiErr := WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists")
			RespondWithAPIErrorGin(c, apiErr)
			return
		}
		sessionParams.WorkspaceID = sql.NullInt32{Int32: defaultWorkspace.ID, Valid: true}
	}
	session, err := h.service.CreateOrUpdateChatSessionByUUID(c.Request.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create or update chat session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, session)
}

// deleteChatSessionByUUID deletes a chat session by its UUID
func (h *ChatSessionHandler) deleteChatSessionByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	err := h.service.DeleteChatSessionByUUID(c.Request.Context(), uuid)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// getSimpleChatSessionsByUserID returns a list of simple chat sessions by user ID
func (h *ChatSessionHandler) getSimpleChatSessionsByUserID(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := ctx.Value(userContextKey).(string)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid user ID")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	sessions, err := h.service.GetSimpleChatSessionsByUserID(ctx, int32(id))
	if err != nil {
		apiErr := ErrResourceNotFound("Chat sessions")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, sessions)
}

// updateChatSessionTopicByUUID updates a chat session topic by its UUID
func (h *ChatSessionHandler) updateChatSessionTopicByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	var sessionParams sqlc_queries.UpdateChatSessionTopicByUUIDParams
	err := c.ShouldBindJSON(&sessionParams)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	sessionParams.Uuid = uuid

	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	sessionParams.UserID = userID

	session, err := h.service.UpdateChatSessionTopicByUUID(c.Request.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat session topic"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, session)
}

// updateSessionMaxLength
func (h *ChatSessionHandler) updateSessionMaxLength(c *gin.Context) {
	uuid := c.Param("uuid")
	var sessionParams sqlc_queries.UpdateSessionMaxLengthParams
	err := c.ShouldBindJSON(&sessionParams)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	sessionParams.Uuid = uuid

	session, err := h.service.UpdateSessionMaxLength(c.Request.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update session max length"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(200, session)
}

// CreateChatSessionFromSnapshot ($uuid)
// create a new session with title of snapshot,
// create a prompt with the first message of snapshot
// create messages based on the rest of messages.
// return the new session uuid

func (h *ChatSessionHandler) createChatSessionFromSnapshot(c *gin.Context) {
	snapshot_uuid := c.Param("uuid")

	userID, err := getUserID(c.Request.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(c.Request.Context(), snapshot_uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	sessionTitle := snapshot.Title
	conversions := snapshot.Conversation
	var conversionsSimpleMessages []SimpleChatMessage
	json.Unmarshal(conversions, &conversionsSimpleMessages)
	promptMsg := conversionsSimpleMessages[0]
	chatPrompt, err := h.service.q.GetChatPromptByUUID(c.Request.Context(), promptMsg.Uuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat prompt")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	originSession, err := h.service.q.GetChatSessionByUUIDWithInActive(c.Request.Context(), chatPrompt.ChatSessionUuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Original chat session")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	sessionUUID := uuid.New().String()

	session, err := h.service.q.CreateOrUpdateChatSessionByUUID(c.Request.Context(), sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid:          sessionUUID,
		UserID:        userID,
		Topic:         sessionTitle,
		MaxLength:     originSession.MaxLength,
		Temperature:   originSession.Temperature,
		Model:         originSession.Model,
		MaxTokens:     originSession.MaxTokens,
		TopP:          originSession.TopP,
		Debug:         originSession.Debug,
		SummarizeMode: originSession.SummarizeMode,
		CodeRunnerEnabled: originSession.CodeRunnerEnabled,
		ExploreMode:   originSession.ExploreMode,
		WorkspaceID:   originSession.WorkspaceID,
		N:             1,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat session from snapshot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	_, err = h.service.q.CreateChatPrompt(c.Request.Context(), sqlc_queries.CreateChatPromptParams{
		Uuid:            NewUUID(),
		ChatSessionUuid: sessionUUID,
		Role:            "system",
		Content:         promptMsg.Text,
		UserID:          userID,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create prompt for chat session from snapshot"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}

	for _, message := range conversionsSimpleMessages[1:] {
		// if inversion is true, the role is user, otherwise assistant
		// Determine the role based on the inversion flag

		messageParam := sqlc_queries.CreateChatMessageParams{
			ChatSessionUuid: sessionUUID,
			Uuid:            NewUUID(),
			Role:            message.GetRole(),
			Content:         message.Text,
			UserID:          userID,
			Raw:             json.RawMessage([]byte("{}")),
		}
		_, err = h.service.q.CreateChatMessage(c.Request.Context(), messageParam)
		if err != nil {
			apiErr := ErrInternalUnexpected
			apiErr.Detail = "Failed to create messages for chat session from snapshot"
			apiErr.DebugInfo = err.Error()
			RespondWithAPIErrorGin(c, apiErr)
			return
		}

	}

	// set active session using simplified service
	activeSessionService := NewUserActiveChatSessionService(h.service.q)
	_, err = activeSessionService.UpsertActiveSession(c.Request.Context(), userID, nil, session.Uuid)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update active session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIErrorGin(c, apiErr)
		return
	}
	c.JSON(http.StatusOK, map[string]string{"SessionUuid": session.Uuid})
}

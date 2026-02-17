package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

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

// GinRegister registers routes with Gin router
func (h *ChatSessionHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/chat_sessions/user", h.GinGetSimpleChatSessionsByUserID)
	rg.PUT("/uuid/chat_sessions/max_length/:uuid", h.GinUpdateSessionMaxLength)
	rg.PUT("/uuid/chat_sessions/topic/:uuid", h.GinUpdateChatSessionTopicByUUID)
	rg.GET("/uuid/chat_sessions/:uuid", h.GinGetChatSessionByUUID)
	rg.PUT("/uuid/chat_sessions/:uuid", h.GinCreateOrUpdateChatSessionByUUID)
	rg.DELETE("/uuid/chat_sessions/:uuid", h.GinDeleteChatSessionByUUID)
	rg.POST("/uuid/chat_sessions", h.GinCreateChatSessionByUUID)
	rg.POST("/uuid/chat_session_from_snapshot/:uuid", h.GinCreateChatSessionFromSnapshot)
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

// =============================================================================
// Gin Handlers
// =============================================================================

func (h *ChatSessionHandler) GinGetChatSessionByUUID(c *gin.Context) {
	sessionUUID := c.Param("uuid")
	session, err := h.service.GetChatSessionByUUID(c.Request.Context(), sessionUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			apiErr := ErrResourceNotFound("Chat session")
			apiErr.Message = "Session not found with UUID: " + sessionUUID
			apiErr.GinResponse(c)
			return
		} else {
			WrapError(MapDatabaseError(err), "Failed to get chat session").GinResponse(c)
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
	c.JSON(http.StatusOK, session_resp)
}

func (h *ChatSessionHandler) GinCreateChatSessionByUUID(c *gin.Context) {
	var sessionParams sqlc_queries.CreateChatSessionParams
	if err := c.ShouldBindJSON(&sessionParams); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	userIDInt, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	workspaceService := NewChatWorkspaceService(h.service.q)
	defaultWorkspace, err := workspaceService.EnsureDefaultWorkspaceExists(c.Request.Context(), userIDInt)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists").GinResponse(c)
		return
	}

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
		WrapError(MapDatabaseError(err), "Failed to create or update chat session").GinResponse(c)
		return
	}

	_, err = h.service.q.UpsertUserActiveSession(c.Request.Context(),
		sqlc_queries.UpsertUserActiveSessionParams{
			UserID:          session.UserID,
			WorkspaceID:     sql.NullInt32{Valid: false},
			ChatSessionUuid: session.Uuid,
		})
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update or create active user session record").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *ChatSessionHandler) GinCreateOrUpdateChatSessionByUUID(c *gin.Context) {
	var sessionReq UpdateChatSessionRequest
	if err := c.ShouldBindJSON(&sessionReq); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	if sessionReq.MaxLength == 0 {
		sessionReq.MaxLength = DefaultMaxLength
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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

	if sessionReq.WorkspaceUUID != "" {
		workspaceService := NewChatWorkspaceService(h.service.q)
		workspace, err := workspaceService.GetWorkspaceByUUID(c.Request.Context(), sessionReq.WorkspaceUUID)
		if err != nil {
			WrapError(MapDatabaseError(err), "Invalid workspace UUID").GinResponse(c)
			return
		}
		sessionParams.WorkspaceID = sql.NullInt32{Int32: workspace.ID, Valid: true}
	} else {
		workspaceService := NewChatWorkspaceService(h.service.q)
		defaultWorkspace, err := workspaceService.EnsureDefaultWorkspaceExists(c.Request.Context(), userID)
		if err != nil {
			WrapError(MapDatabaseError(err), "Failed to ensure default workspace exists").GinResponse(c)
			return
		}
		sessionParams.WorkspaceID = sql.NullInt32{Int32: defaultWorkspace.ID, Valid: true}
	}

	session, err := h.service.CreateOrUpdateChatSessionByUUID(c.Request.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create or update chat session"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *ChatSessionHandler) GinDeleteChatSessionByUUID(c *gin.Context) {
	sessionUUID := c.Param("uuid")
	err := h.service.DeleteChatSessionByUUID(c.Request.Context(), sessionUUID)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat session"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatSessionHandler) GinGetSimpleChatSessionsByUserID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid user ID")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	sessions, err := h.service.GetSimpleChatSessionsByUserID(c.Request.Context(), userID)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat sessions")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, sessions)
}

func (h *ChatSessionHandler) GinUpdateChatSessionTopicByUUID(c *gin.Context) {
	sessionUUID := c.Param("uuid")
	var sessionParams sqlc_queries.UpdateChatSessionTopicByUUIDParams
	if err := c.ShouldBindJSON(&sessionParams); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	sessionParams.Uuid = sessionUUID

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	sessionParams.UserID = userID

	session, err := h.service.UpdateChatSessionTopicByUUID(c.Request.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat session topic"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *ChatSessionHandler) GinUpdateSessionMaxLength(c *gin.Context) {
	sessionUUID := c.Param("uuid")
	var sessionParams sqlc_queries.UpdateSessionMaxLengthParams
	if err := c.ShouldBindJSON(&sessionParams); err != nil {
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	sessionParams.Uuid = sessionUUID

	session, err := h.service.UpdateSessionMaxLength(c.Request.Context(), sessionParams)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update session max length"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *ChatSessionHandler) GinCreateChatSessionFromSnapshot(c *gin.Context) {
	snapshotUUID := c.Param("uuid")

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	snapshot, err := h.service.q.ChatSnapshotByUUID(c.Request.Context(), snapshotUUID)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr.GinResponse(c)
		return
	}
	originSession, err := h.service.q.GetChatSessionByUUIDWithInActive(c.Request.Context(), chatPrompt.ChatSessionUuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Original chat session")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr.GinResponse(c)
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
		apiErr.GinResponse(c)
		return
	}

	for _, message := range conversionsSimpleMessages[1:] {
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
			apiErr.GinResponse(c)
			return
		}
	}

	activeSessionService := NewUserActiveChatSessionService(h.service.q)
	_, err = activeSessionService.UpsertActiveSession(c.Request.Context(), userID, nil, session.Uuid)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update active session"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, map[string]string{"SessionUuid": session.Uuid})
}

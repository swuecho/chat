package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatMessageHandler struct {
	service *ChatMessageService
}

func NewChatMessageHandler(sqlc_q *sqlc_queries.Queries) *ChatMessageHandler {
	chatMessageService := NewChatMessageService(sqlc_q)
	return &ChatMessageHandler{
		service: chatMessageService,
	}
}

// GinRegister registers routes with Gin router
func (h *ChatMessageHandler) GinRegister(rg *gin.RouterGroup) {
	rg.POST("/chat_messages", h.GinCreateChatMessage)
	rg.GET("/chat_messages/:id", h.GinGetChatMessageByID)
	rg.PUT("/chat_messages/:id", h.GinUpdateChatMessage)
	rg.DELETE("/chat_messages/:id", h.GinDeleteChatMessage)
	rg.GET("/chat_messages", h.GinGetAllChatMessages)

	rg.GET("/uuid/chat_messages/:uuid", h.GinGetChatMessageByUUID)
	rg.PUT("/uuid/chat_messages/:uuid", h.GinUpdateChatMessageByUUID)
	rg.DELETE("/uuid/chat_messages/:uuid", h.GinDeleteChatMessageByUUID)
	rg.POST("/uuid/chat_messages/:uuid/generate-suggestions", h.GinGenerateMoreSuggestions)
	rg.GET("/uuid/chat_messages/chat_sessions/:uuid", h.GinGetChatHistoryBySessionUUID)
	rg.DELETE("/uuid/chat_messages/chat_sessions/:uuid", h.GinDeleteChatMessagesBySessionUUID)
}

// =============================================================================
// Gin Handlers
// =============================================================================

func (h *ChatMessageHandler) GinCreateChatMessage(c *gin.Context) {
	var messageParams sqlc_queries.CreateChatMessageParams
	if err := c.ShouldBindJSON(&messageParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	message, err := h.service.CreateChatMessage(c.Request.Context(), messageParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to create chat message").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *ChatMessageHandler) GinGetChatMessageByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid chat message ID").GinResponse(c)
		return
	}
	message, err := h.service.GetChatMessageByID(c.Request.Context(), int32(id))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat message").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *ChatMessageHandler) GinUpdateChatMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid chat message ID").GinResponse(c)
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageParams
	if err := c.ShouldBindJSON(&messageParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	messageParams.ID = int32(id)
	message, err := h.service.UpdateChatMessage(c.Request.Context(), messageParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update chat message").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *ChatMessageHandler) GinDeleteChatMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid chat message ID").GinResponse(c)
		return
	}
	err = h.service.DeleteChatMessage(c.Request.Context(), int32(id))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to delete chat message").GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatMessageHandler) GinGetAllChatMessages(c *gin.Context) {
	messages, err := h.service.GetAllChatMessages(c.Request.Context())
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat messages").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, messages)
}

func (h *ChatMessageHandler) GinGetChatMessageByUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	message, err := h.service.GetChatMessageByUUID(c.Request.Context(), uuidStr)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat message").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *ChatMessageHandler) GinUpdateChatMessageByUUID(c *gin.Context) {
	var simple_msg SimpleChatMessage
	if err := c.ShouldBindJSON(&simple_msg); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageByUUIDParams
	messageParams.Uuid = simple_msg.Uuid
	messageParams.Content = simple_msg.Text
	tokenCount, _ := getTokenCount(simple_msg.Text)
	messageParams.TokenCount = int32(tokenCount)
	messageParams.IsPin = simple_msg.IsPin
	message, err := h.service.UpdateChatMessageByUUID(c.Request.Context(), messageParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update chat message").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *ChatMessageHandler) GinDeleteChatMessageByUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	err := h.service.DeleteChatMessageByUUID(c.Request.Context(), uuidStr)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to delete chat message").GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatMessageHandler) GinGetChatHistoryBySessionUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	pageNum, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "200"))
	if err != nil {
		pageSize = 200
	}
	simple_msgs, err := h.service.q.GetChatHistoryBySessionUUID(c.Request.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat history").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, simple_msgs)
}

func (h *ChatMessageHandler) GinDeleteChatMessagesBySessionUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	err := h.service.DeleteChatMessagesBySesionUUID(c.Request.Context(), uuidStr)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to delete chat messages").GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatMessageHandler) GinGenerateMoreSuggestions(c *gin.Context) {
	messageUUID := c.Param("uuid")

	message, err := h.service.q.GetChatMessageByUUID(c.Request.Context(), messageUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrChatMessageNotFound.WithMessage("Message not found").WithDebugInfo(err.Error()).GinResponse(c)
		} else {
			WrapError(MapDatabaseError(err), "Failed to get message").GinResponse(c)
		}
		return
	}

	if message.Role != "assistant" {
		ErrValidationInvalidInput("Suggestions can only be generated for assistant messages").GinResponse(c)
		return
	}

	session, err := h.service.q.GetChatSessionByUUID(c.Request.Context(), message.ChatSessionUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ErrChatSessionNotFound.WithMessage("Session not found").WithDebugInfo(err.Error()).GinResponse(c)
		} else {
			WrapError(MapDatabaseError(err), "Failed to get session").GinResponse(c)
		}
		return
	}

	if !session.ExploreMode {
		ErrValidationInvalidInput("Suggestions are only available in explore mode").GinResponse(c)
		return
	}

	contextMessages, err := h.service.q.GetLatestMessagesBySessionUUID(c.Request.Context(),
		sqlc_queries.GetLatestMessagesBySessionUUIDParams{
			ChatSessionUuid: session.Uuid,
			Limit:           6,
		})
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get conversation context").GinResponse(c)
		return
	}

	var msgs []models.Message
	for _, msg := range contextMessages {
		msgs = append(msgs, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	chatService := NewChatService(h.service.q)
	newSuggestions := chatService.generateSuggestedQuestions(message.Content, msgs)
	if len(newSuggestions) == 0 {
		createAPIError(ErrInternalUnexpected, "Failed to generate suggestions", "no suggestions returned").GinResponse(c)
		return
	}

	var existingSuggestions []string
	if len(message.SuggestedQuestions) > 0 {
		if err := json.Unmarshal(message.SuggestedQuestions, &existingSuggestions); err != nil {
			existingSuggestions = []string{}
		}
	}

	allSuggestions := append(existingSuggestions, newSuggestions...)
	seenSuggestions := make(map[string]bool)
	var uniqueSuggestions []string
	for _, suggestion := range allSuggestions {
		if !seenSuggestions[suggestion] {
			seenSuggestions[suggestion] = true
			uniqueSuggestions = append(uniqueSuggestions, suggestion)
		}
	}

	suggestionsJSON, err := json.Marshal(uniqueSuggestions)
	if err != nil {
		createAPIError(ErrInternalUnexpected, "Failed to serialize suggestions", err.Error()).GinResponse(c)
		return
	}

	_, err = h.service.q.UpdateChatMessageSuggestions(c.Request.Context(),
		sqlc_queries.UpdateChatMessageSuggestionsParams{
			Uuid:               messageUUID,
			SuggestedQuestions: suggestionsJSON,
		})
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update message with suggestions").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"newSuggestions": newSuggestions,
		"allSuggestions": uniqueSuggestions,
	})
}

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
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

func (h *ChatMessageHandler) Register(router *gin.RouterGroup) {
	router.POST("/chat_messages", h.CreateChatMessage)
	router.GET("/chat_messages/:id", h.GetChatMessageByID)
	router.PUT("/chat_messages/:id", h.UpdateChatMessage)
	router.DELETE("/chat_messages/:id", h.DeleteChatMessage)
	router.GET("/chat_messages", h.GetAllChatMessages)

	router.GET("/uuid/chat_messages/:uuid", h.GetChatMessageByUUID)
	router.PUT("/uuid/chat_messages/:uuid", h.UpdateChatMessageByUUID)
	router.DELETE("/uuid/chat_messages/:uuid", h.DeleteChatMessageByUUID)
	router.POST("/uuid/chat_messages/:uuid/generate-suggestions", h.GenerateMoreSuggestions)
	router.GET("/uuid/chat_messages/chat_sessions/:uuid", h.GetChatHistoryBySessionUUID)
	router.DELETE("/uuid/chat_messages/chat_sessions/:uuid", h.DeleteChatMessagesBySesionUUID)
}

//type userIdContextKey string

//const userIDKey = userIdContextKey("userID")

func (h *ChatMessageHandler) CreateChatMessage(c *gin.Context) {
	var messageParams sqlc_queries.CreateChatMessageParams
	err := c.ShouldBindJSON(&messageParams)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	message, err := h.service.CreateChatMessage(c.Request.Context(), messageParams)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to create chat message"))
		return
	}
	c.JSON(200, message)
}

func (h *ChatMessageHandler) GetChatMessageByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid chat message ID"))
		return
	}
	message, err := h.service.GetChatMessageByID(c.Request.Context(), int32(id))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat message"))
		return
	}
	c.JSON(200, message)
}

func (h *ChatMessageHandler) UpdateChatMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid chat message ID"))
		return
	}
	var messageParams sqlc_queries.UpdateChatMessageParams
	err = c.ShouldBindJSON(&messageParams)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	messageParams.ID = int32(id)
	message, err := h.service.UpdateChatMessage(c.Request.Context(), messageParams)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update chat message"))
		return
	}
	c.JSON(200, message)
}

func (h *ChatMessageHandler) DeleteChatMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid chat message ID"))
		return
	}
	err = h.service.DeleteChatMessage(c.Request.Context(), int32(id))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to delete chat message"))
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (h *ChatMessageHandler) GetAllChatMessages(c *gin.Context) {
	messages, err := h.service.GetAllChatMessages(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat messages"))
		return
	}
	c.JSON(200, messages)
}

// GetChatMessageByUUID get chat message by uuid
func (h *ChatMessageHandler) GetChatMessageByUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	message, err := h.service.GetChatMessageByUUID(c.Request.Context(), uuidStr)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat message"))
		return
	}

	c.JSON(200, message)
}

// UpdateChatMessageByUUID update chat message by uuid
func (h *ChatMessageHandler) UpdateChatMessageByUUID(c *gin.Context) {
	var simple_msg SimpleChatMessage
	err := c.ShouldBindJSON(&simple_msg)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
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
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update chat message"))
		return
	}
	c.JSON(200, message)
}

// DeleteChatMessageByUUID delete chat message by uuid
func (h *ChatMessageHandler) DeleteChatMessageByUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	err := h.service.DeleteChatMessageByUUID(c.Request.Context(), uuidStr)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to delete chat message"))
		return
	}
	c.JSON(http.StatusOK, nil)
}

// GetChatMessagesBySessionUUID get chat messages by session uuid
func (h *ChatMessageHandler) GetChatMessagesBySessionUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	pageNum, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil {
		pageSize = 200
	}

	messages, err := h.service.GetChatMessagesBySessionUUID(c.Request.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat messages"))
		return
	}

	simple_msgs := lo.Map(messages, func(message sqlc_queries.ChatMessage, _ int) SimpleChatMessage {
		// Extract artifacts from database
		var artifacts []Artifact
		if message.Artifacts != nil {
			err := json.Unmarshal(message.Artifacts, &artifacts)
			if err != nil {
				// Log error but don't fail the request
				artifacts = []Artifact{}
			}
		}

		return SimpleChatMessage{
			DateTime:  message.UpdatedAt.Format(time.RFC3339),
			Text:      message.Content,
			Inversion: message.Role != "user",
			Error:     false,
			Loading:   false,
			Artifacts: artifacts,
		}
	})
	c.JSON(200, simple_msgs)
}

// GetChatHistoryBySessionUUID get chat messages by session uuid
func (h *ChatMessageHandler) GetChatHistoryBySessionUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	pageNum, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		pageNum = 1
	}
	pageSize, err := strconv.Atoi(c.Query("page_size"))
	if err != nil {
		pageSize = 200
	}
	simple_msgs, err := h.service.q.GetChatHistoryBySessionUUID(c.Request.Context(), uuidStr, int32(pageNum), int32(pageSize))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat history"))
		return
	}
	c.JSON(200, simple_msgs)
}

// DeleteChatMessagesBySesionUUID delete chat messages by session uuid
func (h *ChatMessageHandler) DeleteChatMessagesBySesionUUID(c *gin.Context) {
	uuidStr := c.Param("uuid")
	err := h.service.DeleteChatMessagesBySesionUUID(c.Request.Context(), uuidStr)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to delete chat messages"))
		return
	}
	c.JSON(http.StatusOK, nil)
}

// GenerateMoreSuggestions generates additional suggested questions for a message
func (h *ChatMessageHandler) GenerateMoreSuggestions(c *gin.Context) {
	messageUUID := c.Param("uuid")

	// Get the existing message
	message, err := h.service.q.GetChatMessageByUUID(c.Request.Context(), messageUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrChatMessageNotFound.WithMessage("Message not found").WithDebugInfo(err.Error()))
		} else {
			RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get message"))
		}
		return
	}

	// Only allow suggestions for assistant messages
	if message.Role != "assistant" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Suggestions can only be generated for assistant messages"))
		return
	}

	// Get the session to check if explore mode is enabled
	session, err := h.service.q.GetChatSessionByUUID(c.Request.Context(), message.ChatSessionUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithAPIErrorGin(c, ErrChatSessionNotFound.WithMessage("Session not found").WithDebugInfo(err.Error()))
		} else {
			RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get session"))
		}
		return
	}

	// Check if explore mode is enabled
	if !session.ExploreMode {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Suggestions are only available in explore mode"))
		return
	}

	// Get conversation context - last 6 messages
	contextMessages, err := h.service.q.GetLatestMessagesBySessionUUID(c.Request.Context(),
		sqlc_queries.GetLatestMessagesBySessionUUIDParams{
			ChatSessionUuid: session.Uuid,
			Limit:           6,
		})
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get conversation context"))
		return
	}

	// Convert to models.Message format for suggestion generation
	var msgs []models.Message
	for _, msg := range contextMessages {
		msgs = append(msgs, models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Create a new ChatService to access suggestion generation methods
	chatService := NewChatService(h.service.q)

	// Generate new suggested questions
	newSuggestions := chatService.generateSuggestedQuestions(message.Content, msgs)
	if len(newSuggestions) == 0 {
		RespondWithAPIErrorGin(c, createAPIError(ErrInternalUnexpected, "Failed to generate suggestions", "no suggestions returned"))
		return
	}

	// Parse existing suggestions
	var existingSuggestions []string
	if len(message.SuggestedQuestions) > 0 {
		if err := json.Unmarshal(message.SuggestedQuestions, &existingSuggestions); err != nil {
			// If unmarshal fails, treat as empty array
			existingSuggestions = []string{}
		}
	}

	// Combine existing and new suggestions (avoiding duplicates)
	allSuggestions := append(existingSuggestions, newSuggestions...)

	// Remove duplicates
	seenSuggestions := make(map[string]bool)
	var uniqueSuggestions []string
	for _, suggestion := range allSuggestions {
		if !seenSuggestions[suggestion] {
			seenSuggestions[suggestion] = true
			uniqueSuggestions = append(uniqueSuggestions, suggestion)
		}
	}

	// Update the message with new suggestions
	suggestionsJSON, err := json.Marshal(uniqueSuggestions)
	if err != nil {
		RespondWithAPIErrorGin(c, createAPIError(ErrInternalUnexpected, "Failed to serialize suggestions", err.Error()))
		return
	}

	_, err = h.service.q.UpdateChatMessageSuggestions(c.Request.Context(),
		sqlc_queries.UpdateChatMessageSuggestionsParams{
			Uuid:               messageUUID,
			SuggestedQuestions: suggestionsJSON,
		})
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update message with suggestions"))
		return
	}

	// Return the new suggestions to the client
	response := map[string]interface{}{
		"newSuggestions": newSuggestions,
		"allSuggestions": uniqueSuggestions,
	}

	c.JSON(200, response)
}

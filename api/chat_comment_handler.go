package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatCommentHandler struct {
	service *ChatCommentService
}

func NewChatCommentHandler(sqlc_q *sqlc_queries.Queries) *ChatCommentHandler {
	chatCommentService := NewChatCommentService(sqlc_q)
	return &ChatCommentHandler{
		service: chatCommentService,
	}
}

// GinRegister registers routes with Gin router
func (h *ChatCommentHandler) GinRegister(rg *gin.RouterGroup) {
	rg.POST("/uuid/chat_sessions/:sessionUUID/chat_messages/:messageUUID/comments", h.GinCreateChatComment)
	rg.GET("/uuid/chat_sessions/:sessionUUID/comments", h.GinGetCommentsBySessionUUID)
	rg.GET("/uuid/chat_messages/:messageUUID/comments", h.GinGetCommentsByMessageUUID)
}

func (h *ChatCommentHandler) GinCreateChatComment(c *gin.Context) {
	sessionUUID := c.Param("sessionUUID")
	messageUUID := c.Param("messageUUID")

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithMessage("unauthorized").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	comment, err := h.service.CreateChatComment(c.Request.Context(), sqlc_queries.CreateChatCommentParams{
		Uuid:            uuid.New().String(),
		ChatSessionUuid: sessionUUID,
		ChatMessageUuid: messageUUID,
		Content:         req.Content,
		CreatedBy:       userID,
	})
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to create chat comment").GinResponse(c)
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *ChatCommentHandler) GinGetCommentsBySessionUUID(c *gin.Context) {
	sessionUUID := c.Param("sessionUUID")

	comments, err := h.service.GetCommentsBySessionUUID(c.Request.Context(), sessionUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get comments by session").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, comments)
}

func (h *ChatCommentHandler) GinGetCommentsByMessageUUID(c *gin.Context) {
	messageUUID := c.Param("messageUUID")

	comments, err := h.service.GetCommentsByMessageUUID(c.Request.Context(), messageUUID)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get comments by message").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, comments)
}

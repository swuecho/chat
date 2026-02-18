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

func (h *ChatCommentHandler) Register(router *gin.RouterGroup) {
	router.POST("/uuid/chat_sessions/:sessionUUID/chat_messages/:messageUUID/comments", h.CreateChatComment)
	router.GET("/uuid/chat_sessions/:sessionUUID/comments", h.GetCommentsBySessionUUID)
	router.GET("/uuid/chat_messages/:messageUUID/comments", h.GetCommentsByMessageUUID)
}

func (h *ChatCommentHandler) CreateChatComment(c *gin.Context) {
	sessionUUID := c.Param("sessionUUID")
	messageUUID := c.Param("messageUUID")

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithMessage("unauthorized").WithDebugInfo(err.Error()))
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
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to create chat comment"))
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *ChatCommentHandler) GetCommentsBySessionUUID(c *gin.Context) {
	sessionUUID := c.Param("sessionUUID")

	comments, err := h.service.GetCommentsBySessionUUID(c.Request.Context(), sessionUUID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get comments by session"))
		return
	}

	c.JSON(200, comments)
}

func (h *ChatCommentHandler) GetCommentsByMessageUUID(c *gin.Context) {
	messageUUID := c.Param("messageUUID")

	comments, err := h.service.GetCommentsByMessageUUID(c.Request.Context(), messageUUID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get comments by message"))
		return
	}

	c.JSON(200, comments)
}

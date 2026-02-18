package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatPromptHandler struct {
	service *ChatPromptService
}

func NewChatPromptHandler(sqlc_q *sqlc_queries.Queries) *ChatPromptHandler {
	promptService := NewChatPromptService(sqlc_q)
	return &ChatPromptHandler{
		service: promptService,
	}
}

func (h *ChatPromptHandler) Register(r *gin.RouterGroup) {
	r.POST("/chat_prompts", h.CreateChatPrompt)
	r.GET("/chat_prompts/users", h.GetChatPromptsByUserID)
	r.GET("/chat_prompts/:id", h.GetChatPromptByID)
	r.PUT("/chat_prompts/:id", h.UpdateChatPrompt)
	r.DELETE("/chat_prompts/:id", h.DeleteChatPrompt)
	r.GET("/chat_prompts", h.GetAllChatPrompts)
	r.DELETE("/uuid/chat_prompts/:uuid", h.DeleteChatPromptByUUID)
	r.PUT("/uuid/chat_prompts/:uuid", h.UpdateChatPromptByUUID)
}

func (h *ChatPromptHandler) CreateChatPrompt(c *gin.Context) {
	var promptParams sqlc_queries.CreateChatPromptParams
	if err := c.ShouldBindJSON(&promptParams); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	prompt, err := h.service.CreateChatPrompt(c.Request.Context(), promptParams)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to create chat prompt"))
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func (h *ChatPromptHandler) GetChatPromptByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	prompt, err := h.service.GetChatPromptByID(c.Request.Context(), int32(id))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat prompt"))
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func (h *ChatPromptHandler) UpdateChatPrompt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	var promptParams sqlc_queries.UpdateChatPromptParams
	if err := c.ShouldBindJSON(&promptParams); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	promptParams.ID = int32(id)
	prompt, err := h.service.UpdateChatPrompt(c.Request.Context(), promptParams)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update chat prompt"))
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func (h *ChatPromptHandler) DeleteChatPrompt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	err = h.service.DeleteChatPrompt(c.Request.Context(), int32(id))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to delete chat prompt"))
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (h *ChatPromptHandler) GetAllChatPrompts(c *gin.Context) {
	prompts, err := h.service.GetAllChatPrompts(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat prompts"))
		return
	}
	c.JSON(http.StatusOK, prompts)
}

func (h *ChatPromptHandler) GetChatPromptsByUserID(c *gin.Context) {
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	prompts, err := h.service.GetChatPromptsByUserID(c.Request.Context(), userID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get chat prompts by user"))
		return
	}
	c.JSON(http.StatusOK, prompts)
}

func (h *ChatPromptHandler) DeleteChatPromptByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	err := h.service.DeleteChatPromptByUUID(c.Request.Context(), uuid)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to delete chat prompt"))
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (h *ChatPromptHandler) UpdateChatPromptByUUID(c *gin.Context) {
	var simpleMsg SimpleChatMessage
	if err := c.ShouldBindJSON(&simpleMsg); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	prompt, err := h.service.UpdateChatPromptByUUID(c.Request.Context(), simpleMsg.Uuid, simpleMsg.Text)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update chat prompt"))
		return
	}
	c.JSON(http.StatusOK, prompt)
}

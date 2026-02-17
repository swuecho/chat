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

// GinRegister is an alias for Register for consistency with other handlers
func (h *ChatPromptHandler) GinRegister(rg *gin.RouterGroup) {
	h.Register(rg)
}

func (h *ChatPromptHandler) Register(rg *gin.RouterGroup) {
	rg.POST("/chat_prompts", h.CreateChatPrompt)
	rg.GET("/chat_prompts/users", h.GetChatPromptsByUserID)
	rg.GET("/chat_prompts/:id", h.GetChatPromptByID)
	rg.PUT("/chat_prompts/:id", h.UpdateChatPrompt)
	rg.DELETE("/chat_prompts/:id", h.DeleteChatPrompt)
	rg.GET("/chat_prompts", h.GetAllChatPrompts)
	rg.DELETE("/uuid/chat_prompts/:uuid", h.DeleteChatPromptByUUID)
	rg.PUT("/uuid/chat_prompts/:uuid", h.UpdateChatPromptByUUID)
}

func (h *ChatPromptHandler) CreateChatPrompt(c *gin.Context) {
	var promptParams sqlc_queries.CreateChatPromptParams
	if err := c.ShouldBindJSON(&promptParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	prompt, err := h.service.CreateChatPrompt(c.Request.Context(), promptParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to create chat prompt").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func (h *ChatPromptHandler) GetChatPromptByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid chat prompt ID").GinResponse(c)
		return
	}
	prompt, err := h.service.GetChatPromptByID(c.Request.Context(), int32(id))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat prompt").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func (h *ChatPromptHandler) UpdateChatPrompt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid chat prompt ID").GinResponse(c)
		return
	}
	var promptParams sqlc_queries.UpdateChatPromptParams
	if err := c.ShouldBindJSON(&promptParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	promptParams.ID = int32(id)
	prompt, err := h.service.UpdateChatPrompt(c.Request.Context(), promptParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update chat prompt").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, prompt)
}

func (h *ChatPromptHandler) DeleteChatPrompt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid chat prompt ID").GinResponse(c)
		return
	}
	err = h.service.DeleteChatPrompt(c.Request.Context(), int32(id))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to delete chat prompt").GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatPromptHandler) GetAllChatPrompts(c *gin.Context) {
	prompts, err := h.service.GetAllChatPrompts(c.Request.Context())
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat prompts").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, prompts)
}

func (h *ChatPromptHandler) GetChatPromptsByUserID(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ErrValidationInvalidInput("invalid user ID").GinResponse(c)
		return
	}
	prompts, err := h.service.GetChatPromptsByUserID(c.Request.Context(), int32(id))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get chat prompts by user").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, prompts)
}

func (h *ChatPromptHandler) DeleteChatPromptByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	err := h.service.DeleteChatPromptByUUID(c.Request.Context(), uuid)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to delete chat prompt").GinResponse(c)
		return
	}
	c.Status(http.StatusOK)
}

func (h *ChatPromptHandler) UpdateChatPromptByUUID(c *gin.Context) {
	var simple_msg SimpleChatMessage
	if err := c.ShouldBindJSON(&simple_msg); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	prompt, err := h.service.UpdateChatPromptByUUID(c.Request.Context(), simple_msg.Uuid, simple_msg.Text)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update chat prompt").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, prompt)
}

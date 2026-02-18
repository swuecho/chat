package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type BotAnswerHistoryHandler struct {
	service *BotAnswerHistoryService
}

func NewBotAnswerHistoryHandler(q *sqlc_queries.Queries) *BotAnswerHistoryHandler {
	service := NewBotAnswerHistoryService(q)
	return &BotAnswerHistoryHandler{service: service}
}

func (h *BotAnswerHistoryHandler) Register(router *gin.RouterGroup) {
	router.POST("/bot_answer_history", h.CreateBotAnswerHistory)
	router.GET("/bot_answer_history/:id", h.GetBotAnswerHistoryByID)
	router.GET("/bot_answer_history/bot/:bot_uuid", h.GetBotAnswerHistoryByBotUUID)
	router.GET("/bot_answer_history/user/:user_id", h.GetBotAnswerHistoryByUserID)
	router.PUT("/bot_answer_history/:id", h.UpdateBotAnswerHistory)
	router.DELETE("/bot_answer_history/:id", h.DeleteBotAnswerHistory)
	router.GET("/bot_answer_history/bot/:bot_uuid/count", h.GetBotAnswerHistoryCountByBotUUID)
	router.GET("/bot_answer_history/user/:user_id/count", h.GetBotAnswerHistoryCountByUserID)
	router.GET("/bot_answer_history/bot/:bot_uuid/latest", h.GetLatestBotAnswerHistoryByBotUUID)
}

func (h *BotAnswerHistoryHandler) CreateBotAnswerHistory(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	var params sqlc_queries.CreateBotAnswerHistoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error()))
		return
	}

	// Set the user ID from context
	params.UserID = userID

	history, err := h.service.CreateBotAnswerHistory(ctx, params)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to create bot answer history"))
		return
	}

	c.JSON(http.StatusCreated, history)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("ID is required"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid ID format"))
		return
	}

	history, err := h.service.GetBotAnswerHistoryByID(c.Request.Context(), int32(idInt))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get bot answer history"))
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByBotUUID(c *gin.Context) {
	botUUID := c.Param("bot_uuid")
	if botUUID == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Bot UUID is required"))
		return
	}

	limit, offset := getPaginationParamsGin(c)
	history, err := h.service.GetBotAnswerHistoryByBotUUID(c.Request.Context(), botUUID, limit, offset)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get bot answer history"))
		return
	}

	// Get total count for pagination
	totalCount, err := h.service.GetBotAnswerHistoryCountByBotUUID(c.Request.Context(), botUUID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get bot answer history count"))
		return
	}

	// Calculate total pages
	totalPages := totalCount / int64(limit)
	if totalCount%int64(limit) > 0 {
		totalPages++
	}

	// Return paginated response
	c.JSON(http.StatusOK, map[string]interface{}{
		"items":      history,
		"totalPages": totalPages,
		"totalCount": totalCount,
	})
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByUserID(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	limit, offset := getPaginationParamsGin(c)
	history, err := h.service.GetBotAnswerHistoryByUserID(ctx, userID, limit, offset)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get bot answer history"))
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) UpdateBotAnswerHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("ID is required"))
		return
	}

	var params sqlc_queries.UpdateBotAnswerHistoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error()))
		return
	}

	history, err := h.service.UpdateBotAnswerHistory(c.Request.Context(), params.ID, params.Answer, params.TokensUsed)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to update bot answer history"))
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) DeleteBotAnswerHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("ID is required"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid ID format"))
		return
	}

	if err := h.service.DeleteBotAnswerHistory(c.Request.Context(), int32(idInt)); err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to delete bot answer history"))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryCountByBotUUID(c *gin.Context) {
	botUUID := c.Param("bot_uuid")
	if botUUID == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Bot UUID is required"))
		return
	}

	count, err := h.service.GetBotAnswerHistoryCountByBotUUID(c.Request.Context(), botUUID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get bot answer history count"))
		return
	}

	c.JSON(http.StatusOK, map[string]int64{"count": count})
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryCountByUserID(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	count, err := h.service.GetBotAnswerHistoryCountByUserID(ctx, userID)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get bot answer history count"))
		return
	}

	c.JSON(http.StatusOK, map[string]int64{"count": count})
}

func (h *BotAnswerHistoryHandler) GetLatestBotAnswerHistoryByBotUUID(c *gin.Context) {
	botUUID := c.Param("bot_uuid")
	if botUUID == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Bot UUID is required"))
		return
	}

	limit := getLimitParamGin(c, 1)
	history, err := h.service.GetLatestBotAnswerHistoryByBotUUID(c.Request.Context(), botUUID, limit)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get latest bot answer history"))
		return
	}

	c.JSON(http.StatusOK, history)
}

// Gin versions of pagination helpers
func getPaginationParamsGin(c *gin.Context) (limit int32, offset int32) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit = 100 // default limit
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(l)
		}
	}

	offset = 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset = int32(o)
		}
	}

	return limit, offset
}

func getLimitParamGin(c *gin.Context, defaultLimit int32) int32 {
	limitStr := c.Query("limit")
	if limitStr == "" {
		return defaultLimit
	}
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		return defaultLimit
	}
	return int32(limit)
}

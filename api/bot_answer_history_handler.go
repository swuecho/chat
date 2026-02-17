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

// GinRegister registers routes with Gin router
func (h *BotAnswerHistoryHandler) GinRegister(rg *gin.RouterGroup) {
	rg.POST("/bot_answer_history", h.GinCreateBotAnswerHistory)
	rg.GET("/bot_answer_history/:id", h.GinGetBotAnswerHistoryByID)
	rg.GET("/bot_answer_history/bot/:bot_uuid", h.GinGetBotAnswerHistoryByBotUUID)
	rg.GET("/bot_answer_history/user/:user_id", h.GinGetBotAnswerHistoryByUserID)
	rg.PUT("/bot_answer_history/:id", h.GinUpdateBotAnswerHistory)
	rg.DELETE("/bot_answer_history/:id", h.GinDeleteBotAnswerHistory)
	rg.GET("/bot_answer_history/bot/:bot_uuid/count", h.GinGetBotAnswerHistoryCountByBotUUID)
	rg.GET("/bot_answer_history/user/:user_id/count", h.GinGetBotAnswerHistoryCountByUserID)
	rg.GET("/bot_answer_history/bot/:bot_uuid/latest", h.GinGetLatestBotAnswerHistoryByBotUUID)
}

func (h *BotAnswerHistoryHandler) GinCreateBotAnswerHistory(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	var params sqlc_queries.CreateBotAnswerHistoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	params.UserID = userID
	history, err := h.service.CreateBotAnswerHistory(c.Request.Context(), params)
	if err != nil {
		WrapError(err, "Failed to create bot answer history").GinResponse(c)
		return
	}

	c.JSON(http.StatusCreated, history)
}

func (h *BotAnswerHistoryHandler) GinGetBotAnswerHistoryByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ErrValidationInvalidInput("ID is required").GinResponse(c)
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ErrValidationInvalidInput("Invalid ID format").GinResponse(c)
		return
	}

	history, err := h.service.GetBotAnswerHistoryByID(c.Request.Context(), int32(idInt))
	if err != nil {
		WrapError(err, "Failed to get bot answer history").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) GinGetBotAnswerHistoryByBotUUID(c *gin.Context) {
	botUUID := c.Param("bot_uuid")
	if botUUID == "" {
		ErrValidationInvalidInput("Bot UUID is required").GinResponse(c)
		return
	}

	limit, offset := getGinPaginationParams(c)
	history, err := h.service.GetBotAnswerHistoryByBotUUID(c.Request.Context(), botUUID, limit, offset)
	if err != nil {
		WrapError(err, "Failed to get bot answer history").GinResponse(c)
		return
	}

	totalCount, err := h.service.GetBotAnswerHistoryCountByBotUUID(c.Request.Context(), botUUID)
	if err != nil {
		WrapError(err, "Failed to get bot answer history count").GinResponse(c)
		return
	}

	totalPages := totalCount / int64(limit)
	if totalCount%int64(limit) > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"items":      history,
		"totalPages": totalPages,
		"totalCount": totalCount,
	})
}

func (h *BotAnswerHistoryHandler) GinGetBotAnswerHistoryByUserID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	limit, offset := getGinPaginationParams(c)
	history, err := h.service.GetBotAnswerHistoryByUserID(c.Request.Context(), userID, limit, offset)
	if err != nil {
		WrapError(err, "Failed to get bot answer history").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) GinUpdateBotAnswerHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ErrValidationInvalidInput("ID is required").GinResponse(c)
		return
	}

	var params sqlc_queries.UpdateBotAnswerHistoryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	history, err := h.service.UpdateBotAnswerHistory(c.Request.Context(), params.ID, params.Answer, params.TokensUsed)
	if err != nil {
		WrapError(err, "Failed to update bot answer history").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) GinDeleteBotAnswerHistory(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ErrValidationInvalidInput("ID is required").GinResponse(c)
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ErrValidationInvalidInput("Invalid ID format").GinResponse(c)
		return
	}

	if err := h.service.DeleteBotAnswerHistory(c.Request.Context(), int32(idInt)); err != nil {
		WrapError(err, "Failed to delete bot answer history").GinResponse(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *BotAnswerHistoryHandler) GinGetBotAnswerHistoryCountByBotUUID(c *gin.Context) {
	botUUID := c.Param("bot_uuid")
	if botUUID == "" {
		ErrValidationInvalidInput("Bot UUID is required").GinResponse(c)
		return
	}

	count, err := h.service.GetBotAnswerHistoryCountByBotUUID(c.Request.Context(), botUUID)
	if err != nil {
		WrapError(err, "Failed to get bot answer history count").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]int64{"count": count})
}

func (h *BotAnswerHistoryHandler) GinGetBotAnswerHistoryCountByUserID(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	count, err := h.service.GetBotAnswerHistoryCountByUserID(c.Request.Context(), userID)
	if err != nil {
		WrapError(err, "Failed to get bot answer history count").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, map[string]int64{"count": count})
}

func (h *BotAnswerHistoryHandler) GinGetLatestBotAnswerHistoryByBotUUID(c *gin.Context) {
	botUUID := c.Param("bot_uuid")
	if botUUID == "" {
		ErrValidationInvalidInput("Bot UUID is required").GinResponse(c)
		return
	}

	limit := getGinLimitParam(c, 1)
	history, err := h.service.GetLatestBotAnswerHistoryByBotUUID(c.Request.Context(), botUUID, limit)
	if err != nil {
		WrapError(err, "Failed to get latest bot answer history").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, history)
}

// Helper functions for Gin pagination
func getGinPaginationParams(c *gin.Context) (int32, int32) {
	limit := int32(20)
	offset := int32(0)

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 32); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	return limit, offset
}

func getGinLimitParam(c *gin.Context, defaultLimit int) int32 {
	limit := defaultLimit
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	return int32(limit)
}

package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type AdminHandler struct {
	service *AuthUserService
}

func NewAdminHandler(service *AuthUserService) *AdminHandler {
	return &AdminHandler{
		service: service,
	}
}

type SessionHistoryResponse struct {
	Data  []SessionHistoryInfo `json:"data"`
	Total int64                `json:"total"`
	Page  int32                `json:"page"`
	Size  int32                `json:"size"`
}

// UserStat holds user statistics for admin dashboard
type UserStat struct {
	Email                            string `json:"email"`
	FirstName                        string `json:"first_name"`
	LastName                         string `json:"last_name"`
	TotalChatMessages                int64  `json:"total_chat_messages"`
	TotalChatMessages3Days           int64  `json:"total_chat_messages_3_days"`
	RateLimit                        int32  `json:"rate_limit"`
	TotalChatMessagesTokenCount      int64  `json:"total_chat_messages_token_count"`
	TotalChatMessages3DaysTokenCount int64  `json:"total_chat_messages_3_days_token_count"`
	AvgChatMessages3DaysTokenCount   int64  `json:"avg_chat_messages_3_days_token_count"`
}

// RateLimitRequest is the request body for updating user rate limits
type RateLimitRequest struct {
	Email     string `json:"email"`
	RateLimit int32  `json:"rate_limit"`
}

// GinRegisterRoutes registers routes with Gin router
func (h *AdminHandler) GinRegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/users", h.GinCreateUser)
	rg.PUT("/users", h.GinUpdateUser)
	rg.POST("/rate_limit", h.GinUpdateRateLimit)
	rg.POST("/user_stats", h.GinUserStatHandler)
	rg.GET("/user_analysis/:email", h.GinUserAnalysisHandler)
	rg.GET("/user_session_history/:email", h.GinUserSessionHistoryHandler)
	rg.GET("/session_messages/:sessionUuid", h.GinSessionMessagesHandler)
}

// =============================================================================
// Gin Handlers
// =============================================================================

func (h *AdminHandler) GinCreateUser(c *gin.Context) {
	var userParams sqlc_queries.CreateAuthUserParams
	if err := c.ShouldBindJSON(&userParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	user, err := h.service.CreateAuthUser(c.Request.Context(), userParams)
	if err != nil {
		WrapError(err, "Failed to create user").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) GinUpdateUser(c *gin.Context) {
	var userParams sqlc_queries.UpdateAuthUserByEmailParams
	if err := c.ShouldBindJSON(&userParams); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	user, err := h.service.q.UpdateAuthUserByEmail(c.Request.Context(), userParams)
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update user").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *AdminHandler) GinUserStatHandler(c *gin.Context) {
	var pagination Pagination
	if err := c.ShouldBindJSON(&pagination); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	userStatsRows, total, err := h.service.GetUserStats(c.Request.Context(), pagination, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to get user stats").GinResponse(c)
		return
	}

	data := make([]interface{}, len(userStatsRows))
	for i, v := range userStatsRows {
		divider := v.TotalChatMessages3Days
		var avg int64
		if divider > 0 {
			avg = v.TotalTokenCount3Days / v.TotalChatMessages3Days
		} else {
			avg = 0
		}
		data[i] = UserStat{
			Email:                            v.UserEmail,
			FirstName:                        v.FirstName,
			LastName:                         v.LastName,
			TotalChatMessages:                v.TotalChatMessages,
			TotalChatMessages3Days:           v.TotalChatMessages3Days,
			RateLimit:                        v.RateLimit,
			TotalChatMessagesTokenCount:      v.TotalTokenCount,
			TotalChatMessages3DaysTokenCount: v.TotalTokenCount3Days,
			AvgChatMessages3DaysTokenCount:   avg,
		}
	}

	c.JSON(http.StatusOK, Pagination{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Total: total,
		Data:  data,
	})
}

func (h *AdminHandler) GinUpdateRateLimit(c *gin.Context) {
	var rateLimitRequest RateLimitRequest
	if err := c.ShouldBindJSON(&rateLimitRequest); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	rate, err := h.service.q.UpdateAuthUserRateLimitByEmail(c.Request.Context(),
		sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
			Email:     rateLimitRequest.Email,
			RateLimit: rateLimitRequest.RateLimit,
		})

	if err != nil {
		WrapError(MapDatabaseError(err), "Failed to update rate limit").GinResponse(c)
		return
	}
	c.JSON(http.StatusOK, map[string]int32{"rate": rate})
}

func (h *AdminHandler) GinUserAnalysisHandler(c *gin.Context) {
	email := c.Param("email")

	if email == "" {
		ErrValidationInvalidInput("Email parameter is required").GinResponse(c)
		return
	}

	analysisData, err := h.service.GetUserAnalysis(c.Request.Context(), email, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		WrapError(err, "Failed to get user analysis").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, analysisData)
}

func (h *AdminHandler) GinUserSessionHistoryHandler(c *gin.Context) {
	email := c.Param("email")

	if email == "" {
		ErrValidationInvalidInput("Email parameter is required").GinResponse(c)
		return
	}

	// Parse pagination parameters
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	page := int32(1)
	size := int32(10)

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}

	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = int32(s)
		}
	}

	sessionHistory, total, err := h.service.GetUserSessionHistory(c.Request.Context(), email, page, size)
	if err != nil {
		WrapError(err, "Failed to get user session history").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, SessionHistoryResponse{
		Data:  sessionHistory,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func (h *AdminHandler) GinSessionMessagesHandler(c *gin.Context) {
	sessionUuid := c.Param("sessionUuid")

	if sessionUuid == "" {
		ErrValidationInvalidInput("Session UUID parameter is required").GinResponse(c)
		return
	}

	messages, err := h.service.q.GetChatMessagesBySessionUUIDForAdmin(c.Request.Context(), sessionUuid)
	if err != nil {
		WrapError(err, "Failed to get session messages").GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, messages)
}

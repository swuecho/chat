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

func (h *AdminHandler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/users", h.CreateUser)
	router.PUT("/users", h.UpdateUser)
	router.POST("/rate_limit", h.UpdateRateLimit)
	router.POST("/user_stats", h.UserStatHandler)
	router.GET("/user_analysis/:email", h.UserAnalysisHandler)
	router.GET("/user_session_history/:email", h.UserSessionHistoryHandler)
	router.GET("/session_messages/:sessionUuid", h.SessionMessagesHandler)
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var userParams sqlc_queries.CreateAuthUserParams
	err := c.ShouldBindJSON(&userParams)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.CreateAuthUser(c.Request.Context(), userParams)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to create user"))
		return
	}
	c.JSON(200, user)
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	var userParams sqlc_queries.UpdateAuthUserByEmailParams
	err := c.ShouldBindJSON(&userParams)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.q.UpdateAuthUserByEmail(c.Request.Context(), userParams)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update user"))
		return
	}
	c.JSON(200, user)
}

func (h *AdminHandler) UserStatHandler(c *gin.Context) {
	var pagination Pagination
	err := c.ShouldBindJSON(&pagination)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userStatsRows, total, err := h.service.GetUserStats(c.Request.Context(), pagination, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to get user stats"))
		return
	}

	// Create a new []interface{} slice with same length as userStatsRows
	data := make([]interface{}, len(userStatsRows))

	// Copy the contents of userStatsRows into data
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

	c.JSON(200, Pagination{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Total: total,
		Data:  data,
	})
}

func (h *AdminHandler) UpdateRateLimit(c *gin.Context) {
	var rateLimitRequest RateLimitRequest
	err := c.ShouldBindJSON(&rateLimitRequest)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	rate, err := h.service.q.UpdateAuthUserRateLimitByEmail(c.Request.Context(),
		sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
			Email:     rateLimitRequest.Email,
			RateLimit: rateLimitRequest.RateLimit,
		})

	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(MapDatabaseError(err), "Failed to update rate limit"))
		return
	}
	c.JSON(200, map[string]int32{
		"rate": rate,
	})
}

func (h *AdminHandler) UserAnalysisHandler(c *gin.Context) {
	email := c.Param("email")

	if email == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Email parameter is required"))
		return
	}

	analysisData, err := h.service.GetUserAnalysis(c.Request.Context(), email, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get user analysis"))
		return
	}

	c.JSON(200, analysisData)
}

type SessionHistoryResponse struct {
	Data  []SessionHistoryInfo `json:"data"`
	Total int64                `json:"total"`
	Page  int32                `json:"page"`
	Size  int32                `json:"size"`
}

func (h *AdminHandler) UserSessionHistoryHandler(c *gin.Context) {
	email := c.Param("email")

	if email == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Email parameter is required"))
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
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get user session history"))
		return
	}

	response := SessionHistoryResponse{
		Data:  sessionHistory,
		Total: total,
		Page:  page,
		Size:  size,
	}

	c.JSON(200, response)
}

func (h *AdminHandler) SessionMessagesHandler(c *gin.Context) {
	sessionUuid := c.Param("sessionUuid")

	if sessionUuid == "" {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Session UUID parameter is required"))
		return
	}

	messages, err := h.service.q.GetChatMessagesBySessionUUIDForAdmin(c.Request.Context(), sessionUuid)
	if err != nil {
		RespondWithAPIErrorGin(c, WrapError(err, "Failed to get session messages"))
		return
	}

	c.JSON(200, messages)
}

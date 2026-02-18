package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelHandler struct {
	db *sqlc_queries.Queries
}

func NewChatModelHandler(db *sqlc_queries.Queries) *ChatModelHandler {
	return &ChatModelHandler{
		db: db,
	}
}

func (h *ChatModelHandler) Register(r *gin.RouterGroup) {
	r.GET("/chat_model", h.ListSystemChatModels)
	r.GET("/chat_model/default", h.GetDefaultChatModel)
	r.GET("/chat_model/:id", h.ChatModelByID)
	r.POST("/chat_model", h.CreateChatModel)
	r.PUT("/chat_model/:id", h.UpdateChatModel)
	r.DELETE("/chat_model/:id", h.DeleteChatModel)
}

func (h *ChatModelHandler) ListSystemChatModels(c *gin.Context) {
	ctx := c.Request.Context()
	ChatModels, err := h.db.ListSystemChatModels(ctx)
	if err != nil {
		RespondWithAPIErrorGin(c, ErrInternalUnexpected.WithDetail("Failed to list chat models").WithDebugInfo(err.Error()))
		return
	}

	latestUsageTimeOfModels, err := h.db.GetLatestUsageTimeOfModel(ctx, "30 days")
	if err != nil {
		RespondWithAPIErrorGin(c, ErrInternalUnexpected.WithDetail("Failed to get model usage data").WithDebugInfo(err.Error()))
		return
	}
	// create a map of model id to usage time
	usageTimeMap := make(map[string]sqlc_queries.GetLatestUsageTimeOfModelRow)
	for _, usageTime := range latestUsageTimeOfModels {
		usageTimeMap[usageTime.Model] = usageTime
	}

	// create a ChatModelWithUsage struct
	type ChatModelWithUsage struct {
		sqlc_queries.ChatModel
		LastUsageTime time.Time `json:"lastUsageTime,omitempty"`
		MessageCount  int64     `json:"messageCount"`
	}

	// merge ChatModels and usageTimeMap with pre-allocated slice
	chatModelsWithUsage := lo.Map(ChatModels, func(model sqlc_queries.ChatModel, _ int) ChatModelWithUsage {
		usage := usageTimeMap[model.Name]
		return ChatModelWithUsage{
			ChatModel:     model,
			LastUsageTime: usage.LatestMessageTime,
			MessageCount:  usage.MessageCount,
		}
	})

	c.JSON(http.StatusOK, chatModelsWithUsage)
}

func (h *ChatModelHandler) ChatModelByID(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	ChatModel, err := h.db.ChatModelByID(ctx, int32(id))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrResourceNotFound("Chat model").WithDebugInfo(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ChatModel)
}

func (h *ChatModelHandler) CreateChatModel(c *gin.Context) {
	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	var input struct {
		Name                   string `json:"name"`
		Label                  string `json:"label"`
		IsDefault              bool   `json:"isDefault"`
		URL                    string `json:"url"`
		ApiAuthHeader          string `json:"apiAuthHeader"`
		ApiAuthKey             string `json:"apiAuthKey"`
		EnablePerModeRatelimit bool   `json:"enablePerModeRatelimit"`
		ApiType                string `json:"apiType"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}

	// Set default api_type if not provided
	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai" // default api type
	}

	// Validate api_type
	validApiTypes := map[string]bool{
		"openai": true,
		"claude": true,
		"gemini": true,
		"ollama": true,
		"custom": true,
	}

	if !validApiTypes[apiType] {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom"))
		return
	}

	ChatModel, err := h.db.CreateChatModel(c.Request.Context(), sqlc_queries.CreateChatModelParams{
		Name:                   input.Name,
		Label:                  input.Label,
		IsDefault:              input.IsDefault,
		Url:                    input.URL,
		ApiAuthHeader:          input.ApiAuthHeader,
		ApiAuthKey:             input.ApiAuthKey,
		UserID:                 userID,
		EnablePerModeRatelimit: input.EnablePerModeRatelimit,
		MaxToken:               4096, // default max token
		DefaultToken:           2048, // default token
		OrderNumber:            0,    // default order
		HttpTimeOut:            120,  // default timeout
		ApiType:                apiType,
	})

	if err != nil {
		RespondWithAPIErrorGin(c, ErrInternalUnexpected.WithDetail("Failed to create chat model").WithDebugInfo(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, ChatModel)
}

func (h *ChatModelHandler) UpdateChatModel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	var input struct {
		Name                   string `json:"name"`
		Label                  string `json:"label"`
		IsDefault              bool   `json:"isDefault"`
		URL                    string `json:"url"`
		ApiAuthHeader          string `json:"apiAuthHeader"`
		ApiAuthKey             string `json:"apiAuthKey"`
		EnablePerModeRatelimit bool   `json:"enablePerModeRatelimit"`
		OrderNumber            int32  `json:"orderNumber"`
		DefaultToken           int32  `json:"defaultToken"`
		MaxToken               int32  `json:"maxToken"`
		HttpTimeOut            int32  `json:"httpTimeOut"`
		IsEnable               bool   `json:"isEnable"`
		ApiType                string `json:"apiType"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}

	// Set default api_type if not provided
	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai" // default api type
	}

	// Validate api_type
	validApiTypes := map[string]bool{
		"openai": true,
		"claude": true,
		"gemini": true,
		"ollama": true,
		"custom": true,
	}

	if !validApiTypes[apiType] {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom"))
		return
	}

	ChatModel, err := h.db.UpdateChatModel(c.Request.Context(), sqlc_queries.UpdateChatModelParams{
		ID:                     int32(id),
		Name:                   input.Name,
		Label:                  input.Label,
		IsDefault:              input.IsDefault,
		Url:                    input.URL,
		ApiAuthHeader:          input.ApiAuthHeader,
		ApiAuthKey:             input.ApiAuthKey,
		UserID:                 userID,
		EnablePerModeRatelimit: input.EnablePerModeRatelimit,
		OrderNumber:            input.OrderNumber,
		DefaultToken:           input.DefaultToken,
		MaxToken:               input.MaxToken,
		HttpTimeOut:            input.HttpTimeOut,
		IsEnable:               input.IsEnable,
		ApiType:                apiType,
	})

	if err != nil {
		RespondWithAPIErrorGin(c, ErrInternalUnexpected.WithDetail("Failed to update chat model").WithDebugInfo(err.Error()))
		return
	}

	c.JSON(http.StatusOK, ChatModel)
}

func (h *ChatModelHandler) DeleteChatModel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		RespondWithAPIErrorGin(c, ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	err = h.db.DeleteChatModel(c.Request.Context(),
		sqlc_queries.DeleteChatModelParams{
			ID:     int32(id),
			UserID: userID,
		})
	if err != nil {
		RespondWithAPIErrorGin(c, ErrInternalUnexpected.WithDetail("Failed to delete chat model").WithDebugInfo(err.Error()))
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *ChatModelHandler) GetDefaultChatModel(c *gin.Context) {
	ChatModel, err := h.db.GetDefaultChatModel(c.Request.Context())
	if err != nil {
		RespondWithAPIErrorGin(c, ErrInternalUnexpected.WithDetail("Failed to retrieve default chat model").WithDebugInfo(err.Error()))
		return
	}
	c.JSON(http.StatusOK, ChatModel)
}

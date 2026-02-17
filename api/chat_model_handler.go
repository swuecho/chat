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

// GinRegister registers routes with Gin router
func (h *ChatModelHandler) GinRegister(rg *gin.RouterGroup) {
	rg.GET("/chat_model", h.GinListSystemChatModels)
	rg.GET("/chat_model/default", h.GinGetDefaultChatModel)
	rg.GET("/chat_model/:id", h.GinChatModelByID)
	rg.POST("/chat_model", h.GinCreateChatModel)
	rg.PUT("/chat_model/:id", h.GinUpdateChatModel)
	rg.DELETE("/chat_model/:id", h.GinDeleteChatModel)
}

func (h *ChatModelHandler) GinListSystemChatModels(c *gin.Context) {
	ctx := c.Request.Context()
	ChatModels, err := h.db.ListSystemChatModels(ctx)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to list chat models"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	latestUsageTimeOfModels, err := h.db.GetLatestUsageTimeOfModel(ctx, "30 days")
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to get model usage data"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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

	// merge ChatModels and usageTimeMap
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

func (h *ChatModelHandler) GinChatModelByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid chat model ID")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	ChatModel, err := h.db.ChatModelByID(c.Request.Context(), int32(id))
	if err != nil {
		apiErr := ErrResourceNotFound("Chat model")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, ChatModel)
}

func (h *ChatModelHandler) GinCreateChatModel(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	// Set default api_type if not provided
	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai"
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
		apiErr := ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom")
		apiErr.GinResponse(c)
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
		MaxToken:               4096,
		DefaultToken:           2048,
		OrderNumber:            0,
		HttpTimeOut:            120,
		ApiType:                apiType,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat model"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	c.JSON(http.StatusCreated, ChatModel)
}

func (h *ChatModelHandler) GinUpdateChatModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid chat model ID")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
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
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	// Set default api_type if not provided
	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai"
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
		apiErr := ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom")
		apiErr.GinResponse(c)
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
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat model"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, ChatModel)
}

func (h *ChatModelHandler) GinDeleteChatModel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid chat model ID")
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	userID, err := GetUserID(c)
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	err = h.db.DeleteChatModel(c.Request.Context(), sqlc_queries.DeleteChatModelParams{
		ID:     int32(id),
		UserID: userID,
	})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat model"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	c.Status(http.StatusOK)
}

func (h *ChatModelHandler) GinGetDefaultChatModel(c *gin.Context) {
	ChatModel, err := h.db.GetDefaultChatModel(c.Request.Context())
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve default chat model"
		apiErr.DebugInfo = err.Error()
		apiErr.GinResponse(c)
		return
	}

	c.JSON(http.StatusOK, ChatModel)
}

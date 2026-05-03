package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelHandler struct {
	db *sqlc_queries.Queries
}

func NewChatModelHandler(db *sqlc_queries.Queries) *ChatModelHandler {
	return &ChatModelHandler{db: db}
}

func (h *ChatModelHandler) Register(r *mux.Router) {
	r.HandleFunc("/chat_model", h.ListSystemChatModels).Methods("GET")
	r.HandleFunc("/chat_model/default", h.GetDefaultChatModel).Methods("GET")
	r.HandleFunc("/chat_model/{id}", h.ChatModelByID).Methods("GET")
	r.HandleFunc("/chat_model", h.CreateChatModel).Methods("POST")
	r.HandleFunc("/chat_model/{id}", h.UpdateChatModel).Methods("PUT")
	r.HandleFunc("/chat_model/{id}", h.DeleteChatModel).Methods("DELETE")
}

func (h *ChatModelHandler) ListSystemChatModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chatModels, err := h.db.ListSystemChatModels(ctx)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to list chat models").WithDebugInfo(err.Error()))
		return
	}

	latestUsageTimeOfModels, err := h.db.GetLatestUsageTimeOfModel(ctx, "30 days")
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to get model usage data").WithDebugInfo(err.Error()))
		return
	}

	usageTimeMap := make(map[string]sqlc_queries.GetLatestUsageTimeOfModelRow)
	for _, usageTime := range latestUsageTimeOfModels {
		usageTimeMap[usageTime.Model] = usageTime
	}

	type ChatModelWithUsage struct {
		sqlc_queries.ChatModel
		LastUsageTime time.Time `json:"lastUsageTime,omitempty"`
		MessageCount  int64     `json:"messageCount"`
	}

	chatModelsWithUsage := lo.Map(chatModels, func(model sqlc_queries.ChatModel, _ int) ChatModelWithUsage {
		usage := usageTimeMap[model.Name]
		return ChatModelWithUsage{
			ChatModel:     model,
			LastUsageTime: usage.LatestMessageTime,
			MessageCount:  usage.MessageCount,
		}
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModelsWithUsage)
}

func (h *ChatModelHandler) ChatModelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	chatModel, err := h.db.ChatModelByID(ctx, int32(id))
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat model").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModel)
}

func (h *ChatModelHandler) CreateChatModel(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
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

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}

	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai"
	}

	validApiTypes := map[string]bool{
		"openai": true, "claude": true, "gemini": true, "ollama": true, "custom": true,
	}
	if !validApiTypes[apiType] {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom"))
		return
	}

	chatModel, err := h.db.CreateChatModel(r.Context(), sqlc_queries.CreateChatModelParams{
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
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to create chat model").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chatModel)
}

func (h *ChatModelHandler) UpdateChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
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
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}

	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai"
	}

	validApiTypes := map[string]bool{
		"openai": true, "claude": true, "gemini": true, "ollama": true, "custom": true,
	}
	if !validApiTypes[apiType] {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom"))
		return
	}

	chatModel, err := h.db.UpdateChatModel(r.Context(), sqlc_queries.UpdateChatModelParams{
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
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update chat model").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModel)
}

func (h *ChatModelHandler) DeleteChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if err := h.db.DeleteChatModel(r.Context(), sqlc_queries.DeleteChatModelParams{
		ID: int32(id), UserID: userID,
	}); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to delete chat model").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ChatModelHandler) GetDefaultChatModel(w http.ResponseWriter, r *http.Request) {
	chatModel, err := h.db.GetDefaultChatModel(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve default chat model").WithDebugInfo(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModel)
}

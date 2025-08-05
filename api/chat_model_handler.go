package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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

func (h *ChatModelHandler) Register(r *mux.Router) {

	// Assuming db is an instance of the SQLC generated DB struct
	//handler := NewChatModelHandler(db)
	// r := mux.NewRouter()

	// TODO: user can read, remove user_id field from the response
	r.HandleFunc("/chat_model", h.ListSystemChatModels).Methods("GET")
	r.HandleFunc("/chat_model/default", h.GetDefaultChatModel).Methods("GET")
	r.HandleFunc("/chat_model/{id}", h.ChatModelByID).Methods("GET")
	// create delete update self's chat model
	r.HandleFunc("/chat_model", h.CreateChatModel).Methods("POST")
	r.HandleFunc("/chat_model/{id}", h.UpdateChatModel).Methods("PUT")
	r.HandleFunc("/chat_model/{id}", h.DeleteChatModel).Methods("DELETE")
}

func (h *ChatModelHandler) ListSystemChatModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ChatModels, err := h.db.ListSystemChatModels(ctx)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to list chat models"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	latestUsageTimeOfModels, err := h.db.GetLatestUsageTimeOfModel(ctx, "30 days")
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to get model usage data"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModelsWithUsage)
}

func (h *ChatModelHandler) ChatModelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid chat model ID")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	ChatModel, err := h.db.ChatModelByID(ctx, int32(id))
	if err != nil {
		apiErr := ErrResourceNotFound("Chat model")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatModel)
}

func (h *ChatModelHandler) CreateChatModel(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
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

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Set default api_type if not provided
	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai" // default api type
	}

	ChatModel, err := h.db.CreateChatModel(r.Context(), sqlc_queries.CreateChatModelParams{
		Name:                   input.Name,
		Label:                  input.Label,
		IsDefault:              input.IsDefault,
		Url:                    input.URL,
		ApiAuthHeader:          input.ApiAuthHeader,
		ApiAuthKey:             input.ApiAuthKey,
		UserID:                 userID,
		EnablePerModeRatelimit: input.EnablePerModeRatelimit,
		MaxToken:               4096,  // default max token
		DefaultToken:           2048,  // default token
		OrderNumber:            0,     // default order
		HttpTimeOut:            120,   // default timeout
		ApiType:               apiType,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create chat model"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ChatModel)
}

func (h *ChatModelHandler) UpdateChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid chat model ID")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
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
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		apiErr := ErrValidationInvalidInput("Failed to parse request body")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// Set default api_type if not provided
	apiType := input.ApiType
	if apiType == "" {
		apiType = "openai" // default api type
	}

	ChatModel, err := h.db.UpdateChatModel(r.Context(), sqlc_queries.UpdateChatModelParams{
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
		ApiType:               apiType,
	})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update chat model"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatModel)
}

func (h *ChatModelHandler) DeleteChatModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		apiErr := ErrValidationInvalidInput("Invalid chat model ID")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	err = h.db.DeleteChatModel(r.Context(),
		sqlc_queries.DeleteChatModelParams{
			ID:     int32(id),
			UserID: userID,
		})
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to delete chat model"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ChatModelHandler) GetDefaultChatModel(w http.ResponseWriter, r *http.Request) {
	ChatModel, err := h.db.GetDefaultChatModel(r.Context())
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to retrieve default chat model"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ChatModel)
}

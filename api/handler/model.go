package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatModelHandler struct {
	service *svc.ChatModelService
}

func NewChatModelHandler(service *svc.ChatModelService) *ChatModelHandler {
	return &ChatModelHandler{service: service}
}

func (h *ChatModelHandler) Register(r *mux.Router) {
	r.HandleFunc("/chat_model", h.ListSystemChatModels).Methods("GET")
	r.HandleFunc("/chat_model/default", h.GetDefaultChatModel).Methods("GET")
	r.HandleFunc("/chat_model/title-default", h.GetTitleChatModel).Methods("GET")
	r.HandleFunc("/chat_model/title-default", h.SetTitleChatModel).Methods("PUT")
	r.HandleFunc("/chat_model/{id}", h.ChatModelByID).Methods("GET")
	r.HandleFunc("/chat_model", h.CreateChatModel).Methods("POST")
	r.HandleFunc("/chat_model/{id}", h.UpdateChatModel).Methods("PUT")
	r.HandleFunc("/chat_model/{id}", h.DeleteChatModel).Methods("DELETE")
}

func (h *ChatModelHandler) ListSystemChatModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chatModels, err := h.service.ListSystemWithUsage(ctx, "30 days")
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to list chat models").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModels)
}

func (h *ChatModelHandler) ChatModelByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ctx := r.Context()
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid chat model ID").WithDebugInfo(err.Error()))
		return
	}

	chatModel, err := h.service.ByID(ctx, int32(id))
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

	var request createChatModelRequest

	if err := DecodeJSON(r, &request); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}

	apiType := request.APIType
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

	chatModel, err := h.service.Create(r.Context(), request.createInput(userID, apiType))
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

	var request createChatModelRequest
	if err := DecodeJSON(r, &request); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error()))
		return
	}

	apiType := request.APIType
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

	chatModel, err := h.service.Update(r.Context(), request.updateInput(int32(id), userID, apiType))
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

	if err := h.service.Delete(r.Context(), int32(id), userID); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to delete chat model").WithDebugInfo(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ChatModelHandler) GetDefaultChatModel(w http.ResponseWriter, r *http.Request) {
	chatModel, err := h.service.Default(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve default chat model").WithDebugInfo(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chatModel)
}

func (h *ChatModelHandler) GetTitleChatModel(w http.ResponseWriter, r *http.Request) {
	chatModel, err := h.service.TitleModel(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Title generation model").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(chatModel)
}

func (h *ChatModelHandler) SetTitleChatModel(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	var input struct {
		ModelID int32 `json:"modelId"`
	}
	if err := DecodeJSON(r, &input); err != nil || input.ModelID <= 0 {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("A valid enabled model is required"))
		return
	}

	chatModel, err := h.service.SetTitleModel(r.Context(), input.ModelID, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("The title model must be an enabled model you manage").WithDebugInfo(err.Error()))
		return
	}
	json.NewEncoder(w).Encode(chatModel)
}

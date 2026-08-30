package handler

import (
	"net/http"

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
	r.HandleFunc("/chat_model", endpoint(h.ListSystemChatModels)).Methods(http.MethodGet)
	r.HandleFunc("/chat_model/default", endpoint(h.GetDefaultChatModel)).Methods(http.MethodGet)
	r.HandleFunc("/chat_model/title-default", endpoint(h.GetTitleChatModel)).Methods(http.MethodGet)
	r.HandleFunc("/chat_model/title-default", endpoint(h.SetTitleChatModel)).Methods(http.MethodPut)
	r.HandleFunc("/chat_model/{id}", endpoint(h.ChatModelByID)).Methods(http.MethodGet)
	r.HandleFunc("/chat_model", endpoint(h.CreateChatModel)).Methods(http.MethodPost)
	r.HandleFunc("/chat_model/{id}", endpoint(h.UpdateChatModel)).Methods(http.MethodPut)
	r.HandleFunc("/chat_model/{id}", endpoint(h.DeleteChatModel)).Methods(http.MethodDelete)
}

func (h *ChatModelHandler) ListSystemChatModels(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	chatModels, err := h.service.ListSystemWithUsage(ctx, "30 days")
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to list chat models").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, chatModelResponses(chatModels))
}

func (h *ChatModelHandler) ChatModelByID(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	chatModel, err := h.service.ByID(r.Context(), id)
	if err != nil {
		return dto.ErrResourceNotFound("Chat model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, chatModelResponse(chatModel))
}

func (h *ChatModelHandler) CreateChatModel(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var request createChatModelRequest

	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error())
	}

	apiType := request.APIType
	if apiType == "" {
		apiType = "openai"
	}

	validApiTypes := map[string]bool{
		"openai": true, "claude": true, "gemini": true, "ollama": true, "custom": true,
	}
	if !validApiTypes[apiType] {
		return dto.ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom")
	}

	chatModel, err := h.service.Create(r.Context(), request.createInput(userID, apiType))
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to create chat model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusCreated, chatModelResponse(chatModel))
}

func (h *ChatModelHandler) UpdateChatModel(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var request createChatModelRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to parse request body").WithDebugInfo(err.Error())
	}

	apiType := request.APIType
	if apiType == "" {
		apiType = "openai"
	}

	validApiTypes := map[string]bool{
		"openai": true, "claude": true, "gemini": true, "ollama": true, "custom": true,
	}
	if !validApiTypes[apiType] {
		return dto.ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom")
	}

	chatModel, err := h.service.Update(r.Context(), request.updateInput(id, userID, apiType))
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to update chat model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, chatModelResponse(chatModel))
}

func (h *ChatModelHandler) DeleteChatModel(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to delete chat model").WithDebugInfo(err.Error())
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatModelHandler) GetDefaultChatModel(w http.ResponseWriter, r *http.Request) error {
	chatModel, err := h.service.Default(r.Context())
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to retrieve default chat model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, chatModelResponse(chatModel))
}

func (h *ChatModelHandler) GetTitleChatModel(w http.ResponseWriter, r *http.Request) error {
	chatModel, err := h.service.TitleModel(r.Context())
	if err != nil {
		return dto.ErrResourceNotFound("Title generation model").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, chatModelResponse(chatModel))
}

func (h *ChatModelHandler) SetTitleChatModel(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var input setTitleModelRequest
	if err := DecodeJSON(r, &input); err != nil {
		return dto.ErrValidationInvalidInput("A valid enabled model is required").WithDebugInfo(err.Error())
	}

	chatModel, err := h.service.SetTitleModel(r.Context(), input.ModelID, userID)
	if err != nil {
		return dto.ErrValidationInvalidInput("The title model must be an enabled model you manage").WithDebugInfo(err.Error())
	}
	return respondJSON(w, http.StatusOK, chatModelResponse(chatModel))
}

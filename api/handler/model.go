package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatModelHandler struct {
	service *svc.ChatModelService
}

func NewChatModelHandler(service *svc.ChatModelService) *ChatModelHandler {
	return &ChatModelHandler{service: service}
}

func (h *ChatModelHandler) Register(r *mux.Router, registry *apicontract.Registry) {
	secured := func(op apicontract.Operation) apicontract.Operation {
		op.Tags = []string{"Models"}
		op.Security = apicontract.BearerAuth()
		return op
	}
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/chat_model", OperationID: "listChatModels", Summary: "List chat models", SuccessStatus: http.StatusOK}), h.listSystemChatModels)
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/chat_model/default", OperationID: "getDefaultChatModel", Summary: "Get the default chat model", SuccessStatus: http.StatusOK}), h.getDefaultChatModel)
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/chat_model/title-default", OperationID: "getTitleChatModel", Summary: "Get the title generation model", SuccessStatus: http.StatusOK}), h.getTitleChatModel)
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodPut, Path: "/chat_model/title-default", OperationID: "setTitleChatModel", Summary: "Set the title generation model", SuccessStatus: http.StatusOK}), h.setTitleChatModel)
	idParameter := []apicontract.Parameter{apicontract.PositiveIntegerPathParameter("id")}
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodGet, Path: "/chat_model/{id}", OperationID: "getChatModel", Summary: "Get a chat model", SuccessStatus: http.StatusOK, Parameters: idParameter}), h.chatModelByID)
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodPost, Path: "/chat_model", OperationID: "createChatModel", Summary: "Create a chat model", SuccessStatus: http.StatusCreated}), h.createChatModel)
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodPut, Path: "/chat_model/{id}", OperationID: "updateChatModel", Summary: "Update a chat model", SuccessStatus: http.StatusOK, Parameters: idParameter}), h.updateChatModel)
	apicontract.RegisterJSON(r, registry, secured(apicontract.Operation{Method: http.MethodDelete, Path: "/chat_model/{id}", OperationID: "deleteChatModel", Summary: "Delete a chat model", SuccessStatus: http.StatusOK, Parameters: idParameter}), h.deleteChatModel)
}

func (h *ChatModelHandler) listSystemChatModels(r *http.Request, _ apicontract.NoBody) ([]chatModelWithUsageHTTPResponse, error) {
	ctx := r.Context()
	chatModels, err := h.service.ListSystemWithUsage(ctx, "30 days")
	if err != nil {
		return nil, dto.ErrInternalUnexpected.WithDetail("Failed to list chat models").WithDebugInfo(err.Error())
	}
	return chatModelResponses(chatModels), nil
}

func (h *ChatModelHandler) chatModelByID(r *http.Request, _ apicontract.NoBody) (chatModelHTTPResponse, error) {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return chatModelHTTPResponse{}, err
	}
	chatModel, err := h.service.ByID(r.Context(), id)
	if err != nil {
		return chatModelHTTPResponse{}, dto.ErrResourceNotFound("Chat model").WithDebugInfo(err.Error())
	}
	return chatModelResponse(chatModel), nil
}

func (h *ChatModelHandler) createChatModel(r *http.Request, request createChatModelRequest) (chatModelHTTPResponse, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatModelHTTPResponse{}, err
	}

	apiType := request.APIType
	if apiType == "" {
		apiType = "openai"
	}

	validApiTypes := map[string]bool{
		"openai": true, "claude": true, "gemini": true, "ollama": true, "custom": true,
	}
	if !validApiTypes[apiType] {
		return chatModelHTTPResponse{}, dto.ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom")
	}

	chatModel, err := h.service.Create(r.Context(), request.createInput(userID, apiType))
	if err != nil {
		return chatModelHTTPResponse{}, dto.ErrInternalUnexpected.WithDetail("Failed to create chat model").WithDebugInfo(err.Error())
	}
	return chatModelResponse(chatModel), nil
}

func (h *ChatModelHandler) updateChatModel(r *http.Request, request createChatModelRequest) (chatModelHTTPResponse, error) {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return chatModelHTTPResponse{}, err
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatModelHTTPResponse{}, err
	}

	apiType := request.APIType
	if apiType == "" {
		apiType = "openai"
	}

	validApiTypes := map[string]bool{
		"openai": true, "claude": true, "gemini": true, "ollama": true, "custom": true,
	}
	if !validApiTypes[apiType] {
		return chatModelHTTPResponse{}, dto.ErrValidationInvalidInput("Invalid API type. Valid types are: openai, claude, gemini, ollama, custom")
	}

	chatModel, err := h.service.Update(r.Context(), request.updateInput(id, userID, apiType))
	if err != nil {
		return chatModelHTTPResponse{}, dto.ErrInternalUnexpected.WithDetail("Failed to update chat model").WithDebugInfo(err.Error())
	}
	return chatModelResponse(chatModel), nil
}

func (h *ChatModelHandler) deleteChatModel(r *http.Request, _ apicontract.NoBody) (apicontract.NoBody, error) {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return apicontract.NoBody{}, err
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return apicontract.NoBody{}, err
	}

	if err := h.service.Delete(r.Context(), id, userID); err != nil {
		return apicontract.NoBody{}, dto.ErrInternalUnexpected.WithDetail("Failed to delete chat model").WithDebugInfo(err.Error())
	}
	return apicontract.NoBody{}, nil
}

func (h *ChatModelHandler) getDefaultChatModel(r *http.Request, _ apicontract.NoBody) (chatModelHTTPResponse, error) {
	chatModel, err := h.service.Default(r.Context())
	if err != nil {
		return chatModelHTTPResponse{}, dto.ErrInternalUnexpected.WithDetail("Failed to retrieve default chat model").WithDebugInfo(err.Error())
	}
	return chatModelResponse(chatModel), nil
}

func (h *ChatModelHandler) getTitleChatModel(r *http.Request, _ apicontract.NoBody) (chatModelHTTPResponse, error) {
	chatModel, err := h.service.TitleModel(r.Context())
	if err != nil {
		return chatModelHTTPResponse{}, dto.ErrResourceNotFound("Title generation model").WithDebugInfo(err.Error())
	}
	return chatModelResponse(chatModel), nil
}

func (h *ChatModelHandler) setTitleChatModel(r *http.Request, input setTitleModelRequest) (chatModelHTTPResponse, error) {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return chatModelHTTPResponse{}, err
	}

	chatModel, err := h.service.SetTitleModel(r.Context(), input.ModelID, userID)
	if err != nil {
		return chatModelHTTPResponse{}, dto.ErrValidationInvalidInput("The title model must be an enabled model you manage").WithDebugInfo(err.Error())
	}
	return chatModelResponse(chatModel), nil
}

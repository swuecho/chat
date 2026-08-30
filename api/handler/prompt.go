package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type ChatPromptHandler struct {
	service *svc.ChatPromptService
}

func NewChatPromptHandler(service *svc.ChatPromptService) *ChatPromptHandler {
	return &ChatPromptHandler{service: service}
}

func (h *ChatPromptHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_prompts", endpoint(h.CreateChatPrompt)).Methods(http.MethodPost)
	router.HandleFunc("/chat_prompts/users", endpoint(h.GetChatPromptsByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/chat_prompts/{id}", endpoint(h.GetChatPromptByID)).Methods(http.MethodGet)
	router.HandleFunc("/chat_prompts/{id}", endpoint(h.UpdateChatPrompt)).Methods(http.MethodPut)
	router.HandleFunc("/chat_prompts/{id}", endpoint(h.DeleteChatPrompt)).Methods(http.MethodDelete)
	router.HandleFunc("/chat_prompts", endpoint(h.GetAllChatPrompts)).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_prompts/{uuid}", endpoint(h.DeleteChatPromptByUUID)).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_prompts/{uuid}", endpoint(h.UpdateChatPromptByUUID)).Methods(http.MethodPut)
}

func (h *ChatPromptHandler) CreateChatPrompt(w http.ResponseWriter, r *http.Request) error {
	var request chatPromptRequest
	err := DecodeJSON(r, &request)
	if err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	promptParams := svc.CreateChatPromptInput{UUID: request.UUID, ChatSessionUUID: request.ChatSessionUUID,
		Role: request.Role, Content: request.Content, TokenCount: request.TokenCount,
		UserID: userID, CreatedBy: userID, UpdatedBy: userID}

	if promptParams.ChatSessionUUID != "" && promptParams.Role == "system" {
		existingPrompt, getErr := h.service.GetOneChatPromptBySessionUUID(r.Context(), promptParams.ChatSessionUUID)
		if getErr == nil {
			return respondJSON(w, http.StatusOK, promptResponse(existingPrompt))
		}
		if !errors.Is(getErr, sql.ErrNoRows) {
			return dto.WrapError(dto.MapDatabaseError(getErr), "Failed to check existing chat prompt")
		}
	}

	prompt, err := h.service.CreateChatPrompt(r.Context(), promptParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if promptParams.ChatSessionUUID != "" && promptParams.Role == "system" &&
			errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existingPrompt, getErr := h.service.GetOneChatPromptBySessionUUID(r.Context(), promptParams.ChatSessionUUID)
			if getErr == nil {
				return respondJSON(w, http.StatusOK, promptResponse(existingPrompt))
			}
		}
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat prompt")
	}
	return respondJSON(w, http.StatusCreated, promptResponse(prompt))
}

func (h *ChatPromptHandler) GetChatPromptByID(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	prompt, err := h.service.GetChatPromptByID(r.Context(), id, userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat prompt")
	}
	return respondJSON(w, http.StatusOK, promptResponse(prompt))
}

func (h *ChatPromptHandler) UpdateChatPrompt(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	var request chatPromptRequest
	err = DecodeJSON(r, &request)
	if err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	promptParams := svc.UpdateChatPromptInput{ID: id, ChatSessionUUID: request.ChatSessionUUID,
		Role: request.Role, Content: request.Content, Score: request.Score}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	promptParams.UserID, promptParams.UpdatedBy = userID, userID
	prompt, err := h.service.UpdateChatPrompt(r.Context(), promptParams)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat prompt")
	}
	return respondJSON(w, http.StatusOK, promptResponse(prompt))
}

func (h *ChatPromptHandler) DeleteChatPrompt(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	err = h.service.DeleteChatPrompt(r.Context(), svc.DeleteChatPromptCommand{ID: id, UserID: userID})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat prompt")
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatPromptHandler) GetAllChatPrompts(w http.ResponseWriter, r *http.Request) error {
	prompts, err := h.service.GetAllChatPrompts(r.Context())
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat prompts")
	}
	return respondJSON(w, http.StatusOK, promptResponses(prompts))
}

func (h *ChatPromptHandler) GetChatPromptsByUserID(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	prompts, err := h.service.GetChatPromptsByUserID(r.Context(), userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat prompts by user")
	}
	return respondJSON(w, http.StatusOK, promptResponses(prompts))
}

func (h *ChatPromptHandler) DeleteChatPromptByUUID(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["uuid"]
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	err = h.service.DeleteChatPromptByUUID(r.Context(), svc.DeleteChatPromptByUUIDCommand{UUID: idStr, UserID: userID})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat prompt")
	}
	return respondStatus(w, http.StatusOK)
}

func (h *ChatPromptHandler) UpdateChatPromptByUUID(w http.ResponseWriter, r *http.Request) error {
	var simpleMsg dto.SimpleChatMessage
	err := DecodeJSON(r, &simpleMsg)
	if err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	prompt, err := h.service.UpdateChatPromptByUUID(r.Context(), svc.UpdateChatPromptByUUIDCommand{UUID: simpleMsg.Uuid, Content: simpleMsg.Text, UserID: userID})
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat prompt")
	}
	return respondJSON(w, http.StatusOK, promptResponse(prompt))
}

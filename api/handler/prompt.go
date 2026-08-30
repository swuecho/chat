package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
	router.HandleFunc("/chat_prompts", h.CreateChatPrompt).Methods(http.MethodPost)
	router.HandleFunc("/chat_prompts/users", h.GetChatPromptsByUserID).Methods(http.MethodGet)
	router.HandleFunc("/chat_prompts/{id}", h.GetChatPromptByID).Methods(http.MethodGet)
	router.HandleFunc("/chat_prompts/{id}", h.UpdateChatPrompt).Methods(http.MethodPut)
	router.HandleFunc("/chat_prompts/{id}", h.DeleteChatPrompt).Methods(http.MethodDelete)
	router.HandleFunc("/chat_prompts", h.GetAllChatPrompts).Methods(http.MethodGet)
	router.HandleFunc("/uuid/chat_prompts/{uuid}", h.DeleteChatPromptByUUID).Methods(http.MethodDelete)
	router.HandleFunc("/uuid/chat_prompts/{uuid}", h.UpdateChatPromptByUUID).Methods(http.MethodPut)
}

func (h *ChatPromptHandler) CreateChatPrompt(w http.ResponseWriter, r *http.Request) {
	var request chatPromptRequest
	err := DecodeJSON(r, &request)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	promptParams := svc.CreateChatPromptInput{UUID: request.UUID, ChatSessionUUID: request.ChatSessionUUID,
		Role: request.Role, Content: request.Content, TokenCount: request.TokenCount,
		UserID: userID, CreatedBy: userID, UpdatedBy: userID}

	if promptParams.ChatSessionUUID != "" && promptParams.Role == "system" {
		existingPrompt, getErr := h.service.GetOneChatPromptBySessionUUID(r.Context(), promptParams.ChatSessionUUID)
		if getErr == nil {
			json.NewEncoder(w).Encode(promptResponse(existingPrompt))
			return
		}
		if !errors.Is(getErr, sql.ErrNoRows) {
			dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(getErr), "Failed to check existing chat prompt"))
			return
		}
	}

	prompt, err := h.service.CreateChatPrompt(r.Context(), promptParams)
	if err != nil {
		var pgErr *pgconn.PgError
		if promptParams.ChatSessionUUID != "" && promptParams.Role == "system" &&
			errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existingPrompt, getErr := h.service.GetOneChatPromptBySessionUUID(r.Context(), promptParams.ChatSessionUUID)
			if getErr == nil {
				json.NewEncoder(w).Encode(promptResponse(existingPrompt))
				return
			}
		}

		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to create chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(promptResponse(prompt))
}

func (h *ChatPromptHandler) GetChatPromptByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	prompt, err := h.service.GetChatPromptByID(r.Context(), int32(id), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(promptResponse(prompt))
}

func (h *ChatPromptHandler) UpdateChatPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	var request chatPromptRequest
	err = DecodeJSON(r, &request)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	promptParams := svc.UpdateChatPromptInput{ID: int32(id), ChatSessionUUID: request.ChatSessionUUID,
		Role: request.Role, Content: request.Content, Score: request.Score}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	promptParams.UserID, promptParams.UpdatedBy = userID, userID
	prompt, err := h.service.UpdateChatPrompt(r.Context(), promptParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(promptResponse(prompt))
}

func (h *ChatPromptHandler) DeleteChatPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	err = h.service.DeleteChatPrompt(r.Context(), svc.DeleteChatPromptCommand{ID: int32(id), UserID: userID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat prompt"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatPromptHandler) GetAllChatPrompts(w http.ResponseWriter, r *http.Request) {
	prompts, err := h.service.GetAllChatPrompts(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat prompts"))
		return
	}
	json.NewEncoder(w).Encode(promptResponses(prompts))
}

func (h *ChatPromptHandler) GetChatPromptsByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	prompts, err := h.service.GetChatPromptsByUserID(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get chat prompts by user"))
		return
	}
	json.NewEncoder(w).Encode(promptResponses(prompts))
}

func (h *ChatPromptHandler) DeleteChatPromptByUUID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["uuid"]
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	err = h.service.DeleteChatPromptByUUID(r.Context(), svc.DeleteChatPromptByUUIDCommand{UUID: idStr, UserID: userID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to delete chat prompt"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatPromptHandler) UpdateChatPromptByUUID(w http.ResponseWriter, r *http.Request) {
	var simpleMsg dto.SimpleChatMessage
	err := DecodeJSON(r, &simpleMsg)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}
	prompt, err := h.service.UpdateChatPromptByUUID(r.Context(), svc.UpdateChatPromptByUUIDCommand{UUID: simpleMsg.Uuid, Content: simpleMsg.Text, UserID: userID})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(promptResponse(prompt))
}

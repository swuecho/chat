package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgconn"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatPromptHandler struct {
	service *ChatPromptService
}

func NewChatPromptHandler(sqlc_q *sqlc_queries.Queries) *ChatPromptHandler {
	promptService := NewChatPromptService(sqlc_q)
	return &ChatPromptHandler{
		service: promptService,
	}
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
	var promptParams sqlc_queries.CreateChatPromptParams
	err := json.NewDecoder(r.Body).Decode(&promptParams)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	// Always trust authenticated user identity over client-provided values.
	promptParams.UserID = userID
	promptParams.CreatedBy = userID
	promptParams.UpdatedBy = userID

	// Idempotent creation for session system prompt:
	// return existing prompt instead of inserting duplicates when concurrent
	// frontend/backend requests race on a fresh session.
	if promptParams.ChatSessionUuid != "" && promptParams.Role == "system" {
		existingPrompt, getErr := h.service.q.GetOneChatPromptBySessionUUID(r.Context(), promptParams.ChatSessionUuid)
		if getErr == nil {
			json.NewEncoder(w).Encode(existingPrompt)
			return
		}
		if !errors.Is(getErr, sql.ErrNoRows) {
			RespondWithAPIError(w, WrapError(MapDatabaseError(getErr), "Failed to check existing chat prompt"))
			return
		}
	}

	prompt, err := h.service.CreateChatPrompt(r.Context(), promptParams)
	if err != nil {
		// Handle race: another request inserted the same session system prompt
		// between our read check and insert attempt.
		var pgErr *pgconn.PgError
		if promptParams.ChatSessionUuid != "" && promptParams.Role == "system" &&
			errors.As(err, &pgErr) && pgErr.Code == "23505" {
			existingPrompt, getErr := h.service.q.GetOneChatPromptBySessionUUID(r.Context(), promptParams.ChatSessionUuid)
			if getErr == nil {
				json.NewEncoder(w).Encode(existingPrompt)
				return
			}
		}

		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to create chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

func (h *ChatPromptHandler) GetChatPromptByID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	prompt, err := h.service.GetChatPromptByID(r.Context(), int32(id))
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

func (h *ChatPromptHandler) UpdateChatPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	var promptParams sqlc_queries.UpdateChatPromptParams
	err = json.NewDecoder(r.Body).Decode(&promptParams)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	promptParams.ID = int32(id)
	prompt, err := h.service.UpdateChatPrompt(r.Context(), promptParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

func (h *ChatPromptHandler) DeleteChatPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid chat prompt ID"))
		return
	}
	err = h.service.DeleteChatPrompt(r.Context(), int32(id))
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to delete chat prompt"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatPromptHandler) GetAllChatPrompts(w http.ResponseWriter, r *http.Request) {
	prompts, err := h.service.GetAllChatPrompts(r.Context())
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get chat prompts"))
		return
	}
	json.NewEncoder(w).Encode(prompts)
}

func (h *ChatPromptHandler) GetChatPromptsByUserID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("invalid user ID"))
		return
	}
	prompts, err := h.service.GetChatPromptsByUserID(r.Context(), int32(id))
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get chat prompts by user"))
		return
	}
	json.NewEncoder(w).Encode(prompts)
}

func (h *ChatPromptHandler) DeleteChatPromptByUUID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["uuid"]
	err := h.service.DeleteChatPromptByUUID(r.Context(), idStr)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to delete chat prompt"))
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *ChatPromptHandler) UpdateChatPromptByUUID(w http.ResponseWriter, r *http.Request) {
	var simple_msg SimpleChatMessage
	err := json.NewDecoder(r.Body).Decode(&simple_msg)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	prompt, err := h.service.UpdateChatPromptByUUID(r.Context(), simple_msg.Uuid, simple_msg.Text)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update chat prompt"))
		return
	}
	json.NewEncoder(w).Encode(prompt)
}

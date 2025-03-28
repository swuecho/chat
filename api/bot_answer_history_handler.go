package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type BotAnswerHistoryHandler struct {
	service *BotAnswerHistoryService
}

func NewBotAnswerHistoryHandler(q *sqlc_queries.Queries) *BotAnswerHistoryHandler {
	service := NewBotAnswerHistoryService(q)
	return &BotAnswerHistoryHandler{service: service}
}

func (h *BotAnswerHistoryHandler) Register(router *mux.Router) {
	router.HandleFunc("/bot_answer_history", h.CreateBotAnswerHistory).Methods(http.MethodPost)
	router.HandleFunc("/bot_answer_history/{id}", h.GetBotAnswerHistoryByID).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/bot/{bot_uuid}", h.GetBotAnswerHistoryByBotUUID).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/user/{user_id}", h.GetBotAnswerHistoryByUserID).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/{id}", h.UpdateBotAnswerHistory).Methods(http.MethodPut)
	router.HandleFunc("/bot_answer_history/{id}", h.DeleteBotAnswerHistory).Methods(http.MethodDelete)
	router.HandleFunc("/bot_answer_history/bot/{bot_uuid}/count", h.GetBotAnswerHistoryCountByBotUUID).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/user/{user_id}/count", h.GetBotAnswerHistoryCountByUserID).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/bot/{bot_uuid}/latest", h.GetLatestBotAnswerHistoryByBotUUID).Methods(http.MethodGet)
}

func (h *BotAnswerHistoryHandler) CreateBotAnswerHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	var params sqlc_queries.CreateBotAnswerHistoryParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error()))
		return
	}

	// Set the user ID from context
	params.UserID = userID

	history, err := h.service.CreateBotAnswerHistory(ctx, params)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to create bot answer history"))
		return
	}

	RespondWithJSON(w, http.StatusCreated, history)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByID(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("ID is required"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid ID format"))
		return
	}

	history, err := h.service.GetBotAnswerHistoryByID(r.Context(), int32(idInt))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get bot answer history"))
		return
	}

	RespondWithJSON(w, http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByBotUUID(w http.ResponseWriter, r *http.Request) {
	botUUID := mux.Vars(r)["bot_uuid"]
	if botUUID == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Bot UUID is required"))
		return
	}

	limit, offset := getPaginationParams(r)
	history, err := h.service.GetBotAnswerHistoryByBotUUID(r.Context(), botUUID, limit, offset)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get bot answer history"))
		return
	}

	// Get total count for pagination
	totalCount, err := h.service.GetBotAnswerHistoryCountByBotUUID(r.Context(), botUUID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get bot answer history count"))
		return
	}

	// Calculate total pages
	totalPages := totalCount / int64(limit)
	if totalCount%int64(limit) > 0 {
		totalPages++
	}

	// Return paginated response
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"items":      history,
		"totalPages": totalPages,
		"totalCount": totalCount,
	})
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	limit, offset := getPaginationParams(r)
	history, err := h.service.GetBotAnswerHistoryByUserID(ctx, userID, limit, offset)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get bot answer history"))
		return
	}

	RespondWithJSON(w, http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) UpdateBotAnswerHistory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("ID is required"))
		return
	}

	var params sqlc_queries.UpdateBotAnswerHistoryParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error()))
		return
	}

	history, err := h.service.UpdateBotAnswerHistory(r.Context(), params.ID, params.Answer, params.TokensUsed)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to update bot answer history"))
		return
	}

	RespondWithJSON(w, http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) DeleteBotAnswerHistory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if id == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("ID is required"))
		return
	}

	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid ID format"))
		return
	}

	if err := h.service.DeleteBotAnswerHistory(r.Context(), int32(idInt)); err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to delete bot answer history"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryCountByBotUUID(w http.ResponseWriter, r *http.Request) {
	botUUID := mux.Vars(r)["bot_uuid"]
	if botUUID == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Bot UUID is required"))
		return
	}

	count, err := h.service.GetBotAnswerHistoryCountByBotUUID(r.Context(), botUUID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get bot answer history count"))
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]int64{"count": count})
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryCountByUserID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	count, err := h.service.GetBotAnswerHistoryCountByUserID(ctx, userID)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get bot answer history count"))
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]int64{"count": count})
}

func (h *BotAnswerHistoryHandler) GetLatestBotAnswerHistoryByBotUUID(w http.ResponseWriter, r *http.Request) {
	botUUID := mux.Vars(r)["bot_uuid"]
	if botUUID == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Bot UUID is required"))
		return
	}

	limit := getLimitParam(r, 1)
	history, err := h.service.GetLatestBotAnswerHistoryByBotUUID(r.Context(), botUUID, limit)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get latest bot answer history"))
		return
	}

	RespondWithJSON(w, http.StatusOK, history)
}

package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/svc"
)

type BotAnswerHistoryHandler struct {
	service *svc.BotAnswerHistoryService
}

func NewBotAnswerHistoryHandler(service *svc.BotAnswerHistoryService) *BotAnswerHistoryHandler {
	return &BotAnswerHistoryHandler{service: service}
}

func (h *BotAnswerHistoryHandler) Register(router *mux.Router) {
	router.HandleFunc("/bot_answer_history", endpoint(h.CreateBotAnswerHistory)).Methods(http.MethodPost)
	router.HandleFunc("/bot_answer_history/{id}", endpoint(h.GetBotAnswerHistoryByID)).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/bot/{bot_uuid}", endpoint(h.GetBotAnswerHistoryByBotUUID)).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/user/{user_id}", endpoint(h.GetBotAnswerHistoryByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/{id}", endpoint(h.UpdateBotAnswerHistory)).Methods(http.MethodPut)
	router.HandleFunc("/bot_answer_history/{id}", endpoint(h.DeleteBotAnswerHistory)).Methods(http.MethodDelete)
	router.HandleFunc("/bot_answer_history/bot/{bot_uuid}/count", endpoint(h.GetBotAnswerHistoryCountByBotUUID)).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/user/{user_id}/count", endpoint(h.GetBotAnswerHistoryCountByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/bot/{bot_uuid}/latest", endpoint(h.GetLatestBotAnswerHistoryByBotUUID)).Methods(http.MethodGet)
}

func (h *BotAnswerHistoryHandler) CreateBotAnswerHistory(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	var request createBotAnswerHistoryRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error())
	}

	history, err := h.service.CreateBotAnswerHistory(ctx, svc.CreateBotAnswerHistoryInput{
		BotUUID: request.BotUUID, UserID: userID, Prompt: request.Prompt,
		Answer: request.Answer, Model: request.Model, TokensUsed: request.TokensUsed,
	})
	if err != nil {
		return dto.WrapError(err, "Failed to create bot answer history")
	}
	return respondJSON(w, http.StatusCreated, history)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByID(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	history, err := h.service.GetBotAnswerHistoryByID(r.Context(), id)
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history")
	}
	return respondJSON(w, http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByBotUUID(w http.ResponseWriter, r *http.Request) error {
	botUUID := mux.Vars(r)["bot_uuid"]
	page, err := httpx.ParsePage(r)
	if err != nil {
		return err
	}
	history, err := h.service.GetBotAnswerHistoryByBotUUID(r.Context(), svc.BotAnswerHistoryPageQuery{BotUUID: botUUID, Page: svc.PageWindow{Limit: page.Limit, Offset: page.Offset}})
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history")
	}

	totalCount, err := h.service.GetBotAnswerHistoryCountByBotUUID(r.Context(), botUUID)
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history count")
	}
	return respondJSON(w, http.StatusOK, pageResponse(history, totalCount, page))
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryByUserID(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}
	page, err := httpx.ParsePage(r)
	if err != nil {
		return err
	}
	history, err := h.service.GetBotAnswerHistoryByUserID(ctx, svc.UserAnswerHistoryPageQuery{UserID: userID, Page: svc.PageWindow{Limit: page.Limit, Offset: page.Offset}})
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history")
	}
	total, err := h.service.GetBotAnswerHistoryCountByUserID(ctx, userID)
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history count")
	}
	return respondJSON(w, http.StatusOK, pageResponse(history, total, page))
}

func (h *BotAnswerHistoryHandler) UpdateBotAnswerHistory(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	var params updateBotAnswerHistoryRequest
	if err := DecodeJSON(r, &params); err != nil {
		return dto.ErrValidationInvalidInput("Invalid request body").WithDebugInfo(err.Error())
	}
	history, err := h.service.UpdateBotAnswerHistory(r.Context(), id, params.Answer, params.TokensUsed)
	if err != nil {
		return dto.WrapError(err, "Failed to update bot answer history")
	}
	return respondJSON(w, http.StatusOK, history)
}

func (h *BotAnswerHistoryHandler) DeleteBotAnswerHistory(w http.ResponseWriter, r *http.Request) error {
	id, err := positiveInt32Param(r, "id")
	if err != nil {
		return err
	}
	if err := h.service.DeleteBotAnswerHistory(r.Context(), id); err != nil {
		return dto.WrapError(err, "Failed to delete bot answer history")
	}
	return noContent(w)
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryCountByBotUUID(w http.ResponseWriter, r *http.Request) error {
	botUUID := mux.Vars(r)["bot_uuid"]
	count, err := h.service.GetBotAnswerHistoryCountByBotUUID(r.Context(), botUUID)
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history count")
	}
	return respondJSON(w, http.StatusOK, countHTTPResponse{Count: count})
}

func (h *BotAnswerHistoryHandler) GetBotAnswerHistoryCountByUserID(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	count, err := h.service.GetBotAnswerHistoryCountByUserID(ctx, userID)
	if err != nil {
		return dto.WrapError(err, "Failed to get bot answer history count")
	}
	return respondJSON(w, http.StatusOK, countHTTPResponse{Count: count})
}

func (h *BotAnswerHistoryHandler) GetLatestBotAnswerHistoryByBotUUID(w http.ResponseWriter, r *http.Request) error {
	botUUID := mux.Vars(r)["bot_uuid"]
	limit, err := httpx.ParseLimit(r, 1)
	if err != nil {
		return err
	}
	history, err := h.service.GetLatestBotAnswerHistoryByBotUUID(r.Context(), svc.LatestBotAnswerHistoryQuery{BotUUID: botUUID, Limit: limit})
	if err != nil {
		return dto.WrapError(err, "Failed to get latest bot answer history")
	}
	return respondJSON(w, http.StatusOK, history)
}

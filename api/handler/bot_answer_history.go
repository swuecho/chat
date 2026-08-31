package handler

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
)

type BotAnswerHistoryHandler struct {
	service *svc.BotAnswerHistoryService
}

func NewBotAnswerHistoryHandler(service *svc.BotAnswerHistoryService) *BotAnswerHistoryHandler {
	return &BotAnswerHistoryHandler{service: service}
}

type botAnswerHistoryResponse struct {
	ID           int32     `json:"id"`
	BotUUID      string    `json:"botUuid"`
	UserID       int32     `json:"userId"`
	Prompt       string    `json:"prompt"`
	Answer       string    `json:"answer"`
	Model        string    `json:"model"`
	TokensUsed   int32     `json:"tokensUsed"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	UserUsername string    `json:"userUsername"`
	UserEmail    string    `json:"userEmail"`
}

type botAnswerHistoryPageResponse struct {
	Items  []botAnswerHistoryResponse `json:"items"`
	Total  int64                      `json:"total"`
	Limit  int32                      `json:"limit"`
	Offset int32                      `json:"offset"`
}

func (h *BotAnswerHistoryHandler) Register(router *mux.Router, registry *apicontract.Registry) {
	router.HandleFunc("/bot_answer_history", endpoint(h.CreateBotAnswerHistory)).Methods(http.MethodPost)
	router.HandleFunc("/bot_answer_history/{id}", endpoint(h.GetBotAnswerHistoryByID)).Methods(http.MethodGet)
	maxPageSize := validation.MaxPageSize
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/bot_answer_history/bot/{bot_uuid}", OperationID: "listBotAnswerHistory",
		Summary: "List answer history for a bot", Tags: []string{"Bot history"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(),
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("bot_uuid"), apicontract.IntegerQueryParameter("limit", 1, &maxPageSize), apicontract.IntegerQueryParameter("offset", 0, nil)},
	}, h.getBotAnswerHistoryByBotUUID)
	router.HandleFunc("/bot_answer_history/user/{user_id}", endpoint(h.GetBotAnswerHistoryByUserID)).Methods(http.MethodGet)
	router.HandleFunc("/bot_answer_history/{id}", endpoint(h.UpdateBotAnswerHistory)).Methods(http.MethodPut)
	router.HandleFunc("/bot_answer_history/{id}", endpoint(h.DeleteBotAnswerHistory)).Methods(http.MethodDelete)
	apicontract.RegisterJSON(router, registry, apicontract.Operation{
		Method: http.MethodGet, Path: "/bot_answer_history/bot/{bot_uuid}/count", OperationID: "countBotAnswerHistory",
		Summary: "Count answer history entries for a bot", Tags: []string{"Bot history"}, SuccessStatus: http.StatusOK, Security: apicontract.BearerAuth(),
		Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("bot_uuid")},
	}, h.getBotAnswerHistoryCountByBotUUID)
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

func (h *BotAnswerHistoryHandler) getBotAnswerHistoryByBotUUID(r *http.Request, _ apicontract.NoBody) (botAnswerHistoryPageResponse, error) {
	botUUID := mux.Vars(r)["bot_uuid"]
	page, err := httpx.ParsePage(r)
	if err != nil {
		return botAnswerHistoryPageResponse{}, err
	}
	history, err := h.service.GetBotAnswerHistoryByBotUUID(r.Context(), svc.BotAnswerHistoryPageQuery{BotUUID: botUUID, Page: svc.PageWindow{Limit: page.Limit, Offset: page.Offset}})
	if err != nil {
		return botAnswerHistoryPageResponse{}, dto.WrapError(err, "Failed to get bot answer history")
	}

	totalCount, err := h.service.GetBotAnswerHistoryCountByBotUUID(r.Context(), botUUID)
	if err != nil {
		return botAnswerHistoryPageResponse{}, dto.WrapError(err, "Failed to get bot answer history count")
	}
	items := make([]botAnswerHistoryResponse, len(history))
	for i, item := range history {
		items[i] = botAnswerHistoryResponse{ID: item.ID, BotUUID: item.BotUuid, UserID: item.UserID, Prompt: item.Prompt, Answer: item.Answer, Model: item.Model, TokensUsed: item.TokensUsed, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, UserUsername: item.UserUsername, UserEmail: item.UserEmail}
	}
	return botAnswerHistoryPageResponse{Items: items, Total: totalCount, Limit: page.Limit, Offset: page.Offset}, nil
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

func (h *BotAnswerHistoryHandler) getBotAnswerHistoryCountByBotUUID(r *http.Request, _ apicontract.NoBody) (countHTTPResponse, error) {
	botUUID := mux.Vars(r)["bot_uuid"]
	count, err := h.service.GetBotAnswerHistoryCountByBotUUID(r.Context(), botUUID)
	if err != nil {
		return countHTTPResponse{}, dto.WrapError(err, "Failed to get bot answer history count")
	}
	return countHTTPResponse{Count: count}, nil
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

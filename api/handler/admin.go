package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/apicontract"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/httpx"
	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
)

type AdminHandler struct {
	service          *svc.AuthUserService
	sessionSvc       *svc.SessionAdminQueryService
	defaultRateLimit int32
}

func NewAdminHandler(service *svc.AuthUserService, sessionSvc *svc.SessionAdminQueryService, defaultRateLimit int32) *AdminHandler {
	return &AdminHandler{
		service:          service,
		sessionSvc:       sessionSvc,
		defaultRateLimit: defaultRateLimit,
	}
}

func (h *AdminHandler) RegisterRoutes(router *mux.Router, registry *apicontract.Registry) {
	router.HandleFunc("/users", endpoint(h.CreateUser)).Methods(http.MethodPost)
	router.HandleFunc("/users", endpoint(h.UpdateUser)).Methods(http.MethodPut)
	router.HandleFunc("/rate_limit", endpoint(h.UpdateRateLimit)).Methods(http.MethodPost)
	router.HandleFunc("/user_stats", endpoint(h.UserStatHandler)).Methods(http.MethodPost)
	router.HandleFunc("/user_analysis/{email}", endpoint(h.UserAnalysisHandler)).Methods(http.MethodGet)
	router.HandleFunc("/user_session_history/{email}", endpoint(h.UserSessionHistoryHandler)).Methods(http.MethodGet)
	router.HandleFunc("/session_messages/{sessionUuid}", endpoint(h.SessionMessagesHandler)).Methods(http.MethodGet)
	security := apicontract.BearerAuth()
	tags := []string{"Admin users"}
	apicontract.DocumentJSON[updateAuthUserRequest, authUserHTTPResponse](registry, apicontract.Operation{Method: http.MethodPut, Path: "/admin/users", OperationID: "updateAdminUser", Summary: "Update a user", Tags: tags, SuccessStatus: http.StatusOK, Security: security})
	apicontract.DocumentJSON[RateLimitRequest, rateHTTPResponse](registry, apicontract.Operation{Method: http.MethodPost, Path: "/admin/rate_limit", OperationID: "updateUserRateLimit", Summary: "Update a user rate limit", Tags: tags, SuccessStatus: http.StatusOK, Security: security})
	apicontract.DocumentJSON[pageRequest, userStatsPageHTTPResponse](registry, apicontract.Operation{Method: http.MethodPost, Path: "/admin/user_stats", OperationID: "getUserStats", Summary: "Get user statistics", Tags: tags, SuccessStatus: http.StatusOK, Security: security})
	emailParameter := []apicontract.Parameter{apicontract.StringPathParameter("email")}
	apicontract.DocumentJSON[apicontract.NoBody, svc.UserAnalysisData](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/user_analysis/{email}", OperationID: "getUserAnalysis", Summary: "Get user analysis", Tags: tags, SuccessStatus: http.StatusOK, Security: security, Parameters: emailParameter})
	maxPageSize := validation.MaxPageSize
	apicontract.DocumentJSON[apicontract.NoBody, sessionHistoryPageHTTPResponse](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/user_session_history/{email}", OperationID: "getUserSessionHistory", Summary: "Get user session history", Tags: tags, SuccessStatus: http.StatusOK, Security: security, Parameters: []apicontract.Parameter{apicontract.StringPathParameter("email"), apicontract.IntegerQueryParameter("limit", 1, &maxPageSize), apicontract.IntegerQueryParameter("offset", 0, nil)}})
	apicontract.DocumentJSON[apicontract.NoBody, []adminMessageHTTPResponse](registry, apicontract.Operation{Method: http.MethodGet, Path: "/admin/session_messages/{sessionUuid}", OperationID: "getAdminSessionMessages", Summary: "Get messages for a session", Tags: tags, SuccessStatus: http.StatusOK, Security: security, Parameters: []apicontract.Parameter{apicontract.UUIDPathParameter("sessionUuid")}})
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var request createAuthUserRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	user, err := h.service.CreateAuthUser(r.Context(), request.input())
	if err != nil {
		return dto.WrapError(err, "Failed to create user")
	}
	return respondJSON(w, http.StatusCreated, user)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	var request updateAuthUserRequest
	if err := DecodeJSON(r, &request); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	user, err := h.service.UpdateAuthUserByEmail(r.Context(), request.emailInput())
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update user")
	}
	return respondJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) error {
	var pagination pageRequest
	if err := DecodeJSON(r, &pagination); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	userStatsRows, total, err := h.service.GetUserStats(r.Context(), svc.PageRequest{Page: pagination.Page, Size: pagination.Size}, h.defaultRateLimit)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get user stats")
	}

	data := make([]UserStat, len(userStatsRows))
	for i, v := range userStatsRows {
		var avg int64
		if v.TotalChatMessages3Days > 0 {
			avg = v.TotalTokenCount3Days / v.TotalChatMessages3Days
		}
		data[i] = UserStat{
			Email:                            v.UserEmail,
			FirstName:                        v.FirstName,
			LastName:                         v.LastName,
			TotalChatMessages:                v.TotalChatMessages,
			TotalChatMessagesTokenCount:      v.TotalTokenCount,
			TotalChatMessages3Days:           v.TotalChatMessages3Days,
			TotalChatMessages3DaysTokenCount: v.TotalTokenCount3Days,
			AvgChatMessages3DaysTokenCount:   avg,
			RateLimit:                        v.RateLimit,
		}
	}

	page := httpx.Page{Limit: pagination.Size, Offset: (pagination.Page - 1) * pagination.Size}
	return respondJSON(w, http.StatusOK, pageResponse(data, total, page))
}

func (h *AdminHandler) UpdateRateLimit(w http.ResponseWriter, r *http.Request) error {
	var rateLimitRequest RateLimitRequest
	if err := DecodeJSON(r, &rateLimitRequest); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	rate, err := h.service.UpdateAuthUserRateLimitByEmail(r.Context(), rateLimitRequest.Email, rateLimitRequest.RateLimit)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update rate limit")
	}
	return respondJSON(w, http.StatusOK, rateHTTPResponse{Rate: rate})
}

func (h *AdminHandler) UserAnalysisHandler(w http.ResponseWriter, r *http.Request) error {
	email := mux.Vars(r)["email"]
	if email == "" {
		return dto.ErrValidationInvalidInput("Email parameter is required")
	}

	analysisData, err := h.service.GetUserAnalysis(r.Context(), email, h.defaultRateLimit)
	if err != nil {
		return dto.WrapError(err, "Failed to get user analysis")
	}
	return respondJSON(w, http.StatusOK, analysisData)
}

type SessionHistoryResponse struct {
	Data  []svc.SessionHistoryInfo `json:"data"`
	Total int64                    `json:"total"`
	Page  int32                    `json:"page"`
	Size  int32                    `json:"size"`
}

func (h *AdminHandler) UserSessionHistoryHandler(w http.ResponseWriter, r *http.Request) error {
	email := mux.Vars(r)["email"]
	if email == "" {
		return dto.ErrValidationInvalidInput("Email parameter is required")
	}
	page, err := httpx.ParsePage(r)
	if err != nil {
		return err
	}

	pageRequest := svc.PageRequest{Page: page.Offset/page.Limit + 1, Size: page.Limit}
	sessionHistory, total, err := h.service.GetUserSessionHistory(r.Context(), svc.UserSessionHistoryQuery{Email: email, Page: pageRequest})
	if err != nil {
		return dto.WrapError(err, "Failed to get user session history")
	}
	return respondJSON(w, http.StatusOK, pageResponse(sessionHistory, total, page))
}

func (h *AdminHandler) SessionMessagesHandler(w http.ResponseWriter, r *http.Request) error {
	sessionUuid := mux.Vars(r)["sessionUuid"]
	if sessionUuid == "" {
		return dto.ErrValidationInvalidInput("Session UUID parameter is required")
	}

	messages, err := h.sessionSvc.Messages(r.Context(), sessionUuid)
	if err != nil {
		return dto.WrapError(err, "Failed to get session messages")
	}

	w.Header().Set("Content-Type", "application/json")
	response := make([]adminMessageHTTPResponse, 0, len(messages))
	for _, message := range messages {
		response = append(response, adminMessageResponse(message))
	}
	return respondJSON(w, http.StatusOK, response)
}

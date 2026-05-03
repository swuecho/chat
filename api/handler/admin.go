package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

type AdminHandler struct {
	service          *svc.AuthUserService
	sessionSvc       *svc.ChatSessionService
	defaultRateLimit int32
}

func NewAdminHandler(service *svc.AuthUserService, defaultRateLimit int32) *AdminHandler {
	return &AdminHandler{
		service:          service,
		sessionSvc:       svc.NewChatSessionService(service.Q()),
		defaultRateLimit: defaultRateLimit,
	}
}

func (h *AdminHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users", h.CreateUser).Methods(http.MethodPost)
	router.HandleFunc("/users", h.UpdateUser).Methods(http.MethodPut)
	router.HandleFunc("/rate_limit", h.UpdateRateLimit).Methods(http.MethodPost)
	router.HandleFunc("/user_stats", h.UserStatHandler).Methods(http.MethodPost)
	router.HandleFunc("/user_analysis/{email}", h.UserAnalysisHandler).Methods(http.MethodGet)
	router.HandleFunc("/user_session_history/{email}", h.UserSessionHistoryHandler).Methods(http.MethodGet)
	router.HandleFunc("/session_messages/{sessionUuid}", h.SessionMessagesHandler).Methods(http.MethodGet)
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userParams sqlc_queries.CreateAuthUserParams
	if err := json.NewDecoder(r.Body).Decode(&userParams); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.CreateAuthUser(r.Context(), userParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to create user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var userParams sqlc_queries.UpdateAuthUserByEmailParams
	if err := json.NewDecoder(r.Body).Decode(&userParams); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.UpdateAuthUserByEmail(r.Context(), userParams)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) {
	var pagination dto.Pagination
	if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userStatsRows, total, err := h.service.GetUserStats(r.Context(), pagination, h.defaultRateLimit)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get user stats"))
		return
	}

	data := make([]interface{}, len(userStatsRows))
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

	json.NewEncoder(w).Encode(dto.Pagination{
		Page: pagination.Page, Size: pagination.Size, Total: total, Data: data,
	})
}

func (h *AdminHandler) UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var rateLimitRequest RateLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&rateLimitRequest); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	rate, err := h.service.UpdateAuthUserRateLimitByEmail(r.Context(),
		sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
			Email:     rateLimitRequest.Email,
			RateLimit: rateLimitRequest.RateLimit,
		})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update rate limit"))
		return
	}
	json.NewEncoder(w).Encode(map[string]int32{"rate": rate})
}

func (h *AdminHandler) UserAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	if email == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Email parameter is required"))
		return
	}

	analysisData, err := h.service.GetUserAnalysis(r.Context(), email, h.defaultRateLimit)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to get user analysis"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysisData)
}

type SessionHistoryResponse struct {
	Data  []svc.SessionHistoryInfo `json:"data"`
	Total int64                    `json:"total"`
	Page  int32                    `json:"page"`
	Size  int32                    `json:"size"`
}

func (h *AdminHandler) UserSessionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	if email == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Email parameter is required"))
		return
	}

	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	page := int32(1)
	size := int32(10)

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}
	if sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = int32(s)
		}
	}

	sessionHistory, total, err := h.service.GetUserSessionHistory(r.Context(), email, page, size)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to get user session history"))
		return
	}

	response := SessionHistoryResponse{
		Data: sessionHistory, Total: total, Page: page, Size: size,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) SessionMessagesHandler(w http.ResponseWriter, r *http.Request) {
	sessionUuid := mux.Vars(r)["sessionUuid"]
	if sessionUuid == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Session UUID parameter is required"))
		return
	}

	messages, err := h.sessionSvc.GetChatMessagesBySessionUUIDForAdmin(r.Context(), sessionUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to get session messages"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

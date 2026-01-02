package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type AdminHandler struct {
	service *AuthUserService
}

func NewAdminHandler(service *AuthUserService) *AdminHandler {
	return &AdminHandler{
		service: service,
	}
}

func (h *AdminHandler) RegisterRoutes(router *mux.Router) {
	// admin routes (without /admin prefix since router already handles it)
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
	err := json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.CreateAuthUser(r.Context(), userParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to create user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var userParams sqlc_queries.UpdateAuthUserByEmailParams
	err := json.NewDecoder(r.Body).Decode(&userParams)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	user, err := h.service.q.UpdateAuthUserByEmail(r.Context(), userParams)
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update user"))
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) {
	var pagination Pagination
	err := json.NewDecoder(r.Body).Decode(&pagination)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	userStatsRows, total, err := h.service.GetUserStats(r.Context(), pagination, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get user stats"))
		return
	}

	// Create a new []interface{} slice with same length as userStatsRows
	data := make([]interface{}, len(userStatsRows))

	// Copy the contents of userStatsRows into data
	for i, v := range userStatsRows {
		divider := v.TotalChatMessages3Days
		var avg int64
		if divider > 0 {
			avg = v.TotalTokenCount3Days / v.TotalChatMessages3Days
		} else {
			avg = 0
		}
		data[i] = UserStat{
			Email:                            v.UserEmail,
			FirstName:                        v.FirstName,
			LastName:                         v.LastName,
			TotalChatMessages:                v.TotalChatMessages,
			TotalChatMessages3Days:           v.TotalChatMessages3Days,
			RateLimit:                        v.RateLimit,
			TotalChatMessagesTokenCount:      v.TotalTokenCount,
			TotalChatMessages3DaysTokenCount: v.TotalTokenCount3Days,
			AvgChatMessages3DaysTokenCount:   avg,
		}
	}

	json.NewEncoder(w).Encode(Pagination{
		Page:  pagination.Page,
		Size:  pagination.Size,
		Total: total,
		Data:  data,
	})
}

func (h *AdminHandler) UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var rateLimitRequest RateLimitRequest
	err := json.NewDecoder(r.Body).Decode(&rateLimitRequest)
	if err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}
	rate, err := h.service.q.UpdateAuthUserRateLimitByEmail(r.Context(),
		sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
			Email:     rateLimitRequest.Email,
			RateLimit: rateLimitRequest.RateLimit,
		})

	if err != nil {
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to update rate limit"))
		return
	}
	json.NewEncoder(w).Encode(
		map[string]int32{
			"rate": rate,
		})
}

func (h *AdminHandler) UserAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	if email == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Email parameter is required"))
		return
	}

	analysisData, err := h.service.GetUserAnalysis(r.Context(), email, int32(appConfig.OPENAI.RATELIMIT))
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get user analysis"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analysisData)
}

type SessionHistoryResponse struct {
	Data  []SessionHistoryInfo `json:"data"`
	Total int64                `json:"total"`
	Page  int32                `json:"page"`
	Size  int32                `json:"size"`
}

func (h *AdminHandler) UserSessionHistoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	email := vars["email"]

	if email == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Email parameter is required"))
		return
	}

	// Parse pagination parameters
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
		RespondWithAPIError(w, WrapError(err, "Failed to get user session history"))
		return
	}

	response := SessionHistoryResponse{
		Data:  sessionHistory,
		Total: total,
		Page:  page,
		Size:  size,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) SessionMessagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionUuid := vars["sessionUuid"]

	if sessionUuid == "" {
		RespondWithAPIError(w, ErrValidationInvalidInput("Session UUID parameter is required"))
		return
	}

	messages, err := h.service.q.GetChatMessagesBySessionUUIDForAdmin(r.Context(), sessionUuid)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to get session messages"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

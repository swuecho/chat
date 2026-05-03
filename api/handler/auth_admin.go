package handler

import (
	"encoding/json"
	"net/http"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// --- Request/response types ---

type UserStat struct {
	Email                            string `json:"email"`
	FirstName                        string `json:"firstName"`
	LastName                         string `json:"lastName"`
	TotalChatMessages                int64  `json:"totalChatMessages"`
	TotalChatMessagesTokenCount      int64  `json:"totalChatMessagesTokenCount"`
	TotalChatMessages3Days           int64  `json:"totalChatMessages3Days"`
	TotalChatMessages3DaysTokenCount int64  `json:"totalChatMessages3DaysTokenCount"`
	AvgChatMessages3DaysTokenCount   int64  `json:"avgChatMessages3DaysTokenCount"`
	RateLimit                        int32  `json:"rateLimit"`
}

type RateLimitRequest struct {
	Email     string `json:"email"`
	RateLimit int32  `json:"rateLimit"`
}

// --- Handlers ---

// UserStatHandler returns paginated user statistics (admin only).
func (h *AuthUserHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) {
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

// UpdateRateLimit updates a user's rate limit (admin only).
func (h *AuthUserHandler) UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	var req RateLimitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	rate, err := h.service.UpdateAuthUserRateLimitByEmail(r.Context(),
		sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
			Email: req.Email, RateLimit: req.RateLimit,
		})
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to update rate limit"))
		return
	}
	json.NewEncoder(w).Encode(map[string]int32{"rate": rate})
}

// GetRateLimit returns the current user's rate limit.
func (h *AuthUserHandler) GetRateLimit(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	rate, err := h.service.GetRateLimit(r.Context(), userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get rate limit"))
		return
	}
	json.NewEncoder(w).Encode(map[string]int32{"rate": rate})
}

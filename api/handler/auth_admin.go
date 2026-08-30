package handler

import (
	"net/http"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
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

func (r *RateLimitRequest) Validate() error {
	if r.Email == "" {
		return dto.ErrValidationInvalidInput("email is required")
	}
	if r.RateLimit <= 0 {
		return dto.ErrValidationInvalidInput("rateLimit must be positive")
	}
	return nil
}

type legacyPageResponse[T any] struct {
	Page  int32 `json:"page"`
	Size  int32 `json:"size"`
	Total int64 `json:"total"`
	Data  []T   `json:"data"`
}

type rateLimitHTTPResponse struct {
	Rate int32 `json:"rate"`
}

// --- Handlers ---

// UserStatHandler returns paginated user statistics (admin only).
func (h *AuthUserHandler) UserStatHandler(w http.ResponseWriter, r *http.Request) error {
	var pagination dto.Pagination
	if err := DecodeJSON(r, &pagination); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}
	if pagination.Page < 1 || pagination.Size < 1 || pagination.Size > 100 {
		return dto.ErrValidationInvalidInput("page must be positive and size must be between 1 and 100")
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

	return respondJSON(w, http.StatusOK, legacyPageResponse[UserStat]{
		Page: pagination.Page, Size: pagination.Size, Total: total, Data: data,
	})
}

// UpdateRateLimit updates a user's rate limit (admin only).
func (h *AuthUserHandler) UpdateRateLimit(w http.ResponseWriter, r *http.Request) error {
	var req RateLimitRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	rate, err := h.service.UpdateAuthUserRateLimitByEmail(r.Context(), req.Email, req.RateLimit)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to update rate limit")
	}
	return respondJSON(w, http.StatusOK, rateLimitHTTPResponse{Rate: rate})
}

// GetRateLimit returns the current user's rate limit.
func (h *AuthUserHandler) GetRateLimit(w http.ResponseWriter, r *http.Request) error {
	userID, err := authenticatedUserID(r)
	if err != nil {
		return err
	}

	rate, err := h.service.GetRateLimit(r.Context(), userID)
	if err != nil {
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get rate limit")
	}
	return respondJSON(w, http.StatusOK, rateLimitHTTPResponse{Rate: rate})
}

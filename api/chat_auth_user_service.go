package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type AuthUserService struct {
	q *sqlc_queries.Queries
}

// NewAuthUserService creates a new AuthUserService.
func NewAuthUserService(q *sqlc_queries.Queries) *AuthUserService {
	return &AuthUserService{q: q}
}

// CreateAuthUser creates a new authentication user record.
func (s *AuthUserService) CreateAuthUser(ctx context.Context, auth_user_params sqlc_queries.CreateAuthUserParams) (sqlc_queries.AuthUser, error) {
	totalUserCount, err := s.q.GetTotalActiveUserCount(ctx)
	if err != nil {
		return sqlc_queries.AuthUser{}, errors.New("failed to retrieve total user count")
	}
	if totalUserCount == 0 {
		auth_user_params.IsSuperuser = true
		fmt.Println("First user is superuser.")
	}
	auth_user, err := s.q.CreateAuthUser(ctx, auth_user_params)
	if err != nil {
		return sqlc_queries.AuthUser{}, err
	}
	return auth_user, nil
}

// GetAuthUserByID returns an authentication user record by ID.
func (s *AuthUserService) GetAuthUserByID(ctx context.Context, id int32) (sqlc_queries.AuthUser, error) {
	auth_user, err := s.q.GetAuthUserByID(ctx, id)
	if err != nil {
		return sqlc_queries.AuthUser{}, errors.New("failed to retrieve authentication user")
	}
	return auth_user, nil
}

// GetAllAuthUsers returns all authentication user records.
func (s *AuthUserService) GetAllAuthUsers(ctx context.Context) ([]sqlc_queries.AuthUser, error) {
	auth_users, err := s.q.GetAllAuthUsers(ctx)
	if err != nil {
		return nil, errors.New("failed to retrieve authentication users")
	}
	return auth_users, nil
}

func (s *AuthUserService) Authenticate(ctx context.Context, email, password string) (sqlc_queries.AuthUser, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc_queries.AuthUser{}, err
	}
	if !auth.ValidatePassword(password, user.Password) {
		return sqlc_queries.AuthUser{}, ErrAuthInvalidCredentials
	}
	return user, nil
}

func (s *AuthUserService) Logout(tokenString string) (*http.Cookie, error) {
	userID, err := auth.ValidateToken(tokenString, jwtSecretAndAud.Secret)
	if err != nil {
		return nil, err
	}
	// Implement a mechanism to track invalidated tokens for the given user ID
	// auth.AddInvalidToken(userID, "insert-invalidated-token-here")
	cookie := auth.GetExpireSecureCookie(strconv.Itoa(int(userID)), false)
	return cookie, nil
}

// backend api
// GetUserStat(page, page_size) -> {data: [{user_email, total_sessions, total_messages, total_sessions_3_days, total_messages_3_days, rate_limit}], total: 100}
// GetTotalUserCount
// GetUserStat(page, page_size) ->[{user_email, total_sessions, total_messages, total_sessions_3_days, total_messages_3_days, rate_limit}]
func (s *AuthUserService) GetUserStats(ctx context.Context, p Pagination, defaultRateLimit int32) ([]sqlc_queries.GetUserStatsRow, int64, error) {
	auth_users_stat, err := s.q.GetUserStats(ctx,
		sqlc_queries.GetUserStatsParams{
			Offset:           p.Offset(),
			Limit:            p.Size,
			DefaultRateLimit: defaultRateLimit,
		})
	if err != nil {
		return nil, 0, eris.Wrap(err, "failed to retrieve user stats ")
	}
	total, err := s.q.GetTotalActiveUserCount(ctx)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve total active user count")
	}
	return auth_users_stat, total, nil
}

// UpdateRateLimit(user_email, rate_limit) -> { rate_limit: 100 }
func (s *AuthUserService) UpdateRateLimit(ctx context.Context, user_email string, rate_limit int32) (int32, error) {
	auth_user_params := sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
		Email:     user_email,
		RateLimit: rate_limit,
	}
	rate, err := s.q.UpdateAuthUserRateLimitByEmail(ctx, auth_user_params)
	if err != nil {
		return -1, errors.New("failed to update authentication user")
	}
	return rate, nil
}

// get ratelimit for user_id
func (s *AuthUserService) GetRateLimit(ctx context.Context, user_id int32) (int32, error) {
	rate, err := s.q.GetRateLimit(ctx, user_id)
	if err != nil {
		return -1, errors.New("failed to get rate limit")

	}
	return rate, nil
}

// UserAnalysisData represents the complete user analysis response
type UserAnalysisData struct {
	UserInfo       UserAnalysisInfo `json:"userInfo"`
	ModelUsage     []ModelUsageInfo `json:"modelUsage"`
	RecentActivity []ActivityInfo   `json:"recentActivity"`
}

type UserAnalysisInfo struct {
	Email         string `json:"email"`
	TotalMessages int64  `json:"totalMessages"`
	TotalTokens   int64  `json:"totalTokens"`
	TotalSessions int64  `json:"totalSessions"`
	Messages3Days int64  `json:"messages3Days"`
	Tokens3Days   int64  `json:"tokens3Days"`
	RateLimit     int32  `json:"rateLimit"`
}

type ModelUsageInfo struct {
	Model        string    `json:"model"`
	MessageCount int64     `json:"messageCount"`
	TokenCount   int64     `json:"tokenCount"`
	Percentage   float64   `json:"percentage"`
	LastUsed     time.Time `json:"lastUsed"`
}

type ActivityInfo struct {
	Date     time.Time `json:"date"`
	Messages int64     `json:"messages"`
	Tokens   int64     `json:"tokens"`
	Sessions int64     `json:"sessions"`
}

type SessionHistoryInfo struct {
	SessionID    string    `json:"sessionId"`
	Model        string    `json:"model"`
	MessageCount int64     `json:"messageCount"`
	TokenCount   int64     `json:"tokenCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// GetUserAnalysis retrieves comprehensive user analysis data
func (s *AuthUserService) GetUserAnalysis(ctx context.Context, email string, defaultRateLimit int32) (*UserAnalysisData, error) {
	// Get basic user info
	userInfo, err := s.q.GetUserAnalysisByEmail(ctx, sqlc_queries.GetUserAnalysisByEmailParams{
		Email:            email,
		DefaultRateLimit: defaultRateLimit,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user analysis")
	}

	// Get model usage
	modelUsageRows, err := s.q.GetUserModelUsageByEmail(ctx, email)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user model usage")
	}

	// Calculate total tokens for percentage calculation
	var totalTokens int64
	for _, row := range modelUsageRows {
		if row.TokenCount != nil {
			if tc, ok := row.TokenCount.(int64); ok {
				totalTokens += tc
			}
		}
	}

	modelUsage := make([]ModelUsageInfo, len(modelUsageRows))
	for i, row := range modelUsageRows {
		// Convert interface{} to int64 safely
		tokenCount := int64(0)
		if row.TokenCount != nil {
			if tc, ok := row.TokenCount.(int64); ok {
				tokenCount = tc
			}
		}

		percentage := float64(0)
		if totalTokens > 0 {
			percentage = float64(tokenCount) / float64(totalTokens) * 100
		}
		modelUsage[i] = ModelUsageInfo{
			Model:        row.Model,
			MessageCount: row.MessageCount,
			TokenCount:   tokenCount,
			Percentage:   percentage,
			LastUsed:     row.LastUsed,
		}
	}

	// Get recent activity
	activityRows, err := s.q.GetUserRecentActivityByEmail(ctx, email)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user recent activity")
	}

	recentActivity := make([]ActivityInfo, len(activityRows))
	for i, row := range activityRows {
		// Convert interface{} to int64 safely
		tokens := int64(0)
		if row.Tokens != nil {
			if t, ok := row.Tokens.(int64); ok {
				tokens = t
			}
		}

		recentActivity[i] = ActivityInfo{
			Date:     row.ActivityDate,
			Messages: row.Messages,
			Tokens:   tokens,
			Sessions: row.Sessions,
		}
	}

	analysisData := &UserAnalysisData{
		UserInfo: UserAnalysisInfo{
			Email:         userInfo.UserEmail,
			TotalMessages: userInfo.TotalMessages,
			TotalTokens:   userInfo.TotalTokens,
			TotalSessions: userInfo.TotalSessions,
			Messages3Days: userInfo.Messages3Days,
			Tokens3Days:   userInfo.Tokens3Days,
			RateLimit:     userInfo.RateLimit,
		},
		ModelUsage:     modelUsage,
		RecentActivity: recentActivity,
	}

	return analysisData, nil
}

// GetUserSessionHistory retrieves paginated session history for a user
func (s *AuthUserService) GetUserSessionHistory(ctx context.Context, email string, page, pageSize int32) ([]SessionHistoryInfo, int64, error) {
	offset := (page - 1) * pageSize

	// Get session history with pagination
	sessionRows, err := s.q.GetUserSessionHistoryByEmail(ctx, sqlc_queries.GetUserSessionHistoryByEmailParams{
		Email:  email,
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, eris.Wrap(err, "failed to get user session history")
	}

	// Get total count
	totalCount, err := s.q.GetUserSessionHistoryCountByEmail(ctx, email)
	if err != nil {
		return nil, 0, eris.Wrap(err, "failed to get user session history count")
	}

	sessionHistory := make([]SessionHistoryInfo, len(sessionRows))
	for i, row := range sessionRows {
		// Convert interface{} to int64 safely
		messageCount := int64(0)
		if row.MessageCount != nil {
			if mc, ok := row.MessageCount.(int64); ok {
				messageCount = mc
			}
		}

		tokenCount := int64(0)
		if row.TokenCount != nil {
			if tc, ok := row.TokenCount.(int64); ok {
				tokenCount = tc
			}
		}

		sessionHistory[i] = SessionHistoryInfo{
			SessionID:    row.SessionID,
			Model:        row.Model,
			MessageCount: messageCount,
			TokenCount:   tokenCount,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
	}

	return sessionHistory, totalCount, nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// AuthService provides authentication and user management.
type AuthService struct {
	q            *sqlc_queries.Queries
	jwtSecret    string
	defaultLimit int32
}

func NewAuthService(q *sqlc_queries.Queries, jwtSecret string, defaultLimit int32) *AuthService {
	return &AuthService{q: q, jwtSecret: jwtSecret, defaultLimit: defaultLimit}
}

func (s *AuthService) Q() *sqlc_queries.Queries { return s.q }

func (s *AuthService) CreateUser(ctx context.Context, params sqlc_queries.CreateAuthUserParams) (sqlc_queries.AuthUser, error) {
	totalUserCount, err := s.q.GetTotalActiveUserCount(ctx)
	if err != nil {
		return sqlc_queries.AuthUser{}, errors.New("failed to retrieve total user count")
	}
	if totalUserCount == 0 {
		params.IsSuperuser = true
		fmt.Println("First user is superuser.")
	}
	return s.q.CreateAuthUser(ctx, params)
}

func (s *AuthService) GetByID(ctx context.Context, id int32) (sqlc_queries.AuthUser, error) {
	return s.q.GetAuthUserByID(ctx, id)
}

func (s *AuthService) GetAll(ctx context.Context) ([]sqlc_queries.AuthUser, error) {
	return s.q.GetAllAuthUsers(ctx)
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (sqlc_queries.AuthUser, error) {
	user, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc_queries.AuthUser{}, err
	}
	if !auth.ValidatePassword(password, user.Password) {
		return sqlc_queries.AuthUser{}, dto.ErrAuthInvalidCredentials
	}
	return user, nil
}

func (s *AuthService) Logout(tokenString string) (*http.Cookie, error) {
	userID, err := auth.ValidateToken(tokenString, s.jwtSecret, auth.TokenTypeAccess)
	if err != nil {
		return nil, err
	}
	cookie := auth.GetExpireSecureCookie(strconv.Itoa(int(userID)), false)
	return cookie, nil
}

// UserStats represents user statistics.
type UserStats struct {
	UserEmail        string
	TotalSessions    int64
	TotalMessages    int64
	TotalSessions3D  int64
	TotalMessages3D  int64
	RateLimit        int32
}

func (s *AuthService) GetUserStats(ctx context.Context, page, size int32) ([]sqlc_queries.GetUserStatsRow, int64, error) {
	offset := (page - 1) * size
	stats, err := s.q.GetUserStats(ctx, sqlc_queries.GetUserStatsParams{
		Offset:           offset,
		Limit:            size,
		DefaultRateLimit: s.defaultLimit,
	})
	if err != nil {
		return nil, 0, eris.Wrap(err, "failed to retrieve user stats")
	}
	total, err := s.q.GetTotalActiveUserCount(ctx)
	if err != nil {
		return nil, 0, errors.New("failed to retrieve total active user count")
	}
	return stats, total, nil
}

func (s *AuthService) UpdateRateLimit(ctx context.Context, email string, rateLimit int32) (int32, error) {
	return s.q.UpdateAuthUserRateLimitByEmail(ctx, sqlc_queries.UpdateAuthUserRateLimitByEmailParams{
		Email:     email,
		RateLimit: rateLimit,
	})
}

func (s *AuthService) GetRateLimit(ctx context.Context, userID int32) (int32, error) {
	return s.q.GetRateLimit(ctx, userID)
}

// UserAnalysisData represents comprehensive user analysis.
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

func (s *AuthService) GetUserAnalysis(ctx context.Context, email string) (*UserAnalysisData, error) {
	userInfo, err := s.q.GetUserAnalysisByEmail(ctx, sqlc_queries.GetUserAnalysisByEmailParams{
		Email:            email,
		DefaultRateLimit: s.defaultLimit,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user analysis")
	}

	modelUsageRows, err := s.q.GetUserModelUsageByEmail(ctx, email)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user model usage")
	}

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

	activityRows, err := s.q.GetUserRecentActivityByEmail(ctx, email)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user recent activity")
	}

	recentActivity := make([]ActivityInfo, len(activityRows))
	for i, row := range activityRows {
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

	return &UserAnalysisData{
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
	}, nil
}

func (s *AuthService) GetUserSessionHistory(ctx context.Context, email string, page, pageSize int32) ([]SessionHistoryInfo, int64, error) {
	offset := (page - 1) * pageSize
	sessionRows, err := s.q.GetUserSessionHistoryByEmail(ctx, sqlc_queries.GetUserSessionHistoryByEmailParams{
		Email:  email,
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, eris.Wrap(err, "failed to get user session history")
	}
	totalCount, err := s.q.GetUserSessionHistoryCountByEmail(ctx, email)
	if err != nil {
		return nil, 0, eris.Wrap(err, "failed to get user session history count")
	}

	sessionHistory := make([]SessionHistoryInfo, len(sessionRows))
	for i, row := range sessionRows {
		messageCount := int64(0)
		tokenCount := int64(0)
		if row.MessageCount != nil {
			if mc, ok := row.MessageCount.(int64); ok {
				messageCount = mc
			}
		}
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

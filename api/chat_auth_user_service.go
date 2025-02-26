package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

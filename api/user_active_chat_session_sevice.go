package main

import (
	"context"
	"errors"
	"fmt"

	sqlc "github.com/swuecho/chatgpt_backend/sqlc_queries"
)

type UserActiveChatSessionService struct {
	q *sqlc.Queries
}

func NewUserActiveChatSessionService(q *sqlc.Queries) *UserActiveChatSessionService {
	return &UserActiveChatSessionService{q: q}
}

// CreateUserActiveChatSession inserts a new user active session and returns the newly-created session.
func (s *UserActiveChatSessionService) CreateUserActiveChatSession(ctx context.Context, sessionParams sqlc.CreateUserActiveChatSessionParams) (sqlc.UserActiveChatSession, error) {
	session, err := s.q.CreateUserActiveChatSession(ctx, sessionParams)
	if err != nil {
		return sqlc.UserActiveChatSession{}, errors.New("failed to create active session")
	}
	return session, nil
}

// UpdateOrCreateUserActiveChatSession updates an existing user active session or creates a new one if it doesn't exist.
func (s *UserActiveChatSessionService) CreateOrUpdateUserActiveChatSession(ctx context.Context, params sqlc.CreateOrUpdateUserActiveChatSessionParams) (sqlc.UserActiveChatSession, error) {
	session, err := s.q.CreateOrUpdateUserActiveChatSession(ctx, params)
	if err != nil {
		return sqlc.UserActiveChatSession{}, fmt.Errorf("failed to update or create active session %w", err)
	}
	return session, nil
}

// GetUserActiveChatSessionByID retrieves a user active session given an ID.
func (s *UserActiveChatSessionService) GetUserActiveChatSession(ctx context.Context, user_id int32) (sqlc.UserActiveChatSession, error) {
	session, err := s.q.GetUserActiveChatSession(ctx, user_id)
	if err != nil {
		return sqlc.UserActiveChatSession{}, err
	}
	return session, nil
}

// UpdateUserActiveChatSession updates an existing user active session.
func (s *UserActiveChatSessionService) UpdateUserActiveChatSession(ctx context.Context, userId int32, chatSessionUuid string) (sqlc.UserActiveChatSession, error) {
	sessionParams := sqlc.UpdateUserActiveChatSessionParams{
		UserID:          userId,
		ChatSessionUuid: chatSessionUuid,
	}
	session, err := s.q.UpdateUserActiveChatSession(ctx, sessionParams)
	if err != nil {
		return sqlc.UserActiveChatSession{}, errors.New("failed to update active session")
	}
	return session, nil
}

// DeleteUserActiveChatSession deletes a user active session given an ID.
func (s *UserActiveChatSessionService) DeleteUserActiveChatSession(ctx context.Context, id int32) error {
	err := s.q.DeleteUserActiveChatSession(ctx, id)
	if err != nil {
		return errors.New("failed to delete active session")
	}
	return nil
}

package main

import (
	"context"
	"database/sql"
	"github.com/rotisserie/eris"
	sqlc "github.com/swuecho/chat_backend/sqlc_queries"
)

type UserActiveChatSessionService struct {
	q *sqlc.Queries
}

func NewUserActiveChatSessionService(q *sqlc.Queries) *UserActiveChatSessionService {
	return &UserActiveChatSessionService{q: q}
}

// Simplified unified methods

// UpsertActiveSession creates or updates an active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) UpsertActiveSession(ctx context.Context, userID int32, workspaceID *int32, sessionUUID string) (sqlc.UserActiveChatSession, error) {
	var nullWorkspaceID sql.NullInt32
	if workspaceID != nil {
		nullWorkspaceID = sql.NullInt32{Int32: *workspaceID, Valid: true}
	}
	
	session, err := s.q.UpsertUserActiveSession(ctx, sqlc.UpsertUserActiveSessionParams{
		UserID:          userID,
		WorkspaceID:     nullWorkspaceID,
		ChatSessionUuid: sessionUUID,
	})
	if err != nil {
		return sqlc.UserActiveChatSession{}, eris.Wrap(err, "failed to upsert active session")
	}
	return session, nil
}

// GetActiveSession retrieves the active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) GetActiveSession(ctx context.Context, userID int32, workspaceID *int32) (sqlc.UserActiveChatSession, error) {
	var workspaceParam int32
	if workspaceID != nil {
		workspaceParam = *workspaceID
	}
	
	session, err := s.q.GetUserActiveSession(ctx, sqlc.GetUserActiveSessionParams{
		UserID:  userID,
		Column2: workspaceParam, // SQLC generated this awkward name due to the complex WHERE clause
	})
	if err != nil {
		return sqlc.UserActiveChatSession{}, eris.Wrap(err, "failed to get active session")
	}
	return session, nil
}

// GetAllActiveSessions retrieves all active sessions for a user (both global and workspace-specific)
func (s *UserActiveChatSessionService) GetAllActiveSessions(ctx context.Context, userID int32) ([]sqlc.UserActiveChatSession, error) {
	sessions, err := s.q.GetAllUserActiveSessions(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get all active sessions")
	}
	return sessions, nil
}

// DeleteActiveSession deletes the active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) DeleteActiveSession(ctx context.Context, userID int32, workspaceID *int32) error {
	var workspaceParam int32
	if workspaceID != nil {
		workspaceParam = *workspaceID
	}
	
	err := s.q.DeleteUserActiveSession(ctx, sqlc.DeleteUserActiveSessionParams{
		UserID:  userID,
		Column2: workspaceParam, // SQLC generated this awkward name
	})
	if err != nil {
		return eris.Wrap(err, "failed to delete active session")
	}
	return nil
}


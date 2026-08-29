package svc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/domain"
	sqlc "github.com/swuecho/chat_backend/sqlc_queries"
)

type ActiveSession struct {
	ID          int32
	UserID      int32
	SessionUUID string
	WorkspaceID *int32
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SetActiveSessionCommand struct {
	UserID      int32
	WorkspaceID *int32
	SessionUUID string
}

type SetWorkspaceActiveSessionCommand struct {
	UserID        int32
	WorkspaceUUID string
	SessionUUID   string
}

type GetActiveSessionQuery struct {
	UserID      int32
	WorkspaceID *int32
}

type DeleteActiveSessionCommand struct {
	UserID      int32
	WorkspaceID *int32
}

func activeSessionFromRecord(record sqlc.UserActiveChatSession) ActiveSession {
	var workspaceID *int32
	if record.WorkspaceID.Valid {
		workspaceID = &record.WorkspaceID.Int32
	}
	return ActiveSession{ID: record.ID, UserID: record.UserID, SessionUUID: record.ChatSessionUuid,
		WorkspaceID: workspaceID, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}

type UserActiveChatSessionService struct {
	q *sqlc.Queries
}

func NewUserActiveChatSessionService(q *sqlc.Queries) *UserActiveChatSessionService {
	return &UserActiveChatSessionService{q: q}
}

// Simplified unified methods

// UpsertActiveSession creates or updates an active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) UpsertActiveSession(ctx context.Context, command SetActiveSessionCommand) (ActiveSession, error) {
	var nullWorkspaceID sql.NullInt32
	if command.WorkspaceID != nil {
		nullWorkspaceID = sql.NullInt32{Int32: *command.WorkspaceID, Valid: true}
	}

	session, err := s.q.UpsertUserActiveSession(ctx, sqlc.UpsertUserActiveSessionParams{
		UserID:          command.UserID,
		WorkspaceID:     nullWorkspaceID,
		ChatSessionUuid: command.SessionUUID,
	})
	if err != nil {
		return ActiveSession{}, eris.Wrap(err, "failed to upsert active session")
	}
	return activeSessionFromRecord(session), nil
}

// SetWorkspaceActiveSession validates ownership and workspace membership before
// changing the user's active session for that workspace.
func (s *UserActiveChatSessionService) SetWorkspaceActiveSession(ctx context.Context, command SetWorkspaceActiveSessionCommand) (ActiveSession, error) {
	if command.UserID <= 0 {
		return ActiveSession{}, domain.Invalid("user ID is required")
	}
	if command.WorkspaceUUID == "" {
		return ActiveSession{}, domain.Invalid("workspace UUID is required")
	}
	if command.SessionUUID == "" {
		return ActiveSession{}, domain.Invalid("session UUID is required")
	}

	workspace, err := s.q.GetWorkspaceByUUID(ctx, command.WorkspaceUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ActiveSession{}, domain.NotFound("Workspace", err)
		}
		return ActiveSession{}, eris.Wrap(err, "failed to retrieve workspace")
	}
	if workspace.UserID != command.UserID {
		return ActiveSession{}, domain.Forbidden("workspace does not belong to user")
	}

	session, err := s.q.GetChatSessionByUUID(ctx, command.SessionUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ActiveSession{}, domain.NotFound("Chat session", err)
		}
		return ActiveSession{}, eris.Wrap(err, "failed to retrieve chat session")
	}
	if session.UserID != command.UserID {
		return ActiveSession{}, domain.Forbidden("chat session does not belong to user")
	}
	if !session.WorkspaceID.Valid || session.WorkspaceID.Int32 != workspace.ID {
		return ActiveSession{}, domain.Invalid("chat session does not belong to workspace")
	}

	workspaceID := workspace.ID
	return s.UpsertActiveSession(ctx, SetActiveSessionCommand{
		UserID: command.UserID, WorkspaceID: &workspaceID, SessionUUID: command.SessionUUID,
	})
}

// GetActiveSession retrieves the active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) GetActiveSession(ctx context.Context, query GetActiveSessionQuery) (ActiveSession, error) {
	var workspaceParam int32
	if query.WorkspaceID != nil {
		workspaceParam = *query.WorkspaceID
	}

	session, err := s.q.GetUserActiveSession(ctx, sqlc.GetUserActiveSessionParams{
		UserID:  query.UserID,
		Column2: workspaceParam, // SQLC generated this awkward name due to the complex WHERE clause
	})
	if err != nil {
		return ActiveSession{}, eris.Wrap(err, "failed to get active session")
	}
	return activeSessionFromRecord(session), nil
}

// GetAllActiveSessions retrieves all active sessions for a user (both global and workspace-specific)
func (s *UserActiveChatSessionService) GetAllActiveSessions(ctx context.Context, userID int32) ([]ActiveSession, error) {
	sessions, err := s.q.GetAllUserActiveSessions(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get all active sessions")
	}
	result := make([]ActiveSession, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, activeSessionFromRecord(session))
	}
	return result, nil
}

// DeleteActiveSession deletes the active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) DeleteActiveSession(ctx context.Context, command DeleteActiveSessionCommand) error {
	var workspaceParam int32
	if command.WorkspaceID != nil {
		workspaceParam = *command.WorkspaceID
	}

	err := s.q.DeleteUserActiveSession(ctx, sqlc.DeleteUserActiveSessionParams{
		UserID:  command.UserID,
		Column2: workspaceParam, // SQLC generated this awkward name
	})
	if err != nil {
		return eris.Wrap(err, "failed to delete active session")
	}
	return nil
}

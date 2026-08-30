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

type SetGlobalActiveSessionCommand struct {
	UserID      int32
	SessionUUID string
}

type GetWorkspaceActiveSessionQuery struct {
	UserID        int32
	WorkspaceUUID string
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

// SetGlobalActiveSession prevents a user from selecting another user's session.
func (s *UserActiveChatSessionService) SetGlobalActiveSession(ctx context.Context, command SetGlobalActiveSessionCommand) (ActiveSession, error) {
	if command.UserID <= 0 {
		return ActiveSession{}, domain.Invalid("user ID is required")
	}
	if command.SessionUUID == "" {
		return ActiveSession{}, domain.Invalid("session UUID is required")
	}
	session, err := s.ownedSession(ctx, command.UserID, command.SessionUUID)
	if err != nil {
		return ActiveSession{}, err
	}
	return s.UpsertActiveSession(ctx, SetActiveSessionCommand{UserID: command.UserID, SessionUUID: session.Uuid})
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

	session, err := s.ownedSession(ctx, command.UserID, command.SessionUUID)
	if err != nil {
		return ActiveSession{}, err
	}
	if !session.WorkspaceID.Valid || session.WorkspaceID.Int32 != workspace.ID {
		return ActiveSession{}, domain.Invalid("chat session does not belong to workspace")
	}

	workspaceID := workspace.ID
	return s.UpsertActiveSession(ctx, SetActiveSessionCommand{
		UserID: command.UserID, WorkspaceID: &workspaceID, SessionUUID: command.SessionUUID,
	})
}

// GetWorkspaceActiveSession resolves and authorizes the workspace before
// returning its active-session selection.
func (s *UserActiveChatSessionService) GetWorkspaceActiveSession(ctx context.Context, query GetWorkspaceActiveSessionQuery) (ActiveSession, error) {
	if query.UserID <= 0 {
		return ActiveSession{}, domain.Invalid("user ID is required")
	}
	if query.WorkspaceUUID == "" {
		return ActiveSession{}, domain.Invalid("workspace UUID is required")
	}
	workspace, err := s.q.GetWorkspaceByUUID(ctx, query.WorkspaceUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ActiveSession{}, domain.NotFound("Workspace", err)
		}
		return ActiveSession{}, eris.Wrap(err, "failed to retrieve workspace")
	}
	if workspace.UserID != query.UserID {
		return ActiveSession{}, domain.Forbidden("workspace does not belong to user")
	}
	workspaceID := workspace.ID
	active, err := s.GetActiveSession(ctx, GetActiveSessionQuery{UserID: query.UserID, WorkspaceID: &workspaceID})
	if errors.Is(err, sql.ErrNoRows) {
		return ActiveSession{}, domain.NotFound("Active session", err)
	}
	return active, err
}

func (s *UserActiveChatSessionService) ownedSession(ctx context.Context, userID int32, sessionUUID string) (sqlc.ChatSession, error) {
	session, err := s.q.GetChatSessionByUUID(ctx, sessionUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc.ChatSession{}, domain.NotFound("Chat session", err)
		}
		return sqlc.ChatSession{}, eris.Wrap(err, "failed to retrieve chat session")
	}
	if session.UserID != userID {
		return sqlc.ChatSession{}, domain.Forbidden("chat session does not belong to user")
	}
	return session, nil
}

// GetActiveSession retrieves the active session for a user in a specific workspace (or global if workspaceID is nil)
func (s *UserActiveChatSessionService) GetActiveSession(ctx context.Context, query GetActiveSessionQuery) (ActiveSession, error) {
	var workspaceParam sql.NullInt32
	if query.WorkspaceID != nil {
		workspaceParam = sql.NullInt32{Int32: *query.WorkspaceID, Valid: true}
	}

	session, err := s.q.GetUserActiveSession(ctx, sqlc.GetUserActiveSessionParams{
		UserID:      query.UserID,
		WorkspaceID: workspaceParam,
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
	var workspaceParam sql.NullInt32
	if command.WorkspaceID != nil {
		workspaceParam = sql.NullInt32{Int32: *command.WorkspaceID, Valid: true}
	}

	err := s.q.DeleteUserActiveSession(ctx, sqlc.DeleteUserActiveSessionParams{
		UserID:      command.UserID,
		WorkspaceID: workspaceParam,
	})
	if err != nil {
		return eris.Wrap(err, "failed to delete active session")
	}
	return nil
}

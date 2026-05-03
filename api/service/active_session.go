package service

import (
	"context"
	"database/sql"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ActiveSessionService manages the active chat session per user.
type ActiveSessionService struct {
	q *sqlc_queries.Queries
}

func NewActiveSessionService(q *sqlc_queries.Queries) *ActiveSessionService {
	return &ActiveSessionService{q: q}
}

func (s *ActiveSessionService) Q() *sqlc_queries.Queries { return s.q }

func (s *ActiveSessionService) Upsert(ctx context.Context, userID int32, workspaceID *int32, sessionUUID string) (*sqlc_queries.UserActiveChatSession, error) {
	var wsID sql.NullInt32
	if workspaceID != nil {
		wsID = sql.NullInt32{Int32: *workspaceID, Valid: true}
	}
	session, err := s.q.UpsertUserActiveSession(ctx, sqlc_queries.UpsertUserActiveSessionParams{
		UserID:          userID,
		WorkspaceID:     wsID,
		ChatSessionUuid: sessionUUID,
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// DeleteBySession deletes the active session by user and session UUID.
func (s *ActiveSessionService) DeleteBySession(ctx context.Context, userID int32, sessionUUID string) error {
	return s.q.DeleteUserActiveSessionBySession(ctx, sqlc_queries.DeleteUserActiveSessionBySessionParams{
		UserID:          userID,
		ChatSessionUuid: sessionUUID,
	})
}

func (s *ActiveSessionService) GetAll(ctx context.Context, userID int32) ([]sqlc_queries.UserActiveChatSession, error) {
	return s.q.GetAllUserActiveSessions(ctx, userID)
}

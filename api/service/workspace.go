package service

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// WorkspaceService provides methods for workspace management.
type WorkspaceService struct {
	q *sqlc_queries.Queries
}

func NewWorkspaceService(q *sqlc_queries.Queries) *WorkspaceService {
	return &WorkspaceService{q: q}
}

func (s *WorkspaceService) Q() *sqlc_queries.Queries { return s.q }

func (s *WorkspaceService) Create(ctx context.Context, params sqlc_queries.CreateWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	return s.q.CreateWorkspace(ctx, params)
}

func (s *WorkspaceService) GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatWorkspace, error) {
	return s.q.GetWorkspaceByUUID(ctx, uuid)
}

func (s *WorkspaceService) GetByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatWorkspace, error) {
	return s.q.GetWorkspacesByUserID(ctx, userID)
}

func (s *WorkspaceService) GetWithSessionCount(ctx context.Context, userID int32) ([]sqlc_queries.GetWorkspaceWithSessionCountRow, error) {
	return s.q.GetWorkspaceWithSessionCount(ctx, userID)
}

func (s *WorkspaceService) Update(ctx context.Context, params sqlc_queries.UpdateWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	return s.q.UpdateWorkspace(ctx, params)
}

func (s *WorkspaceService) UpdateOrder(ctx context.Context, params sqlc_queries.UpdateWorkspaceOrderParams) (sqlc_queries.ChatWorkspace, error) {
	return s.q.UpdateWorkspaceOrder(ctx, params)
}

func (s *WorkspaceService) Delete(ctx context.Context, uuid string) error {
	return s.q.DeleteWorkspace(ctx, uuid)
}

func (s *WorkspaceService) GetDefaultByUserID(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	return s.q.GetDefaultWorkspaceByUserID(ctx, userID)
}

func (s *WorkspaceService) SetDefault(ctx context.Context, params sqlc_queries.SetDefaultWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	return s.q.SetDefaultWorkspace(ctx, params)
}

func (s *WorkspaceService) CreateDefault(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	return s.q.CreateDefaultWorkspace(ctx, sqlc_queries.CreateDefaultWorkspaceParams{
		Uuid:   uuid.New().String(),
		UserID: userID,
	})
}

func (s *WorkspaceService) EnsureDefaultExists(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.GetDefaultByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.CreateDefault(ctx, userID)
		}
		return sqlc_queries.ChatWorkspace{}, err
	}
	return workspace, nil
}

func (s *WorkspaceService) HasPermission(ctx context.Context, uuid string, userID int32) (bool, error) {
	log.Printf("Checking permission for workspace=%s, user=%d", uuid, userID)
	result, err := s.q.HasWorkspacePermission(ctx, sqlc_queries.HasWorkspacePermissionParams{
		Uuid:   uuid,
		UserID: userID,
	})
	if err != nil {
		return false, eris.Wrap(err, "failed to check workspace permission")
	}
	return result, nil
}

func (s *WorkspaceService) MigrateSessionsToDefault(ctx context.Context, userID int32, workspaceID int32) error {
	return s.q.MigrateSessionsToDefaultWorkspace(ctx, sqlc_queries.MigrateSessionsToDefaultWorkspaceParams{
		UserID:      userID,
		WorkspaceID: sql.NullInt32{Int32: workspaceID, Valid: true},
	})
}

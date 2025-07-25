package main

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatWorkspaceService provides methods for interacting with chat workspaces.
type ChatWorkspaceService struct {
	q *sqlc_queries.Queries
}

// NewChatWorkspaceService creates a new ChatWorkspaceService.
func NewChatWorkspaceService(q *sqlc_queries.Queries) *ChatWorkspaceService {
	return &ChatWorkspaceService{q: q}
}

// CreateWorkspace creates a new workspace.
func (s *ChatWorkspaceService) CreateWorkspace(ctx context.Context, params sqlc_queries.CreateWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.q.CreateWorkspace(ctx, params)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to create workspace")
	}
	return workspace, nil
}

// GetWorkspaceByUUID returns a workspace by UUID.
func (s *ChatWorkspaceService) GetWorkspaceByUUID(ctx context.Context, workspaceUUID string) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.q.GetWorkspaceByUUID(ctx, workspaceUUID)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to retrieve workspace")
	}
	return workspace, nil
}

// GetWorkspacesByUserID returns all workspaces for a user.
func (s *ChatWorkspaceService) GetWorkspacesByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatWorkspace, error) {
	workspaces, err := s.q.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve workspaces")
	}
	return workspaces, nil
}

// GetWorkspaceWithSessionCount returns all workspaces with session counts for a user.
func (s *ChatWorkspaceService) GetWorkspaceWithSessionCount(ctx context.Context, userID int32) ([]sqlc_queries.GetWorkspaceWithSessionCountRow, error) {
	workspaces, err := s.q.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve workspaces with session count")
	}
	return workspaces, nil
}

// UpdateWorkspace updates an existing workspace.
func (s *ChatWorkspaceService) UpdateWorkspace(ctx context.Context, params sqlc_queries.UpdateWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.q.UpdateWorkspace(ctx, params)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to update workspace")
	}
	return workspace, nil
}

// UpdateWorkspaceOrder updates the order position of a workspace.
func (s *ChatWorkspaceService) UpdateWorkspaceOrder(ctx context.Context, params sqlc_queries.UpdateWorkspaceOrderParams) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.q.UpdateWorkspaceOrder(ctx, params)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to update workspace order")
	}
	return workspace, nil
}

// DeleteWorkspace deletes a workspace by UUID.
func (s *ChatWorkspaceService) DeleteWorkspace(ctx context.Context, workspaceUUID string) error {
	err := s.q.DeleteWorkspace(ctx, workspaceUUID)
	if err != nil {
		return eris.Wrap(err, "failed to delete workspace")
	}
	return nil
}

// GetDefaultWorkspaceByUserID returns the default workspace for a user.
func (s *ChatWorkspaceService) GetDefaultWorkspaceByUserID(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.q.GetDefaultWorkspaceByUserID(ctx, userID)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to retrieve default workspace")
	}
	return workspace, nil
}

// SetDefaultWorkspace sets a workspace as the default.
func (s *ChatWorkspaceService) SetDefaultWorkspace(ctx context.Context, params sqlc_queries.SetDefaultWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	workspace, err := s.q.SetDefaultWorkspace(ctx, params)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to set default workspace")
	}
	return workspace, nil
}

// CreateDefaultWorkspace creates a default workspace for a user.
func (s *ChatWorkspaceService) CreateDefaultWorkspace(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	workspaceUUID := uuid.New().String()
	params := sqlc_queries.CreateDefaultWorkspaceParams{
		Uuid:   workspaceUUID,
		UserID: userID,
	}
	workspace, err := s.q.CreateDefaultWorkspace(ctx, params)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, eris.Wrap(err, "failed to create default workspace")
	}
	return workspace, nil
}

// EnsureDefaultWorkspaceExists ensures a user has a default workspace, creating one if needed.
func (s *ChatWorkspaceService) EnsureDefaultWorkspaceExists(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	// Try to get existing default workspace
	workspace, err := s.GetDefaultWorkspaceByUserID(ctx, userID)
	if err != nil {
		// If no default workspace exists, create one
		if err == sql.ErrNoRows {
			return s.CreateDefaultWorkspace(ctx, userID)
		}
		return sqlc_queries.ChatWorkspace{}, err
	}
	return workspace, nil
}

// HasWorkspacePermission checks if a user has permission to access a workspace.
func (s *ChatWorkspaceService) HasWorkspacePermission(ctx context.Context, workspaceUUID string, userID int32) (bool, error) {
	result, err := s.q.HasWorkspacePermission(ctx, sqlc_queries.HasWorkspacePermissionParams{
		Uuid:   workspaceUUID,
		UserID: userID,
	})
	if err != nil {
		return false, eris.Wrap(err, "failed to check workspace permission")
	}
	return result, nil
}

// MigrateSessionsToDefaultWorkspace migrates all sessions without workspace to default workspace.
func (s *ChatWorkspaceService) MigrateSessionsToDefaultWorkspace(ctx context.Context, userID int32, workspaceID int32) error {
	err := s.q.MigrateSessionsToDefaultWorkspace(ctx, sqlc_queries.MigrateSessionsToDefaultWorkspaceParams{
		UserID:      userID,
		WorkspaceID: sql.NullInt32{Int32: workspaceID, Valid: true},
	})
	if err != nil {
		return eris.Wrap(err, "failed to migrate sessions to default workspace")
	}
	return nil
}

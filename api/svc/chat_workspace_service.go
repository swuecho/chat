package svc

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatWorkspaceService provides all workspace-related business logic.
type ChatWorkspaceService struct {
	q     *sqlc_queries.Queries
	tx    TransactionManager
	newID func() string
}

type CreateWorkspaceInput struct {
	Uuid          string
	UserID        int32
	Name          string
	Description   string
	Color         string
	Icon          string
	IsDefault     bool
	OrderPosition int32
}

type UpdateWorkspaceInput struct {
	Uuid        string
	UserID      int32
	Name        string
	Description string
	Color       string
	Icon        string
}

type UpdateWorkspaceOrderCommand struct {
	WorkspaceUUID         string
	UserID, OrderPosition int32
}
type DeleteWorkspaceCommand struct {
	WorkspaceUUID string
	UserID        int32
}

type SetDefaultWorkspaceCommand struct {
	UserID        int32
	WorkspaceUUID string
}

// Workspace is an application model; generated database records stay private
// to the persistence implementation.
type Workspace struct {
	ID            int32
	UUID          string
	UserID        int32
	Name          string
	Description   string
	Color         string
	Icon          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	IsDefault     bool
	OrderPosition int32
}

type WorkspaceSummary struct {
	Workspace
	SessionCount int64
}

func workspaceFromRecord(w sqlc_queries.ChatWorkspace) Workspace {
	return Workspace{ID: w.ID, UUID: w.Uuid, UserID: w.UserID, Name: w.Name,
		Description: w.Description, Color: w.Color, Icon: w.Icon, CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt, IsDefault: w.IsDefault, OrderPosition: w.OrderPosition}
}

// NewChatWorkspaceService creates a new ChatWorkspaceService.
func NewChatWorkspaceService(q *sqlc_queries.Queries) *ChatWorkspaceService {
	return &ChatWorkspaceService{q: q, tx: newSQLCTransactionManager(q), newID: util.NewUUID}
}

// --- Workspace CRUD ---

func (s *ChatWorkspaceService) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (Workspace, error) {
	w, err := s.q.CreateWorkspace(ctx, sqlc_queries.CreateWorkspaceParams(input))
	return workspaceFromRecord(w), eris.Wrap(err, "failed to create workspace")
}

func (s *ChatWorkspaceService) GetWorkspaceByUUID(ctx context.Context, uuid string) (Workspace, error) {
	w, err := s.q.GetWorkspaceByUUID(ctx, uuid)
	return workspaceFromRecord(w), eris.Wrap(err, "failed to retrieve workspace")
}

func (s *ChatWorkspaceService) GetWorkspacesByUserID(ctx context.Context, userID int32) ([]Workspace, error) {
	ws, err := s.q.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve workspaces")
	}
	result := make([]Workspace, 0, len(ws))
	for _, w := range ws {
		result = append(result, workspaceFromRecord(w))
	}
	return result, nil
}

func (s *ChatWorkspaceService) GetWorkspaceWithSessionCount(ctx context.Context, userID int32) ([]WorkspaceSummary, error) {
	ws, err := s.q.GetWorkspaceWithSessionCount(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve workspaces with session count")
	}
	result := make([]WorkspaceSummary, 0, len(ws))
	for _, w := range ws {
		result = append(result, WorkspaceSummary{Workspace: Workspace{ID: w.ID, UUID: w.Uuid, UserID: w.UserID, Name: w.Name, Description: w.Description, Color: w.Color, Icon: w.Icon, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt, IsDefault: w.IsDefault, OrderPosition: w.OrderPosition}, SessionCount: w.SessionCount})
	}
	return result, nil
}

func (s *ChatWorkspaceService) UpdateWorkspace(ctx context.Context, input UpdateWorkspaceInput) (Workspace, error) {
	w, err := s.q.UpdateWorkspace(ctx, sqlc_queries.UpdateWorkspaceParams{
		Uuid: input.Uuid, UserID: input.UserID, Name: input.Name,
		Description: input.Description, Color: input.Color, Icon: input.Icon,
	})
	return workspaceFromRecord(w), eris.Wrap(err, "failed to update workspace")
}

func (s *ChatWorkspaceService) UpdateWorkspaceOrder(ctx context.Context, command UpdateWorkspaceOrderCommand) (Workspace, error) {
	w, err := s.q.UpdateWorkspaceOrder(ctx, sqlc_queries.UpdateWorkspaceOrderParams{Uuid: command.WorkspaceUUID, UserID: command.UserID, OrderPosition: command.OrderPosition})
	return workspaceFromRecord(w), eris.Wrap(err, "failed to update workspace order")
}

func (s *ChatWorkspaceService) DeleteWorkspace(ctx context.Context, command DeleteWorkspaceCommand) error {
	rows, err := s.q.DeleteWorkspace(ctx, sqlc_queries.DeleteWorkspaceParams{Uuid: command.WorkspaceUUID, UserID: command.UserID})
	if err != nil {
		return eris.Wrap(err, "failed to delete workspace")
	}
	if rows == 0 {
		return domain.NotFound("Workspace", sql.ErrNoRows)
	}
	return nil
}

// --- Default workspace ---

func (s *ChatWorkspaceService) GetDefaultWorkspaceByUserID(ctx context.Context, userID int32) (Workspace, error) {
	w, err := s.q.GetDefaultWorkspaceByUserID(ctx, userID)
	return workspaceFromRecord(w), eris.Wrap(err, "failed to retrieve default workspace")
}

func (s *ChatWorkspaceService) CreateDefaultWorkspace(ctx context.Context, userID int32) (Workspace, error) {
	w, err := s.q.CreateDefaultWorkspace(ctx, sqlc_queries.CreateDefaultWorkspaceParams{
		Uuid: util.NewUUID(), UserID: userID,
	})
	return workspaceFromRecord(w), eris.Wrap(err, "failed to create default workspace")
}

func (s *ChatWorkspaceService) EnsureDefaultWorkspaceExists(ctx context.Context, userID int32) (Workspace, error) {
	w, err := s.GetDefaultWorkspaceByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.CreateDefaultWorkspace(ctx, userID)
		}
		return Workspace{}, err
	}
	return w, nil
}

// SetWorkspaceAsDefaultForUser changes a user's default workspace atomically.
func (s *ChatWorkspaceService) SetWorkspaceAsDefaultForUser(ctx context.Context, command SetDefaultWorkspaceCommand) (Workspace, error) {
	if command.UserID <= 0 {
		return Workspace{}, domain.Invalid("user ID is required")
	}
	if command.WorkspaceUUID == "" {
		return Workspace{}, domain.Invalid("workspace UUID is required")
	}

	var result Workspace
	err := s.tx.WithinTransaction(ctx, func(uow UnitOfWork) error {
		workspace, err := uow.WorkspaceByUUID(ctx, command.WorkspaceUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.NotFound("Workspace", err)
			}
			return eris.Wrap(err, "failed to retrieve workspace")
		}
		if workspace.UserID != command.UserID {
			return domain.Forbidden("workspace does not belong to user")
		}
		if err := uow.ClearDefaultWorkspaces(ctx, command.UserID); err != nil {
			return eris.Wrap(err, "failed to clear default workspace")
		}
		var updated Workspace
		updated, err = uow.SetDefaultWorkspace(ctx, command.UserID, command.WorkspaceUUID)
		if err != nil {
			return eris.Wrap(err, "failed to set default workspace")
		}
		result = updated
		return nil
	})
	if err != nil {
		return Workspace{}, err
	}
	return result, nil
}

// --- Permission ---

func (s *ChatWorkspaceService) HasWorkspacePermission(ctx context.Context, uuid string, userID int32) (bool, error) {
	slog.Info("Checking workspace permission", "workspace", uuid, "user", userID)
	result, err := s.q.HasWorkspacePermission(ctx, sqlc_queries.HasWorkspacePermissionParams{
		Uuid: uuid, UserID: userID,
	})
	if err != nil {
		return false, eris.Wrap(err, "failed to check workspace permission")
	}
	return result, nil
}

// GetSessionsByWorkspaceID returns all sessions in a workspace.
func (s *ChatWorkspaceService) GetSessionsByWorkspaceID(ctx context.Context, workspaceID int32) ([]ChatSession, error) {
	sessions, err := s.q.GetSessionsByWorkspaceID(ctx, sql.NullInt32{Int32: workspaceID, Valid: true})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get sessions by workspace")
	}
	result := make([]ChatSession, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, chatSessionFromRecord(session))
	}
	return result, nil
}

// --- Legacy migration ---

// AutoMigrateLegacySessionsResult holds the result of the migration operation.
type AutoMigrateLegacySessionsResult struct {
	HasLegacySessions bool
	MigratedCount     int
	DefaultWorkspace  Workspace
}

// AutoMigrateLegacySessions migrates sessions without a workspace_id to the default workspace.
func (s *ChatWorkspaceService) AutoMigrateLegacySessions(ctx context.Context, userID int32) (*AutoMigrateLegacySessionsResult, error) {
	legacySessions, err := s.q.GetSessionsWithoutWorkspace(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to check for legacy sessions")
	}

	result := &AutoMigrateLegacySessionsResult{
		HasLegacySessions: len(legacySessions) > 0,
	}

	if !result.HasLegacySessions {
		return result, nil
	}

	defaultWS, err := s.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to ensure default workspace")
	}
	result.DefaultWorkspace = defaultWS

	if err := s.q.MigrateSessionsToDefaultWorkspace(ctx, sqlc_queries.MigrateSessionsToDefaultWorkspaceParams{
		UserID:      userID,
		WorkspaceID: sql.NullInt32{Int32: defaultWS.ID, Valid: true},
	}); err != nil {
		return nil, eris.Wrap(err, "failed to migrate legacy sessions")
	}

	result.MigratedCount = len(legacySessions)
	return result, nil
}

// MigrateLegacyActiveSessions migrates active sessions without workspace context.
func (s *ChatWorkspaceService) MigrateLegacyActiveSessions(ctx context.Context, userID int32, defaultWorkspaceID int32) error {
	activeSessions, err := s.q.GetAllUserActiveSessions(ctx, userID)
	if err != nil {
		return eris.Wrap(err, "failed to get legacy active sessions")
	}

	for _, session := range activeSessions {
		if !session.WorkspaceID.Valid {
			_, err := s.q.UpsertUserActiveSession(ctx, sqlc_queries.UpsertUserActiveSessionParams{
				UserID:          userID,
				WorkspaceID:     sql.NullInt32{Int32: defaultWorkspaceID, Valid: true},
				ChatSessionUuid: session.ChatSessionUuid,
			})
			if err != nil {
				slog.Warn("failed to migrate active session", "sessionUuid", session.ChatSessionUuid, "error", err)
				continue
			}
			// Delete old global active session
			_ = s.q.DeleteUserActiveSessionBySession(ctx, sqlc_queries.DeleteUserActiveSessionBySessionParams{
				UserID: userID, ChatSessionUuid: session.ChatSessionUuid,
			})
		}
	}
	return nil
}

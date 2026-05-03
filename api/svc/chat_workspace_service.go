package svc

import (
	"fmt"
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatWorkspaceService provides all workspace-related business logic.
type ChatWorkspaceService struct {
	q *sqlc_queries.Queries
}

// NewChatWorkspaceService creates a new ChatWorkspaceService.
func NewChatWorkspaceService(q *sqlc_queries.Queries) *ChatWorkspaceService {
	return &ChatWorkspaceService{q: q}
}

// Q returns the underlying queries.
func (s *ChatWorkspaceService) Q() *sqlc_queries.Queries { return s.q }

// --- Workspace CRUD ---

func (s *ChatWorkspaceService) CreateWorkspace(ctx context.Context, params sqlc_queries.CreateWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.CreateWorkspace(ctx, params)
	return w, fmt.Errorf("failed to create workspace: %w", err)
}

func (s *ChatWorkspaceService) GetWorkspaceByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.GetWorkspaceByUUID(ctx, uuid)
	return w, fmt.Errorf("failed to retrieve workspace: %w", err)
}

func (s *ChatWorkspaceService) GetWorkspacesByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatWorkspace, error) {
	ws, err := s.q.GetWorkspacesByUserID(ctx, userID)
	return ws, fmt.Errorf("failed to retrieve workspaces: %w", err)
}

func (s *ChatWorkspaceService) GetWorkspaceWithSessionCount(ctx context.Context, userID int32) ([]sqlc_queries.GetWorkspaceWithSessionCountRow, error) {
	ws, err := s.q.GetWorkspaceWithSessionCount(ctx, userID)
	return ws, fmt.Errorf("failed to retrieve workspaces with session count: %w", err)
}

func (s *ChatWorkspaceService) UpdateWorkspace(ctx context.Context, params sqlc_queries.UpdateWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.UpdateWorkspace(ctx, params)
	return w, fmt.Errorf("failed to update workspace: %w", err)
}

func (s *ChatWorkspaceService) UpdateWorkspaceOrder(ctx context.Context, params sqlc_queries.UpdateWorkspaceOrderParams) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.UpdateWorkspaceOrder(ctx, params)
	return w, fmt.Errorf("failed to update workspace order: %w", err)
}

func (s *ChatWorkspaceService) DeleteWorkspace(ctx context.Context, uuid string) error {
	return fmt.Errorf("failed to delete workspace: %w", s.q.DeleteWorkspace(ctx, uuid))
}

// --- Default workspace ---

func (s *ChatWorkspaceService) GetDefaultWorkspaceByUserID(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.GetDefaultWorkspaceByUserID(ctx, userID)
	return w, fmt.Errorf("failed to retrieve default workspace: %w", err)
}

func (s *ChatWorkspaceService) SetDefaultWorkspace(ctx context.Context, params sqlc_queries.SetDefaultWorkspaceParams) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.SetDefaultWorkspace(ctx, params)
	return w, fmt.Errorf("failed to set default workspace: %w", err)
}

func (s *ChatWorkspaceService) CreateDefaultWorkspace(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.q.CreateDefaultWorkspace(ctx, sqlc_queries.CreateDefaultWorkspaceParams{
		Uuid: uuid.New().String(), UserID: userID,
	})
	return w, fmt.Errorf("failed to create default workspace: %w", err)
}

func (s *ChatWorkspaceService) EnsureDefaultWorkspaceExists(ctx context.Context, userID int32) (sqlc_queries.ChatWorkspace, error) {
	w, err := s.GetDefaultWorkspaceByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.CreateDefaultWorkspace(ctx, userID)
		}
		return sqlc_queries.ChatWorkspace{}, err
	}
	return w, nil
}

// SetWorkspaceAsDefaultForUser clears any existing default then sets the target as default.
// This is a business operation that should live in the service, not the handler.
func (s *ChatWorkspaceService) SetWorkspaceAsDefaultForUser(ctx context.Context, userID int32, workspaceUUID string) (sqlc_queries.ChatWorkspace, error) {
	workspaces, err := s.GetWorkspacesByUserID(ctx, userID)
	if err != nil {
		return sqlc_queries.ChatWorkspace{}, err
	}

	for _, ws := range workspaces {
		if ws.IsDefault && ws.Uuid != workspaceUUID {
			if _, err := s.SetDefaultWorkspace(ctx, sqlc_queries.SetDefaultWorkspaceParams{
				Uuid: ws.Uuid, IsDefault: false,
			}); err != nil {
				return sqlc_queries.ChatWorkspace{}, err
			}
		}
	}

	return s.SetDefaultWorkspace(ctx, sqlc_queries.SetDefaultWorkspaceParams{
		Uuid: workspaceUUID, IsDefault: true,
	})
}

// --- Permission ---

func (s *ChatWorkspaceService) HasWorkspacePermission(ctx context.Context, uuid string, userID int32) (bool, error) {
	log.Printf("Checking permission for workspace=%s, user=%d", uuid, userID)
	result, err := s.q.HasWorkspacePermission(ctx, sqlc_queries.HasWorkspacePermissionParams{
		Uuid: uuid, UserID: userID,
	})
	if err != nil {
		return false, fmt.Errorf("failed to check workspace permission: %w", err)
	}
	return result, nil
}

// --- Session creation inside workspace ---

// CreateSessionInWorkspace creates a new chat session inside a workspace and sets it as active.
func (s *ChatWorkspaceService) CreateSessionInWorkspace(ctx context.Context, userID int32, workspaceID int32, topic, model, defaultSystemPrompt string) (sqlc_queries.ChatSession, error) {
	sessionUUID := uuid.New().String()

	session, err := s.q.CreateChatSessionInWorkspace(ctx, sqlc_queries.CreateChatSessionInWorkspaceParams{
		UserID:      userID,
		Uuid:        sessionUUID,
		Topic:       topic,
		Model:       model,
		MaxLength:   10,
		Active:      true,
		WorkspaceID: sql.NullInt32{Int32: workspaceID, Valid: true},
	})
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to create session in workspace: %w", err)
	}

	return session, nil
}

// GetSessionsByWorkspaceID returns all sessions in a workspace.
func (s *ChatWorkspaceService) GetSessionsByWorkspaceID(ctx context.Context, workspaceID int32) ([]sqlc_queries.ChatSession, error) {
	sessions, err := s.q.GetSessionsByWorkspaceID(ctx, sql.NullInt32{Int32: workspaceID, Valid: true})
	return sessions, fmt.Errorf("failed to get sessions by workspace: %w", err)
}

// --- Legacy migration ---

// AutoMigrateLegacySessionsResult holds the result of the migration operation.
type AutoMigrateLegacySessionsResult struct {
	HasLegacySessions bool
	MigratedCount     int
	DefaultWorkspace  sqlc_queries.ChatWorkspace
}

// AutoMigrateLegacySessions migrates sessions without a workspace_id to the default workspace.
func (s *ChatWorkspaceService) AutoMigrateLegacySessions(ctx context.Context, userID int32) (*AutoMigrateLegacySessionsResult, error) {
	legacySessions, err := s.q.GetSessionsWithoutWorkspace(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check for legacy sessions: %w", err)
	}

	result := &AutoMigrateLegacySessionsResult{
		HasLegacySessions: len(legacySessions) > 0,
	}

	if !result.HasLegacySessions {
		return result, nil
	}

	defaultWS, err := s.EnsureDefaultWorkspaceExists(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure default workspace: %w", err)
	}
	result.DefaultWorkspace = defaultWS

	if err := s.q.MigrateSessionsToDefaultWorkspace(ctx, sqlc_queries.MigrateSessionsToDefaultWorkspaceParams{
		UserID:      userID,
		WorkspaceID: sql.NullInt32{Int32: defaultWS.ID, Valid: true},
	}); err != nil {
		return nil, fmt.Errorf("failed to migrate legacy sessions: %w", err)
	}

	result.MigratedCount = len(legacySessions)
	return result, nil
}

// MigrateLegacyActiveSessions migrates active sessions without workspace context.
func (s *ChatWorkspaceService) MigrateLegacyActiveSessions(ctx context.Context, userID int32, defaultWorkspaceID int32) error {
	activeSessions, err := s.q.GetAllUserActiveSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get legacy active sessions: %w", err)
	}

	for _, session := range activeSessions {
		if !session.WorkspaceID.Valid {
			_, err := s.q.UpsertUserActiveSession(ctx, sqlc_queries.UpsertUserActiveSessionParams{
				UserID:          userID,
				WorkspaceID:     sql.NullInt32{Int32: defaultWorkspaceID, Valid: true},
				ChatSessionUuid: session.ChatSessionUuid,
			})
			if err != nil {
				log.Printf("Warning: failed to migrate active session %s: %v", session.ChatSessionUuid, err)
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

package svc

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/domain"
)

type SaveSessionCommand struct {
	UserID              int32
	SessionUUID         string
	WorkspaceUUID       string
	Topic               string
	Model               string
	MaxLength           int32
	Temperature         float64
	MaxTokens           int32
	TopP                float64
	N                   int32
	Debug               bool
	SummarizeMode       bool
	ExploreMode         bool
	ArtifactEnabled     bool
	EnsureSystemPrompt  bool
	DefaultSystemPrompt string
	ActivateGlobally    bool
}

// SaveSession resolves the workspace, validates ownership, saves the session,
// optionally creates its system prompt, and selects it as active atomically.
func (s *ChatSessionService) SaveSession(ctx context.Context, command SaveSessionCommand) (ChatSession, error) {
	if command.UserID <= 0 {
		return ChatSession{}, domain.Invalid("user ID is required")
	}
	if command.SessionUUID == "" {
		return ChatSession{}, domain.Invalid("session UUID is required")
	}
	if command.MaxLength <= 0 {
		return ChatSession{}, domain.Invalid("max length must be positive")
	}

	var saved ChatSession
	err := s.tx.WithinSessionSaveTransaction(ctx, func(uow SessionSaveUnitOfWork) error {
		workspace, err := resolveSaveSessionWorkspace(ctx, s, uow, command)
		if err != nil {
			return err
		}

		existing, err := uow.InactiveSessionByUUID(ctx, command.SessionUUID)
		if err == nil && existing.UserID != command.UserID {
			return domain.Forbidden("chat session does not belong to user")
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return eris.Wrap(err, "failed to check existing chat session")
		}

		workspaceID := workspace.ID
		saved, err = uow.CreateOrUpdateSession(ctx, CreateOrUpdateChatSessionInput{
			UUID: command.SessionUUID, UserID: command.UserID, Topic: command.Topic,
			MaxLength: command.MaxLength, Temperature: command.Temperature, Model: command.Model,
			MaxTokens: command.MaxTokens, TopP: command.TopP, N: command.N, Debug: command.Debug,
			SummarizeMode: command.SummarizeMode, WorkspaceID: &workspaceID,
			ExploreMode: command.ExploreMode, ArtifactEnabled: command.ArtifactEnabled,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.Forbidden("chat session does not belong to user")
			}
			return eris.Wrap(err, "failed to save chat session")
		}

		if command.EnsureSystemPrompt {
			prompt := normalizedSystemPrompt(command.DefaultSystemPrompt)
			if err := uow.EnsureSystemPrompt(ctx, saved.UUID, command.UserID, prompt, s.newID()); err != nil {
				return err
			}
		}
		if command.ActivateGlobally {
			if err := uow.SetActiveSession(ctx, command.UserID, nil, saved.UUID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ChatSession{}, err
	}
	return saved, nil
}

func resolveSaveSessionWorkspace(ctx context.Context, service *ChatSessionService, uow SessionSaveUnitOfWork, command SaveSessionCommand) (Workspace, error) {
	if command.WorkspaceUUID != "" {
		workspace, err := uow.WorkspaceByUUID(ctx, command.WorkspaceUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Workspace{}, domain.NotFound("Workspace", err)
			}
			return Workspace{}, eris.Wrap(err, "failed to retrieve workspace")
		}
		if workspace.UserID != command.UserID {
			return Workspace{}, domain.Forbidden("workspace does not belong to user")
		}
		return workspace, nil
	}

	workspace, err := uow.DefaultWorkspaceByUserID(ctx, command.UserID)
	if err == nil {
		return workspace, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Workspace{}, eris.Wrap(err, "failed to retrieve default workspace")
	}
	workspace, err = uow.CreateDefaultWorkspace(ctx, command.UserID, service.newID())
	return workspace, eris.Wrap(err, "failed to create default workspace")
}

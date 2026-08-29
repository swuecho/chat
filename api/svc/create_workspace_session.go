package svc

import (
	"context"
	"strings"

	"github.com/swuecho/chat_backend/domain"
)

type CreateWorkspaceSessionCommand struct {
	UserID              int32
	WorkspaceID         int32
	WorkspaceUUID       string
	SessionUUID         string
	Topic               string
	Model               string
	DefaultSystemPrompt string
}

type CreateWorkspaceSessionResult struct {
	Session       ChatSession
	WorkspaceUUID string
}

// CreateWorkspaceSession creates the session, system prompt, and active-session
// selection atomically after checking workspace ownership.
func (s *ChatWorkspaceService) CreateWorkspaceSession(ctx context.Context, command CreateWorkspaceSessionCommand) (CreateWorkspaceSessionResult, error) {
	if command.UserID <= 0 || command.WorkspaceUUID == "" {
		return CreateWorkspaceSessionResult{}, domain.Invalid("user and workspace are required")
	}
	command.SessionUUID = s.newID()
	command.DefaultSystemPrompt = normalizedSystemPrompt(command.DefaultSystemPrompt)

	var result CreateWorkspaceSessionResult
	err := s.tx.WithinWorkspaceTransaction(ctx, func(uow WorkspaceUnitOfWork) error {
		workspace, err := uow.WorkspaceByUUID(ctx, command.WorkspaceUUID)
		if err != nil {
			return err
		}
		if workspace.UserID != command.UserID {
			return domain.Forbidden("workspace does not belong to user")
		}
		command.WorkspaceID = workspace.ID
		session, err := uow.CreateWorkspaceSession(ctx, command)
		if err != nil {
			return err
		}
		if err := uow.EnsureSystemPrompt(ctx, session.UUID, command.UserID, command.DefaultSystemPrompt, s.newID()); err != nil {
			return err
		}
		if err := uow.SetActiveSession(ctx, command.UserID, &workspace.ID, session.UUID); err != nil {
			return err
		}
		result = CreateWorkspaceSessionResult{Session: session, WorkspaceUUID: workspace.UUID}
		return nil
	})
	return result, err
}

func normalizedSystemPrompt(text string) string {
	if text = strings.TrimSpace(text); text != "" {
		return text
	}
	return defaultSystemPromptText
}

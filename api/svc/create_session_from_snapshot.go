package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/domain"
)

type CreateSessionFromSnapshotCommand struct {
	SnapshotUUID string
	UserID       int32
}

type CreateSessionFromSnapshotResult struct {
	SessionUUID string
}

type snapshotConversationMessage struct {
	UUID      string `json:"uuid"`
	Text      string `json:"text"`
	Inversion bool   `json:"inversion"`
}

func (m snapshotConversationMessage) role() string {
	if m.Inversion {
		return "user"
	}
	return "assistant"
}

// CreateSessionFromSnapshot copies a snapshot into a new session atomically.
// No session, prompt, message, or active-session change is visible unless the
// complete operation succeeds.
func (s *ChatSessionService) CreateSessionFromSnapshot(ctx context.Context, command CreateSessionFromSnapshotCommand) (CreateSessionFromSnapshotResult, error) {
	if command.SnapshotUUID == "" {
		return CreateSessionFromSnapshotResult{}, domain.Invalid("snapshot UUID is required")
	}
	if command.UserID <= 0 {
		return CreateSessionFromSnapshotResult{}, domain.Invalid("user ID is required")
	}

	result := CreateSessionFromSnapshotResult{}
	err := s.tx.WithinSnapshotCopyTransaction(ctx, func(uow SnapshotCopyUnitOfWork) error {
		snapshot, err := uow.SnapshotByUUID(ctx, command.SnapshotUUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.NotFound("Chat snapshot", err)
			}
			return eris.Wrap(err, "failed to retrieve snapshot")
		}

		var messages []snapshotConversationMessage
		if err := json.Unmarshal(snapshot.Conversation, &messages); err != nil || len(messages) == 0 {
			return domain.Invalid("snapshot has no messages")
		}

		prompt, err := uow.PromptByUUID(ctx, messages[0].UUID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.NotFound("Chat prompt", err)
			}
			return eris.Wrap(err, "failed to retrieve source prompt")
		}
		source, err := uow.InactiveSessionByUUID(ctx, prompt.ChatSessionUuid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.NotFound("Original chat session", err)
			}
			return eris.Wrap(err, "failed to retrieve source session")
		}

		sessionUUID := s.newID()
		session, err := uow.CreateOrUpdateSession(ctx, CreateOrUpdateChatSessionInput{
			Uuid: sessionUUID, UserID: command.UserID, Topic: snapshot.Title,
			MaxLength: source.MaxLength, Temperature: source.Temperature,
			Model: source.Model, MaxTokens: source.MaxTokens, TopP: source.TopP, N: 1,
			Debug: source.Debug, SummarizeMode: source.SummarizeMode,
			WorkspaceID: source.WorkspaceID, ExploreMode: source.ExploreMode,
			ArtifactEnabled: source.ArtifactEnabled,
		})
		if err != nil {
			return eris.Wrap(err, "failed to create session from snapshot")
		}

		if err := uow.CreatePrompt(ctx, CreateChatPromptInput{
			Uuid: s.newID(), ChatSessionUuid: sessionUUID, Role: "system",
			Content: messages[0].Text, UserID: command.UserID,
			CreatedBy: command.UserID, UpdatedBy: command.UserID,
		}); err != nil {
			return eris.Wrap(err, "failed to copy snapshot prompt")
		}

		for _, message := range messages[1:] {
			if err := uow.CreateMessage(ctx, CreateChatMessageInput{
				ChatSessionUuid: sessionUUID, Uuid: s.newID(), Role: message.role(),
				Content: message.Text, UserID: command.UserID,
				CreatedBy: command.UserID, UpdatedBy: command.UserID,
				Raw: json.RawMessage(`{}`), Artifacts: json.RawMessage(`[]`),
				SuggestedQuestions: json.RawMessage(`[]`),
			}); err != nil {
				return eris.Wrap(err, "failed to copy snapshot message")
			}
		}

		if err := uow.SetActiveSession(ctx, command.UserID, source.WorkspaceID, session.Uuid); err != nil {
			return eris.Wrap(err, "failed to set active session")
		}
		result.SessionUUID = session.Uuid
		return nil
	})
	if err != nil {
		return CreateSessionFromSnapshotResult{}, err
	}
	return result, nil
}

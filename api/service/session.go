package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// SessionService provides methods for chat session management.
type SessionService struct {
	q *sqlc_queries.Queries
}

func NewSessionService(q *sqlc_queries.Queries) *SessionService {
	return &SessionService{q: q}
}

func (s *SessionService) Q() *sqlc_queries.Queries { return s.q }

func (s *SessionService) CreateSession(ctx context.Context, params sqlc_queries.CreateChatSessionParams) (sqlc_queries.ChatSession, error) {
	return s.q.CreateChatSession(ctx, params)
}

func (s *SessionService) GetByID(ctx context.Context, id int32) (sqlc_queries.ChatSession, error) {
	session, err := s.q.GetChatSessionByID(ctx, id)
	if err != nil {
		return sqlc_queries.ChatSession{}, eris.Wrap(err, "failed to retrieve session")
	}
	return session, nil
}

func (s *SessionService) Update(ctx context.Context, params sqlc_queries.UpdateChatSessionParams) (sqlc_queries.ChatSession, error) {
	return s.q.UpdateChatSession(ctx, params)
}

func (s *SessionService) Delete(ctx context.Context, id int32) error {
	return eris.Wrap(s.q.DeleteChatSession(ctx, id), "failed to delete session")
}

func (s *SessionService) GetByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatSession, error) {
	return s.q.GetChatSessionsByUserID(ctx, userID)
}

func (s *SessionService) GetSimpleSessionsByUserID(ctx context.Context, userID int32) ([]dto.SimpleChatSession, error) {
	sessions, err := s.q.GetSessionsGroupedByWorkspace(ctx, userID)
	if err != nil {
		return nil, err
	}
	return lo.Map(sessions, func(session sqlc_queries.GetSessionsGroupedByWorkspaceRow, _ int) dto.SimpleChatSession {
		workspaceUuid := ""
		if session.WorkspaceUuid.Valid {
			workspaceUuid = session.WorkspaceUuid.String
		}
		return dto.SimpleChatSession{
			Uuid:            session.Uuid,
			Title:           session.Topic,
			MaxLength:       int(session.MaxLength),
			Temperature:     float64(session.Temperature),
			TopP:            float64(session.TopP),
			N:               session.N,
			MaxTokens:       session.MaxTokens,
			Debug:           session.Debug,
			Model:           session.Model,
			SummarizeMode:   session.SummarizeMode,
			ArtifactEnabled: session.ArtifactEnabled,
			WorkspaceUuid:   workspaceUuid,
		}
	}), nil
}

func (s *SessionService) GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatSession, error) {
	return s.q.GetChatSessionByUUID(ctx, uuid)
}

func (s *SessionService) UpdateByUUID(ctx context.Context, params sqlc_queries.UpdateChatSessionByUUIDParams) (sqlc_queries.ChatSession, error) {
	return s.q.UpdateChatSessionByUUID(ctx, params)
}

func (s *SessionService) UpdateTopicByUUID(ctx context.Context, params sqlc_queries.UpdateChatSessionTopicByUUIDParams) (sqlc_queries.ChatSession, error) {
	return s.q.UpdateChatSessionTopicByUUID(ctx, params)
}

func (s *SessionService) CreateOrUpdateByUUID(ctx context.Context, params sqlc_queries.CreateOrUpdateChatSessionByUUIDParams) (sqlc_queries.ChatSession, error) {
	return s.q.CreateOrUpdateChatSessionByUUID(ctx, params)
}

func (s *SessionService) DeleteByUUID(ctx context.Context, uuid string) error {
	return s.q.DeleteChatSessionByUUID(ctx, uuid)
}

func (s *SessionService) UpdateMaxLength(ctx context.Context, params sqlc_queries.UpdateSessionMaxLengthParams) (sqlc_queries.ChatSession, error) {
	return s.q.UpdateSessionMaxLength(ctx, params)
}

// EnsureDefaultSystemPrompt ensures a session has exactly one active system prompt.
func (s *SessionService) EnsureDefaultSystemPrompt(ctx context.Context, chatSessionUUID string, userID int32, systemPrompt string) (sqlc_queries.ChatPrompt, error) {
	existingPrompt, err := s.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUUID)
	if err == nil {
		return existingPrompt, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sqlc_queries.ChatPrompt{}, eris.Wrap(err, "failed to check existing session prompt")
	}

	promptText := strings.TrimSpace(systemPrompt)
	if promptText == "" {
		promptText = dto.DefaultSystemPromptText
	}

	tokenCount, tokenErr := getTokenCount(promptText)
	if tokenErr != nil {
		tokenCount = len(promptText) / dto.TokenEstimateRatio
	}
	if tokenCount <= 0 {
		tokenCount = 1
	}

	prompt, createErr := s.q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams{
		Uuid:            newUUID(),
		ChatSessionUuid: chatSessionUUID,
		Role:            "system",
		Content:         promptText,
		TokenCount:      int32(tokenCount),
		UserID:          userID,
		CreatedBy:       userID,
		UpdatedBy:       userID,
	})
	if createErr == nil {
		return prompt, nil
	}

	existingPrompt, err = s.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUUID)
	if err == nil {
		return existingPrompt, nil
	}

	return sqlc_queries.ChatPrompt{}, eris.Wrap(createErr, "failed to create default system prompt")
}

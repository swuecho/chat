package svc

import (
	"fmt"
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/dto"
)

// ChatSessionService provides methods for interacting with chat sessions.
type ChatSessionService struct {
	q *sqlc_queries.Queries
}

// NewChatSessionService creates a new ChatSessionService.
func NewChatSessionService(q *sqlc_queries.Queries) *ChatSessionService {
	return &ChatSessionService{q: q}
}

// Q returns the underlying queries.
func (s *ChatSessionService) Q() *sqlc_queries.Queries { return s.q }

// CreateChatSession creates a new chat session.
func (s *ChatSessionService) CreateChatSession(ctx context.Context, session_params sqlc_queries.CreateChatSessionParams) (sqlc_queries.ChatSession, error) {
	session, err := s.q.CreateChatSession(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, err
	}
	return session, nil
}

// GetChatSessionByID returns a chat session by ID.
func (s *ChatSessionService) GetChatSessionByID(ctx context.Context, id int32) (sqlc_queries.ChatSession, error) {
	session, err := s.q.GetChatSessionByID(ctx, id)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to retrieve session: : %w", err)
	}
	return session, nil
}

// UpdateChatSession updates an existing chat session.
func (s *ChatSessionService) UpdateChatSession(ctx context.Context, session_params sqlc_queries.UpdateChatSessionParams) (sqlc_queries.ChatSession, error) {
	session_u, err := s.q.UpdateChatSession(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to update session: %w", err)
	}
	return session_u, nil
}

// DeleteChatSession deletes a chat session by ID.
func (s *ChatSessionService) DeleteChatSession(ctx context.Context, id int32) error {
	err := s.q.DeleteChatSession(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete session by id: %w", err)
	}
	return nil
}

// GetAllChatSessions returns all chat sessions.
func (s *ChatSessionService) GetAllChatSessions(ctx context.Context) ([]sqlc_queries.ChatSession, error) {
	sessions, err := s.q.GetAllChatSessions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}
	return sessions, nil
}

func (s *ChatSessionService) GetChatSessionsByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatSession, error) {
	sessions, err := s.q.GetChatSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sessions: %w", err)
	}
	return sessions, nil
}

func (s *ChatSessionService) GetSimpleChatSessionsByUserID(ctx context.Context, userID int32) ([]dto.SimpleChatSession, error) {
	sessions, err := s.q.GetSessionsGroupedByWorkspace(ctx, userID)
	if err != nil {
		return nil, err
	}

	simple_sessions := lo.Map(sessions, func(session sqlc_queries.GetSessionsGroupedByWorkspaceRow, _idx int) dto.SimpleChatSession {
		workspaceUuid := ""
		if session.WorkspaceUuid.Valid {
			workspaceUuid = session.WorkspaceUuid.String
		}

		return dto.SimpleChatSession{
			Uuid:            session.Uuid,
			IsEdit:          false,
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
	})
	return simple_sessions, nil
}

// GetChatSessionByUUID returns an authentication user record by ID.
func (s *ChatSessionService) GetChatSessionByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatSession, error) {
	chatSession, err := s.q.GetChatSessionByUUID(ctx, uuid)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to retrieve session by uuid, : %w", err)
	}
	return chatSession, nil
}

// UpdateChatSessionByUUID updates an existing chat session.
func (s *ChatSessionService) UpdateChatSessionByUUID(ctx context.Context, session_params sqlc_queries.UpdateChatSessionByUUIDParams) (sqlc_queries.ChatSession, error) {
	session_u, err := s.q.UpdateChatSessionByUUID(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to update session, : %w", err)
	}
	return session_u, nil
}

// UpdateChatSessionTopicByUUID updates an existing chat session topic.
func (s *ChatSessionService) UpdateChatSessionTopicByUUID(ctx context.Context, session_params sqlc_queries.UpdateChatSessionTopicByUUIDParams) (sqlc_queries.ChatSession, error) {
	session_u, err := s.q.UpdateChatSessionTopicByUUID(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to update session, : %w", err)
	}
	return session_u, nil
}

// CreateOrUpdateChatSessionByUUID updates an existing chat session.
func (s *ChatSessionService) CreateOrUpdateChatSessionByUUID(ctx context.Context, session_params sqlc_queries.CreateOrUpdateChatSessionByUUIDParams) (sqlc_queries.ChatSession, error) {
	session_u, err := s.q.CreateOrUpdateChatSessionByUUID(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to update session, : %w", err)
	}
	return session_u, nil
}

// DeleteChatSessionByUUID deletes a chat session by UUID.
func (s *ChatSessionService) DeleteChatSessionByUUID(ctx context.Context, uuid string) error {
	err := s.q.DeleteChatSessionByUUID(ctx, uuid)
	if err != nil {
		return fmt.Errorf("failed to delete session by uuid, : %w", err)

	}
	return nil
}

// UpdateSessionMaxLength
func (s *ChatSessionService) UpdateSessionMaxLength(ctx context.Context, session_params sqlc_queries.UpdateSessionMaxLengthParams) (sqlc_queries.ChatSession, error) {
	session_u, err := s.q.UpdateSessionMaxLength(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, fmt.Errorf("failed to update session, : %w", err)
	}
	return session_u, nil
}

// ChatModelByName returns a chat model by name.
func (s *ChatSessionService) ChatModelByName(ctx context.Context, name string) (sqlc_queries.ChatModel, error) {
	m, err := s.q.ChatModelByName(ctx, name)
	return m, fmt.Errorf("failed to get chat model: %w", err)
}

// GetChatSessionByUUIDWithInActive returns a session by UUID including inactive ones.
func (s *ChatSessionService) GetChatSessionByUUIDWithInActive(ctx context.Context, uuid string) (sqlc_queries.ChatSession, error) {
	session, err := s.q.GetChatSessionByUUIDWithInActive(ctx, uuid)
	return session, fmt.Errorf("failed to get session with inactive: %w", err)
}

// GetOneChatPromptBySessionUUID returns the single prompt for a session.
func (s *ChatSessionService) GetOneChatPromptBySessionUUID(ctx context.Context, uuid string) (sqlc_queries.ChatPrompt, error) {
	p, err := s.q.GetOneChatPromptBySessionUUID(ctx, uuid)
	return p, fmt.Errorf("failed to get chat prompt: %w", err)
}

// GetChatMessagesBySessionUUID returns paginated messages for a session.
func (s *ChatSessionService) GetChatMessagesBySessionUUID(ctx context.Context, params sqlc_queries.GetChatMessagesBySessionUUIDParams) ([]sqlc_queries.ChatMessage, error) {
	msgs, err := s.q.GetChatMessagesBySessionUUID(ctx, params)
	return msgs, fmt.Errorf("failed to get chat messages: %w", err)
}

// RateLimitByUserAndSessionUUID checks per-model rate limits.
func (s *ChatSessionService) RateLimitByUserAndSessionUUID(ctx context.Context, params sqlc_queries.RateLimiteByUserAndSessionUUIDParams) (sqlc_queries.RateLimiteByUserAndSessionUUIDRow, error) {
	r, err := s.q.RateLimiteByUserAndSessionUUID(ctx, params)
	return r, err
}

// GetChatMessagesCountByUserAndModel returns message count for rate limiting.
func (s *ChatSessionService) GetChatMessagesCountByUserAndModel(ctx context.Context, params sqlc_queries.GetChatMessagesCountByUserAndModelParams) (int64, error) {
	return s.q.GetChatMessagesCountByUserAndModel(ctx, params)
}

// ChatSnapshotByUUID returns a snapshot by UUID.
func (s *ChatSessionService) ChatSnapshotByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatSnapshot, error) {
	sn, err := s.q.ChatSnapshotByUUID(ctx, uuid)
	return sn, fmt.Errorf("failed to get snapshot: %w", err)
}

// ChatSnapshotByUserIdAndUuid returns a user's snapshot by UUID.
func (s *ChatSessionService) ChatSnapshotByUserIdAndUuid(ctx context.Context, params sqlc_queries.ChatSnapshotByUserIdAndUuidParams) (sqlc_queries.ChatSnapshot, error) {
	sn, err := s.q.ChatSnapshotByUserIdAndUuid(ctx, params)
	return sn, fmt.Errorf("failed to get snapshot: %w", err)
}

// GetChatPromptByUUID returns a prompt by UUID.
func (s *ChatSessionService) GetChatPromptByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatPrompt, error) {
	p, err := s.q.GetChatPromptByUUID(ctx, uuid)
	return p, fmt.Errorf("failed to get chat prompt: %w", err)
}

// CreateChatPrompt creates a new chat prompt.
func (s *ChatSessionService) CreateChatPrompt(ctx context.Context, params sqlc_queries.CreateChatPromptParams) (sqlc_queries.ChatPrompt, error) {
	p, err := s.q.CreateChatPrompt(ctx, params)
	return p, fmt.Errorf("failed to create chat prompt: %w", err)
}

// CreateChatMessage creates a new chat message.
func (s *ChatSessionService) CreateChatMessage(ctx context.Context, params sqlc_queries.CreateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	m, err := s.q.CreateChatMessage(ctx, params)
	return m, fmt.Errorf("failed to create chat message: %w", err)
}

// CreateBotAnswerHistory creates a bot answer history entry.
func (s *ChatSessionService) CreateBotAnswerHistory(ctx context.Context, params sqlc_queries.CreateBotAnswerHistoryParams) (sqlc_queries.BotAnswerHistory, error) {
	h, err := s.q.CreateBotAnswerHistory(ctx, params)
	return h, fmt.Errorf("failed to create bot answer history: %w", err)
}

// UpdateChatMessageSuggestions updates suggested questions.
func (s *ChatSessionService) UpdateChatMessageSuggestions(ctx context.Context, params sqlc_queries.UpdateChatMessageSuggestionsParams) (sqlc_queries.ChatMessage, error) {
	return s.q.UpdateChatMessageSuggestions(ctx, params)
}

// UpsertUserActiveSession creates or updates an active session.
func (s *ChatSessionService) UpsertUserActiveSession(ctx context.Context, params sqlc_queries.UpsertUserActiveSessionParams) (sqlc_queries.UserActiveChatSession, error) {
	sess, err := s.q.UpsertUserActiveSession(ctx, params)
	return sess, err
}

// GetChatMessagesBySessionUUIDForAdmin returns messages for admin view.
func (s *ChatSessionService) GetChatMessagesBySessionUUIDForAdmin(ctx context.Context, uuid string) ([]sqlc_queries.GetChatMessagesBySessionUUIDForAdminRow, error) {
	return s.q.GetChatMessagesBySessionUUIDForAdmin(ctx, uuid)
}

// GetChatHistoryBySessionUUID returns chat history as simple messages.
func (s *ChatSessionService) GetChatHistoryBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]sqlc_queries.SimpleChatMessage, error) {
	return s.q.GetChatHistoryBySessionUUID(ctx, uuid, pageNum, pageSize)
}

// EnsureDefaultSystemPrompt ensures a session has exactly one active system prompt.
// It is safe to call repeatedly and tolerates concurrent callers.
func (s *ChatSessionService) EnsureDefaultSystemPrompt(ctx context.Context, chatSessionUUID string, userID int32, systemPrompt string) (sqlc_queries.ChatPrompt, error) {
	existingPrompt, err := s.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUUID)
	if err == nil {
		return existingPrompt, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sqlc_queries.ChatPrompt{}, fmt.Errorf("failed to check existing session prompt: %w", err)
	}

	promptText := strings.TrimSpace(systemPrompt)
	if promptText == "" {
		promptText = dto.DefaultSystemPromptText
	}

	tokenCount, tokenErr := provider.GetTokenCount(promptText)
	if tokenErr != nil {
		tokenCount = len(promptText) / dto.TokenEstimateRatio
	}
	if tokenCount <= 0 {
		tokenCount = 1
	}

	prompt, createErr := s.q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams{
		Uuid:            provider.NewUUID(),
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

	// Handle concurrent creation race by returning the now-existing prompt.
	existingPrompt, err = s.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUUID)
	if err == nil {
		return existingPrompt, nil
	}

	return sqlc_queries.ChatPrompt{}, fmt.Errorf("failed to create default system prompt: %w", createErr)
}

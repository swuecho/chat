package svc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatPrompt struct {
	ID                                   int32
	Uuid, ChatSessionUuid, Role, Content string
	Score                                float64
	UserID                               int32
	CreatedAt, UpdatedAt                 time.Time
	CreatedBy, UpdatedBy                 int32
	IsDeleted                            bool
	TokenCount                           int32
}

func chatPromptFromRecord(p sqlc_queries.ChatPrompt) ChatPrompt {
	return ChatPrompt{ID: p.ID, Uuid: p.Uuid, ChatSessionUuid: p.ChatSessionUuid, Role: p.Role,
		Content: p.Content, Score: p.Score, UserID: p.UserID, CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt, CreatedBy: p.CreatedBy, UpdatedBy: p.UpdatedBy,
		IsDeleted: p.IsDeleted, TokenCount: p.TokenCount}
}

func chatPromptsFromRecords(records []sqlc_queries.ChatPrompt) []ChatPrompt {
	result := make([]ChatPrompt, 0, len(records))
	for _, record := range records {
		result = append(result, chatPromptFromRecord(record))
	}
	return result
}

type ChatPromptService struct {
	q *sqlc_queries.Queries
}

type CreateChatPromptInput struct {
	Uuid            string
	ChatSessionUuid string
	Role            string
	Content         string
	TokenCount      int32
	UserID          int32
	CreatedBy       int32
	UpdatedBy       int32
}

type UpdateChatPromptInput struct {
	ID              int32
	ChatSessionUuid string
	Role            string
	Content         string
	Score           float64
	UserID          int32
	UpdatedBy       int32
}

type DeleteChatPromptCommand struct{ ID, UserID int32 }
type DeleteChatPromptByUUIDCommand struct {
	UUID   string
	UserID int32
}
type UpdateChatPromptByUUIDCommand struct {
	UUID, Content string
	UserID        int32
}

// NewChatPromptService creates a new ChatPromptService.
func NewChatPromptService(q *sqlc_queries.Queries) *ChatPromptService {
	return &ChatPromptService{q: q}
}

// CreateChatPrompt creates a new chat prompt.
func (s *ChatPromptService) CreateChatPrompt(ctx context.Context, input CreateChatPromptInput) (ChatPrompt, error) {
	prompt, err := s.q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams(input))
	if err != nil {
		return ChatPrompt{}, eris.Wrap(err, "failed to create prompt: ")
	}
	return chatPromptFromRecord(prompt), nil
}

func (s *ChatPromptService) CreateChatPromptWithUUID(ctx context.Context, uuid string, role, content string) (ChatPrompt, error) {
	params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: uuid,
		Role:            role,
		Content:         content,
	}
	prompt, err := s.q.CreateChatPrompt(ctx, params)
	return chatPromptFromRecord(prompt), err
}

// GetChatPromptByID returns a chat prompt by ID.
func (s *ChatPromptService) GetChatPromptByID(ctx context.Context, id, userID int32) (ChatPrompt, error) {
	prompt, err := s.q.GetChatPromptByID(ctx, sqlc_queries.GetChatPromptByIDParams{ID: id, UserID: userID})
	if err != nil {
		return ChatPrompt{}, eris.Wrap(err, "failed to create prompt: ")
	}
	return chatPromptFromRecord(prompt), nil
}

// UpdateChatPrompt updates an existing chat prompt.
func (s *ChatPromptService) UpdateChatPrompt(ctx context.Context, input UpdateChatPromptInput) (ChatPrompt, error) {
	prompt_u, err := s.q.UpdateChatPrompt(ctx, sqlc_queries.UpdateChatPromptParams(input))
	if err != nil {
		return ChatPrompt{}, errors.New("failed to update prompt")
	}
	return chatPromptFromRecord(prompt_u), nil
}

// DeleteChatPrompt deletes a chat prompt by ID.
func (s *ChatPromptService) DeleteChatPrompt(ctx context.Context, command DeleteChatPromptCommand) error {
	rows, err := s.q.DeleteChatPrompt(ctx, sqlc_queries.DeleteChatPromptParams{ID: command.ID, UserID: command.UserID})
	if err != nil {
		return errors.New("failed to delete prompt")
	}
	if rows == 0 {
		return domain.NotFound("Chat prompt", sql.ErrNoRows)
	}
	return nil
}

// GetAllChatPrompts returns all chat prompts.
func (s *ChatPromptService) GetAllChatPrompts(ctx context.Context) ([]ChatPrompt, error) {
	prompts, err := s.q.GetAllChatPrompts(ctx)
	if err != nil {
		return nil, errors.New("failed to retrieve prompts")
	}
	return chatPromptsFromRecords(prompts), nil
}

func (s *ChatPromptService) GetChatPromptsByUserID(ctx context.Context, userID int32) ([]ChatPrompt, error) {
	prompts, err := s.q.GetChatPromptsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return chatPromptsFromRecords(prompts), nil
}

func (s *ChatPromptService) GetChatPromptsBySessionUUID(ctx context.Context, session_uuid string) ([]ChatPrompt, error) {
	prompts, err := s.q.GetChatPromptsBySessionUUID(ctx, session_uuid)
	if err != nil {
		return nil, err
	}
	return chatPromptsFromRecords(prompts), nil
}

// GetOneChatPromptBySessionUUID returns the first prompt for a session.
func (s *ChatPromptService) GetOneChatPromptBySessionUUID(ctx context.Context, uuid string) (ChatPrompt, error) {
	prompt, err := s.q.GetOneChatPromptBySessionUUID(ctx, uuid)
	return chatPromptFromRecord(prompt), err
}

// DeleteChatPromptByUUID
func (s *ChatPromptService) DeleteChatPromptByUUID(ctx context.Context, command DeleteChatPromptByUUIDCommand) error {
	rows, err := s.q.DeleteChatPromptByUUID(ctx, sqlc_queries.DeleteChatPromptByUUIDParams{Uuid: command.UUID, UserID: command.UserID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.NotFound("Chat prompt", sql.ErrNoRows)
	}
	return nil
}

// UpdateChatPromptByUUID
func (s *ChatPromptService) UpdateChatPromptByUUID(ctx context.Context, command UpdateChatPromptByUUIDCommand) (ChatPrompt, error) {
	tokenCount, _ := provider.GetTokenCount(command.Content)
	params := sqlc_queries.UpdateChatPromptByUUIDParams{
		Uuid:       command.UUID,
		Content:    command.Content,
		TokenCount: int32(tokenCount),
		UserID:     command.UserID,
	}
	prompt, err := s.q.UpdateChatPromptByUUID(ctx, params)
	return chatPromptFromRecord(prompt), err
}

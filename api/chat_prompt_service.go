package main

import (
	"context"
	"errors"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatPromptService struct {
	q *sqlc_queries.Queries
}

// NewChatPromptService creates a new ChatPromptService.
func NewChatPromptService(q *sqlc_queries.Queries) *ChatPromptService {
	return &ChatPromptService{q: q}
}

// CreateChatPrompt creates a new chat prompt.
func (s *ChatPromptService) CreateChatPrompt(ctx context.Context, prompt_params sqlc_queries.CreateChatPromptParams) (sqlc_queries.ChatPrompt, error) {
	prompt, err := s.q.CreateChatPrompt(ctx, prompt_params)
	if err != nil {
		return sqlc_queries.ChatPrompt{}, eris.Wrap(err, "failed to create prompt: ")
	}
	return prompt, nil
}

func (s *ChatPromptService) CreateChatPromptWithUUID(ctx context.Context, uuid string, role, content string) (sqlc_queries.ChatPrompt, error) {
	params := sqlc_queries.CreateChatPromptParams{
		ChatSessionUuid: uuid,
		Role:            role,
		Content:         content,
	}
	prompt, err := s.q.CreateChatPrompt(ctx, params)
	return prompt, err
}

// GetChatPromptByID returns a chat prompt by ID.
func (s *ChatPromptService) GetChatPromptByID(ctx context.Context, id int32) (sqlc_queries.ChatPrompt, error) {
	prompt, err := s.q.GetChatPromptByID(ctx, id)
	if err != nil {
		return sqlc_queries.ChatPrompt{}, eris.Wrap(err, "failed to create prompt: ")
	}
	return prompt, nil
}

// UpdateChatPrompt updates an existing chat prompt.
func (s *ChatPromptService) UpdateChatPrompt(ctx context.Context, prompt_params sqlc_queries.UpdateChatPromptParams) (sqlc_queries.ChatPrompt, error) {
	prompt_u, err := s.q.UpdateChatPrompt(ctx, prompt_params)
	if err != nil {
		return sqlc_queries.ChatPrompt{}, errors.New("failed to update prompt")
	}
	return prompt_u, nil
}

// DeleteChatPrompt deletes a chat prompt by ID.
func (s *ChatPromptService) DeleteChatPrompt(ctx context.Context, id int32) error {
	err := s.q.DeleteChatPrompt(ctx, id)
	if err != nil {
		return errors.New("failed to delete prompt")
	}
	return nil
}

// GetAllChatPrompts returns all chat prompts.
func (s *ChatPromptService) GetAllChatPrompts(ctx context.Context) ([]sqlc_queries.ChatPrompt, error) {
	prompts, err := s.q.GetAllChatPrompts(ctx)
	if err != nil {
		return nil, errors.New("failed to retrieve prompts")
	}
	return prompts, nil
}

func (s *ChatPromptService) GetChatPromptsByUserID(ctx context.Context, userID int32) ([]sqlc_queries.ChatPrompt, error) {
	prompts, err := s.q.GetChatPromptsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return prompts, nil
}

func (s *ChatPromptService) GetChatPromptsBySessionUUID(ctx context.Context, session_uuid string) ([]sqlc_queries.ChatPrompt, error) {
	prompts, err := s.q.GetChatPromptsBySessionUUID(ctx, session_uuid)
	if err != nil {
		return nil, err
	}
	return prompts, nil
}

// DeleteChatPromptByUUID
func (s *ChatPromptService) DeleteChatPromptByUUID(ctx context.Context, uuid string) error {
	err := s.q.DeleteChatPromptByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	return nil
}

// UpdateChatPromptByUUID
func (s *ChatPromptService) UpdateChatPromptByUUID(ctx context.Context, uuid string, content string) (sqlc_queries.ChatPrompt, error) {
	tokenCount, _ := getTokenCount(content)
	params := sqlc_queries.UpdateChatPromptByUUIDParams{
		Uuid:       uuid,
		Content:    content,
		TokenCount: int32(tokenCount),
	}
	prompt, err := s.q.UpdateChatPromptByUUID(ctx, params)
	return prompt, err
}

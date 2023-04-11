package main

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/ai"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatMessageService struct {
	q *sqlc_queries.Queries
}

// NewChatMessageService creates a new ChatMessageService.
func NewChatMessageService(q *sqlc_queries.Queries) *ChatMessageService {
	return &ChatMessageService{q: q}
}

// CreateChatMessage creates a new chat message.
func (s *ChatMessageService) CreateChatMessage(ctx context.Context, message_params sqlc_queries.CreateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	message, err := s.q.CreateChatMessage(ctx, message_params)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
}

// GetChatMessageByID returns a chat message by ID.
func (s *ChatMessageService) GetChatMessageByID(ctx context.Context, id int32) (sqlc_queries.ChatMessage, error) {
	message, err := s.q.GetChatMessageByID(ctx, id)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
}

// UpdateChatMessage updates an existing chat message.
func (s *ChatMessageService) UpdateChatMessage(ctx context.Context, message_params sqlc_queries.UpdateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	message_u, err := s.q.UpdateChatMessage(ctx, message_params)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to update message ")
	}
	return message_u, nil
}

// DeleteChatMessage deletes a chat message by ID.
func (s *ChatMessageService) DeleteChatMessage(ctx context.Context, id int32) error {
	err := s.q.DeleteChatMessage(ctx, id)
	if err != nil {
		return eris.Wrap(err, "failed to delete message ")
	}
	return nil
}

// DeleteChatMessageByUUID deletes a chat message by uuid
func (s *ChatMessageService) DeleteChatMessageByUUID(ctx context.Context, uuid string) error {
	err := s.q.DeleteChatMessageByUUID(ctx, uuid)
	if err != nil {
		return eris.Wrap(err, "failed to delete message ")
	}
	return nil
}

// GetAllChatMessages returns all chat messages.
func (s *ChatMessageService) GetAllChatMessages(ctx context.Context) ([]sqlc_queries.ChatMessage, error) {
	messages, err := s.q.GetAllChatMessages(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve messages ")
	}
	return messages, nil
}

func (s *ChatMessageService) GetLatestMessagesBySessionID(ctx context.Context, chatSessionUuid string, limit int32) ([]sqlc_queries.ChatMessage, error) {
	params := sqlc_queries.GetLatestMessagesBySessionUUIDParams{ChatSessionUuid: chatSessionUuid, Limit: limit}
	msgs, err := s.q.GetLatestMessagesBySessionUUID(ctx, params)
	if err != nil {
		return []sqlc_queries.ChatMessage{}, err
	}
	return msgs, nil
}

func (s *ChatMessageService) GetFirstMessageBySessionUUID(ctx context.Context, chatSessionUuid string) (sqlc_queries.ChatMessage, error) {
	msg, err := s.q.GetFirstMessageBySessionUUID(ctx, chatSessionUuid)
	if err != nil {
		return sqlc_queries.ChatMessage{}, err
	}
	return msg, nil
}

func (s *ChatMessageService) AddMessage(ctx context.Context, chatSessionUuid string, uuid string, role ai.Role, content string, raw []byte) (sqlc_queries.ChatMessage, error) {
	params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid: chatSessionUuid,
		Uuid:            uuid,
		Role:            role.String(),
		Content:         content,
		Raw:             json.RawMessage(raw),
	}
	msg, err := s.q.CreateChatMessage(ctx, params)
	if err != nil {
		return sqlc_queries.ChatMessage{}, err
	}
	return msg, nil
}

// GetChatMessageByUUID returns a chat message by ID.
func (s *ChatMessageService) GetChatMessageByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatMessage, error) {
	message, err := s.q.GetChatMessageByUUID(ctx, uuid)
	if err != nil {
		return sqlc_queries.ChatMessage{}, errors.New("failed to retrieve message")
	}
	return message, nil
}

// UpdateChatMessageByUUID updates an existing chat message.
func (s *ChatMessageService) UpdateChatMessageByUUID(ctx context.Context, message_params sqlc_queries.UpdateChatMessageByUUIDParams) (sqlc_queries.ChatMessage, error) {
	message_u, err := s.q.UpdateChatMessageByUUID(ctx, message_params)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to update message ")
	}
	return message_u, nil
}

// GetChatMessagesBySessionUUID returns a chat message by session uuid.
func (s *ChatMessageService) GetChatMessagesBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]sqlc_queries.ChatMessage, error) {
	param := sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid:   uuid,
		Offset: pageNum - 1,
		Limit:  pageSize,
	}
	message, err := s.q.GetChatMessagesBySessionUUID(ctx, param)
	if err != nil {
		return []sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to retrieve message ")
	}
	return message, nil
}

// GetChatHistoryBySessionUUID returns chat message related by session uuid.
func (s *ChatMessageService) GetChatHistoryBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]SimpleChatMessage, error) {

	chat_prompts, err := s.q.GetChatPromptsBySessionUUID(ctx, uuid)
	if err != nil {
		return nil, eris.Wrap(err, "fail to get prompt: ")
	}

	simple_prompts := lo.Map(chat_prompts, func(prompt sqlc_queries.ChatPrompt, idx int) SimpleChatMessage {
		return SimpleChatMessage{
			Uuid:      prompt.Uuid,
			DateTime:  prompt.UpdatedAt.Format("2006-01-02 15:04:05PM"),
			Text:      prompt.Content,
			Inversion: idx%2 == 0,
			Error:     false,
			Loading:   false,
			IsPrompt:  true,
			RequestOptions: RequestOption{
				Prompt: prompt.Content,
			},
		}
	})

	messages, err := s.q.GetChatMessagesBySessionUUID(ctx,
		sqlc_queries.GetChatMessagesBySessionUUIDParams{
			Uuid:   uuid,
			Offset: pageNum - 1,
			Limit:  pageSize,
		})
	if err != nil {
		return nil, eris.Wrap(err, "fail to get message: ")
	}

	simple_msgs := lo.Map(messages, func(message sqlc_queries.ChatMessage, _ int) SimpleChatMessage {
		return SimpleChatMessage{
			Uuid:      message.Uuid,
			DateTime:  message.UpdatedAt.Format("2006-01-02 15:04:05PM"),
			Text:      message.Content,
			Inversion: message.Role == "user",
			Error:     false,
			Loading:   false,
			RequestOptions: RequestOption{
				Prompt: message.Content,
			},
		}
	})

	msgs := append(simple_prompts, simple_msgs...)
	return msgs, nil
}

// DeleteChatMessagesBySesionUUID deletes chat messages by session uuid.
func (s *ChatMessageService) DeleteChatMessagesBySesionUUID(ctx context.Context, uuid string) error {
	err := s.q.DeleteChatMessagesBySesionUUID(ctx, uuid)
	return err
}

func (s *ChatMessageService) GetChatMessagesCount(ctx context.Context, userID int32) (int32, error) {
	count, err := s.q.GetChatMessagesCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

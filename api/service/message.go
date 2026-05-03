package service

import (
	"context"
	"encoding/json"

	"github.com/swuecho/chat_backend/ai"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// MessageService provides methods for chat message management.
type MessageService struct {
	q *sqlc_queries.Queries
}

func NewMessageService(q *sqlc_queries.Queries) *MessageService {
	return &MessageService{q: q}
}

func (s *MessageService) Q() *sqlc_queries.Queries { return s.q }

func (s *MessageService) Create(ctx context.Context, params sqlc_queries.CreateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	return s.q.CreateChatMessage(ctx, params)
}

func (s *MessageService) GetByID(ctx context.Context, id int32) (sqlc_queries.ChatMessage, error) {
	return s.q.GetChatMessageByID(ctx, id)
}

func (s *MessageService) Update(ctx context.Context, params sqlc_queries.UpdateChatMessageParams) (sqlc_queries.ChatMessage, error) {
	return s.q.UpdateChatMessage(ctx, params)
}

func (s *MessageService) Delete(ctx context.Context, id int32) error {
	return s.q.DeleteChatMessage(ctx, id)
}

func (s *MessageService) GetAll(ctx context.Context) ([]sqlc_queries.ChatMessage, error) {
	return s.q.GetAllChatMessages(ctx)
}

func (s *MessageService) GetByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatMessage, error) {
	return s.q.GetChatMessageByUUID(ctx, uuid)
}

func (s *MessageService) UpdateByUUID(ctx context.Context, params sqlc_queries.UpdateChatMessageByUUIDParams) (sqlc_queries.ChatMessage, error) {
	return s.q.UpdateChatMessageByUUID(ctx, params)
}

func (s *MessageService) DeleteByUUID(ctx context.Context, uuid string) error {
	return s.q.DeleteChatMessageByUUID(ctx, uuid)
}

func (s *MessageService) GetLatestBySessionUUID(ctx context.Context, uuid string, limit int32) ([]sqlc_queries.ChatMessage, error) {
	return s.q.GetLatestMessagesBySessionUUID(ctx, sqlc_queries.GetLatestMessagesBySessionUUIDParams{
		ChatSessionUuid: uuid,
		Limit:           limit,
	})
}

func (s *MessageService) GetFirstBySessionUUID(ctx context.Context, uuid string) (sqlc_queries.ChatMessage, error) {
	return s.q.GetFirstMessageBySessionUUID(ctx, uuid)
}

func (s *MessageService) GetBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]sqlc_queries.ChatMessage, error) {
	return s.q.GetChatMessagesBySessionUUID(ctx, sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid:   uuid,
		Offset: pageNum - 1,
		Limit:  pageSize,
	})
}

func (s *MessageService) DeleteBySessionUUID(ctx context.Context, uuid string) error {
	return s.q.DeleteChatMessagesBySesionUUID(ctx, uuid)
}

func (s *MessageService) GetCount(ctx context.Context, userID int32) (int32, error) {
	count, err := s.q.GetChatMessagesCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (s *MessageService) AddMessage(ctx context.Context, sessionUUID, msgUUID string, role ai.Role, content string, raw []byte) (sqlc_queries.ChatMessage, error) {
	return s.q.CreateChatMessage(ctx, sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid: sessionUUID,
		Uuid:            msgUUID,
		Role:            role.String(),
		Content:         content,
		Raw:             json.RawMessage(raw),
	})
}

// UpdateContent updates the content and token count of a message.
func (s *MessageService) UpdateContent(ctx context.Context, uuid, content string) error {
	numTokens, err := getTokenCount(content)
	if err != nil {
		numTokens = len(content) / 4
	}
	return s.q.UpdateChatMessageContent(ctx, sqlc_queries.UpdateChatMessageContentParams{
		Uuid:       uuid,
		Content:    content,
		TokenCount: int32(numTokens),
	})
}

// UpdateSuggestions updates the suggested questions for a message.
func (s *MessageService) UpdateSuggestions(ctx context.Context, uuid string, suggestedQuestions json.RawMessage) error {
	_, err := s.q.UpdateChatMessageSuggestions(ctx, sqlc_queries.UpdateChatMessageSuggestionsParams{
		Uuid:               uuid,
		SuggestedQuestions: suggestedQuestions,
	})
	return err
}

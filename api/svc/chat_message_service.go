package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/ai"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatMessage struct {
	ID                                                                        int32
	Uuid, ChatSessionUuid, Role, Content, ReasoningContent, Model, LlmSummary string
	Score                                                                     float64
	UserID                                                                    int32
	CreatedAt, UpdatedAt                                                      time.Time
	CreatedBy, UpdatedBy                                                      int32
	IsDeleted, IsPin                                                          bool
	TokenCount                                                                int32
	Raw, Artifacts, SuggestedQuestions                                        json.RawMessage
}

func chatMessageFromRecord(m sqlc_queries.ChatMessage) ChatMessage {
	return ChatMessage{ID: m.ID, Uuid: m.Uuid, ChatSessionUuid: m.ChatSessionUuid, Role: m.Role,
		Content: m.Content, ReasoningContent: m.ReasoningContent, Model: m.Model, LlmSummary: m.LlmSummary,
		Score: m.Score, UserID: m.UserID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
		CreatedBy: m.CreatedBy, UpdatedBy: m.UpdatedBy, IsDeleted: m.IsDeleted, IsPin: m.IsPin,
		TokenCount: m.TokenCount, Raw: m.Raw, Artifacts: m.Artifacts, SuggestedQuestions: m.SuggestedQuestions}
}

func chatMessagesFromRecords(records []sqlc_queries.ChatMessage) []ChatMessage {
	result := make([]ChatMessage, 0, len(records))
	for _, record := range records {
		result = append(result, chatMessageFromRecord(record))
	}
	return result
}

type ChatMessageService struct {
	q *sqlc_queries.Queries
}

type CreateChatMessageInput struct {
	ChatSessionUuid    string
	Uuid               string
	Role               string
	Content            string
	ReasoningContent   string
	Model              string
	TokenCount         int32
	Score              float64
	UserID             int32
	CreatedBy          int32
	UpdatedBy          int32
	LlmSummary         string
	Raw                json.RawMessage
	Artifacts          json.RawMessage
	SuggestedQuestions json.RawMessage
}

type UpdateChatMessageInput struct {
	ID                 int32
	Role               string
	Content            string
	Score              float64
	UserID             int32
	UpdatedBy          int32
	Artifacts          json.RawMessage
	SuggestedQuestions json.RawMessage
}

type UpdateChatMessageByUUIDInput struct {
	Uuid               string
	Content            string
	IsPin              bool
	TokenCount         int32
	Artifacts          json.RawMessage
	SuggestedQuestions json.RawMessage
	UserID             int32
}

type DeleteChatMessageCommand struct{ ID, UserID int32 }
type DeleteChatMessageByUUIDCommand struct {
	UUID   string
	UserID int32
}
type DeleteSessionMessagesCommand struct {
	SessionUUID string
	UserID      int32
}

// NewChatMessageService creates a new ChatMessageService.
func NewChatMessageService(q *sqlc_queries.Queries) *ChatMessageService {
	return &ChatMessageService{q: q}
}

// CreateChatMessage creates a new chat message.
func (s *ChatMessageService) CreateChatMessage(ctx context.Context, input CreateChatMessageInput) (ChatMessage, error) {
	message, err := s.q.CreateChatMessage(ctx, sqlc_queries.CreateChatMessageParams(input))
	if err != nil {
		return ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return chatMessageFromRecord(message), nil
}

// GetChatMessageByID returns a chat message by ID.
func (s *ChatMessageService) GetChatMessageByID(ctx context.Context, id, userID int32) (ChatMessage, error) {
	message, err := s.q.GetChatMessageByID(ctx, sqlc_queries.GetChatMessageByIDParams{ID: id, UserID: userID})
	if err != nil {
		return ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return chatMessageFromRecord(message), nil
}

// UpdateChatMessage updates an existing chat message.
func (s *ChatMessageService) UpdateChatMessage(ctx context.Context, input UpdateChatMessageInput) (ChatMessage, error) {
	message_u, err := s.q.UpdateChatMessage(ctx, sqlc_queries.UpdateChatMessageParams(input))
	if err != nil {
		return ChatMessage{}, eris.Wrap(err, "failed to update message ")
	}
	return chatMessageFromRecord(message_u), nil
}

// DeleteChatMessage deletes a chat message by ID.
func (s *ChatMessageService) DeleteChatMessage(ctx context.Context, command DeleteChatMessageCommand) error {
	rows, err := s.q.DeleteChatMessage(ctx, sqlc_queries.DeleteChatMessageParams{ID: command.ID, UserID: command.UserID})
	if err != nil {
		return eris.Wrap(err, "failed to delete message ")
	}
	if rows == 0 {
		return domain.NotFound("Chat message", sql.ErrNoRows)
	}
	return nil
}

// DeleteChatMessageByUUID deletes a chat message by uuid
func (s *ChatMessageService) DeleteChatMessageByUUID(ctx context.Context, command DeleteChatMessageByUUIDCommand) error {
	rows, err := s.q.DeleteChatMessageByUUID(ctx, sqlc_queries.DeleteChatMessageByUUIDParams{Uuid: command.UUID, UserID: command.UserID})
	if err != nil {
		return eris.Wrap(err, "failed to delete message ")
	}
	if rows == 0 {
		return domain.NotFound("Chat message", sql.ErrNoRows)
	}
	return nil
}

// GetAllChatMessages returns all chat messages.
func (s *ChatMessageService) GetAllChatMessages(ctx context.Context, userID int32) ([]ChatMessage, error) {
	messages, err := s.q.GetAllChatMessages(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve messages ")
	}
	return chatMessagesFromRecords(messages), nil
}

func (s *ChatMessageService) GetLatestMessagesBySessionID(ctx context.Context, chatSessionUuid string, limit int32) ([]ChatMessage, error) {
	params := sqlc_queries.GetLatestMessagesBySessionUUIDParams{ChatSessionUuid: chatSessionUuid, Limit: limit}
	msgs, err := s.q.GetLatestMessagesBySessionUUID(ctx, params)
	if err != nil {
		return []ChatMessage{}, err
	}
	return chatMessagesFromRecords(msgs), nil
}

func (s *ChatMessageService) GetFirstMessageBySessionUUID(ctx context.Context, chatSessionUuid string) (ChatMessage, error) {
	msg, err := s.q.GetFirstMessageBySessionUUID(ctx, chatSessionUuid)
	if err != nil {
		return ChatMessage{}, err
	}
	return chatMessageFromRecord(msg), nil
}

func (s *ChatMessageService) AddMessage(ctx context.Context, chatSessionUuid string, uuid string, role ai.Role, content string, raw []byte) (ChatMessage, error) {
	params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid: chatSessionUuid,
		Uuid:            uuid,
		Role:            role.String(),
		Content:         content,
		Raw:             json.RawMessage(raw),
	}
	msg, err := s.q.CreateChatMessage(ctx, params)
	if err != nil {
		return ChatMessage{}, err
	}
	return chatMessageFromRecord(msg), nil
}

// GetChatMessageByUUID returns a chat message by ID.
func (s *ChatMessageService) GetChatMessageByUUID(ctx context.Context, uuid string, userID int32) (ChatMessage, error) {
	message, err := s.q.GetChatMessageByUUID(ctx, sqlc_queries.GetChatMessageByUUIDParams{Uuid: uuid, UserID: userID})
	if err != nil {
		return ChatMessage{}, errors.New("failed to retrieve message")
	}
	return chatMessageFromRecord(message), nil
}

// UpdateChatMessageByUUID updates an existing chat message.
func (s *ChatMessageService) UpdateChatMessageByUUID(ctx context.Context, input UpdateChatMessageByUUIDInput) (ChatMessage, error) {
	message_u, err := s.q.UpdateChatMessageByUUID(ctx, sqlc_queries.UpdateChatMessageByUUIDParams(input))
	if err != nil {
		return ChatMessage{}, eris.Wrap(err, "failed to update message ")
	}
	return chatMessageFromRecord(message_u), nil
}

func (s *ChatMessageService) UpdateSuggestedQuestions(ctx context.Context, uuid string, userID int32, suggestions json.RawMessage) (ChatMessage, error) {
	message, err := s.q.UpdateChatMessageSuggestions(ctx, sqlc_queries.UpdateChatMessageSuggestionsParams{
		Uuid:               uuid,
		SuggestedQuestions: suggestions,
		UserID:             userID,
	})
	return chatMessageFromRecord(message), err
}

// GetChatMessagesBySessionUUID returns a chat message by session uuid.
func (s *ChatMessageService) GetChatMessagesBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]ChatMessage, error) {
	param := sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid:   uuid,
		Offset: pageNum - 1,
		Limit:  pageSize,
	}
	message, err := s.q.GetChatMessagesBySessionUUID(ctx, param)
	if err != nil {
		return []ChatMessage{}, eris.Wrap(err, "failed to retrieve message ")
	}
	return chatMessagesFromRecords(message), nil
}

// DeleteChatMessagesBySesionUUID deletes chat messages by session uuid.
func (s *ChatMessageService) DeleteChatMessagesBySesionUUID(ctx context.Context, command DeleteSessionMessagesCommand) error {
	rows, err := s.q.DeleteChatMessagesBySesionUUID(ctx, sqlc_queries.DeleteChatMessagesBySesionUUIDParams{ChatSessionUuid: command.SessionUUID, UserID: command.UserID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.NotFound("Chat messages", sql.ErrNoRows)
	}
	return nil
}

func (s *ChatMessageService) GetChatMessagesCount(ctx context.Context, userID int32) (int32, error) {
	count, err := s.q.GetChatMessagesCount(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

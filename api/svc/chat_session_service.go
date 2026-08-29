package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatSessionService provides methods for interacting with chat sessions.
type ChatSessionService struct {
	q     *sqlc_queries.Queries
	newID func() string
	tx    SnapshotCopyTransactionManager
}

type ChatSession struct {
	ID              int32
	UserID          int32
	Uuid            string
	Topic           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Active          bool
	Model           string
	MaxLength       int32
	Temperature     float64
	TopP            float64
	MaxTokens       int32
	N               int32
	SummarizeMode   bool
	WorkspaceID     *int32
	ArtifactEnabled bool
	Debug           bool
	ExploreMode     bool
}

type CreateChatSessionCommand struct {
	UserID    int32
	Topic     string
	MaxLength int32
	UUID      string
	Model     string
}

type UpdateChatSessionCommand struct {
	ID     int32
	UserID int32
	Topic  string
	Active bool
}

type UpdateChatSessionByUUIDCommand struct {
	UUID   string
	UserID int32
	Topic  string
}

type DeleteChatSessionCommand struct {
	UUID   string
	UserID int32
}

type UpdateSessionMaxLengthCommand struct {
	UUID      string
	UserID    int32
	MaxLength int32
}

func chatSessionFromRecord(s sqlc_queries.ChatSession) ChatSession {
	var workspaceID *int32
	if s.WorkspaceID.Valid {
		value := s.WorkspaceID.Int32
		workspaceID = &value
	}
	return ChatSession{ID: s.ID, UserID: s.UserID, Uuid: s.Uuid, Topic: s.Topic,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt, Active: s.Active, Model: s.Model,
		MaxLength: s.MaxLength, Temperature: s.Temperature, TopP: s.TopP,
		MaxTokens: s.MaxTokens, N: s.N, SummarizeMode: s.SummarizeMode,
		WorkspaceID: workspaceID, ArtifactEnabled: s.ArtifactEnabled, Debug: s.Debug,
		ExploreMode: s.ExploreMode}
}

func (s ChatSession) ToRawMessage() *json.RawMessage {
	encoded, err := json.Marshal(map[string]any{
		"id": s.ID, "userId": s.UserID, "uuid": s.Uuid, "topic": s.Topic,
		"createdAt": s.CreatedAt, "updatedAt": s.UpdatedAt, "active": s.Active,
		"model": s.Model, "maxLength": s.MaxLength, "temperature": s.Temperature,
		"topP": s.TopP, "maxTokens": s.MaxTokens, "n": s.N,
		"summarizeMode": s.SummarizeMode, "workspaceId": s.WorkspaceID,
		"artifactEnabled": s.ArtifactEnabled, "debug": s.Debug, "exploreMode": s.ExploreMode,
	})
	if err != nil {
		return nil
	}
	raw := json.RawMessage(encoded)
	return &raw
}

type RateLimit struct {
	ChatModelName string
	RateLimit     int32
}

type CreateOrUpdateChatSessionInput struct {
	Uuid            string
	UserID          int32
	Topic           string
	MaxLength       int32
	Temperature     float64
	Model           string
	MaxTokens       int32
	TopP            float64
	N               int32
	Debug           bool
	SummarizeMode   bool
	WorkspaceID     *int32
	ExploreMode     bool
	ArtifactEnabled bool
}

// NewChatSessionService creates a new ChatSessionService.
func NewChatSessionService(q *sqlc_queries.Queries) *ChatSessionService {
	return &ChatSessionService{q: q, newID: util.NewUUID, tx: newSQLCTransactionManager(q)}
}

// CreateChatSession creates a new chat session.
func (s *ChatSessionService) CreateChatSession(ctx context.Context, command CreateChatSessionCommand) (ChatSession, error) {
	session, err := s.q.CreateChatSession(ctx, sqlc_queries.CreateChatSessionParams{UserID: command.UserID, Topic: command.Topic, MaxLength: command.MaxLength, Uuid: command.UUID, Model: command.Model})
	if err != nil {
		return ChatSession{}, err
	}
	return chatSessionFromRecord(session), nil
}

// GetChatSessionByID returns a chat session by ID.
func (s *ChatSessionService) GetChatSessionByID(ctx context.Context, id int32) (ChatSession, error) {
	session, err := s.q.GetChatSessionByID(ctx, id)
	if err != nil {
		return ChatSession{}, eris.Wrap(err, "failed to retrieve session: ")
	}
	return chatSessionFromRecord(session), nil
}

// UpdateChatSession updates an existing chat session.
func (s *ChatSessionService) UpdateChatSession(ctx context.Context, command UpdateChatSessionCommand) (ChatSession, error) {
	session_u, err := s.q.UpdateChatSession(ctx, sqlc_queries.UpdateChatSessionParams{ID: command.ID, UserID: command.UserID, Topic: command.Topic, Active: command.Active})
	if err != nil {
		return ChatSession{}, eris.Wrap(err, "failed to update session")
	}
	return chatSessionFromRecord(session_u), nil
}

// DeleteChatSession deletes a chat session by ID.
func (s *ChatSessionService) DeleteChatSession(ctx context.Context, id int32) error {
	err := s.q.DeleteChatSession(ctx, id)
	if err != nil {
		return eris.Wrap(err, "failed to delete session by id")
	}
	return nil
}

// GetAllChatSessions returns all chat sessions.
func (s *ChatSessionService) GetAllChatSessions(ctx context.Context) ([]ChatSession, error) {
	sessions, err := s.q.GetAllChatSessions(ctx)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve sessions")
	}
	result := make([]ChatSession, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, chatSessionFromRecord(session))
	}
	return result, nil
}

func (s *ChatSessionService) GetChatSessionsByUserID(ctx context.Context, userID int32) ([]ChatSession, error) {
	sessions, err := s.q.GetChatSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve sessions")
	}
	result := make([]ChatSession, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, chatSessionFromRecord(session))
	}
	return result, nil
}

func (s *ChatSessionService) GetSimpleChatSessionsByUserID(ctx context.Context, userID int32) ([]SimpleChatSession, error) {
	sessions, err := s.q.GetSessionsGroupedByWorkspace(ctx, userID)
	if err != nil {
		return nil, err
	}

	simpleSessions := make([]SimpleChatSession, 0, len(sessions))
	for _, session := range sessions {
		workspaceUuid := ""
		if session.WorkspaceUuid.Valid {
			workspaceUuid = session.WorkspaceUuid.String
		}

		simpleSessions = append(simpleSessions, SimpleChatSession{
			UUID:            session.Uuid,
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
			WorkspaceUUID:   workspaceUuid,
		})
	}
	return simpleSessions, nil
}

// GetChatSessionByUUID returns an authentication user record by ID.
func (s *ChatSessionService) GetChatSessionByUUID(ctx context.Context, uuid string) (ChatSession, error) {
	chatSession, err := s.q.GetChatSessionByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChatSession{}, domain.NotFound("Chat session", err)
		}
		return ChatSession{}, eris.Wrap(err, "failed to retrieve session by uuid, ")
	}
	return chatSessionFromRecord(chatSession), nil
}

// UpdateChatSessionByUUID updates an existing chat session.
func (s *ChatSessionService) UpdateChatSessionByUUID(ctx context.Context, command UpdateChatSessionByUUIDCommand) (ChatSession, error) {
	session_u, err := s.q.UpdateChatSessionByUUID(ctx, sqlc_queries.UpdateChatSessionByUUIDParams{Uuid: command.UUID, UserID: command.UserID, Topic: command.Topic})
	if err != nil {
		return ChatSession{}, eris.Wrap(err, "failed to update session, ")
	}
	return chatSessionFromRecord(session_u), nil
}

// UpdateChatSessionTopicByUUID updates an existing chat session topic.
func (s *ChatSessionService) UpdateChatSessionTopicByUUID(ctx context.Context, uuid string, userID int32, topic string) (ChatSession, error) {
	session_u, err := s.q.UpdateChatSessionTopicByUUID(ctx, sqlc_queries.UpdateChatSessionTopicByUUIDParams{
		Uuid: uuid, UserID: userID, Topic: topic,
	})
	if err != nil {
		return ChatSession{}, eris.Wrap(err, "failed to update session, ")
	}
	return chatSessionFromRecord(session_u), nil
}

// CreateOrUpdateChatSessionByUUID updates an existing chat session.
func (s *ChatSessionService) CreateOrUpdateChatSessionByUUID(ctx context.Context, input CreateOrUpdateChatSessionInput) (ChatSession, error) {
	workspaceID := sql.NullInt32{}
	if input.WorkspaceID != nil {
		workspaceID = sql.NullInt32{Int32: *input.WorkspaceID, Valid: true}
	}
	session_u, err := s.q.CreateOrUpdateChatSessionByUUID(ctx, sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid: input.Uuid, UserID: input.UserID, Topic: input.Topic,
		MaxLength: input.MaxLength, Temperature: input.Temperature, Model: input.Model,
		MaxTokens: input.MaxTokens, TopP: input.TopP, N: input.N, Debug: input.Debug,
		SummarizeMode: input.SummarizeMode, WorkspaceID: workspaceID,
		ExploreMode: input.ExploreMode, ArtifactEnabled: input.ArtifactEnabled,
	})
	if err != nil {
		return ChatSession{}, eris.Wrap(err, "failed to update session, ")
	}
	return chatSessionFromRecord(session_u), nil
}

// DeleteChatSessionByUUID deletes a chat session by UUID.
func (s *ChatSessionService) DeleteChatSessionByUUID(ctx context.Context, command DeleteChatSessionCommand) error {
	rows, err := s.q.DeleteChatSessionByUUID(ctx, sqlc_queries.DeleteChatSessionByUUIDParams{Uuid: command.UUID, UserID: command.UserID})
	if err != nil {
		return eris.Wrap(err, "failed to delete session by uuid, ")
	}
	if rows == 0 {
		return domain.NotFound("Chat session", sql.ErrNoRows)
	}
	return nil
}

// UpdateSessionMaxLength
func (s *ChatSessionService) UpdateSessionMaxLength(ctx context.Context, command UpdateSessionMaxLengthCommand) (ChatSession, error) {
	session_u, err := s.q.UpdateSessionMaxLength(ctx, sqlc_queries.UpdateSessionMaxLengthParams{
		Uuid: command.UUID, MaxLength: command.MaxLength, UserID: command.UserID,
	})
	if err != nil {
		return ChatSession{}, eris.Wrap(err, "failed to update session, ")
	}
	return chatSessionFromRecord(session_u), nil
}

// GetChatSessionByUUIDWithInActive returns a session by UUID including inactive ones.
func (s *ChatSessionService) GetChatSessionByUUIDWithInActive(ctx context.Context, uuid string) (ChatSession, error) {
	session, err := s.q.GetChatSessionByUUIDWithInActive(ctx, uuid)
	return chatSessionFromRecord(session), eris.Wrap(err, "failed to get session with inactive")
}

func estimatePromptTokens(text string) int32 {
	tokenCount, err := provider.GetTokenCount(text)
	if err != nil {
		tokenCount = len(text) / tokenEstimateRatio
	}
	if tokenCount <= 0 {
		tokenCount = 1
	}
	return int32(tokenCount)
}

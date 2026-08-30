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

type ConversationMessage struct{ Role, Content string }
type SessionHistoryMessage struct {
	UUID, DateTime, Text, Model                string
	Inversion, Error, Loading, IsPin, IsPrompt bool
	Artifacts                                  []domain.Artifact
	SuggestedQuestions                         []string
}
type SessionSnapshot struct{ Session, Conversation json.RawMessage }
type SessionHistoryQuery struct {
	SessionUUID string
	Page        PageWindow
}
type ConversationMessagesPageQuery struct {
	SessionUUID string
	Page        PageWindow
}
type AdminSessionMessage struct {
	ID                                           int32
	UUID, Role, Content, ReasoningContent, Model string
	TokenCount, UserID                           int32
	CreatedAt, UpdatedAt                         time.Time
}
type RuntimeModel struct {
	Name, URL, APIAuthHeader, APIAuthKey, APIType string
	EnablePerModelRateLimit                       bool
}

type SessionConversationService struct {
	q     *sqlc_queries.Queries
	newID func() string
}

func NewSessionConversationService(q *sqlc_queries.Queries) *SessionConversationService {
	return &SessionConversationService{q: q, newID: util.NewUUID}
}

func (s *SessionConversationService) HasSystemPrompt(ctx context.Context, sessionUUID string) (bool, error) {
	_, err := s.q.GetOneChatPromptBySessionUUID(ctx, sessionUUID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, eris.Wrap(err, "failed to get chat prompt")
}

func (s *SessionConversationService) MessagesPage(ctx context.Context, query ConversationMessagesPageQuery) ([]ConversationMessage, error) {
	msgs, err := s.q.GetChatMessagesBySessionUUID(ctx, sqlc_queries.GetChatMessagesBySessionUUIDParams{Uuid: query.SessionUUID, Offset: query.Page.Offset, Limit: query.Page.Limit})
	if err != nil {
		return nil, eris.Wrap(err, "failed to get chat messages")
	}
	result := make([]ConversationMessage, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, ConversationMessage{Role: m.Role, Content: m.Content})
	}
	return result, nil
}

func (s *SessionConversationService) History(ctx context.Context, query SessionHistoryQuery) ([]SessionHistoryMessage, error) {
	page := PageRequest{Page: query.Page.Offset/query.Page.Limit + 1, Size: query.Page.Limit}
	rows, err := s.q.GetChatHistoryBySessionUUID(ctx, query.SessionUUID, page.Page, page.Size)
	if err != nil {
		return nil, err
	}
	result := make([]SessionHistoryMessage, 0, len(rows))
	for _, row := range rows {
		arts := make([]domain.Artifact, 0, len(row.Artifacts))
		for _, a := range row.Artifacts {
			arts = append(arts, domain.Artifact{UUID: a.UUID, Type: a.Type, Title: a.Title, Content: a.Content, Language: a.Language})
		}
		result = append(result, SessionHistoryMessage{UUID: row.Uuid, DateTime: row.DateTime, Text: row.Text, Model: row.Model, Inversion: row.Inversion, Error: row.Error, Loading: row.Loading, IsPin: row.IsPin, IsPrompt: row.IsPrompt, Artifacts: arts, SuggestedQuestions: row.SuggestedQuestions})
	}
	return result, nil
}

func (s *SessionConversationService) EnsureSystemPrompt(ctx context.Context, sessionUUID string, userID int32, text string) error {
	if exists, err := s.HasSystemPrompt(ctx, sessionUUID); err != nil || exists {
		return err
	}
	text = normalizedSystemPrompt(text)
	_, err := s.q.CreateChatPrompt(ctx, sqlc_queries.CreateChatPromptParams{Uuid: s.newID(), ChatSessionUuid: sessionUUID, Role: "system", Content: text, TokenCount: estimatePromptTokens(text), UserID: userID, CreatedBy: userID, UpdatedBy: userID})
	return eris.Wrap(err, "failed to create default system prompt")
}

type SessionRateLimitService struct{ q *sqlc_queries.Queries }

func NewSessionRateLimitService(q *sqlc_queries.Queries) *SessionRateLimitService {
	return &SessionRateLimitService{q: q}
}
func (s *SessionRateLimitService) Check(ctx context.Context, sessionUUID string, userID int32) (RateLimit, error) {
	rate, err := s.q.RateLimiteByUserAndSessionUUID(ctx, sqlc_queries.RateLimiteByUserAndSessionUUIDParams{Uuid: sessionUUID, UserID: userID})
	return RateLimit{ChatModelName: rate.ChatModelName, RateLimit: rate.RateLimit}, err
}
func (s *SessionRateLimitService) Usage(ctx context.Context, userID int32, model string) (int64, error) {
	return s.q.GetChatMessagesCountByUserAndModel(ctx, sqlc_queries.GetChatMessagesCountByUserAndModelParams{UserID: userID, Model: model})
}

type SessionSnapshotQueryService struct{ q *sqlc_queries.Queries }

func NewSessionSnapshotQueryService(q *sqlc_queries.Queries) *SessionSnapshotQueryService {
	return &SessionSnapshotQueryService{q: q}
}
func (s *SessionSnapshotQueryService) ByUserAndUUID(ctx context.Context, userID int32, uuid string) (SessionSnapshot, error) {
	row, err := s.q.ChatSnapshotByUserIdAndUuid(ctx, sqlc_queries.ChatSnapshotByUserIdAndUuidParams{UserID: userID, Uuid: uuid})
	return SessionSnapshot{Session: row.Session, Conversation: row.Conversation}, err
}

type SessionAdminQueryService struct{ q *sqlc_queries.Queries }

func NewSessionAdminQueryService(q *sqlc_queries.Queries) *SessionAdminQueryService {
	return &SessionAdminQueryService{q: q}
}
func (s *SessionAdminQueryService) Messages(ctx context.Context, uuid string) ([]AdminSessionMessage, error) {
	rows, err := s.q.GetChatMessagesBySessionUUIDForAdmin(ctx, uuid)
	if err != nil {
		return nil, err
	}
	result := make([]AdminSessionMessage, 0, len(rows))
	for _, r := range rows {
		result = append(result, AdminSessionMessage{ID: r.ID, UUID: r.Uuid, Role: r.Role, Content: r.Content, ReasoningContent: r.ReasoningContent, Model: r.Model, TokenCount: r.TokenCount, UserID: r.UserID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt})
	}
	return result, nil
}

type SessionBotHistoryService struct{ q *sqlc_queries.Queries }

func NewSessionBotHistoryService(q *sqlc_queries.Queries) *SessionBotHistoryService {
	return &SessionBotHistoryService{q: q}
}
func (s *SessionBotHistoryService) Save(ctx context.Context, input CreateBotAnswerHistoryInput) error {
	_, err := s.q.CreateBotAnswerHistory(ctx, createBotAnswerHistoryParams(input))
	return eris.Wrap(err, "failed to save bot answer history")
}

type SessionModelService struct{ q *sqlc_queries.Queries }

func NewSessionModelService(q *sqlc_queries.Queries) *SessionModelService {
	return &SessionModelService{q: q}
}
func (s *SessionModelService) ByName(ctx context.Context, name string) (RuntimeModel, error) {
	m, err := s.q.ChatModelByName(ctx, name)
	return runtimeModelFromRecord(m), eris.Wrap(err, "failed to get chat model")
}
func (s *SessionModelService) GenerateTitle(ctx context.Context, text string) (string, error) {
	m, err := s.q.GetTitleChatModel(ctx)
	if err != nil {
		return "", err
	}
	return provider.GenerateChatTitle(ctx, providerRuntimeModel(runtimeModelFromRecord(m)), text)
}

func runtimeModelFromRecord(m sqlc_queries.ChatModel) RuntimeModel {
	return RuntimeModel{Name: m.Name, URL: m.Url, APIAuthHeader: m.ApiAuthHeader, APIAuthKey: m.ApiAuthKey, APIType: m.ApiType, EnablePerModelRateLimit: m.EnablePerModeRatelimit}
}
func providerRuntimeModel(m RuntimeModel) provider.ModelConfig {
	return provider.ModelConfig{Name: m.Name, URL: m.URL, APIAuthHeader: m.APIAuthHeader, APIAuthKey: m.APIAuthKey, APIType: m.APIType, EnablePerModelRateLimit: m.EnablePerModelRateLimit}
}

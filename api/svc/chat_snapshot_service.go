package svc

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/pkg/util"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatSnapshotService provides methods for chat snapshot management.
type ChatSnapshotService struct {
	q *sqlc_queries.Queries
}

type ChatSnapshot struct {
	ID                          int32
	Type, UUID                  string
	UserID                      int32
	Title, Summary, Model       string
	Tags, Session, Conversation json.RawMessage
	CreatedAt                   time.Time
	Text                        string
}

type ChatSnapshotSummary struct {
	UUID, Title, Summary string
	Tags                 json.RawMessage
	CreatedAt            time.Time
	Type                 string
}
type ChatSnapshotSearchResult struct {
	UUID, Title string
	Rank        float32
}
type UpdateChatBotSettingsCommand struct {
	UUID                  string
	UserID                int32
	Title, Summary, Model string
}
type UpdateChatBotModelCommand struct {
	UUID   string
	UserID int32
	Model  string
}

func chatSnapshotFromRecord(r sqlc_queries.ChatSnapshot) ChatSnapshot {
	return ChatSnapshot{ID: r.ID, Type: r.Typ, UUID: r.Uuid, UserID: r.UserID, Title: r.Title, Summary: r.Summary, Model: r.Model, Tags: r.Tags, Session: r.Session, Conversation: r.Conversation, CreatedAt: r.CreatedAt, Text: r.Text}
}

// NewChatSnapshotService creates a new ChatSnapshotService.
func NewChatSnapshotService(q *sqlc_queries.Queries) *ChatSnapshotService {
	return &ChatSnapshotService{q: q}
}

// --- Query wrappers ---

func (s *ChatSnapshotService) ChatSnapshotByUUID(ctx context.Context, uuid string) (ChatSnapshot, error) {
	r, err := s.q.ChatSnapshotByUUID(ctx, uuid)
	return chatSnapshotFromRecord(r), err
}

func (s *ChatSnapshotService) ChatSnapshotMetaByUserID(ctx context.Context, userID int32, typ string, limit, offset int32) ([]ChatSnapshotSummary, error) {
	rows, err := s.q.ChatSnapshotMetaByUserID(ctx, sqlc_queries.ChatSnapshotMetaByUserIDParams{UserID: userID, Typ: typ, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	result := make([]ChatSnapshotSummary, 0, len(rows))
	for _, r := range rows {
		result = append(result, ChatSnapshotSummary{UUID: r.Uuid, Title: r.Title, Summary: r.Summary, Tags: r.Tags, CreatedAt: r.CreatedAt, Type: r.Typ})
	}
	return result, nil
}

func (s *ChatSnapshotService) ChatSnapshotCountByUserIDAndType(ctx context.Context, userID int32, typ string) (int64, error) {
	return s.q.ChatSnapshotCountByUserIDAndType(ctx, sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams{UserID: userID, Column2: typ})
}

func (s *ChatSnapshotService) UpdateChatSnapshotMetaByUUID(ctx context.Context, uuid, title, summary string, userID int32) error {
	return s.q.UpdateChatSnapshotMetaByUUID(ctx, sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{Uuid: uuid, Title: title, Summary: summary, UserID: userID})
}

func (s *ChatSnapshotService) UpdateChatBotModel(ctx context.Context, command UpdateChatBotModelCommand) (ChatSnapshot, error) {
	r, err := s.q.UpdateChatBotModel(ctx, sqlc_queries.UpdateChatBotModelParams{Uuid: command.UUID, BotUserID: command.UserID, InputModel: command.Model})
	return chatSnapshotFromRecord(r), err
}

func (s *ChatSnapshotService) UpdateChatBotSettings(ctx context.Context, command UpdateChatBotSettingsCommand) (ChatSnapshot, error) {
	r, err := s.q.UpdateChatBotSettings(ctx, sqlc_queries.UpdateChatBotSettingsParams{Uuid: command.UUID, BotUserID: command.UserID, InputTitle: command.Title, InputSummary: command.Summary, InputModel: command.Model})
	return chatSnapshotFromRecord(r), err
}

func (s *ChatSnapshotService) DeleteChatSnapshot(ctx context.Context, uuid string, userID int32) error {
	_, err := s.q.DeleteChatSnapshot(ctx, sqlc_queries.DeleteChatSnapshotParams{Uuid: uuid, UserID: userID})
	return err
}

func (s *ChatSnapshotService) ChatSnapshotSearch(ctx context.Context, userID int32, search string) ([]ChatSnapshotSearchResult, error) {
	rows, err := s.q.ChatSnapshotSearch(ctx, sqlc_queries.ChatSnapshotSearchParams{UserID: userID, Search: search})
	if err != nil {
		return nil, err
	}
	result := make([]ChatSnapshotSearchResult, 0, len(rows))
	for _, r := range rows {
		result = append(result, ChatSnapshotSearchResult{UUID: r.Uuid, Title: r.Title, Rank: r.Rank})
	}
	return result, nil
}

// --- Business operations ---

func (s *ChatSnapshotService) CreateChatSnapshot(ctx context.Context, chatSessionUuid string, userId int32) (string, error) {
	chatSession, err := s.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		return "", err
	}
	if chatSession.UserID != userId {
		return "", domain.Forbidden("chat session does not belong to user")
	}
	simple_msgs, err := s.q.GetChatHistoryBySessionUUID(ctx, chatSessionUuid, int32(1), int32(10000))
	if err != nil {
		return "", err
	}
	text := lo.Reduce(simple_msgs, func(acc string, curr sqlc_queries.SimpleChatMessage, _ int) string {
		return acc + curr.Text
	}, "")
	title := GenTitle(s.q, ctx, chatSession, text)
	simple_msgs_raw, err := json.Marshal(simple_msgs)
	if err != nil {
		return "", err
	}
	snapshot_uuid := util.NewUUID()
	chatSessionMsg, err := json.Marshal(chatSession)
	if err != nil {
		return "", err
	}
	one, err := s.q.CreateChatSnapshot(ctx, sqlc_queries.CreateChatSnapshotParams{
		Uuid: snapshot_uuid, Model: chatSession.Model, Title: title, UserID: userId,
		Session: chatSessionMsg, Tags: json.RawMessage([]byte("{}")),
		Text: text, Conversation: simple_msgs_raw,
	})
	if err != nil {
		slog.Info("error", "error", err)
		return "", err
	}
	return one.Uuid, nil
}

func GenTitle(q *sqlc_queries.Queries, ctx context.Context, chatSession sqlc_queries.ChatSession, text string) string {
	title := provider.FirstN(chatSession.Topic, 100)
	model, err := q.GetTitleChatModel(ctx)
	if err == nil {
		genTitle, err := provider.GenerateChatTitle(ctx, providerModel(model), text)
		if err != nil {
			slog.Info("error", "error", err)
		}
		if genTitle != "" {
			title = genTitle
		}
	}
	return title
}

func (s *ChatSnapshotService) CreateChatBot(ctx context.Context, chatSessionUuid string, userId int32) (string, error) {
	chatSession, err := s.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		return "", err
	}
	if chatSession.UserID != userId {
		return "", domain.Forbidden("chat session does not belong to user")
	}
	simple_msgs, err := s.q.GetChatHistoryBySessionUUID(ctx, chatSessionUuid, int32(1), int32(10000))
	text := lo.Reduce(simple_msgs, func(acc string, curr sqlc_queries.SimpleChatMessage, _ int) string {
		return acc + curr.Text
	}, "")
	if err != nil {
		return "", err
	}
	simple_msgs_raw, err := json.Marshal(simple_msgs)
	if err != nil {
		return "", err
	}
	snapshot_uuid := util.NewUUID()
	chatSessionMsg, err := json.Marshal(chatSession)
	if err != nil {
		return "", err
	}
	title := GenTitle(s.q, ctx, chatSession, text)
	one, err := s.q.CreateChatBot(ctx, sqlc_queries.CreateChatBotParams{
		Uuid: snapshot_uuid, Model: chatSession.Model, Typ: "chatbot",
		Title: title, UserID: userId, Session: chatSessionMsg,
		Tags: json.RawMessage([]byte("{}")), Text: text, Conversation: simple_msgs_raw,
	})
	if err != nil {
		slog.Info("error", "error", err)
		return "", err
	}
	return one.Uuid, nil
}

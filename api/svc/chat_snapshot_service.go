package svc

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatSnapshotService provides methods for chat snapshot management.
type ChatSnapshotService struct {
	q *sqlc_queries.Queries
}

// NewChatSnapshotService creates a new ChatSnapshotService.
func NewChatSnapshotService(q *sqlc_queries.Queries) *ChatSnapshotService {
	return &ChatSnapshotService{q: q}
}

// --- Query wrappers ---

func (s *ChatSnapshotService) ChatSnapshotByUUID(ctx context.Context, uuid string) (sqlc_queries.ChatSnapshot, error) {
	return s.q.ChatSnapshotByUUID(ctx, uuid)
}

func (s *ChatSnapshotService) ChatSnapshotMetaByUserID(ctx context.Context, userID int32, typ string, limit, offset int32) ([]sqlc_queries.ChatSnapshotMetaByUserIDRow, error) {
	return s.q.ChatSnapshotMetaByUserID(ctx, sqlc_queries.ChatSnapshotMetaByUserIDParams{UserID: userID, Typ: typ, Limit: limit, Offset: offset})
}

func (s *ChatSnapshotService) ChatSnapshotCountByUserIDAndType(ctx context.Context, userID int32, typ string) (int64, error) {
	return s.q.ChatSnapshotCountByUserIDAndType(ctx, sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams{UserID: userID, Column2: typ})
}

func (s *ChatSnapshotService) UpdateChatSnapshotMetaByUUID(ctx context.Context, uuid, title, summary string, userID int32) error {
	return s.q.UpdateChatSnapshotMetaByUUID(ctx, sqlc_queries.UpdateChatSnapshotMetaByUUIDParams{Uuid: uuid, Title: title, Summary: summary, UserID: userID})
}

func (s *ChatSnapshotService) UpdateChatBotModel(ctx context.Context, uuid string, userID int32, model string) (sqlc_queries.ChatSnapshot, error) {
	return s.q.UpdateChatBotModel(ctx, sqlc_queries.UpdateChatBotModelParams{Uuid: uuid, BotUserID: userID, InputModel: model})
}

func (s *ChatSnapshotService) UpdateChatBotSettings(ctx context.Context, uuid string, userID int32, title, summary, model string) (sqlc_queries.ChatSnapshot, error) {
	return s.q.UpdateChatBotSettings(ctx, sqlc_queries.UpdateChatBotSettingsParams{Uuid: uuid, BotUserID: userID, InputTitle: title, InputSummary: summary, InputModel: model})
}

func (s *ChatSnapshotService) DeleteChatSnapshot(ctx context.Context, uuid string, userID int32) error {
	_, err := s.q.DeleteChatSnapshot(ctx, sqlc_queries.DeleteChatSnapshotParams{Uuid: uuid, UserID: userID})
	return err
}

func (s *ChatSnapshotService) ChatSnapshotSearch(ctx context.Context, userID int32, search string) ([]sqlc_queries.ChatSnapshotSearchRow, error) {
	return s.q.ChatSnapshotSearch(ctx, sqlc_queries.ChatSnapshotSearchParams{UserID: userID, Search: search})
}

// --- Business operations ---

func (s *ChatSnapshotService) CreateChatSnapshot(ctx context.Context, chatSessionUuid string, userId int32) (string, error) {
	chatSession, err := s.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		return "", err
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
	snapshot_uuid := uuid.New().String()
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
	snapshot_uuid := uuid.New().String()
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

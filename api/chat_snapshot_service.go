package main

import (
	"context"
	"encoding/json"
	"log"

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

func (s *ChatSnapshotService) ChatSnapshotMetaByUserID(ctx context.Context, params sqlc_queries.ChatSnapshotMetaByUserIDParams) ([]sqlc_queries.ChatSnapshotMetaByUserIDRow, error) {
	return s.q.ChatSnapshotMetaByUserID(ctx, params)
}

func (s *ChatSnapshotService) ChatSnapshotCountByUserIDAndType(ctx context.Context, params sqlc_queries.ChatSnapshotCountByUserIDAndTypeParams) (int64, error) {
	return s.q.ChatSnapshotCountByUserIDAndType(ctx, params)
}

func (s *ChatSnapshotService) UpdateChatSnapshotMetaByUUID(ctx context.Context, params sqlc_queries.UpdateChatSnapshotMetaByUUIDParams) error {
	return s.q.UpdateChatSnapshotMetaByUUID(ctx, params)
}

func (s *ChatSnapshotService) DeleteChatSnapshot(ctx context.Context, params sqlc_queries.DeleteChatSnapshotParams) error {
	_, err := s.q.DeleteChatSnapshot(ctx, params)
	return err
}

func (s *ChatSnapshotService) ChatSnapshotSearch(ctx context.Context, params sqlc_queries.ChatSnapshotSearchParams) ([]sqlc_queries.ChatSnapshotSearchRow, error) {
	return s.q.ChatSnapshotSearch(ctx, params)
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
		log.Println(err)
		return "", err
	}
	return one.Uuid, nil
}

func GenTitle(q *sqlc_queries.Queries, ctx context.Context, chatSession sqlc_queries.ChatSession, text string) string {
	title := firstN(chatSession.Topic, 100)
	model := "gemini-2.0-flash"
	_, err := q.ChatModelByName(ctx, model)
	if err == nil {
		genTitle, err := provider.GenerateChatTitle(ctx, model, text)
		if err != nil {
			log.Println(err)
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
		log.Println(err)
		return "", err
	}
	return one.Uuid, nil
}

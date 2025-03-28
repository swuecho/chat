package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatSnapshotService provides methods for interacting with chat sessions.
type ChatSnapshotService struct {
	q *sqlc_queries.Queries
}

// NewChatSnapshotService creates a new ChatSnapshotService.
func NewChatSnapshotService(q *sqlc_queries.Queries) *ChatSnapshotService {
	return &ChatSnapshotService{q: q}
}

func (s *ChatSnapshotService) CreateChatSnapshot(ctx context.Context, chatSessionUuid string, userId int32) (string, error) {
	chatSession, err := s.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		return "", err
	}
	// TODO: fix hardcode
	// Get chat history
	simple_msgs, err := s.q.GetChatHistoryBySessionUUID(ctx, chatSessionUuid, int32(1), int32(10000))
	if err != nil {
		return "", err
	}
	text := lo.Reduce(simple_msgs, func(acc string, curr sqlc_queries.SimpleChatMessage, _ int) string {
		return acc + curr.Text
	}, "")
	title := GenTitle(s.q, ctx, chatSession, text)
	// simple_msgs to RawMessage
	simple_msgs_raw, err := json.Marshal(simple_msgs)
	if err != nil {
		return "", err
	}
	snapshot_uuid := uuid.New().String()
	chatSessionMessage, err := json.Marshal(chatSession)
	if err != nil {
		return "", err
	}
	one, err := s.q.CreateChatSnapshot(ctx, sqlc_queries.CreateChatSnapshotParams{
		Uuid:         snapshot_uuid,
		Model:        chatSession.Model,
		Title:        title,
		UserID:       userId,
		Session:      chatSessionMessage,
		Tags:         json.RawMessage([]byte("{}")),
		Text:         text,
		Conversation: simple_msgs_raw,
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	return one.Uuid, nil

}

func GenTitle(q *sqlc_queries.Queries, ctx context.Context, chatSession sqlc_queries.ChatSession, text string) string {
	title := firstN(chatSession.Topic, 100)
	// generate title using
	model := "gemini-2.0-flash"
	_, err := q.ChatModelByName(ctx, model)
	if err == nil {
		genTitle, err := GenerateChatTitle(ctx, model, text)
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
	// TODO: fix hardcode
	simple_msgs, err := s.q.GetChatHistoryBySessionUUID(ctx, chatSessionUuid, int32(1), int32(10000))
	text := lo.Reduce(simple_msgs, func(acc string, curr sqlc_queries.SimpleChatMessage, _ int) string {
		return acc + curr.Text
	}, "")
	// save all simple_msgs to a jsonb field in chat_snapshot
	if err != nil {
		return "", err
	}
	// simple_msgs to RawMessage
	simple_msgs_raw, err := json.Marshal(simple_msgs)
	if err != nil {
		return "", err
	}
	snapshot_uuid := uuid.New().String()
	chatSessionMessage, err := json.Marshal(chatSession)
	if err != nil {
		return "", err
	}

	title := GenTitle(s.q, ctx, chatSession, text)
	one, err := s.q.CreateChatBot(ctx, sqlc_queries.CreateChatBotParams{
		Uuid:         snapshot_uuid,
		Model:        chatSession.Model,
		Typ:          "chatbot",
		Title:        title,
		UserID:       userId,
		Session:      chatSessionMessage,
		Tags:         json.RawMessage([]byte("{}")),
		Text:         text,
		Conversation: simple_msgs_raw,
	})
	if err != nil {
		log.Println(err)
		return "", err
	}
	return one.Uuid, nil

}

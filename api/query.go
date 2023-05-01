package main

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func GetChatHistoryBySessionUUID(q *sqlc_queries.Queries, ctx context.Context, uuid string, pageNum, pageSize int32) ([]SimpleChatMessage, error) {

	chat_prompts, err := q.GetChatPromptsBySessionUUID(ctx, uuid)
	if err != nil {
		return nil, eris.Wrap(err, "fail to get prompt: ")
	}

	simple_prompts := lo.Map(chat_prompts, func(prompt sqlc_queries.ChatPrompt, idx int) SimpleChatMessage {
		return SimpleChatMessage{
			Uuid:      prompt.Uuid,
			DateTime:  prompt.UpdatedAt.Format(time.RFC3339),
			Text:      prompt.Content,
			Inversion: idx%2 == 0,
			Error:     false,
			Loading:   false,
			IsPin:     false,
			IsPrompt:  true,
		}
	})

	messages, err := q.GetChatMessagesBySessionUUID(ctx,
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
			DateTime:  message.UpdatedAt.Format(time.RFC3339),
			Text:      message.Content,
			Inversion: message.Role == "user",
			Error:     false,
			Loading:   false,
			IsPin:     message.IsPin,
		}
	})

	msgs := append(simple_prompts, simple_msgs...)
	return msgs, nil
}

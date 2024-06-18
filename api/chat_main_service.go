package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	models "github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatService struct {
	q *sqlc_queries.Queries
}

// NewChatSessionService creates a new ChatSessionService.
func NewChatService(q *sqlc_queries.Queries) *ChatService {
	return &ChatService{q: q}
}

func (s *ChatService) getAskMessages(chatSession sqlc_queries.ChatSession, chatUuid string, regenerate bool) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	chatSessionUuid := chatSession.Uuid

	lastN := chatSession.MaxLength
	if chatSession.MaxLength == 0 {
		lastN = 10
	}

	chat_prompts, err := s.q.GetChatPromptsBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		return nil, eris.Wrap(err, "fail to get prompt: ")
	}

	var chat_massages []sqlc_queries.ChatMessage
	if regenerate {
		chat_massages, err = s.q.GetLastNChatMessages(ctx,
			sqlc_queries.GetLastNChatMessagesParams{
				ChatSessionUuid: chatSessionUuid,
				Uuid:            chatUuid,
				Limit:           lastN,
			})

	} else {
		chat_massages, err = s.q.GetLatestMessagesBySessionUUID(ctx,
			sqlc_queries.GetLatestMessagesBySessionUUIDParams{ChatSessionUuid: chatSession.Uuid, Limit: lastN})
	}

	if err != nil {
		return nil, eris.Wrap(err, "fail to get messages: ")
	}
	chat_prompt_msgs := lo.Map(chat_prompts, func(m sqlc_queries.ChatPrompt, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})
	chat_message_msgs := lo.Map(chat_massages, func(m sqlc_queries.ChatMessage, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})
	msgs := append(chat_prompt_msgs, chat_message_msgs...)
	return msgs, nil
}

func (s *ChatService) CreateChatPromptSimple(chatSessionUuid string, newQuestion string, userID int32) (sqlc_queries.ChatPrompt, error) {
	tokenCount, _ := getTokenCount(newQuestion)
	chatPrompt, err := s.q.CreateChatPrompt(context.Background(),
		sqlc_queries.CreateChatPromptParams{
			Uuid:            NewUUID(),
			ChatSessionUuid: chatSessionUuid,
			Role:            "system",
			Content:         newQuestion,
			UserID:          userID,
			CreatedBy:       userID,
			UpdatedBy:       userID,
			TokenCount:      int32(tokenCount),
		})
	return chatPrompt, err
}

// CreateChatMessage creates a new chat message.
func (s *ChatService) CreateChatMessageSimple(ctx context.Context, sessionUuid, uuid, role, content string, userId int32, baseURL string, is_summarize_mode bool) (sqlc_queries.ChatMessage, error) {
	numTokens, err := getTokenCount(content)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to get token count: "))
	}

	summary := ""

	if is_summarize_mode && numTokens > 300 {
		log.Println("summarizing")
		summary = llm_summarize_with_timeout(baseURL, content)
		log.Println("summarizing: " + summary)
	}

	chatMessage := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid: sessionUuid,
		Uuid:            uuid,
		Role:            role,
		Content:         content,
		UserID:          userId,
		CreatedBy:       userId,
		UpdatedBy:       userId,
		LlmSummary:      summary,
		TokenCount:      int32(numTokens),
		Raw:             json.RawMessage([]byte("{}")),
	}
	message, err := s.q.CreateChatMessage(ctx, chatMessage)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
}

// UpdateChatMessageContent
func (s *ChatService) UpdateChatMessageContent(ctx context.Context, uuid, content string) error {
	// encode
	// num_tokens
	num_tokens, err := getTokenCount(content)
	if err != nil {
		log.Println(eris.Wrap(err, "getTokenCount: "))
	}

	err = s.q.UpdateChatMessageContent(ctx, sqlc_queries.UpdateChatMessageContentParams{
		Uuid:       uuid,
		Content:    content,
		TokenCount: int32(num_tokens),
	})
	return err
}

func (s *ChatService) logChat(chatSession sqlc_queries.ChatSession, msgs []models.Message, answerText string) {
	// log chat
	sessionRaw := chatSession.ToRawMessage()
	if sessionRaw == nil {
		log.Println("failed to marshal chat session")
		return
	}
	question, err := json.Marshal(msgs)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to marshal chat messages"))
	}
	answerRaw, err := json.Marshal(answerText)
	if err != nil {
		log.Println(eris.Wrap(err, "failed to marshal answer"))
	}

	s.q.CreateChatLog(context.Background(), sqlc_queries.CreateChatLogParams{
		Session:  *sessionRaw,
		Question: question,
		Answer:   answerRaw,
	})
}

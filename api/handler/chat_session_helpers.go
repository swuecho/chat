package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/svc"
)

var titleGenSemaphore = make(chan struct{}, 5)

type answerEventChunkSink struct{ events *provider.AnswerEventWriter }

func (s answerEventChunkSink) WriteAnswerChunk(_ context.Context, chunk provider.StreamChunk) error {
	return s.events.Emit(provider.AnswerEvent{Type: provider.AnswerEventDelta, AnswerID: chunk.ID, Delta: chunk.Content})
}

func (h *ChatHandler) generateSessionTitle(session *svc.ChatSession, userID int32) {
	ctx, cancel := context.WithTimeout(context.Background(), sessionTitleGenerationTimeout)
	defer cancel()
	messages, err := h.conversationSvc.MessagesPage(ctx, svc.ConversationMessagesPageQuery{SessionUUID: session.UUID, Page: svc.PageWindow{Limit: 100}})
	if err != nil {
		slog.Warn("Failed to get messages for title generation", "error", err)
		return
	}
	var chatText strings.Builder
	for _, message := range messages {
		fmt.Fprintf(&chatText, "%s: %s\n", message.Role, message.Content)
	}
	if strings.TrimSpace(chatText.String()) == "" {
		return
	}
	title, err := h.modelSvc.GenerateTitle(ctx, chatText.String())
	if err != nil || title == "" {
		return
	}
	if _, err := h.sessionSvc.UpdateChatSessionTopicByUUID(ctx, session.UUID, userID, title); err != nil {
		slog.Warn("Failed to update session title", "error", err)
		return
	}
	slog.Info("Generated LLM title", "sessionUUID", session.UUID, "title", title)
}

func (h *ChatHandler) launchSessionTitleGeneration(session *svc.ChatSession, userID int32) {
	select {
	case titleGenSemaphore <- struct{}{}:
		go func() {
			defer func() { <-titleGenSemaphore }()
			h.generateSessionTitle(session, userID)
		}()
	default:
		slog.Info("Skipping title generation because all worker slots are busy", "sessionUUID", session.UUID)
	}
}

func (h *ChatHandler) sendSuggestedQuestionsStream(events *provider.AnswerEventWriter, answerID string, raw json.RawMessage) {
	var questions []string
	if err := json.Unmarshal(raw, &questions); err != nil || len(questions) == 0 {
		return
	}
	_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventSuggested, AnswerID: answerID, SuggestedQuestions: questions})
}

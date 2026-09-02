package svc

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/swuecho/chat_backend/provider"
)

type CompleteChatCommand struct {
	SessionUUID, RequestUUID, Prompt string
	UserID                           int32
	Stream                           bool
}

type CompleteChatResult struct {
	Session    ChatSession
	Answer     GenerateAnswerResult
	Replay     *ChatMessage
	InProgress bool
}

type CompleteChatUseCase struct {
	chat         *ChatService
	lifecycle    ChatRequestLifecycle
	sessions     *ChatSessionService
	conversation *SessionConversationService
	models       ModelSelector
	chunks       AnswerChunkSink
	logPolicy    ChatLogPolicy
	audit        ChatAuditLogger
}

func NewCompleteChatUseCase(chat *ChatService, sessions *ChatSessionService, conversation *SessionConversationService, models ModelSelector, chunks AnswerChunkSink, logPolicy ChatLogPolicy, audit ChatAuditLogger) *CompleteChatUseCase {
	return &CompleteChatUseCase{chat: chat, lifecycle: chat, sessions: sessions, conversation: conversation, models: models, chunks: chunks, logPolicy: logPolicy, audit: audit}
}

func (u *CompleteChatUseCase) Execute(ctx context.Context, command CompleteChatCommand) (CompleteChatResult, error) {
	session, err := u.sessions.GetOwnedChatSession(ctx, GetOwnedChatSessionQuery{UUID: command.SessionUUID, UserID: command.UserID})
	if err != nil {
		return CompleteChatResult{}, err
	}
	model, err := u.chat.ProviderModel(ctx, session.Model)
	if err != nil {
		return CompleteChatResult{}, err
	}
	baseURL, _ := provider.GetModelBaseURL(model.URL)

	if err := u.lifecycle.Claim(ctx, command.RequestUUID, session.UUID, command.UserID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return CompleteChatResult{}, err
		}
		request, reconcileErr := u.lifecycle.State(ctx, command.RequestUUID, session.UUID, command.UserID)
		if reconcileErr != nil {
			return CompleteChatResult{}, reconcileErr
		}
		if request.Status != ChatRequestCompleted || request.AssistantUUID == "" {
			return CompleteChatResult{Session: session, InProgress: true}, nil
		}
		message, loadErr := u.chat.GetChatMessageByUUID(ctx, request.AssistantUUID, command.UserID)
		if loadErr != nil {
			return CompleteChatResult{}, loadErr
		}
		return CompleteChatResult{Session: session, Replay: &message}, nil
	}

	hasPrompt, err := u.conversation.HasSystemPrompt(ctx, session.UUID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = u.lifecycle.Fail(ctx, command.RequestUUID, session.UUID, command.UserID, "prompt_persistence_failed")
		return CompleteChatResult{}, err
	}
	if !hasPrompt {
		if _, err := u.chat.CreateChatPromptSimple(ctx, session.UUID, defaultSystemPromptText, command.UserID); err != nil {
			_ = u.lifecycle.Fail(ctx, command.RequestUUID, session.UUID, command.UserID, "prompt_persistence_failed")
			return CompleteChatResult{}, err
		}
	}
	if command.Prompt != "" {
		if _, err := u.chat.CreateChatMessageSimple(ctx, session.UUID, command.RequestUUID, "user", command.Prompt, "", session.Model, command.UserID, baseURL, session.SummarizeMode); err != nil {
			_ = u.lifecycle.Fail(ctx, command.RequestUUID, session.UUID, command.UserID, "prompt_persistence_failed")
			return CompleteChatResult{}, err
		}
		if !hasPrompt {
			words := strings.Fields(command.Prompt)
			if len(words) > 10 {
				words = words[:10]
			}
			if title := strings.Join(words, " "); title != "" {
				_, _ = u.sessions.UpdateChatSessionTopicByUUID(ctx, session.UUID, command.UserID, title)
			}
		}
	}
	answer, err := NewGenerateAnswerUseCase(u.chat, u.models, u.chunks, u.logPolicy, u.audit).Execute(ctx, GenerateAnswerCommand{Session: session, RequestUUID: command.RequestUUID, UserID: command.UserID, BaseURL: baseURL, Stream: command.Stream})
	if err != nil {
		return CompleteChatResult{}, err
	}
	return CompleteChatResult{Session: session, Answer: answer}, nil
}

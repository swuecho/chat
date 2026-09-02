package svc

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
)

type GenerateBotAnswerCommand struct {
	Session      ChatSession
	Messages     []models.Message
	SnapshotUUID string
	Question     string
	UserID       int32
	Stream       bool
}

type BotAnswerHistoryWriter interface {
	Save(context.Context, CreateBotAnswerHistoryInput) error
}

type GenerateBotAnswerUseCase struct {
	chat      *ChatService
	models    ModelSelector
	chunks    AnswerChunkSink
	logPolicy ChatLogPolicy
	history   BotAnswerHistoryWriter
}

func NewGenerateBotAnswerUseCase(chat *ChatService, models ModelSelector, chunks AnswerChunkSink, logPolicy ChatLogPolicy, history BotAnswerHistoryWriter) *GenerateBotAnswerUseCase {
	if chunks == nil {
		chunks = discardAnswerChunks{}
	}
	if logPolicy == nil {
		logPolicy = allowChatLogs{}
	}
	return &GenerateBotAnswerUseCase{chat: chat, models: models, chunks: chunks, logPolicy: logPolicy, history: history}
}

func (u *GenerateBotAnswerUseCase) Execute(ctx context.Context, command GenerateBotAnswerCommand) (models.LLMAnswer, error) {
	if u.models == nil || u.history == nil {
		return models.LLMAnswer{}, errors.New("bot answer dependencies are incomplete")
	}
	messages := append(append([]models.Message(nil), command.Messages...), models.Message{Role: "user", Content: command.Question})
	request, err := u.chat.ProviderRequest(ctx, command.Session, messages, "", false, command.Stream)
	if err != nil {
		return models.LLMAnswer{}, err
	}
	model, err := u.models.SelectModel(ctx, command.Session, messages)
	if err != nil {
		return models.LLMAnswer{}, err
	}
	answer, err := provider.ConsumeStream(ctx, model, request, func(chunk provider.StreamChunk) error {
		return u.chunks.WriteAnswerChunk(ctx, chunk)
	})
	if err != nil {
		return models.LLMAnswer{}, err
	}
	if answer == nil {
		return models.LLMAnswer{}, errors.New("provider returned no final answer")
	}
	if err := u.history.Save(ctx, CreateBotAnswerHistoryInput{BotUUID: command.SnapshotUUID, UserID: command.UserID,
		Prompt: command.Question, Answer: answer.Answer, Model: command.Session.Model, TokensUsed: int32(len(answer.Answer)) / tokenEstimateRatio}); err != nil {
		return models.LLMAnswer{}, err
	}
	if u.logPolicy.ShouldLogChat(messages) {
		u.chat.LogChat(ctx, command.Session, messages, answer.Answer)
	}
	return *answer, nil
}

type RegenerateAnswerCommand struct {
	SessionUUID string
	MessageUUID string
	UserID      int32
	Stream      bool
}

type RegenerateAnswerResult struct {
	Answer             models.LLMAnswer
	SuggestedQuestions json.RawMessage
}

type RegenerateAnswerUseCase struct {
	chat      *ChatService
	sessions  *ChatSessionService
	models    ModelSelector
	chunks    AnswerChunkSink
	logPolicy ChatLogPolicy
}

func NewRegenerateAnswerUseCase(chat *ChatService, sessions *ChatSessionService, models ModelSelector, chunks AnswerChunkSink, logPolicy ChatLogPolicy) *RegenerateAnswerUseCase {
	if chunks == nil {
		chunks = discardAnswerChunks{}
	}
	if logPolicy == nil {
		logPolicy = allowChatLogs{}
	}
	return &RegenerateAnswerUseCase{chat: chat, sessions: sessions, models: models, chunks: chunks, logPolicy: logPolicy}
}

func (u *RegenerateAnswerUseCase) Execute(ctx context.Context, command RegenerateAnswerCommand) (RegenerateAnswerResult, error) {
	if u.models == nil {
		return RegenerateAnswerResult{}, errors.New("model selector is required")
	}
	session, err := u.sessions.GetOwnedChatSession(ctx, GetOwnedChatSessionQuery{UUID: command.SessionUUID, UserID: command.UserID})
	if err != nil {
		return RegenerateAnswerResult{}, err
	}
	messages, err := u.chat.GetAskMessages(ctx, session, command.MessageUUID, true)
	if err != nil {
		return RegenerateAnswerResult{}, err
	}
	request, err := u.chat.ProviderRequest(ctx, session, messages, command.MessageUUID, true, command.Stream)
	if err != nil {
		return RegenerateAnswerResult{}, err
	}
	model, err := u.models.SelectModel(ctx, session, messages)
	if err != nil {
		return RegenerateAnswerResult{}, err
	}
	answer, err := provider.ConsumeStream(ctx, model, request, func(chunk provider.StreamChunk) error {
		return u.chunks.WriteAnswerChunk(ctx, chunk)
	})
	if err != nil {
		return RegenerateAnswerResult{}, err
	}
	if answer == nil {
		return RegenerateAnswerResult{}, errors.New("provider returned no final answer")
	}
	if err := u.chat.UpdateChatMessageContent(ctx, command.MessageUUID, command.UserID, answer.Answer); err != nil {
		return RegenerateAnswerResult{}, err
	}
	if u.logPolicy.ShouldLogChat(messages) {
		u.chat.LogChat(ctx, session, messages, answer.Answer)
	}
	var suggested json.RawMessage
	if session.ExploreMode {
		questions := u.chat.GenerateSuggestedQuestions(ctx, answer.Answer, messages)
		if len(questions) > 0 {
			suggested, err = json.Marshal(questions)
			if err != nil {
				return RegenerateAnswerResult{}, err
			}
			if err := u.chat.UpdateChatMessageSuggestions(ctx, command.MessageUUID, command.UserID, suggested); err != nil {
				return RegenerateAnswerResult{}, err
			}
		}
	}
	return RegenerateAnswerResult{Answer: *answer, SuggestedQuestions: suggested}, nil
}

package svc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
)

// GenerateAnswerCommand contains the application inputs for one normal answer.
type GenerateAnswerCommand struct {
	Session     ChatSession
	RequestUUID string
	UserID      int32
	BaseURL     string
	Stream      bool
}

// ModelSelector resolves the provider implementation for an application
// request. Implementations may enforce access and rate-limit policy.
type ModelSelector interface {
	SelectModel(context.Context, ChatSession, []models.Message) (provider.ChatModel, error)
}

// AnswerChunkSink receives provider chunks without coupling the application
// workflow to HTTP or SSE.
type AnswerChunkSink interface {
	WriteAnswerChunk(context.Context, provider.StreamChunk) error
}

// ChatLogPolicy decides whether a conversation should be written to the audit
// log (for example, demo/test conversations are excluded).
type ChatLogPolicy interface {
	ShouldLogChat([]models.Message) bool
}

type discardAnswerChunks struct{}

func (discardAnswerChunks) WriteAnswerChunk(context.Context, provider.StreamChunk) error { return nil }

type allowChatLogs struct{}

func (allowChatLogs) ShouldLogChat([]models.Message) bool { return true }

// GenerateAnswerUseCase owns the provider-to-persistence lifecycle.
type GenerateAnswerUseCase struct {
	chat      *ChatService
	models    ModelSelector
	chunks    AnswerChunkSink
	logPolicy ChatLogPolicy
}

func NewGenerateAnswerUseCase(chat *ChatService, models ModelSelector, chunks AnswerChunkSink, logPolicy ChatLogPolicy) *GenerateAnswerUseCase {
	if chunks == nil {
		chunks = discardAnswerChunks{}
	}
	if logPolicy == nil {
		logPolicy = allowChatLogs{}
	}
	return &GenerateAnswerUseCase{chat: chat, models: models, chunks: chunks, logPolicy: logPolicy}
}

type GenerateAnswerResult struct {
	Answer             models.LLMAnswer
	Messages           []models.Message
	SuggestedQuestions json.RawMessage
}

type GenerateAnswerError struct {
	Code string
	Err  error
}

func (e *GenerateAnswerError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *GenerateAnswerError) Unwrap() error { return e.Err }

func generationError(code string, err error) error {
	return &GenerateAnswerError{Code: code, Err: err}
}

// GenerateAnswer owns the provider-to-persistence lifecycle. Successful return
// guarantees that the final answer has been durably stored.
func (u *GenerateAnswerUseCase) Execute(ctx context.Context, command GenerateAnswerCommand) (GenerateAnswerResult, error) {
	fail := func(code string, err error) (GenerateAnswerResult, error) {
		failureCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = u.chat.MarkChatRequestFailed(failureCtx, command.RequestUUID, command.Session.UUID, command.UserID, code)
		return GenerateAnswerResult{}, generationError(code, err)
	}

	if u.models == nil {
		return fail("provider_selection_failed", errors.New("model selector is required"))
	}
	messages, err := u.chat.GetAskMessages(ctx, command.Session, command.RequestUUID, false)
	if err != nil {
		return fail("message_collection_failed", err)
	}
	if err := u.chat.MarkChatRequestStreaming(ctx, command.RequestUUID, command.Session.UUID, command.UserID); err != nil {
		return fail("request_state_failed", err)
	}

	request, err := u.chat.ProviderRequest(ctx, command.Session, messages, command.RequestUUID, false, command.Stream)
	if err != nil {
		return fail("provider_request_failed", err)
	}
	model, err := u.models.SelectModel(ctx, command.Session, messages)
	if err != nil {
		return fail("provider_selection_failed", err)
	}
	answer, err := provider.ConsumeStream(ctx, model, request, func(chunk provider.StreamChunk) error {
		return u.chunks.WriteAnswerChunk(ctx, chunk)
	})
	if err != nil {
		code := "generation_failed"
		if errors.Is(err, context.Canceled) {
			code = "canceled"
		} else if errors.Is(err, context.DeadlineExceeded) {
			code = "timeout"
		}
		return fail(code, err)
	}
	if answer == nil {
		return fail("empty_answer", errors.New("provider returned no final answer"))
	}

	if u.logPolicy.ShouldLogChat(messages) {
		u.chat.LogChat(ctx, command.Session, messages, answer.ReasoningContent+answer.Answer)
	}
	message, err := u.chat.CompleteChatRequestWithSuggestedQuestions(ctx, command.RequestUUID, command.Session.UUID, answer.AnswerId, answer.Answer, answer.ReasoningContent, command.Session.Model, command.UserID, command.BaseURL, command.Session.SummarizeMode, command.Session.ExploreMode, messages)
	if err != nil {
		return fail("persistence_failed", err)
	}
	return GenerateAnswerResult{Answer: *answer, Messages: messages, SuggestedQuestions: message.SuggestedQuestions}, nil
}

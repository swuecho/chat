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

// GenerateAnswerDependencies are request-scoped capabilities supplied by the
// transport composition root. They keep HTTP details out of the use case.
type GenerateAnswerDependencies struct {
	SelectModel func([]models.Message) provider.ChatModel
	OnChunk     func(provider.StreamChunk) error
	ShouldLog   func([]models.Message) bool
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
func (s *ChatService) GenerateAnswer(ctx context.Context, command GenerateAnswerCommand, deps GenerateAnswerDependencies) (GenerateAnswerResult, error) {
	fail := func(code string, err error) (GenerateAnswerResult, error) {
		failureCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = s.MarkChatRequestFailed(failureCtx, command.RequestUUID, command.Session.UUID, command.UserID, code)
		return GenerateAnswerResult{}, generationError(code, err)
	}

	if deps.SelectModel == nil {
		return fail("provider_selection_failed", errors.New("model selector is required"))
	}
	messages, err := s.GetAskMessages(ctx, command.Session, command.RequestUUID, false)
	if err != nil {
		return fail("message_collection_failed", err)
	}
	if err := s.MarkChatRequestStreaming(ctx, command.RequestUUID, command.Session.UUID, command.UserID); err != nil {
		return fail("request_state_failed", err)
	}

	request, err := s.ProviderRequest(ctx, command.Session, messages, command.RequestUUID, false, command.Stream)
	if err != nil {
		return fail("provider_request_failed", err)
	}
	answer, err := provider.ConsumeStream(ctx, deps.SelectModel(messages), request, deps.OnChunk)
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

	if deps.ShouldLog == nil || deps.ShouldLog(messages) {
		s.LogChat(ctx, command.Session, messages, answer.ReasoningContent+answer.Answer)
	}
	message, err := s.CompleteChatRequestWithSuggestedQuestions(ctx, command.RequestUUID, command.Session.UUID, answer.AnswerId, answer.Answer, answer.ReasoningContent, command.Session.Model, command.UserID, command.BaseURL, command.Session.SummarizeMode, command.Session.ExploreMode, messages)
	if err != nil {
		return fail("persistence_failed", err)
	}
	return GenerateAnswerResult{Answer: *answer, Messages: messages, SuggestedQuestions: message.SuggestedQuestions}, nil
}

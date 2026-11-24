package provider

import (
	"context"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/swuecho/chat_backend/domain"
)

// CompleteOptions configures an app-initiated, one-shot completion.
type CompleteOptions struct {
	MaxTokens   int32
	Temperature float64
	// ChatUUID labels the request in provider logs.
	ChatUUID    string
	OpenAIProxy string
}

// Complete sends messages to model through the provider for its API type and
// returns the full answer text. Retryable provider failures are retried twice.
func Complete(ctx context.Context, model ModelConfig, messages []domain.Message, opts CompleteOptions) (string, error) {
	chatModel := newInternalChatModel(model, newInternalCompletionHandler(opts.OpenAIProxy))

	exponential := backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(200*time.Millisecond),
		backoff.WithMaxInterval(2*time.Second),
		backoff.WithMultiplier(2),
		backoff.WithRandomizationFactor(0.2),
		backoff.WithMaxElapsedTime(0),
	)
	retryBackoff := backoff.WithContext(backoff.WithMaxRetries(exponential, 2), ctx)
	return backoff.RetryWithData(func() (string, error) {
		stream, err := chatModel.Stream(ctx, Request{
			Session: Session{Model: model.Name, MaxTokens: opts.MaxTokens, Temperature: opts.Temperature, TopP: 1, N: 1},
			Model:   model, Messages: messages, ChatUUID: opts.ChatUUID,
		})
		if err != nil {
			return "", stopRetryingUnlessProviderFailureIsRetryable(err)
		}
		var value string
		for chunk := range stream {
			if chunk.Err != nil {
				return "", stopRetryingUnlessProviderFailureIsRetryable(chunk.Err)
			}
			if chunk.FinalAnswer != nil {
				value = chunk.FinalAnswer.Answer
			} else if chunk.Content != "" {
				value += chunk.Content
			}
		}
		return value, nil
	}, retryBackoff)
}

func newInternalChatModel(model ModelConfig, h Handler) ChatModel {
	switch model.APIType {
	case "claude":
		return NewClaude3ChatModel(h)
	case "gemini":
		return NewGeminiChatModel(h)
	case "ollama":
		return NewOllamaChatModel(h)
	case "custom":
		return NewCustomChatModel(h)
	}
	url := strings.TrimSuffix(model.URL, "/")
	if strings.HasSuffix(url, "/completions") && !strings.HasSuffix(url, "/chat/completions") {
		return NewCompletionChatModel(h)
	}
	return NewOpenAIChatModel(h)
}

package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"log/slog"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/svc"
)

// chooseChatModel returns the appropriate ChatModel implementation based on session config.
func (h *ChatHandler) chooseChatModel(ctx context.Context, session svc.ChatSession, msgs []models.Message) provider.ChatModel {
	if isTest(msgs) {
		return provider.NewTestChatModel(h)
	}

	chatModel, err := h.service.ProviderModel(ctx, session.Model)
	if err != nil {
		return provider.NewOpenAIChatModel(h) // fallback
	}

	completionModels := mapset.NewSet[string]()
	isCompletion := completionModels.Contains(session.Model)

	switch chatModel.APIType {
	case "claude":
		return provider.NewClaude3ChatModel(h)
	case "ollama":
		return provider.NewOllamaChatModel(h)
	case "gemini":
		return provider.NewGeminiChatModel(h)
	case "custom":
		return provider.NewCustomChatModel(h)
	case "openai":
		if isCompletion {
			return provider.NewCompletionChatModel(h)
		}
		return provider.NewOpenAIChatModel(h)
	default:
		return provider.NewOpenAIChatModel(h)
	}
}

// SelectModel implements svc.ModelSelector for chat-generation use cases.
func (h *ChatHandler) SelectModel(ctx context.Context, session svc.ChatSession, msgs []models.Message) (provider.ChatModel, error) {
	return h.chooseChatModel(ctx, session, msgs), nil
}

// isTest returns true if any message starts with the test demo prefix.
func isTest(msgs []models.Message) bool {
	for _, msg := range msgs {
		if len(msg.Content) >= dto.TestPrefixLength && msg.Content[:dto.TestPrefixLength] == dto.TestDemoPrefix {
			return true
		}
	}
	return false
}

// ShouldLogChat implements svc.ChatLogPolicy.
func (h *ChatHandler) ShouldLogChat(msgs []models.Message) bool { return !isTest(msgs) }

// CheckModelAccess verifies the user hasn't exceeded per-model rate limits.
// Returns nil if access is allowed, or an error (dto.APIError) if denied.
func (h *ChatHandler) CheckModelAccess(ctx context.Context, chatSessionUuid, model string, userID int32) error {
	chatModel, err := h.modelSvc.ByName(ctx, model)
	if err != nil {
		slog.Error("Chat model not found", "error", err, "model", model)
		apiErr := dto.ErrResourceNotFound("chat model: " + model)
		return apiErr
	}

	if !chatModel.EnablePerModelRateLimit {
		return nil
	}

	rate, err := h.rateLimitSvc.Check(ctx, chatSessionUuid, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return dto.WrapError(dto.MapDatabaseError(err), "Failed to get rate limit")
	}

	usage10Min, err := h.rateLimitSvc.Usage(ctx, userID, rate.ChatModelName)
	if err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to get usage data").WithDebugInfo(err.Error())
	}

	if int32(usage10Min) > rate.RateLimit {
		apiErr := dto.ErrTooManyRequests
		apiErr.Message = fmt.Sprintf("Rate limit exceeded for %s", rate.ChatModelName)
		apiErr.Detail = fmt.Sprintf("Usage: %d, Limit: %d", usage10Min, rate.RateLimit)
		return apiErr
	}

	return nil
}

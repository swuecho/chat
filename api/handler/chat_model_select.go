package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"log/slog"
	mapset "github.com/deckarep/golang-set/v2"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// chooseChatModel returns the appropriate ChatModel implementation based on session config.
func (h *ChatHandler) chooseChatModel(session sqlc_queries.ChatSession, msgs []models.Message) provider.ChatModel {
	if isTest(msgs) {
		return provider.NewTestChatModel(h)
	}

	chatModel, err := provider.GetChatModel(h.RequestContext(), h.Queries(), session.Model)
	if err != nil {
		return provider.NewOpenAIChatModel(h) // fallback
	}

	completionModels := mapset.NewSet[string]()
	isCompletion := completionModels.Contains(session.Model)

	switch chatModel.ApiType {
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

// isTest returns true if any message starts with the test demo prefix.
func isTest(msgs []models.Message) bool {
	for _, msg := range msgs {
		if len(msg.Content) >= dto.TestPrefixLength && msg.Content[:dto.TestPrefixLength] == dto.TestDemoPrefix {
			return true
		}
	}
	return false
}

// CheckModelAccess verifies the user hasn't exceeded per-model rate limits.
func (h *ChatHandler) CheckModelAccess(w http.ResponseWriter, chatSessionUuid, model string, userID int32) bool {
	ctx := h.RequestContext()

	chatModel, err := h.sessionSvc.ChatModelByName(ctx, model)
	if err != nil {
		slog.Error("Chat model not found", "error", err, "model", model)
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model: "+model))
		return true
	}

	if !chatModel.EnablePerModeRatelimit {
		return false
	}

	rate, err := h.sessionSvc.RateLimitByUserAndSessionUUID(ctx, sqlc_queries.RateLimiteByUserAndSessionUUIDParams{
		Uuid: chatSessionUuid, UserID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		dto.RespondWithAPIError(w, dto.WrapError(dto.MapDatabaseError(err), "Failed to get rate limit"))
		return true
	}

	usage10Min, err := h.sessionSvc.GetChatMessagesCountByUserAndModel(ctx, sqlc_queries.GetChatMessagesCountByUserAndModelParams{
		UserID: userID, Model: rate.ChatModelName,
	})
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to get usage data").WithDebugInfo(err.Error()))
		return true
	}

	if int32(usage10Min) > rate.RateLimit {
		apiErr := dto.ErrTooManyRequests
		apiErr.Message = fmt.Sprintf("Rate limit exceeded for %s", rate.ChatModelName)
		apiErr.Detail = fmt.Sprintf("Usage: %d, Limit: %d", usage10Min, rate.RateLimit)
		dto.RespondWithAPIError(w, apiErr)
		return true
	}

	return false
}

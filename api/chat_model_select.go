// Package main — Chat model selection and rate-limit access checks.
package main

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// chooseChatModel returns the appropriate ChatModel implementation based on session config.
func (h *ChatHandler) chooseChatModel(session sqlc_queries.ChatSession, msgs []models.Message) ChatModel {
	if isTest(msgs) {
		return &TestChatModel{h: h}
	}

	chatModel, err := GetChatModel(h.GetRequestContext(), h.service.q, session.Model)
	if err != nil {
		return &OpenAIChatModel{h: h} // fallback
	}

	completionModels := mapset.NewSet[string]()
	isCompletion := completionModels.Contains(session.Model)

	switch chatModel.ApiType {
	case "claude":
		return &Claude3ChatModel{h: h}
	case "ollama":
		return &OllamaChatModel{h: h}
	case "gemini":
		return NewGeminiChatModel(h)
	case "custom":
		return &CustomChatModel{h: h}
	case "openai":
		if isCompletion {
			return &CompletionChatModel{h: h}
		}
		return &OpenAIChatModel{h: h}
	default:
		return &OpenAIChatModel{h: h}
	}
}

// isTest returns true if any message starts with the test demo prefix.
func isTest(msgs []models.Message) bool {
	for _, msg := range msgs {
		if len(msg.Content) >= TestPrefixLength && msg.Content[:TestPrefixLength] == TestDemoPrefix {
			return true
		}
	}
	return false
}

// CheckModelAccess verifies the user hasn't exceeded per-model rate limits.
// Returns true if rate limit is exceeded (caller should abort).
func (h *ChatHandler) CheckModelAccess(w http.ResponseWriter, chatSessionUuid, model string, userID int32) bool {
	ctx := h.GetRequestContext()

	chatModel, err := h.sessionSvc.ChatModelByName(ctx, model)
	if err != nil {
		log.WithError(err).WithField("model", model).Error("Chat model not found")
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+model))
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
		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get rate limit"))
		return true
	}

	usage10Min, err := h.sessionSvc.GetChatMessagesCountByUserAndModel(ctx, sqlc_queries.GetChatMessagesCountByUserAndModelParams{
		UserID: userID, Model: rate.ChatModelName,
	})
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to get usage data").WithDebugInfo(err.Error()))
		return true
	}

	if int32(usage10Min) > rate.RateLimit {
		apiErr := ErrTooManyRequests
		apiErr.Message = fmt.Sprintf("Rate limit exceeded for %s", rate.ChatModelName)
		apiErr.Detail = fmt.Sprintf("Usage: %d, Limit: %d", usage10Min, rate.RateLimit)
		RespondWithAPIError(w, apiErr)
		return true
	}

	return false
}

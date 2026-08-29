package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
	"log/slog"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/svc"
)

// --- Request types used by chat handlers ---

type ChatRequest struct {
	Prompt      string `json:"prompt"`
	SessionUuid string `json:"sessionUuid"`
	ChatUuid    string `json:"chatUuid"`
	Regenerate  bool   `json:"regenerate"`
	Stream      bool   `json:"stream,omitempty"`
}

type BotRequest struct {
	Message      string `json:"message"`
	SnapshotUuid string `json:"snapshot_uuid"`
	Stream       bool   `json:"stream"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message      openai.ChatCompletionMessage `json:"message"`
	FinishReason any                          `json:"finish_reason"`
	Index        int                          `json:"index"`
}

// --- Handler methods ---

// ChatBotCompletionHandler handles bot chat completion via snapshot.
func (h *ChatHandler) ChatBotCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req BotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		slog.Error("error getting user ID", "error", err)
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	chatSnapshot, err := h.snapshotSvc.ByUserAndUUID(ctx, userID, req.SnapshotUuid)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()))
		return
	}

	var session svc.ChatSession
	if err := json.Unmarshal(chatSnapshot.Session, &session); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to deserialize chat session").WithDebugInfo(err.Error()))
		return
	}

	var messages []dto.SimpleChatMessage
	if err := json.Unmarshal(chatSnapshot.Conversation, &messages); err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to deserialize conversation").WithDebugInfo(err.Error()))
		return
	}

	genBotAnswer(ctx, h, w, session, messages, req.SnapshotUuid, req.Message, userID, req.Stream)
}

// ChatCompletionHandler handles regular chat completion with streaming support.
func (h *ChatHandler) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("error decoding request", "error", err)
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}
	if req.SessionUuid == "" || req.ChatUuid == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("sessionUuid and chatUuid are required"))
		return
	}
	if !req.Regenerate && req.Prompt == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("prompt is required"))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		slog.Error("error getting user ID", "error", err)
		dto.RespondWithAPIError(w, dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if req.Regenerate {
		regenerateAnswer(h, w, ctx, req.SessionUuid, req.ChatUuid, userID, req.Stream)
	} else {
		genAnswer(h, w, ctx, req.SessionUuid, req.ChatUuid, req.Prompt, userID, req.Stream)
	}
}

// genAnswer orchestrates the full chat completion flow.
func genAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, sessionUuid, chatUuid, question string, userID int32, streamOutput bool) {
	chatSession, _, baseURL, ok := h.validateChatSession(ctx, w, sessionUuid, userID)
	if !ok {
		return
	}
	slog.Info("Processing chat session", "sessionUUID", chatSession.Uuid, "userID", userID, "model", chatSession.Model)

	if !h.claimOrReplayChatRequest(ctx, w, *chatSession, chatUuid, userID, streamOutput) {
		return
	}

	if !h.handlePromptCreation(ctx, w, chatSession, chatUuid, question, userID, baseURL) {
		h.failChatRequest(chatSession.Uuid, chatUuid, userID, "prompt_persistence_failed")
		return
	}

	h.generateAndSaveAnswer(ctx, w, chatSession, chatUuid, userID, baseURL, streamOutput)
}

// genBotAnswer generates a bot answer from a snapshot conversation.
func genBotAnswer(ctx context.Context, h *ChatHandler, w http.ResponseWriter, session svc.ChatSession, messages []dto.SimpleChatMessage, snapshotUuid, question string, userID int32, streamOutput bool) {
	if _, err := h.modelSvc.ByName(ctx, session.Model); err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("Chat model: "+session.Model).WithDebugInfo(err.Error()))
		return
	}

	msgs := simpleChatMessagesToMessages(messages)
	msgs = append(msgs, models.Message{Role: "user", Content: question})

	model := h.chooseChatModel(ctx, session, msgs)
	providerRequest, err := h.service.ProviderRequest(ctx, session, msgs, "", false, streamOutput)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to prepare model request"))
		return
	}
	LLMAnswer, err := streamFromModel(model, ctx, w, providerRequest)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to generate answer"))
		return
	}

	if err := h.botHistorySvc.Save(ctx, svc.CreateBotAnswerHistoryInput{
		BotUuid:    snapshotUuid,
		UserID:     userID,
		Prompt:     question,
		Answer:     LLMAnswer.Answer,
		Model:      session.Model,
		TokensUsed: int32(len(LLMAnswer.Answer)) / 4,
	}); err != nil {
		slog.Info("Failed to save bot answer history", "error", err)
	}

	if !isTest(msgs) {
		h.service.LogChat(session, msgs, LLMAnswer.Answer)
	}
}

// regenerateAnswer regenerates the last assistant response.
func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, sessionUuid, chatUuid string, userID int32, stream bool) {
	chatSession, _, _, ok := h.validateChatSession(ctx, w, sessionUuid, userID)
	if !ok {
		return
	}

	msgs, err := h.service.GetAskMessages(*chatSession, chatUuid, true)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to get chat messages").WithDebugInfo(err.Error()))
		return
	}

	model := h.chooseChatModel(ctx, *chatSession, msgs)
	providerRequest, err := h.service.ProviderRequest(ctx, *chatSession, msgs, chatUuid, true, stream)
	if err != nil {
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to prepare model request"))
		return
	}
	LLMAnswer, err := streamFromModel(model, ctx, w, providerRequest)
	if err != nil {
		slog.Error("error regenerating answer", "error", err)
		if stream {
			_ = provider.FlushStreamEvent(w, "failed", provider.StreamEvent{
				Type: "failed", AnswerID: chatUuid, Code: "generation_failed", Message: "Failed to regenerate answer",
			})
			return
		}
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to regenerate answer"))
		return
	}

	h.service.LogChat(*chatSession, msgs, LLMAnswer.Answer)

	if err := h.service.UpdateChatMessageContent(ctx, chatUuid, userID, LLMAnswer.Answer); err != nil {
		if stream {
			_ = provider.FlushStreamEvent(w, "failed", provider.StreamEvent{
				Type: "failed", AnswerID: chatUuid, Code: "persistence_failed", Message: "Failed to save regenerated answer",
			})
			return
		}
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithDetail("Failed to update message").WithDebugInfo(err.Error()))
		return
	}

	if chatSession.ExploreMode {
		suggested := h.service.GenerateSuggestedQuestions(LLMAnswer.Answer, msgs)
		if len(suggested) > 0 {
			if questionsJSON, err := json.Marshal(suggested); err == nil {
				h.service.UpdateChatMessageSuggestions(ctx, chatUuid, userID, questionsJSON)
				if stream {
					h.sendSuggestedQuestionsStream(w, LLMAnswer.AnswerId, questionsJSON)
				}
			}
		}
	}
	if stream {
		_ = provider.FlushStreamEvent(w, "completed", provider.StreamEvent{
			Type: "completed", AnswerID: chatUuid, Persisted: true,
		})
	}
}

// simpleChatMessagesToMessages converts SimpleChatMessage to LLM Message format.
func simpleChatMessagesToMessages(simpleChatMessages []dto.SimpleChatMessage) []models.Message {
	messages := make([]models.Message, len(simpleChatMessages))
	for i, scm := range simpleChatMessages {
		role := scm.GetRole()
		if i == 0 {
			role = "system"
		}
		messages[i] = models.Message{Role: role, Content: scm.Text}
	}
	return messages
}

// Ensure fmt is referenced (used transitively by imported packages for debug prints).
var _ = fmt.Println

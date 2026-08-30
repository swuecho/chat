package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"log/slog"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/svc"
)

// titleGenSemaphore limits concurrent title generation goroutines to prevent unbounded resource usage.
var titleGenSemaphore = make(chan struct{}, 5)

func (h *ChatHandler) failChatRequest(sessionUuid, requestUuid string, userID int32, code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := h.service.MarkChatRequestFailed(ctx, requestUuid, sessionUuid, userID, code); err != nil {
		slog.Warn("Failed to persist chat request failure", "requestUUID", requestUuid, "error", err)
	}
}

func (h *ChatHandler) claimOrReplayChatRequest(ctx context.Context, w http.ResponseWriter, session svc.ChatSession, requestUuid string, userID int32, streamOutput bool) bool {
	if _, err := h.service.ClaimChatRequest(ctx, requestUuid, session.UUID, userID); err == nil {
		return true
	} else if !errors.Is(err, sql.ErrNoRows) {
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to claim chat request", err.Error()))
		return false
	}

	request, err := h.service.GetChatRequest(ctx, requestUuid, session.UUID, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to reconcile chat request", err.Error()))
		return false
	}

	if request.Status != "completed" || request.AssistantUuid == "" {
		if streamOutput {
			if _, err := setupSSEStream(w); err == nil {
				_ = provider.FlushStreamEvent(w, "failed", provider.StreamEvent{
					Type: "failed", Code: "request_in_progress", Message: "This request is already being processed",
				})
			}
		} else {
			dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("This request is already being processed"))
		}
		return false
	}

	message, err := h.service.GetChatMessageByUUID(ctx, request.AssistantUuid, userID)
	if err != nil {
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Completed response could not be loaded", err.Error()))
		return false
	}

	if streamOutput {
		if _, err := setupSSEStream(w); err != nil {
			return false
		}
		events := provider.NewAnswerEventWriter(w)
		_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventDelta, AnswerID: message.Uuid, Delta: message.Content})
		_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventCompleted, AnswerID: message.Uuid, Persisted: true})
		return false
	}

	_ = json.NewEncoder(w).Encode(ChatCompletionResponse{
		ID: message.Uuid, Object: "chat.completion",
		Choices: []Choice{{Message: openai.ChatCompletionMessage{Content: message.Content}}},
	})
	return false
}

// validateChatSession validates the session UUID and retrieves session + model info.
func (h *ChatHandler) validateChatSession(ctx context.Context, w http.ResponseWriter, chatSessionUuid string, userID int32) (*svc.ChatSession, *svc.RuntimeModel, string, bool) {
	chatSession, err := h.sessionSvc.GetOwnedChatSession(ctx, svc.GetOwnedChatSessionQuery{UUID: chatSessionUuid, UserID: userID})
	if err != nil {
		slog.Info("Invalid session UUID", "uuid", chatSessionUuid, "error", err)
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat session").WithMessage(chatSessionUuid))
		return nil, nil, "", false
	}
	chatModel, err := h.modelSvc.ByName(ctx, chatSession.Model)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, nil, "", false
	}

	baseURL, _ := provider.GetModelBaseURL(chatModel.URL)

	if chatSession.UUID == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid session UUID"))
		return nil, nil, "", false
	}

	return &chatSession, &chatModel, baseURL, true
}

// handlePromptCreation creates or reuses the system prompt and adds the user message.
func (h *ChatHandler) handlePromptCreation(ctx context.Context, w http.ResponseWriter, chatSession *svc.ChatSession, chatUuid, newQuestion string, userID int32, baseURL string) bool {
	existingPrompt := true
	hasPrompt, err := h.conversationSvc.HasSystemPrompt(ctx, chatSession.UUID)
	if err != nil || !hasPrompt {
		if errors.Is(err, sql.ErrNoRows) {
			existingPrompt = false
		} else {
			slog.Error("error checking prompt", "session", chatSession.UUID, "error", err)
			dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to get prompt", err.Error()))
			return false
		}
	}

	if existingPrompt {
		if newQuestion != "" {
			if _, err := h.service.CreateChatMessageSimple(ctx, chatSession.UUID, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
				dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create message", err.Error()))
				return false
			}
		}
	} else {
		if _, err := h.service.CreateChatPromptSimple(ctx, chatSession.UUID, dto.DefaultSystemPromptText, userID); err != nil {
			dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create prompt", err.Error()))
			return false
		}

		if newQuestion != "" {
			if _, err := h.service.CreateChatMessageSimple(ctx, chatSession.UUID, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
				dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create message", err.Error()))
				return false
			}

			if title := firstNWords(newQuestion, 10); title != "" {
				if _, err := h.sessionSvc.UpdateChatSessionTopicByUUID(ctx, chatSession.UUID, userID, title); err != nil {
					slog.Warn("Failed to update session title", "error", err)
				}
			}
		}
	}
	return true
}

// generateAndSaveAnswer calls the LLM, streams the response, and persists the answer.
func (h *ChatHandler) generateAndSaveAnswer(ctx context.Context, w http.ResponseWriter, chatSession *svc.ChatSession, chatUuid string, userID int32, baseURL string, streamOutput bool) bool {
	events := provider.NewAnswerEventWriter(w)
	if streamOutput {
		if _, err := setupSSEStream(w); err != nil {
			return false
		}
	}
	var onChunk func(provider.StreamChunk) error
	if streamOutput {
		onChunk = func(chunk provider.StreamChunk) error {
			return events.Emit(provider.AnswerEvent{Type: provider.AnswerEventDelta, AnswerID: chunk.ID, Delta: chunk.Content})
		}
	}
	result, err := h.service.GenerateAnswer(ctx, svc.GenerateAnswerCommand{
		Session: *chatSession, RequestUUID: chatUuid, UserID: userID, BaseURL: baseURL,
		Stream: streamOutput,
	}, svc.GenerateAnswerDependencies{
		SelectModel: func(msgs []models.Message) provider.ChatModel {
			return h.chooseChatModel(ctx, *chatSession, msgs)
		},
		OnChunk:   onChunk,
		ShouldLog: func(msgs []models.Message) bool { return !isTest(msgs) },
	})
	if err != nil {
		slog.Error("error generating answer", "error", err)
		code := "generation_failed"
		var generationErr *svc.GenerateAnswerError
		if errors.As(err, &generationErr) {
			code = generationErr.Code
		}
		if streamOutput {
			eventType := provider.AnswerEventFailed
			message := "Failed to generate answer"
			if code == "canceled" {
				eventType = provider.AnswerEventCanceled
				message = "Answer generation was canceled"
			}
			_ = events.Emit(provider.AnswerEvent{
				Type: eventType, Code: code, Message: message,
			})
			return false
		}
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to generate answer"))
		return false
	}

	if !streamOutput {
		if err := json.NewEncoder(w).Encode(ChatCompletionResponse{ID: result.Answer.AnswerId, Object: "chat.completion", Choices: []Choice{{Message: openai.ChatCompletionMessage{Content: result.Answer.Answer}}}}); err != nil {
			return false
		}
	}
	if streamOutput && chatSession.ExploreMode && result.SuggestedQuestions != nil {
		h.sendSuggestedQuestionsStream(events, result.Answer.AnswerId, result.SuggestedQuestions)
	}
	if streamOutput {
		if err := events.Emit(provider.AnswerEvent{
			Type: provider.AnswerEventCompleted, AnswerID: result.Answer.AnswerId, Persisted: true,
		}); err != nil {
			slog.Warn("Failed to send stream completion event", "error", err)
		}
	}

	// Launch best-effort title generation only when a bounded slot is available.
	select {
	case titleGenSemaphore <- struct{}{}:
		go func() {
			defer func() { <-titleGenSemaphore }()
			h.generateSessionTitle(chatSession, userID)
		}()
	default:
		slog.Info("Skipping title generation because all worker slots are busy", "sessionUUID", chatSession.UUID)
	}
	return true
}

// streamFromModel calls model.Stream() and consumes the channel, writing SSE or JSON to w.
// Returns the final answer or an error.
func streamFromModel(model provider.ChatModel, ctx context.Context, w http.ResponseWriter, input provider.Request, events *provider.AnswerEventWriter) (*models.LLMAnswer, error) {
	var onChunk func(provider.StreamChunk) error
	if input.Stream {
		if events == nil {
			return nil, errors.New("typed answer event writer is required for streaming")
		}
		if _, err := setupSSEStream(w); err != nil {
			return nil, err
		}
		onChunk = func(chunk provider.StreamChunk) error {
			return events.Emit(provider.AnswerEvent{Type: provider.AnswerEventDelta, AnswerID: chunk.ID, Delta: chunk.Content})
		}
	}
	answer, err := provider.ConsumeStream(ctx, model, input, onChunk)
	if err != nil {
		return nil, err
	}
	if !input.Stream {
		if err := json.NewEncoder(w).Encode(ChatCompletionResponse{ID: answer.AnswerId, Object: "chat.completion", Choices: []Choice{{Message: openai.ChatCompletionMessage{Content: answer.Answer}}}}); err != nil {
			return nil, err
		}
	}
	return answer, nil
}

// generateSessionTitle asynchronously updates the session topic using an LLM.
func (h *ChatHandler) generateSessionTitle(chatSession *svc.ChatSession, userID int32) {
	ctx, cancel := context.WithTimeout(context.Background(), sessionTitleGenerationTimeout)
	defer cancel()

	messages, err := h.conversationSvc.MessagesPage(ctx, svc.ConversationMessagesPageQuery{SessionUUID: chatSession.UUID, Page: svc.PageWindow{Limit: 100}})
	if err != nil {
		slog.Warn("Failed to get messages for title generation", "error", err)
		return
	}

	var chatText strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&chatText, "%s: %s\n", msg.Role, msg.Content)
	}

	if strings.TrimSpace(chatText.String()) == "" {
		return
	}

	genTitle, err := h.modelSvc.GenerateTitle(ctx, chatText.String())
	if err != nil || genTitle == "" {
		return
	}

	if _, err := h.sessionSvc.UpdateChatSessionTopicByUUID(ctx, chatSession.UUID, userID, genTitle); err != nil {
		slog.Warn("Failed to update session title", "error", err)
		return
	}

	slog.Info("Generated LLM title", "sessionUUID", chatSession.UUID, "title", genTitle)
}

// sendSuggestedQuestionsStream sends suggested questions as an SSE event.
func (h *ChatHandler) sendSuggestedQuestionsStream(events *provider.AnswerEventWriter, answerID string, suggestedQuestionsJSON json.RawMessage) {
	var suggestedQuestions []string
	if err := json.Unmarshal(suggestedQuestionsJSON, &suggestedQuestions); err != nil || len(suggestedQuestions) == 0 {
		return
	}

	_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventSuggested, AnswerID: answerID, SuggestedQuestions: suggestedQuestions})
}

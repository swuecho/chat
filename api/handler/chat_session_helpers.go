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
		flusher, err := setupSSEStream(w)
		if err != nil {
			return false
		}
		_ = provider.FlushResponse(w, flusher, provider.StreamingResponse{
			AnswerID: message.Uuid, Content: message.Content,
		})
		_ = provider.FlushStreamEvent(w, "completed", provider.StreamEvent{
			Type: "completed", AnswerID: message.Uuid, Persisted: true,
		})
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
	chatSession, err := h.sessionSvc.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		slog.Info("Invalid session UUID", "uuid", chatSessionUuid, "error", err)
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat session").WithMessage(chatSessionUuid))
		return nil, nil, "", false
	}
	if chatSession.UserID != userID {
		dto.RespondWithAPIError(w, dto.ErrAuthAccessDenied.WithMessage("You do not own this session"))
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
	msgs, err := h.service.GetAskMessages(*chatSession, chatUuid, false)
	if err != nil {
		h.failChatRequest(chatSession.UUID, chatUuid, userID, "message_collection_failed")
		slog.Error("error collecting messages", "session", chatSession.UUID, "error", err)
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to collect messages", err.Error()))
		return false
	}
	slog.Info("Collected messages", "sessionUUID", chatSession.UUID, "count", len(msgs), "model", chatSession.Model)

	if err := h.service.MarkChatRequestStreaming(ctx, chatUuid, chatSession.UUID, userID); err != nil {
		h.failChatRequest(chatSession.UUID, chatUuid, userID, "request_state_failed")
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to start chat request", err.Error()))
		return false
	}

	model := h.chooseChatModel(ctx, *chatSession, msgs)
	providerRequest, err := h.service.ProviderRequest(ctx, *chatSession, msgs, chatUuid, false, streamOutput)
	if err != nil {
		h.failChatRequest(chatSession.UUID, chatUuid, userID, "provider_request_failed")
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to prepare model request"))
		return false
	}
	LLMAnswer, err := streamFromModel(model, ctx, w, providerRequest)
	if err != nil {
		h.failChatRequest(chatSession.UUID, chatUuid, userID, "generation_failed")
		slog.Error("error generating answer", "error", err)
		if streamOutput {
			_ = provider.FlushStreamEvent(w, "failed", provider.StreamEvent{
				Type: "failed", Code: "generation_failed", Message: "Failed to generate answer",
			})
			return false
		}
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to generate answer"))
		return false
	}
	if LLMAnswer == nil {
		h.failChatRequest(chatSession.UUID, chatUuid, userID, "empty_answer")
		if streamOutput {
			_ = provider.FlushStreamEvent(w, "failed", provider.StreamEvent{
				Type: "failed", Code: "empty_answer", Message: "The model returned no final answer",
			})
			return false
		}
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("LLMAnswer is nil"))
		return false
	}

	if !isTest(msgs) {
		h.service.LogChat(*chatSession, msgs, LLMAnswer.ReasoningContent+LLMAnswer.Answer)
	}

	chatMessage, err := h.service.CompleteChatRequestWithSuggestedQuestions(ctx, chatUuid, chatSession.UUID, LLMAnswer.AnswerId, LLMAnswer.Answer, LLMAnswer.ReasoningContent, chatSession.Model, userID, baseURL, chatSession.SummarizeMode, chatSession.ExploreMode, msgs)
	if err != nil {
		h.failChatRequest(chatSession.UUID, chatUuid, userID, "persistence_failed")
		if streamOutput {
			_ = provider.FlushStreamEvent(w, "failed", provider.StreamEvent{
				Type: "failed", AnswerID: LLMAnswer.AnswerId, Code: "persistence_failed", Message: "Failed to save answer",
			})
			return false
		}
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create message", err.Error()))
		return false
	}

	if streamOutput && chatSession.ExploreMode && chatMessage.SuggestedQuestions != nil {
		h.sendSuggestedQuestionsStream(w, LLMAnswer.AnswerId, chatMessage.SuggestedQuestions)
	}
	if streamOutput {
		if err := provider.FlushStreamEvent(w, "completed", provider.StreamEvent{
			Type: "completed", AnswerID: LLMAnswer.AnswerId, Persisted: true,
		}); err != nil {
			slog.Warn("Failed to send stream completion event", "error", err)
		}
	}

	// Launch title generation with bounded concurrency
	go func() {
		titleGenSemaphore <- struct{}{}
		defer func() { <-titleGenSemaphore }()
		h.generateSessionTitle(chatSession, userID)
	}()
	return true
}

// streamFromModel calls model.Stream() and consumes the channel, writing SSE or JSON to w.
// Returns the final answer or an error.
func streamFromModel(model provider.ChatModel, ctx context.Context, w http.ResponseWriter, input provider.Request) (*models.LLMAnswer, error) {
	ch, err := model.Stream(ctx, input)
	if err != nil {
		return nil, err
	}

	var lastAnswer *models.LLMAnswer

	if input.Stream {
		flusher, err := setupSSEStream(w)
		if err != nil {
			return nil, err
		}
		for chunk := range ch {
			if chunk.Err != nil {
				return nil, chunk.Err
			}
			if chunk.Done {
				lastAnswer = chunk.FinalAnswer
				break
			}
			if chunk.Content != "" {
				provider.FlushResponse(w, flusher, provider.StreamingResponse{
					AnswerID: chunk.ID,
					Content:  chunk.Content,
					IsFinal:  false,
				})
			}
		}
	} else {
		for chunk := range ch {
			if chunk.Err != nil {
				return nil, chunk.Err
			}
			if chunk.Done {
				lastAnswer = chunk.FinalAnswer
				break
			}
		}
		// Write non-streaming JSON response
		if lastAnswer != nil {
			json.NewEncoder(w).Encode(ChatCompletionResponse{
				ID:     lastAnswer.AnswerId,
				Object: "chat.completion",
				Choices: []Choice{{
					Message: openai.ChatCompletionMessage{Content: lastAnswer.Answer},
				}},
			})
		}
	}

	return lastAnswer, nil
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
func (h *ChatHandler) sendSuggestedQuestionsStream(w http.ResponseWriter, answerID string, suggestedQuestionsJSON json.RawMessage) {
	var suggestedQuestions []string
	if err := json.Unmarshal(suggestedQuestionsJSON, &suggestedQuestions); err != nil || len(suggestedQuestions) == 0 {
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	response := suggestedQuestionsChunk{ID: answerID, Object: "chat.completion.chunk",
		Choices: []suggestedQuestionsChoice{{Index: 0, Delta: suggestedQuestionsDelta{SuggestedQuestions: suggestedQuestions}}}}

	data, _ := json.Marshal(response)
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()
}

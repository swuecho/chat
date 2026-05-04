package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"log/slog"
	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// validateChatSession validates the session UUID and retrieves session + model info.
func (h *ChatHandler) validateChatSession(ctx context.Context, w http.ResponseWriter, chatSessionUuid string) (*sqlc_queries.ChatSession, *sqlc_queries.ChatModel, string, bool) {
	chatSession, err := h.sessionSvc.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		slog.Info("Invalid session UUID: %s, error: %v", chatSessionUuid, err)
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat session").WithMessage(chatSessionUuid))
		return nil, nil, "", false
	}

	chatModel, err := h.sessionSvc.ChatModelByName(ctx, chatSession.Model)
	if err != nil {
		dto.RespondWithAPIError(w, dto.ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, nil, "", false
	}

	baseURL, _ := provider.GetModelBaseURL(chatModel.Url)

	if chatSession.Uuid == "" {
		dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("Invalid session UUID"))
		return nil, nil, "", false
	}

	return &chatSession, &chatModel, baseURL, true
}

// handlePromptCreation creates or reuses the system prompt and adds the user message.
func (h *ChatHandler) handlePromptCreation(ctx context.Context, w http.ResponseWriter, chatSession *sqlc_queries.ChatSession, chatUuid, newQuestion string, userID int32, baseURL string) bool {
	existingPrompt := true
	_, err := h.sessionSvc.GetOneChatPromptBySessionUUID(ctx, chatSession.Uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existingPrompt = false
		} else {
			slog.Error("error: checking prompt for session %s: %v", chatSession.Uuid, err)
			dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to get prompt", err.Error()))
			return false
		}
	}

	if existingPrompt {
		if newQuestion != "" {
			if _, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
				dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create message", err.Error()))
				return false
			}
		}
	} else {
		if _, err := h.service.CreateChatPromptSimple(ctx, chatSession.Uuid, dto.DefaultSystemPromptText, userID); err != nil {
			dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create prompt", err.Error()))
			return false
		}

		if newQuestion != "" {
			if _, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
				dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create message", err.Error()))
				return false
			}

			if title := firstNWords(newQuestion, 10); title != "" {
				params := sqlc_queries.UpdateChatSessionTopicByUUIDParams{
					Uuid: chatSession.Uuid, UserID: userID, Topic: title,
				}
				if _, err := h.sessionSvc.UpdateChatSessionTopicByUUID(ctx, params); err != nil {
					slog.Warn("Failed to update session title: %v", err)
				}
			}
		}
	}
	return true
}

// generateAndSaveAnswer calls the LLM, streams the response, and persists the answer.
func (h *ChatHandler) generateAndSaveAnswer(ctx context.Context, w http.ResponseWriter, chatSession *sqlc_queries.ChatSession, chatUuid string, userID int32, baseURL string, streamOutput bool) bool {
	msgs, err := h.service.GetAskMessages(*chatSession, chatUuid, false)
	if err != nil {
		slog.Error("error: collecting messages for session %s: %v", chatSession.Uuid, err)
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to collect messages", err.Error()))
		return false
	}
	slog.Info("Collected messages - SessionUUID: %s, Count: %d, Model: %s", chatSession.Uuid, len(msgs), chatSession.Model)

	model := h.chooseChatModel(ctx, *chatSession, msgs)
	LLMAnswer, err := streamFromModel(model, ctx, w, *chatSession, msgs, chatUuid, false, streamOutput)
	if err != nil {
		slog.Error("error: generating answer: %v", err)
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to generate answer"))
		return false
	}
	if LLMAnswer == nil {
		dto.RespondWithAPIError(w, dto.ErrInternalUnexpected.WithMessage("LLMAnswer is nil"))
		return false
	}

	if !isTest(msgs) {
		h.service.LogChat(*chatSession, msgs, LLMAnswer.ReasoningContent+LLMAnswer.Answer)
	}

	chatMessage, err := h.service.CreateChatMessageWithSuggestedQuestions(ctx, chatSession.Uuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, chatSession.Model, userID, baseURL, chatSession.SummarizeMode, chatSession.ExploreMode, msgs)
	if err != nil {
		dto.RespondWithAPIError(w, dto.CreateAPIError(dto.ErrInternalUnexpected, "Failed to create message", err.Error()))
		return false
	}

	if streamOutput && chatSession.ExploreMode && chatMessage.SuggestedQuestions != nil {
		h.sendSuggestedQuestionsStream(w, LLMAnswer.AnswerId, chatMessage.SuggestedQuestions)
	}

	go h.generateSessionTitle(chatSession, userID)
	return true
}

// streamFromModel calls model.Stream() and consumes the channel, writing SSE or JSON to w.
// Returns the final answer or an error.
func streamFromModel(model provider.ChatModel, ctx context.Context, w http.ResponseWriter, session sqlc_queries.ChatSession, msgs []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	ch, err := model.Stream(ctx, session, msgs, chatUuid, regenerate, streamOutput)
	if err != nil {
		return nil, err
	}

	var lastAnswer *models.LLMAnswer

	if streamOutput {
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
func (h *ChatHandler) generateSessionTitle(chatSession *sqlc_queries.ChatSession, userID int32) {
	ctx, cancel := context.WithTimeout(context.Background(), sessionTitleGenerationTimeout)
	defer cancel()

	messages, err := h.sessionSvc.GetChatMessagesBySessionUUID(ctx, sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid: chatSession.Uuid, Offset: 0, Limit: 100,
	})
	if err != nil {
		slog.Warn("Failed to get messages for title generation: %v", err)
		return
	}

	var chatText strings.Builder
	for _, msg := range messages {
		fmt.Fprintf(&chatText, "%s: %s\n", msg.Role, msg.Content)
	}

	if strings.TrimSpace(chatText.String()) == "" {
		return
	}

	model := "gemini-2.0-flash"
	if _, err := h.sessionSvc.ChatModelByName(ctx, model); err != nil {
		return
	}

	genTitle, err := provider.GenerateChatTitle(ctx, model, chatText.String())
	if err != nil || genTitle == "" {
		return
	}

	if _, err := h.sessionSvc.UpdateChatSessionTopicByUUID(ctx, sqlc_queries.UpdateChatSessionTopicByUUIDParams{
		Uuid: chatSession.Uuid, UserID: userID, Topic: genTitle,
	}); err != nil {
		slog.Warn("Failed to update session title: %v", err)
		return
	}

	slog.Info("Generated LLM title for session %s: %s", chatSession.Uuid, genTitle)
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

	response := map[string]interface{}{
		"id":     answerID,
		"object": "chat.completion.chunk",
		"choices": []map[string]interface{}{{
			"index": 0,
			"delta": map[string]interface{}{
				"content":            "",
				"suggestedQuestions": suggestedQuestions,
			},
			"finish_reason": nil,
		}},
	}

	data, _ := json.Marshal(response)
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()
}

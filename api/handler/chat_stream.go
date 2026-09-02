package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	openai "github.com/sashabaranov/go-openai"
	"log/slog"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
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

func (r *BotRequest) Validate() error {
	return validation.UUID("snapshot_uuid", r.SnapshotUuid, true)
}

func (r *ChatRequest) Validate() error {
	if err := validation.UUID("sessionUuid", r.SessionUuid, true); err != nil {
		return err
	}
	return validation.UUID("chatUuid", r.ChatUuid, true)
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
func (h *ChatHandler) ChatBotCompletionHandler(w http.ResponseWriter, r *http.Request) error {
	var req BotRequest
	if err := DecodeJSON(r, &req); err != nil {
		return dto.ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error())
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		slog.Error("error getting user ID", "error", err)
		return dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error())
	}

	chatSnapshot, err := h.snapshotSvc.ByUserAndUUID(ctx, userID, req.SnapshotUuid)
	if err != nil {
		return dto.ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error())
	}

	var session svc.ChatSession
	if err := json.Unmarshal(chatSnapshot.Session, &session); err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to deserialize chat session").WithDebugInfo(err.Error())
	}

	var messages []dto.SimpleChatMessage
	if err := json.Unmarshal(chatSnapshot.Conversation, &messages); err != nil {
		return dto.ErrInternalUnexpected.WithDetail("Failed to deserialize conversation").WithDebugInfo(err.Error())
	}

	genBotAnswer(ctx, h, w, session, messages, req.SnapshotUuid, req.Message, userID, req.Stream)
	return nil
}

// ChatCompletionHandler handles regular chat completion with streaming support.
func (h *ChatHandler) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) error {
	var req ChatRequest
	if err := DecodeJSON(r, &req); err != nil {
		slog.Error("error decoding request", "error", err)
		return dto.ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error())
	}
	if req.SessionUuid == "" || req.ChatUuid == "" {
		return dto.ErrValidationInvalidInput("sessionUuid and chatUuid are required")
	}
	if !req.Regenerate && req.Prompt == "" {
		return dto.ErrValidationInvalidInput("prompt is required")
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		slog.Error("error getting user ID", "error", err)
		return dto.ErrAuthInvalidCredentials.WithDebugInfo(err.Error())
	}

	if req.Regenerate {
		regenerateAnswer(h, w, ctx, req.SessionUuid, req.ChatUuid, userID, req.Stream)
	} else {
		genAnswer(h, w, ctx, req.SessionUuid, req.ChatUuid, req.Prompt, userID, req.Stream)
	}
	return nil
}

// genAnswer orchestrates the full chat completion flow.
func genAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, sessionUuid, chatUuid, question string, userID int32, streamOutput bool) {
	var events *provider.AnswerEventWriter
	var chunks svc.AnswerChunkSink
	if streamOutput {
		if _, err := setupSSEStream(w); err != nil {
			return
		}
		events = provider.NewAnswerEventWriter(w)
		chunks = answerEventChunkSink{events: events}
	}
	useCase := svc.NewCompleteChatUseCase(h.service, h.sessionSvc, h.conversationSvc, h, chunks, h)
	result, err := useCase.Execute(ctx, svc.CompleteChatCommand{SessionUUID: sessionUuid, RequestUUID: chatUuid, Prompt: question, UserID: userID, Stream: streamOutput})
	if err != nil {
		slog.Error("error completing chat", "error", err)
		if streamOutput {
			_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventFailed, Code: generationErrorCode(err), Message: "Failed to generate answer"})
			return
		}
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to generate answer"))
		return
	}
	if result.InProgress {
		if streamOutput {
			_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventFailed, Code: "request_in_progress", Message: "This request is already being processed"})
		} else {
			dto.RespondWithAPIError(w, dto.ErrValidationInvalidInput("This request is already being processed"))
		}
		return
	}
	answerID, content := result.Answer.Answer.AnswerId, result.Answer.Answer.Answer
	if result.Replay != nil {
		answerID, content = result.Replay.UUID, result.Replay.Content
		if streamOutput {
			_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventDelta, AnswerID: answerID, Delta: content})
		}
	}
	if !streamOutput {
		_ = json.NewEncoder(w).Encode(ChatCompletionResponse{ID: answerID, Object: "chat.completion", Choices: []Choice{{Message: openai.ChatCompletionMessage{Content: content}}}})
		return
	}
	if len(result.Answer.SuggestedQuestions) > 0 {
		h.sendSuggestedQuestionsStream(events, answerID, result.Answer.SuggestedQuestions)
	}
	_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventCompleted, AnswerID: answerID, Persisted: true})
	if result.Replay == nil {
		h.launchSessionTitleGeneration(&result.Session, userID)
	}
}

func generationErrorCode(err error) string {
	var generationErr *svc.GenerateAnswerError
	if errors.As(err, &generationErr) {
		return generationErr.Code
	}
	return "generation_failed"
}

// genBotAnswer generates a bot answer from a snapshot conversation.
func genBotAnswer(ctx context.Context, h *ChatHandler, w http.ResponseWriter, session svc.ChatSession, messages []dto.SimpleChatMessage, snapshotUuid, question string, userID int32, streamOutput bool) {
	msgs := simpleChatMessagesToMessages(messages)
	var events *provider.AnswerEventWriter
	var chunks svc.AnswerChunkSink
	if streamOutput {
		if _, err := setupSSEStream(w); err != nil {
			return
		}
		events = provider.NewAnswerEventWriter(w)
		chunks = answerEventChunkSink{events: events}
	}
	useCase := svc.NewGenerateBotAnswerUseCase(h.service, h, chunks, h, h.botHistorySvc)
	answer, err := useCase.Execute(ctx, svc.GenerateBotAnswerCommand{Session: session, Messages: msgs,
		SnapshotUUID: snapshotUuid, Question: question, UserID: userID, Stream: streamOutput})
	if err != nil {
		if streamOutput {
			eventType := provider.AnswerEventFailed
			if errors.Is(err, context.Canceled) {
				eventType = provider.AnswerEventCanceled
			}
			_ = events.Emit(provider.AnswerEvent{Type: eventType, Code: "generation_failed", Message: "Failed to generate bot answer"})
			return
		}
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to generate answer"))
		return
	}

	if streamOutput {
		_ = events.Emit(provider.AnswerEvent{Type: provider.AnswerEventCompleted, AnswerID: answer.AnswerId, Persisted: true})
	} else {
		_ = json.NewEncoder(w).Encode(ChatCompletionResponse{ID: answer.AnswerId, Object: "chat.completion",
			Choices: []Choice{{Message: openai.ChatCompletionMessage{Content: answer.Answer}}}})
	}
}

// regenerateAnswer regenerates the last assistant response.
func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, sessionUuid, chatUuid string, userID int32, stream bool) {
	var events *provider.AnswerEventWriter
	var chunks svc.AnswerChunkSink
	if stream {
		if _, err := setupSSEStream(w); err != nil {
			return
		}
		events = provider.NewAnswerEventWriter(w)
		chunks = answerEventChunkSink{events: events}
	}
	useCase := svc.NewRegenerateAnswerUseCase(h.service, h.sessionSvc, h, chunks, h)
	result, err := useCase.Execute(ctx, svc.RegenerateAnswerCommand{SessionUUID: sessionUuid, MessageUUID: chatUuid, UserID: userID, Stream: stream})
	if err != nil {
		slog.Error("error regenerating answer", "error", err)
		if stream {
			eventType := provider.AnswerEventFailed
			code := "generation_failed"
			message := "Failed to regenerate answer"
			if errors.Is(err, context.Canceled) {
				eventType = provider.AnswerEventCanceled
				code = "canceled"
				message = "Answer regeneration was canceled"
			}
			_ = events.Emit(provider.AnswerEvent{
				Type: eventType, AnswerID: chatUuid, Code: code, Message: message,
			})
			return
		}
		dto.RespondWithAPIError(w, dto.WrapError(err, "Failed to regenerate answer"))
		return
	}

	if !stream {
		_ = json.NewEncoder(w).Encode(ChatCompletionResponse{ID: result.Answer.AnswerId, Object: "chat.completion",
			Choices: []Choice{{Message: openai.ChatCompletionMessage{Content: result.Answer.Answer}}}})
	}
	if stream && len(result.SuggestedQuestions) > 0 {
		h.sendSuggestedQuestionsStream(events, result.Answer.AnswerId, result.SuggestedQuestions)
	}
	if stream {
		_ = events.Emit(provider.AnswerEvent{
			Type: provider.AnswerEventCompleted, AnswerID: chatUuid, Persisted: true,
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

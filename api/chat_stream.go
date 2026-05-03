// Package main — Chat completion and streaming handlers.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"

	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
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
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Model   string   `json:"model"`
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

// GetChatInstructions returns artifact instruction text.
func (h *ChatHandler) GetChatInstructions(w http.ResponseWriter, r *http.Request) {
	artifactInstruction, err := loadArtifactInstruction()
	if err != nil {
		log.Printf("Warning: Failed to load artifact instruction: %v", err)
		artifactInstruction = ""
	}
	json.NewEncoder(w).Encode(ChatInstructionResponse{
		ArtifactInstruction: artifactInstruction,
	})
}

// ChatBotCompletionHandler handles bot chat completion via snapshot.
func (h *ChatHandler) ChatBotCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req BotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	chatSnapshot, err := h.sessionSvc.ChatSnapshotByUserIdAndUuid(ctx, sqlc_queries.ChatSnapshotByUserIdAndUuidParams{
		UserID: userID, Uuid: req.SnapshotUuid,
	})
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()))
		return
	}

	var session sqlc_queries.ChatSession
	if err := json.Unmarshal(chatSnapshot.Session, &session); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to deserialize chat session").WithDebugInfo(err.Error()))
		return
	}

	var messages []SimpleChatMessage
	if err := json.Unmarshal(chatSnapshot.Conversation, &messages); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to deserialize conversation").WithDebugInfo(err.Error()))
		return
	}

	genBotAnswer(h, w, session, messages, req.SnapshotUuid, req.Message, userID, req.Stream)
}

// ChatCompletionHandler handles regular chat completion with streaming support.
func (h *ChatHandler) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()))
		return
	}

	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		RespondWithAPIError(w, ErrAuthInvalidCredentials.WithDebugInfo(err.Error()))
		return
	}

	if req.Regenerate {
		regenerateAnswer(h, w, ctx, req.SessionUuid, req.ChatUuid, req.Stream)
	} else {
		genAnswer(h, w, ctx, req.SessionUuid, req.ChatUuid, req.Prompt, userID, req.Stream)
	}
}

// genAnswer orchestrates the full chat completion flow.
func genAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, sessionUuid, chatUuid, question string, userID int32, streamOutput bool) {
	chatSession, _, baseURL, ok := h.validateChatSession(ctx, w, sessionUuid)
	if !ok {
		return
	}
	log.Printf("Processing chat session - SessionUUID: %s, UserID: %d, Model: %s", chatSession.Uuid, userID, chatSession.Model)

	if !h.handlePromptCreation(ctx, w, chatSession, chatUuid, question, userID, baseURL) {
		return
	}

	h.generateAndSaveAnswer(ctx, w, chatSession, chatUuid, userID, baseURL, streamOutput)
}

// genBotAnswer generates a bot answer from a snapshot conversation.
func genBotAnswer(h *ChatHandler, w http.ResponseWriter, session sqlc_queries.ChatSession, messages []SimpleChatMessage, snapshotUuid, question string, userID int32, streamOutput bool) {
	ctx := h.GetRequestContext()
	if _, err := h.sessionSvc.ChatModelByName(ctx, session.Model); err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("Chat model: "+session.Model).WithDebugInfo(err.Error()))
		return
	}

	msgs := simpleChatMessagesToMessages(messages)
	msgs = append(msgs, models.Message{Role: "user", Content: question})

	model := h.chooseChatModel(session, msgs)
	LLMAnswer, err := model.Stream(w, session, msgs, "", false, streamOutput)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to generate answer"))
		return
	}

	if _, err := h.sessionSvc.CreateBotAnswerHistory(ctx, sqlc_queries.CreateBotAnswerHistoryParams{
		BotUuid:    snapshotUuid,
		UserID:     userID,
		Prompt:     question,
		Answer:     LLMAnswer.Answer,
		Model:      session.Model,
		TokensUsed: int32(len(LLMAnswer.Answer)) / 4,
	}); err != nil {
		log.Printf("Failed to save bot answer history: %v", err)
	}

	if !isTest(msgs) {
		h.service.logChat(session, msgs, LLMAnswer.Answer)
	}
}

// regenerateAnswer regenerates the last assistant response.
func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, sessionUuid, chatUuid string, stream bool) {
	chatSession, _, _, ok := h.validateChatSession(ctx, w, sessionUuid)
	if !ok {
		return
	}

	msgs, err := h.service.getAskMessages(*chatSession, chatUuid, true)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to get chat messages"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	h.requestCtx = ctx
	model := h.chooseChatModel(*chatSession, msgs)
	LLMAnswer, err := model.Stream(w, *chatSession, msgs, chatUuid, true, stream)
	if err != nil {
		log.Printf("Error regenerating answer: %v", err)
		return
	}

	h.service.logChat(*chatSession, msgs, LLMAnswer.Answer)

	if err := h.service.UpdateChatMessageContent(ctx, chatUuid, LLMAnswer.Answer); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to update message").WithDebugInfo(err.Error()))
		return
	}

	if chatSession.ExploreMode {
		suggested := h.service.generateSuggestedQuestions(LLMAnswer.Answer, msgs)
		if len(suggested) > 0 {
			if questionsJSON, err := json.Marshal(suggested); err == nil {
				h.service.UpdateChatMessageSuggestions(ctx, chatUuid, questionsJSON)
				if stream {
					h.sendSuggestedQuestionsStream(w, LLMAnswer.AnswerId, questionsJSON)
				}
			}
		}
	}
}

// simpleChatMessagesToMessages converts SimpleChatMessage to LLM Message format.
func simpleChatMessagesToMessages(simpleChatMessages []SimpleChatMessage) []models.Message {
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

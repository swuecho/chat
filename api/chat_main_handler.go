package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	openai "github.com/sashabaranov/go-openai"

	"github.com/gorilla/mux"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatHandler struct {
	service         *ChatService
	chatfileService *ChatFileService
}

func NewChatHandler(sqlc_q *sqlc_queries.Queries) *ChatHandler {
	// create a new ChatService instance
	chatService := NewChatService(sqlc_q)
	ChatFileService := NewChatFileService(sqlc_q)
	return &ChatHandler{
		service:         chatService,
		chatfileService: ChatFileService,
	}
}

func (h *ChatHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat_stream", h.ChatCompletionHandler).Methods(http.MethodPost)
	// for bot
	// given a chat_uuid, a user message, return the answer
	//
	router.HandleFunc("/chatbot", h.ChatBotCompletionHandler).Methods(http.MethodPost)
}

type ChatRequest struct {
	Prompt      string
	SessionUuid string
	ChatUuid    string
	Regenerate  bool
	Stream      bool `json:"stream,omitempty"`
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

type OpenaiChatRequest struct {
	Model    string                         `json:"model"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

type BotRequest struct {
	Message      string `json:"message"`
	SnapshotUuid string `json:"snapshot_uuid"`
	Stream       bool   `json:"stream"`
}

// ChatCompletionHandler is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
func (h *ChatHandler) ChatBotCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req BotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithAPIError(w, ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()))
		return
	}

	snapshotUuid := req.SnapshotUuid
	newQuestion := req.Message

	log.Printf("snapshotUuid: %s", snapshotUuid)
	log.Printf("newQuestion: %s", newQuestion)

	ctx := r.Context()

	userID, err := getUserID(ctx)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	fmt.Printf("userID: %d", userID)

	chatSnapshot, err := h.service.q.ChatSnapshotByUserIdAndUuid(ctx, sqlc_queries.ChatSnapshotByUserIdAndUuidParams{
		UserID: userID,
		Uuid:   snapshotUuid,
	})
	if err != nil {
		apiErr := ErrResourceNotFound("Chat snapshot")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	fmt.Printf("chatSnapshot: %+v", chatSnapshot)

	var session sqlc_queries.ChatSession
	err = json.Unmarshal(chatSnapshot.Session, &session)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to deserialize chat session"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
	var simpleChatMessages []SimpleChatMessage
	err = json.Unmarshal(chatSnapshot.Conversation, &simpleChatMessages)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to deserialize conversation"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	genBotAnswer(h, w, session, simpleChatMessages, snapshotUuid, newQuestion, userID, req.Stream)

}

// ChatCompletionHandler is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
func (h *ChatHandler) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		apiErr := ErrValidationInvalidInput("Invalid request format")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	chatSessionUuid := req.SessionUuid
	chatUuid := req.ChatUuid
	newQuestion := req.Prompt

	log.Printf("chatSessionUuid: %s", chatSessionUuid)
	log.Printf("chatUuid: %s", chatUuid)
	log.Printf("newQuestion: %s", newQuestion)

	ctx := r.Context()

	userID, err := getUserID(ctx)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		apiErr := ErrAuthInvalidCredentials
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	if req.Regenerate {
		regenerateAnswer(h, w, chatSessionUuid, chatUuid, req.Stream)
	} else {
		genAnswer(h, w, chatSessionUuid, chatUuid, newQuestion, userID, req.Stream)
	}

}

// validateChatSession validates the chat session and returns the session and model info.
// It performs comprehensive validation including:
// - Session existence check
// - Model availability verification  
// - Base URL extraction
// - UUID validation
// Returns: session, model, baseURL, success
func (h *ChatHandler) validateChatSession(ctx context.Context, w http.ResponseWriter, chatSessionUuid string) (*sqlc_queries.ChatSession, *sqlc_queries.ChatModel, string, bool) {
	chatSession, err := h.service.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		log.Printf("Invalid session UUID: %s, error: %v", chatSessionUuid, err)
		RespondWithAPIError(w, ErrResourceNotFound("chat session").WithMessage(chatSessionUuid))
		return nil, nil, "", false
	}

	chatModel, err := h.service.q.ChatModelByName(ctx, chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, nil, "", false
	}

	baseURL, _ := getModelBaseUrl(chatModel.Url)

	if chatSession.Uuid == "" {
		log.Printf("Empty session UUID for chat: %s", chatSessionUuid)
		RespondWithAPIError(w, ErrValidationInvalidInput("Invalid session UUID"))
		return nil, nil, "", false
	}

	return &chatSession, &chatModel, baseURL, true
}

// handlePromptCreation handles creating new prompt or adding user message to existing conversation.
// This function manages the logic for:
// - Detecting existing prompts in the session
// - Creating new prompts for fresh conversations
// - Adding user messages to ongoing conversations
// - Handling empty questions for regeneration scenarios
func (h *ChatHandler) handlePromptCreation(ctx context.Context, w http.ResponseWriter, chatSession *sqlc_queries.ChatSession, chatUuid, newQuestion string, userID int32, baseURL string) bool {
	existingPrompt := true
	prompt, err := h.service.q.GetOneChatPromptBySessionUUID(ctx, chatSession.Uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("No existing prompt found for session: %s", chatSession.Uuid)
			existingPrompt = false
		} else {
			log.Printf("Error checking prompt for session %s: %v", chatSession.Uuid, err)
			RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to get prompt", err.Error()))
			return false
		}
	} else {
		log.Printf("Found existing prompt ID %d for session %s", prompt.ID, chatSession.Uuid)
	}

	if existingPrompt {
		if newQuestion != "" {
			_, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode)
			if err != nil {
				RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to create message", err.Error()))
				return false
			}
		} else {
			log.Println("no new question, regenerate answer")
		}
	} else {
		chatPrompt, err := h.service.CreateChatPromptSimple(ctx, chatSession.Uuid, newQuestion, userID)
		if err != nil {
			RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to create prompt", err.Error()))
			return false
		}
		log.Printf("%+v\n", chatPrompt)
	}
	return true
}

// generateAndSaveAnswer generates the LLM response and saves it to the database
func (h *ChatHandler) generateAndSaveAnswer(ctx context.Context, w http.ResponseWriter, chatSession *sqlc_queries.ChatSession, chatUuid string, userID int32, baseURL string, streamOutput bool) bool {
	msgs, err := h.service.getAskMessages(*chatSession, chatUuid, false)
	if err != nil {
		log.Printf("Error collecting messages for session %s: %v", chatSession.Uuid, err)
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to collect messages", err.Error()))
		return false
	}
	log.Printf("Collected messages for processing - SessionUUID: %s, MessageCount: %d, Model: %s", chatSession.Uuid, len(msgs), chatSession.Model)

	model := h.chooseChatModel(*chatSession, msgs)
	LLMAnswer, err := model.Stream(w, *chatSession, msgs, chatUuid, false, streamOutput)
	if err != nil {
		log.Printf("Error generating answer: %v", err)
		RespondWithAPIError(w, WrapError(err, "Failed to generate answer"))
		return false
	}
	if LLMAnswer == nil {
		log.Printf("Error generating answer: LLMAnswer is nil")
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("LLMAnswer is nil"))
		return false
	}

	if !isTest(msgs) {
		log.Printf("LLMAnswer: %+v", LLMAnswer)
		h.service.logChat(*chatSession, msgs, LLMAnswer.ReasoningContent+LLMAnswer.Answer)
	}

	if _, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to create message", err.Error()))
		return false
	}
	return true
}

// genAnswer is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
// if there is no prompt yet, it will create a new prompt and use it as request
// otherwise, it will create a message, use prompt + get latest N message + newQuestion as request
func genAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string, newQuestion string, userID int32, streamOutput bool) {
	ctx := context.Background()

	// Validate chat session and get model info
	chatSession, _, baseURL, ok := h.validateChatSession(ctx, w, chatSessionUuid)
	if !ok {
		return
	}
	log.Printf("Processing chat session - SessionUUID: %s, UserID: %d, Model: %s", chatSession.Uuid, userID, chatSession.Model)

	// Handle prompt creation or user message addition
	if !h.handlePromptCreation(ctx, w, chatSession, chatUuid, newQuestion, userID, baseURL) {
		return
	}

	// Generate and save the answer
	h.generateAndSaveAnswer(ctx, w, chatSession, chatUuid, userID, baseURL, streamOutput)
}

func genBotAnswer(h *ChatHandler, w http.ResponseWriter, session sqlc_queries.ChatSession, simpleChatMessages []SimpleChatMessage, snapshotUuid, newQuestion string, userID int32, streamOutput bool) {
	_, err := h.service.q.ChatModelByName(context.Background(), session.Model)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat model: " + session.Model)
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	messages := simpleChatMessagesToMessages(simpleChatMessages)
	messages = append(messages, models.Message{
		Role:    "user",
		Content: newQuestion,
	})
	model := h.chooseChatModel(session, messages)

	LLMAnswer, err := model.Stream(w, session, messages, "", false, streamOutput)
	if err != nil {
		RespondWithAPIError(w, WrapError(err, "Failed to generate answer"))
		return
	}

	ctx := context.Background()

	// Save to bot answer history
	historyParams := sqlc_queries.CreateBotAnswerHistoryParams{
		BotUuid:    snapshotUuid,
		UserID:     userID,
		Prompt:     newQuestion,
		Answer:     LLMAnswer.Answer,
		Model:      session.Model,
		TokensUsed: int32(len(LLMAnswer.Answer)) / 4, // Approximate token count
	}
	if _, err := h.service.q.CreateBotAnswerHistory(ctx, historyParams); err != nil {
		log.Printf("Failed to save bot answer history: %v", err)
		// Don't fail the request, just log the error
	}

	if !isTest(messages) {
		h.service.logChat(session, messages, LLMAnswer.Answer)
	}
}

// Helper function to convert SimpleChatMessage to Message
func simpleChatMessagesToMessages(simpleChatMessages []SimpleChatMessage) []models.Message {
	messages := make([]models.Message, len(simpleChatMessages))
	for i, scm := range simpleChatMessages {
		role := "user"
		if scm.Inversion {
			role = "assistant"
		}
		if i == 0 {
			role = "system"
		}
		messages[i] = models.Message{
			Role:    role,
			Content: scm.Text,
		}
	}
	return messages
}

func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string, stream bool) {
	ctx := context.Background()

	// Validate chat session
	chatSession, _, _, ok := h.validateChatSession(ctx, w, chatSessionUuid)
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

	model := h.chooseChatModel(*chatSession, msgs)
	LLMAnswer, err := model.Stream(w, *chatSession, msgs, chatUuid, true, stream)
	if err != nil {
		log.Printf("Error regenerating answer: %v", err)
		return
	}

	h.service.logChat(*chatSession, msgs, LLMAnswer.Answer)

	if err := h.service.UpdateChatMessageContent(ctx, chatUuid, LLMAnswer.Answer); err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to update message"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
}

func (h *ChatHandler) chooseChatModel(chat_session sqlc_queries.ChatSession, msgs []models.Message) ChatModel {
	model := chat_session.Model
	isTestChat := isTest(msgs)
	isClaude3 := strings.HasPrefix(model, "claude-3") || strings.HasPrefix(model, "claude-sonnet-4")
	isOllama := strings.HasPrefix(model, "ollama-")
	isGemini := strings.HasPrefix(model, "gemini")

	completionModel := mapset.NewSet[string]()

	// completionModel.Add(openai.GPT3TextDavinci002)
	isCompletion := completionModel.Contains(model)
	isCustom := strings.HasPrefix(model, "custom-")

	var chatModel ChatModel
	if isTestChat {
		chatModel = &TestChatModel{h: h}
	} else if isClaude3 {
		chatModel = &Claude3ChatModel{h: h}
	} else if isOllama {
		chatModel = &OllamaChatModel{h: h}
	} else if isCompletion {
		chatModel = &CompletionChatModel{h: h}
	} else if isGemini {
		chatModel = NewGeminiChatModel(h)
	} else if isCustom {
		chatModel = &CustomChatModel{h: h}
	} else {
		chatModel = &OpenAIChatModel{h: h}
	}
	return chatModel
}

// isTest determines if the chat messages indicate this is a test scenario
func isTest(msgs []models.Message) bool {
	if len(msgs) == 0 {
		return false
	}
	
	lastMsgs := msgs[len(msgs)-1]
	promptMsg := msgs[0]
	
	// Check if either first or last message contains test demo marker
	return (len(promptMsg.Content) >= TestPrefixLength && promptMsg.Content[:TestPrefixLength] == TestDemoPrefix) ||
		   (len(lastMsgs.Content) >= TestPrefixLength && lastMsgs.Content[:TestPrefixLength] == TestDemoPrefix)
}

func (h *ChatHandler) CheckModelAccess(w http.ResponseWriter, chatSessionUuid string, model string, userID int32) bool {
	chatModel, err := h.service.q.ChatModelByName(context.Background(), model)
	log.Printf("%+v", chatModel)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model"+chatModel.Name))
		return true
	}
	if !chatModel.EnablePerModeRatelimit {
		return false
	}
	ctx := context.Background()
	rate, err := h.service.q.RateLimiteByUserAndSessionUUID(ctx,
		sqlc_queries.RateLimiteByUserAndSessionUUIDParams{
			Uuid:   chatSessionUuid,
			UserID: userID,
		})
	log.Printf("%+v", rate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rate limit is found, use a default value instead of returning an error
			log.Printf("No rate limit found for user %d and session %s, using default", userID, chatSessionUuid)
			return false
		}

		apiErr := WrapError(MapDatabaseError(err), "Failed to get rate limit")
		RespondWithAPIError(w, apiErr)
		return true
	}

	// get last model usage in 10min
	usage10Min, err := h.service.q.GetChatMessagesCountByUserAndModel(ctx,
		sqlc_queries.GetChatMessagesCountByUserAndModelParams{
			UserID: userID,
			Model:  rate.ChatModelName,
		})

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to get usage data"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return true
	}

	log.Printf("%+v", usage10Min)

	if int32(usage10Min) > rate.RateLimit {
		apiErr := ErrTooManyRequests
		apiErr.Message = fmt.Sprintf("Rate limit exceeded for %s", rate.ChatModelName)
		apiErr.Detail = fmt.Sprintf("Usage: %d, Limit: %d", usage10Min, rate.RateLimit)
		RespondWithAPIError(w, apiErr)
		return true
	}
	return false
}

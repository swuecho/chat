package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	openai "github.com/sashabaranov/go-openai"

	"github.com/gin-gonic/gin"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatHandler struct {
	service         *ChatService
	chatfileService *ChatFileService
	requestCtx      context.Context // Store the request context for streaming
}

func NewChatHandler(sqlc_q *sqlc_queries.Queries) *ChatHandler {
	// create a new ChatService instance
	chatService := NewChatService(sqlc_q)
	ChatFileService := NewChatFileService(sqlc_q)
	return &ChatHandler{
		service:         chatService,
		chatfileService: ChatFileService,
		requestCtx:      context.Background(),
	}
}

// GinRegister registers routes with Gin router
func (h *ChatHandler) GinRegister(rg *gin.RouterGroup) {
	rg.POST("/chat_stream", h.GinChatCompletionHandler)
	rg.POST("/chatbot", h.GinChatBotCompletionHandler)
	rg.GET("/chat_instructions", h.GinGetChatInstructions)
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

type ChatInstructionResponse struct {
	ArtifactInstruction string `json:"artifactInstruction"`
	ToolInstruction     string `json:"toolInstruction"`
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

		// Update session title with first 10 words of the prompt
		if newQuestion != "" {
			sessionTitle := firstNWords(newQuestion, 10)
			if sessionTitle != "" {
				updateParams := sqlc_queries.UpdateChatSessionTopicByUUIDParams{
					Uuid:   chatSession.Uuid,
					UserID: userID,
					Topic:  sessionTitle,
				}
				_, err := h.service.q.UpdateChatSessionTopicByUUID(ctx, updateParams)
				if err != nil {
					log.Printf("Warning: Failed to update session title for session %s: %v", chatSession.Uuid, err)
				} else {
					log.Printf("Updated session %s title to: %s", chatSession.Uuid, sessionTitle)
				}
			}
		}
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

	// Store the request context so models can access it
	h.requestCtx = ctx
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

	chatMessage, err := h.service.CreateChatMessageWithSuggestedQuestions(ctx, chatSession.Uuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, chatSession.Model, userID, baseURL, chatSession.SummarizeMode, chatSession.ExploreMode, msgs)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to create message", err.Error()))
		return false
	}

	// Send suggested questions as a separate streaming event if streaming is enabled and exploreMode is on
	if streamOutput && chatSession.ExploreMode && chatMessage.SuggestedQuestions != nil {
		h.sendSuggestedQuestionsStream(w, LLMAnswer.AnswerId, chatMessage.SuggestedQuestions)
	}

	// Generate a better title using LLM for the first exchange (async, non-blocking)
	go h.generateSessionTitle(ctx, chatSession, userID)

	return true
}

// generateSessionTitle generates a better title using LLM for the first message exchange
// It checks if this is the first assistant message in the session, and if so,
// generates a more descriptive title using Gemini
func (h *ChatHandler) generateSessionTitle(ctx context.Context, chatSession *sqlc_queries.ChatSession, userID int32) {
	// Skip if topic is already set (non-empty and not default)
	if chatSession.Topic != "" && !strings.HasPrefix(chatSession.Topic, "New Chat") {
		return
	}

	// Only generate title for the first assistant message
	// Get all messages to check if this is the first exchange
	messages, err := h.service.q.GetChatMessagesBySessionUUID(ctx, sqlc_queries.GetChatMessagesBySessionUUIDParams{
		Uuid:   chatSession.Uuid,
		Offset: 0,
		Limit:  100,
	})
	if err != nil {
		log.Printf("Warning: Failed to get messages for title generation: %v", err)
		return
	}

	// Count user and assistant messages
	var userCount, assistantCount int
	var chatText string
	for _, msg := range messages {
		if msg.Role == "user" {
			userCount++
			chatText += "user: " + msg.Content + "\n"
		} else if msg.Role == "assistant" {
			assistantCount++
			chatText += "assistant: " + msg.Content + "\n"
		}
	}

	// Only generate title if we have exactly 1 user message and 1 assistant message (first exchange)
	if userCount != 1 || assistantCount != 1 {
		return
	}

	// Use the same approach as chat_snapshot_service.go - check if gemini-2.0-flash is available
	model := "gemini-2.0-flash"
	_, err = h.service.q.ChatModelByName(ctx, model)
	if err != nil {
		// Model not available, skip title generation
		return
	}

	// Generate title using Gemini
	genTitle, err := GenerateChatTitle(ctx, model, chatText)
	if err != nil {
		log.Printf("Warning: Failed to generate session title: %v", err)
		return
	}

	if genTitle == "" {
		return
	}

	// Update the session title
	updateParams := sqlc_queries.UpdateChatSessionTopicByUUIDParams{
		Uuid:   chatSession.Uuid,
		UserID: userID,
		Topic:  genTitle,
	}
	_, err = h.service.q.UpdateChatSessionTopicByUUID(ctx, updateParams)
	if err != nil {
		log.Printf("Warning: Failed to update session title: %v", err)
		return
	}

	log.Printf("Generated LLM title for session %s: %s", chatSession.Uuid, genTitle)
}

// sendSuggestedQuestionsStream sends suggested questions as a separate streaming event
func (h *ChatHandler) sendSuggestedQuestionsStream(w http.ResponseWriter, answerID string, suggestedQuestionsJSON json.RawMessage) {
	// Parse the suggested questions JSON
	var suggestedQuestions []string
	if err := json.Unmarshal(suggestedQuestionsJSON, &suggestedQuestions); err != nil {
		log.Printf("Warning: Failed to parse suggested questions for streaming: %v", err)
		return
	}

	// Only send if we have questions
	if len(suggestedQuestions) == 0 {
		return
	}

	// Get the flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Printf("Warning: Response writer does not support flushing, cannot send suggested questions stream")
		return
	}

	// Create a special response with suggested questions
	suggestedQuestionsResponse := map[string]interface{}{
		"id":     answerID,
		"object": "chat.completion.chunk",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"delta": map[string]interface{}{
					"content":            "", // Empty content
					"suggestedQuestions": suggestedQuestions,
				},
				"finish_reason": nil,
			},
		},
	}

	data, err := json.Marshal(suggestedQuestionsResponse)
	if err != nil {
		log.Printf("Warning: Failed to marshal suggested questions response: %v", err)
		return
	}

	// Send the streaming event
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()

	log.Printf("Sent suggested questions stream for answer ID: %s, questions: %v", answerID, suggestedQuestions)
}

// genAnswer is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
// if there is no prompt yet, it will create a new prompt and use it as request
// otherwise, it will create a message, use prompt + get latest N message + newQuestion as request
func genAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, chatSessionUuid string, chatUuid string, newQuestion string, userID int32, streamOutput bool) {

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

func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, ctx context.Context, chatSessionUuid string, chatUuid string, stream bool) {

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

	// Store the request context so models can access it
	h.requestCtx = ctx
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

	// Generate suggested questions if explore mode is enabled
	if chatSession.ExploreMode {
		suggestedQuestions := h.service.generateSuggestedQuestions(LLMAnswer.Answer, msgs)
		if len(suggestedQuestions) > 0 {
			// Update the message with suggested questions in database
			questionsJSON, err := json.Marshal(suggestedQuestions)
			if err == nil {
				h.service.UpdateChatMessageSuggestions(ctx, chatUuid, questionsJSON)

				// Stream suggested questions to frontend
				if stream {
					h.sendSuggestedQuestionsStream(w, LLMAnswer.AnswerId, questionsJSON)
				}
			}
		}
	}
}

// GetRequestContext returns the current request context for streaming operations
func (h *ChatHandler) GetRequestContext() context.Context {
	return h.requestCtx
}

func (h *ChatHandler) chooseChatModel(chat_session sqlc_queries.ChatSession, msgs []models.Message) ChatModel {
	model := chat_session.Model
	isTestChat := isTest(msgs)

	// If this is a test chat, return the test model immediately
	if isTestChat {
		return &TestChatModel{h: h}
	}

	// Get the chat model from database to access api_type field
	chatModel, err := GetChatModel(h.service.q, model)
	if err != nil {
		// Fallback to OpenAI if model not found in database
		return &OpenAIChatModel{h: h}
	}

	// Use api_type field from database instead of string prefix matching
	apiType := chatModel.ApiType

	completionModel := mapset.NewSet[string]()
	// completionModel.Add(openai.GPT3TextDavinci002)
	isCompletion := completionModel.Contains(model)

	var chatModelImpl ChatModel
	switch apiType {
	case "claude":
		chatModelImpl = &Claude3ChatModel{h: h}
	case "ollama":
		chatModelImpl = &OllamaChatModel{h: h}
	case "gemini":
		chatModelImpl = NewGeminiChatModel(h)
	case "custom":
		chatModelImpl = &CustomChatModel{h: h}
	case "openai":
		if isCompletion {
			chatModelImpl = &CompletionChatModel{h: h}
		} else {
			chatModelImpl = &OpenAIChatModel{h: h}
		}
	default:
		// Default to OpenAI for unknown api types
		chatModelImpl = &OpenAIChatModel{h: h}
	}
	return chatModelImpl
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

// =============================================================================
// Gin Handlers
// =============================================================================

// GinGetChatInstructions handles GET requests for chat instructions
func (h *ChatHandler) GinGetChatInstructions(c *gin.Context) {
	artifactInstruction, err := loadArtifactInstruction()
	if err != nil {
		log.Printf("Warning: Failed to load artifact instruction: %v", err)
		artifactInstruction = ""
	}

	toolInstruction, err := loadToolInstruction()
	if err != nil {
		log.Printf("Warning: Failed to load tool instruction: %v", err)
		toolInstruction = ""
	}

	c.JSON(http.StatusOK, ChatInstructionResponse{
		ArtifactInstruction: artifactInstruction,
		ToolInstruction:     toolInstruction,
	})
}

// GinChatBotCompletionHandler handles POST requests for bot chat with SSE streaming
func (h *ChatHandler) GinChatBotCompletionHandler(c *gin.Context) {
	var req BotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrValidationInvalidInput("Failed to decode request body").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	snapshotUuid := req.SnapshotUuid
	newQuestion := req.Message

	log.Printf("snapshotUuid: %s", snapshotUuid)
	log.Printf("newQuestion: %s", newQuestion)

	ctx := c.Request.Context()

	userID, err := GetUserID(c)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	fmt.Printf("userID: %d", userID)

	chatSnapshot, err := h.service.q.ChatSnapshotByUserIdAndUuid(ctx, sqlc_queries.ChatSnapshotByUserIdAndUuidParams{
		UserID: userID,
		Uuid:   snapshotUuid,
	})
	if err != nil {
		ErrResourceNotFound("Chat snapshot").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	fmt.Printf("chatSnapshot: %+v", chatSnapshot)

	var session sqlc_queries.ChatSession
	err = json.Unmarshal(chatSnapshot.Session, &session)
	if err != nil {
		ErrInternalUnexpected.WithDetail("Failed to deserialize chat session").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}
	var simpleChatMessages []SimpleChatMessage
	err = json.Unmarshal(chatSnapshot.Conversation, &simpleChatMessages)
	if err != nil {
		ErrInternalUnexpected.WithDetail("Failed to deserialize conversation").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	// Use c.Writer for SSE streaming (compatible with http.ResponseWriter)
	genBotAnswer(h, c.Writer, session, simpleChatMessages, snapshotUuid, newQuestion, userID, req.Stream)
}

// GinChatCompletionHandler handles POST requests for chat with SSE streaming
func (h *ChatHandler) GinChatCompletionHandler(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		ErrValidationInvalidInput("Invalid request format").WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	chatSessionUuid := req.SessionUuid
	chatUuid := req.ChatUuid
	newQuestion := req.Prompt

	log.Printf("chatSessionUuid: %s", chatSessionUuid)
	log.Printf("chatUuid: %s", chatUuid)
	log.Printf("newQuestion: %s", newQuestion)

	ctx := c.Request.Context()

	userID, err := GetUserID(c)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		ErrAuthInvalidCredentials.WithDebugInfo(err.Error()).GinResponse(c)
		return
	}

	// Use c.Writer for SSE streaming (compatible with http.ResponseWriter)
	if req.Regenerate {
		regenerateAnswer(h, c.Writer, ctx, chatSessionUuid, chatUuid, req.Stream)
	} else {
		genAnswer(h, c.Writer, ctx, chatSessionUuid, chatUuid, newQuestion, userID, req.Stream)
	}
}

// GinCheckModelAccess checks model access for Gin context (helper for future use)
func (h *ChatHandler) GinCheckModelAccess(c *gin.Context, chatSessionUuid string, model string, userID int32) bool {
	chatModel, err := h.service.q.ChatModelByName(context.Background(), model)
	if err != nil {
		log.WithError(err).WithField("model", model).Error("Chat model not found")
		ErrResourceNotFound("chat model: " + model).GinResponse(c)
		return true
	}
	log.Printf("%+v", chatModel)
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
			log.Printf("No rate limit found for user %d and session %s, using default", userID, chatSessionUuid)
			return false
		}

		WrapError(MapDatabaseError(err), "Failed to get rate limit").GinResponse(c)
		return true
	}

	usage10Min, err := h.service.q.GetChatMessagesCountByUserAndModel(ctx,
		sqlc_queries.GetChatMessagesCountByUserAndModelParams{
			UserID: userID,
			Model:  rate.ChatModelName,
		})

	if err != nil {
		ErrInternalUnexpected.WithDetail("Failed to get usage data").WithDebugInfo(err.Error()).GinResponse(c)
		return true
	}

	log.Printf("%+v", usage10Min)

	if int32(usage10Min) > rate.RateLimit {
		ErrTooManyRequests.WithMessage(fmt.Sprintf("Rate limit exceeded for %s", rate.ChatModelName)).
			WithDetail(fmt.Sprintf("Usage: %d, Limit: %d", usage10Min, rate.RateLimit)).GinResponse(c)
		return true
	}
	return false
}

// CheckModelAccess checks model access using http.ResponseWriter (for use with SSE streaming)
func (h *ChatHandler) CheckModelAccess(w http.ResponseWriter, chatSessionUuid string, model string, userID int32) bool {
	chatModel, err := h.service.q.ChatModelByName(context.Background(), model)
	if err != nil {
		log.WithError(err).WithField("model", model).Error("Chat model not found")
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+model))
		return true
	}
	log.Printf("%+v", chatModel)
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
			log.Printf("No rate limit found for user %d and session %s, using default", userID, chatSessionUuid)
			return false
		}

		RespondWithAPIError(w, WrapError(MapDatabaseError(err), "Failed to get rate limit"))
		return true
	}

	usage10Min, err := h.service.q.GetChatMessagesCountByUserAndModel(ctx,
		sqlc_queries.GetChatMessagesCountByUserAndModelParams{
			UserID: userID,
			Model:  rate.ChatModelName,
		})

	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to get usage data").WithDebugInfo(err.Error()))
		return true
	}

	log.Printf("%+v", usage10Min)

	if int32(usage10Min) > rate.RateLimit {
		RespondWithAPIError(w, ErrTooManyRequests.WithMessage(fmt.Sprintf("Rate limit exceeded for %s", rate.ChatModelName)).
			WithDetail(fmt.Sprintf("Usage: %d, Limit: %d", usage10Min, rate.RateLimit)))
		return true
	}
	return false
}

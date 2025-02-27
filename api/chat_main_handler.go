package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rotisserie/eris"
	openai "github.com/sashabaranov/go-openai"

	"github.com/gorilla/mux"

	claude "github.com/swuecho/chat_backend/llm/claude"
	"github.com/swuecho/chat_backend/llm/gemini"
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
	FinishReason interface{}                  `json:"finish_reason"`
	Index        int                          `json:"index"`
}

type OpenaiChatRequest struct {
	Model    string                         `json:"model"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

func NewUserMessage(content string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{Role: "user", Content: content}
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
		log.Printf("Error decoding request: %v", err)
		apiErr := ErrValidationInvalidInput("Failed to decode request body")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
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

	genBotAnswer(h, w, session, simpleChatMessages, newQuestion, userID, req.Stream)

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

// regenerateAnswer is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
// if there is no prompt yet, it will create a new prompt and use it as request
// otherwise,
//
//	it will create a message, use prompt + get latest N message + newQuestion as request
func genAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string, newQuestion string, userID int32, streamOutput bool) {
	ctx := context.Background()
	chatSession, err := h.service.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	fmt.Printf("chatSession: %+v ", chatSession)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat session: "+chatSessionUuid))
		return
	}

	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return
	}
	baseURL, _ := getModelBaseUrl(chatModel.Url)

	existingPrompt := true

	_, err = h.service.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existingPrompt = false
		} else {
			http.Error(w, eris.Wrap(err, "fail to get prompt: ").Error(), http.StatusInternalServerError)
		}
	}

	if existingPrompt {
		if newQuestion != "" {
			_, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode)
			if err != nil {
				apiErr := ErrInternalUnexpected
				apiErr.Detail = "Failed to create message"
				apiErr.DebugInfo = err.Error()
				RespondWithAPIError(w, apiErr)
				return
			}
		} else {
			log.Println("no new question, regenerate answer")
		}
	} else {
		chatPrompt, err := h.service.CreateChatPromptSimple(chatSessionUuid, newQuestion, userID)
		if err != nil {
			apiErr := ErrInternalUnexpected
			apiErr.Detail = "Failed to create prompt"
			apiErr.DebugInfo = err.Error()
			RespondWithAPIError(w, apiErr)
			return
		}
		log.Printf("%+v\n", chatPrompt)
	}

	msgs, err := h.service.getAskMessages(chatSession, chatUuid, false)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to collect messages"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	model := h.chooseChatModel(chatSession, msgs)
	LLMAnswer, err := model.Stream(w, chatSession, msgs, chatUuid, false, streamOutput)
	if err != nil {
		log.Printf("Error generating answer: %v", err)
		return
	}
	if !isTest(msgs) {
		log.Printf("LLMAnswer: %+v", LLMAnswer)
		h.service.logChat(chatSession, msgs, LLMAnswer.ReasoningContent+LLMAnswer.Answer)
	}

	if _, err := h.service.CreateChatMessageSimple(ctx, chatSessionUuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create message"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}
}

func genBotAnswer(h *ChatHandler, w http.ResponseWriter, session sqlc_queries.ChatSession, simpleChatMessages []SimpleChatMessage, newQuestion string, userID int32, streamOutput bool) {
	chatModel, err := h.service.q.ChatModelByName(context.Background(), session.Model)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat model: " + session.Model)
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	baseURL, _ := getModelBaseUrl(chatModel.Url)

	messages := simpleChatMessagesToMessages(simpleChatMessages)
	messages = append(messages, models.Message{
		Role:    "user",
		Content: newQuestion,
	})
	model := h.chooseChatModel(session, messages)

	LLMAnswer, err := model.Stream(w, session, messages, "", false, streamOutput)
	if err != nil {
		log.Printf("Error generating answer: %v", err)
		return
	}

	if !isTest(messages) {
		h.service.logChat(session, messages, LLMAnswer.Answer)
	}

	ctx := context.Background()
	if _, err := h.service.CreateChatMessageSimple(ctx, session.Uuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, session.Model, userID, baseURL, session.SummarizeMode); err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create message"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
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
	chatSession, err := h.service.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		apiErr := ErrResourceNotFound("Chat session")
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	msgs, err := h.service.getAskMessages(chatSession, chatUuid, true)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to get chat messages"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return
	}

	// // calc total tokens
	// totalTokens := totalInputToken(msgs)

	// if totalTokens > chatSession.MaxTokens*2/3 {
	// 	RespondWithError(w, http.StatusRequestEntityTooLarge, "error.token_length_exceed_limit",
	// 		map[string]interface{}{
	// 			"max_tokens":   chatSession.MaxTokens,
	// 			"total_tokens": totalTokens,
	// 		})
	// 	return
	// }

	// Determine whether the chat is a test or not
	model := h.chooseChatModel(chatSession, msgs)

	LLMAnswer, err := model.Stream(w, chatSession, msgs, chatUuid, true, stream)
	if err != nil {
		log.Printf("Error regenerating answer: %v", err)
		return
	}

	h.service.logChat(chatSession, msgs, LLMAnswer.Answer)

	// Delete previous message and create new one
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
	isClaude := strings.HasPrefix(model, "claude")
	isClaude3 := false
	if strings.HasPrefix(model, "claude-3") {
		isClaude = false
		isClaude3 = true
	}
	isOllama := strings.HasPrefix(model, "ollama-")
	isGemini := strings.HasPrefix(model, "gemini")

	completionModel := mapset.NewSet[string]()

	// completionModel.Add(openai.GPT3TextDavinci002)
	isCompletion := completionModel.Contains(model)
	isCustom := strings.HasPrefix(model, "custom-")

	var chatModel ChatModel
	if isClaude {
		chatModel = &ClaudeChatModel{h: h}
	} else if isClaude3 {
		chatModel = &Claude3ChatModel{h: h}
	} else if isTestChat {
		chatModel = &TestChatModel{h: h}
	} else if isOllama {
		chatModel = &OllamaChatModel{h: h}
	} else if isCompletion {
		chatModel = &CompletionChatModel{h: h}
	} else if isGemini {
		chatModel = &GeminiChatModel{h: h}
	} else if isCustom {
		chatModel = &CustomChatModel{h: h}
	} else {
		chatModel = &OpenAIChatModel{h: h}
	}
	return chatModel
}

func isTest(msgs []models.Message) bool {
	lastMsgs := msgs[len(msgs)-1]
	promptMsg := msgs[0]
	return promptMsg.Content == "test_demo_bestqa" || lastMsgs.Content == "test_demo_bestqa"
}

func (h *ChatHandler) CheckModelAccess(w http.ResponseWriter, chatSessionUuid string, model string, userID int32) bool {
	// userID, err := getUserID(r.Context())
	// if err != nil {
	// 	RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
	// 	return true
	// }
	// get chatModel, check the per model rate limit is Enabled
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

func (h *ChatHandler) CompletionStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	// check per chat_model limit

	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return nil, eris.New("exceed per mode rate limit")
	}

	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model "+chatSession.Model))
		return nil, err
	}

	config, err := genOpenAIConfig(chatModel)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to generate OpenAI configuration"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return nil, err
	}

	client := openai.NewClientWithConfig(config)
	// latest message contents
	prompt := chat_compeletion_messages[len(chat_compeletion_messages)-1].Content

	//totalInputToken := chat_compeletion_messages[len(chat_compeletion_messages)-1].TokenCount()
	// max - input = max possible output
	//maxOutputToken := int(chatSession.MaxTokens - totalInputToken) - 500

	N := chatSession.N
	req := openai.CompletionRequest{
		Model: chatSession.Model,
		// MaxTokens:   maxOutputToken,
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
		N:           int(N),
		Prompt:      prompt,
		Stream:      true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	stream, err := client.CreateCompletionStream(ctx, req)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create completion stream"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return nil, err
	}
	defer stream.Close()

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Streaming unsupported by the client"
		RespondWithAPIError(w, apiErr)
		return nil, eris.New("Streaming unsupported!")
	}

	var answer string
	var answer_id string
	textBuffer := newTextBuffer(N, "```\n"+prompt, "\n```\n") // create slice of string builders
	if regenerate {
		answer_id = chatUuid
	}

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// send the last message
			if len(answer) > 0 {
				final_resp := constructChatCompletionStreamReponse(answer_id, answer)
				data, _ := json.Marshal(final_resp)
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
			if chatSession.Debug {
				req_j, _ := json.Marshal(req)
				log.Println(string(req_j))
				answer = answer + "\n" + string(req_j)
				req_as_resp := constructChatCompletionStreamReponse(answer_id, answer)
				data, _ := json.Marshal(req_as_resp)
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
			break
		}
		if err != nil {
			RespondWithErrorMessage(w, http.StatusInternalServerError, fmt.Sprintf("Stream error: %v", err), nil)
			return nil, err
		}
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Text
		textBuffer.appendByIndex(textIdx, delta)
		// log.Println(delta)
		if chatSession.Debug {
			log.Printf("%d: %s", textIdx, delta)
		}
		if answer_id == "" {
			answer_id = response.ID
		}
		// concatenate all string builders into a single string
		answer = textBuffer.String("\n\n")

		perWordStreamLimit := getPerWordStreamLimit()
		if strings.HasSuffix(delta, "\n") || len(answer) < perWordStreamLimit {
			if len(answer) == 0 {
				log.Printf("%s", "no content in answer")
			} else {
				response := constructChatCompletionStreamReponse(answer_id, answer)
				data, _ := json.Marshal(response)
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
		}
	}
	return &models.LLMAnswer{AnswerId: answer_id, Answer: answer}, nil
}

type ClaudeResponse struct {
	Completion string      `json:"completion"`
	Stop       string      `json:"stop"`
	StopReason string      `json:"stop_reason"`
	Truncated  bool        `json:"truncated"`
	LogID      string      `json:"log_id"`
	Model      string      `json:"model"`
	Exception  interface{} `json:"exception"`
}

func (h *ChatHandler) chatStreamClaude(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}

	// OPENAI_API_KEY

	// create a new strings.Builder
	// iterate through the messages and format them
	// print the user's question
	// convert assistant's response to json format
	prompt := claude.FormatClaudePrompt(chat_compeletion_messages)
	// create the json data
	jsonData := map[string]interface{}{
		"prompt":               prompt,
		"model":                chatSession.Model,
		"max_tokens_to_sample": chatSession.MaxTokens,
		"temperature":          chatSession.Temperature,
		"stop_sequences":       []string{"\n\nHuman:"},
		"stream":               true,
	}

	// convert data to json format
	jsonValue, _ := json.Marshal(jsonData)
	// create the request
	req, err := http.NewRequest("POST", chatModel.Url, bytes.NewBuffer(jsonValue))

	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to create request"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return nil, err
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)
	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	// set the streaming flag
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		var apiErr APIError

		// Check for specific HTTP client errors
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded") {
			apiErr = ErrExternalTimeout
			apiErr.Detail = "The AI model service took too long to respond"
			apiErr.DebugInfo = err.Error()
		} else if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			apiErr = ErrExternalUnavailable
			apiErr.Detail = "Could not connect to the AI model service"
			apiErr.DebugInfo = err.Error()
		} else {
			apiErr = ErrInternalUnexpected
			apiErr.Detail = "Failed to create request to AI service"
			apiErr.DebugInfo = err.Error()
		}

		RespondWithAPIError(w, apiErr)
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Streaming unsupported by the client"
		RespondWithAPIError(w, apiErr)
		return nil, err
	}

	var answer string
	var answer_id string

	if regenerate {
		answer_id = chatUuid
	}

	var headerData = []byte("data: ")
	count := 0
	for {
		count++
		// prevent infinite loop
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
			return nil, err
		}
		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("[DONE]")) {
			// stream.isFinished = true
			fmt.Println("DONE break")
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			break
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}
		var response ClaudeResponse
		_ = json.Unmarshal(line, &response)
		answer = response.Completion
		if len(answer) < 200 || len(answer)%2 == 0 {
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

// Claude3 ChatModel implementation
type Claude3ChatModel struct {
	h *ChatHandler
}

func (m *Claude3ChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	log.Printf("%+v", chatSession)
	// Release the API token
	// set the api key
	chatModel, err := m.h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	log.Printf("%+v", chatModel)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}
	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat files "+chatSession.Uuid))
		return nil, err
	}

	// create a new strings.Builder
	// iterate through the messages and format them
	// print the user's question
	// convert assistant's response to json format
	//     "messages": [
	//	{"role": "user", "content": "Hello, world"}
	//	]
	// first message is user instead of system
	var messages []openai.ChatCompletionMessage
	if len(chat_compeletion_messages) > 1 {
		// first message used as system message
		// messages start with second message
		// drop the first assistant message if it is an assistant message
		claude_messages := chat_compeletion_messages[1:]

		if len(claude_messages) > 0 && claude_messages[0].Role == "assistant" {
			claude_messages = claude_messages[1:]
		}
		messages = messagesToOpenAIMesages(claude_messages, chatFiles)
	} else {
		// only system message, return and do nothing
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.system_message_notice", err)
		return nil, err
	}
	// create the json data
	jsonData := map[string]interface{}{
		"system":      chat_compeletion_messages[0].Content,
		"model":       chatSession.Model,
		"messages":    messages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"top_p":       chatSession.TopP,
		"stream":      stream,
	}
	log.Printf("%+v", jsonData)

	// convert data to json format
	jsonValue, _ := json.Marshal(jsonData)
	log.Printf("%+v", string(jsonValue))
	// create the request
	req, err := http.NewRequest("POST", chatModel.Url, bytes.NewBuffer(jsonValue))

	if err != nil {
		log.Printf("%+v", err)
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_make_request", err)
		return nil, err
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)
	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	if !stream {
		req.Header.Set("Accept", "application/json")
		return doGenerateClaude3(w, req)
	} else {
		// set the streaming flag
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		return m.h.chatStreamClaude3(w, req, chatUuid, regenerate)
	}
}

func doGenerateClaude3(w http.ResponseWriter, req *http.Request) (*models.LLMAnswer, error) {

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%+v", err)
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}

	// Unmarshal directly from resp.Body
	var message claude.Response
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_unmarshal_response", err)
		return nil, err
	}
	defer resp.Body.Close()
	uuid := message.ID
	firstMessage := message.Content[0].Text
	answer := constructChatCompletionStreamReponse(uuid, firstMessage)
	data, _ := json.Marshal(answer)
	fmt.Fprint(w, string(data))
	return &models.LLMAnswer{
		AnswerId: uuid,
		Answer:   firstMessage,
	}, nil
}

// claude-3-opus-20240229
// claude-3-sonnet-20240229
// claude-3-haiku-20240307
func (h *ChatHandler) chatStreamClaude3(w http.ResponseWriter, req *http.Request, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%+v", err)
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Streaming unsupported by the client"
		RespondWithAPIError(w, apiErr)
		return nil, err
	}

	var answer string
	var answer_id string

	if regenerate {
		answer_id = chatUuid
	}

	var headerData = []byte("data: ")
	count := 0
	for {
		count++
		// prevent infinite loop
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		log.Printf("%+v", string(line))
		if err != nil {
			if errors.Is(err, io.EOF) {
				if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
					log.Println(string(line))
					data, _ := json.Marshal(constructChatCompletionStreamReponse(NewUUID(), string(line)))
					fmt.Fprintf(w, "data: %v\n\n", string(data))
					flusher.Flush()
				}
				fmt.Println("End of stream reached")
				return nil, err
			}
			return nil, err
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("event: message_stop")) {
			// stream.isFinished = true
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			break
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
			log.Println(string(line))
			RespondWithErrorMessage(w, http.StatusInternalServerError, string(line), nil)
			return nil, err
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_start\"")) {
			answer = claude.AnswerFromBlockStart(line)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_delta\"")) {
			answer += claude.AnswerFromBlockDelta(line)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}
	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

type OllamaResponse struct {
	Model              string         `json:"model"`
	CreatedAt          time.Time      `json:"created_at"`
	Done               bool           `json:"done"`
	Message            models.Message `json:"message"`
	TotalDuration      int64          `json:"total_duration"`
	LoadDuration       int64          `json:"load_duration"`
	PromptEvalCount    int            `json:"prompt_eval_count"`
	PromptEvalDuration int64          `json:"prompt_eval_duration"`
	EvalCount          int            `json:"eval_count"`
	EvalDuration       int64          `json:"eval_duration"`
}

func (h *ChatHandler) chatOllamStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}
	jsonData := map[string]interface{}{
		"model":    strings.Replace(chatSession.Model, "ollama-", "", 1),
		"messages": chat_compeletion_messages,
	}
	// convert data to json format
	jsonValue, _ := json.Marshal(jsonData)
	// create the request
	req, err := http.NewRequest("POST", chatModel.Url, bytes.NewBuffer(jsonValue))

	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_make_request", err)
		return nil, err
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)
	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	// set the streaming flag
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return nil, err
	}

	var answer string
	var answer_id string

	if regenerate {
		answer_id = chatUuid
	}

	count := 0
	for {
		count++
		// prevent infinite loop
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
			return nil, err
		}
		var streamResp OllamaResponse
		err = json.Unmarshal(line, &streamResp)
		if err != nil {
			return nil, err
		}
		answer += strings.ReplaceAll(streamResp.Message.Content, "<0x0A>", "\n")
		if streamResp.Done {
			// stream.isFinished = true
			fmt.Println("DONE break")
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			break
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}

		if len(answer) < 200 || len(answer)%2 == 0 {
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

type CustomModelResponse struct {
	Completion string      `json:"completion"`
	Stop       string      `json:"stop"`
	StopReason string      `json:"stop_reason"`
	Truncated  bool        `json:"truncated"`
	LogID      string      `json:"log_id"`
	Model      string      `json:"model"`
	Exception  interface{} `json:"exception"`
}

func (h *ChatHandler) customChatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	// set the api key
	chat_model, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}
	apiKey := os.Getenv(chat_model.ApiAuthKey)
	// set the url
	url := chat_model.Url

	// create a new strings.Builder
	// iterate through the messages and format them
	// print the user's question
	// convert assistant's response to json format
	prompt := claude.FormatClaudePrompt(chat_compeletion_messages)
	// create the json data
	jsonData := map[string]interface{}{
		"prompt":               prompt,
		"model":                chatSession.Model,
		"max_tokens_to_sample": chatSession.MaxTokens,
		"temperature":          chatSession.Temperature,
		"stop_sequences":       []string{"\n\nHuman:"},
		"stream":               true,
	}

	// convert data to json format
	jsonValue, _ := json.Marshal(jsonData)
	// create the request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		fmt.Println("Error while creating request: ", err)
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "post to claude api").Error(), err)
		return nil, err
	}

	authHeaderName := chat_model.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	// set the streaming flag
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// create the http client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while sending request: ", err)
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return nil, err
	}

	var answer string
	var answer_id string

	if regenerate {
		answer_id = chatUuid
	}

	var headerData = []byte("data: ")
	count := 0
	for {
		count++
		// prevent infinite loop
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
			return nil, err
		}
		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("[DONE]")) {
			// stream.isFinished = true
			fmt.Println("DONE break")
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			break
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}
		var response CustomModelResponse
		_ = json.Unmarshal(line, &response)
		answer = response.Completion
		if len(answer) < 200 || len(answer)%2 == 0 {
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

func (h *ChatHandler) chatStreamTest(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	//message := Message{Role: "assitant", Content:}
	chatFiles, err := h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Failed to get chat files"
		apiErr.DebugInfo = err.Error()
		RespondWithAPIError(w, apiErr)
		return nil, err
	}

	answer_id := chatUuid
	if !regenerate {
		answer_id = NewUUID()
	}
	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)

	if !ok {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return nil, err
	}
	answer := "Hi, I am a chatbot. I can help you to find the best answer for your question. Please ask me a question."
	resp := constructChatCompletionStreamReponse(answer_id, answer)
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()

	if chatSession.Debug {
		openai_req := NewChatCompletionRequest(chatSession, chat_compeletion_messages, chatFiles, false)

		req_j, _ := json.Marshal(openai_req)
		answer = answer + "\n" + string(req_j)
		req_as_resp := constructChatCompletionStreamReponse(answer_id, answer)
		data, _ := json.Marshal(req_as_resp)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil

}

func NewChatCompletionRequest(chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatFiles []sqlc_queries.ChatFile, streamOutput bool) openai.ChatCompletionRequest {

	openai_message := messagesToOpenAIMesages(chat_compeletion_messages, chatFiles)
	//totalInputToken := lo.SumBy(chat_compeletion_messages, func(m Message) int32 {
	//	return m.TokenCount()
	//})
	// max - input = max possible output
	//maxOutputToken := int(chatSession.MaxTokens - totalInputToken) - 500 // offset
	for _, m := range openai_message {
		b, _ := m.MarshalJSON()
		log.Printf("messages: %+v\n", string(b))
	}

	log.Printf("messages: %+v\n", openai_message)
	openai_req := openai.ChatCompletionRequest{
		Model:    chatSession.Model,
		Messages: openai_message,
		//MaxTokens:   maxOutputToken,
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP) - 0.01,
		N:           int(chatSession.N),
		Stream:      streamOutput,
	}
	return openai_req
}

func constructChatCompletionStreamReponse(answer_id string, answer string) openai.ChatCompletionStreamResponse {
	resp := openai.ChatCompletionStreamResponse{
		ID: answer_id,
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Content: answer,
				},
			},
		},
	}
	return resp
}

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

// curl https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=$API_KEY \
//     -H 'Content-Type: application/json' \
//     -X POST \
//     -d '{
//       "contents": [{
//         "parts":[{
//           "text": "Write a story about a magic backpack."}]}]}' 2> /dev/null

func (h *ChatHandler) chatStreamGemini(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	chatFiles, err := h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting chat files").Error(), err)
		return nil, err
	}
	payloadBytes, err := gemini.GenGemminPayload(chat_compeletion_messages, chatFiles)
	if err != nil {
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "Error generating gemmi payload").Error(), err)
		return nil, err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=$GEMINI_API_KEY", chatSession.Model)
	if stream {
		url = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=$GEMINI_API_KEY", chatSession.Model)
	}

	url = os.ExpandEnv(url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error while creating request: ", err)
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "create request to gemini api").Error(), err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while sending request: ", err)
		RespondWithErrorMessage(w, http.StatusInternalServerError, eris.Wrap(err, "post to gemini api").Error(), err)
		return nil, err
	}
	defer resp.Body.Close()

	answer_id := chatUuid
	if !regenerate {
		answer_id = NewUUID()
	}

	if !stream {
		// Handle non-streaming response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_read_response", err)
			return nil, err
		}
		// body to GeminiResponse
		var geminiResp gemini.ResponseBody
		err = json.Unmarshal(body, &geminiResp)
		if err != nil {
			RespondWithErrorMessage(w, http.StatusInternalServerError, "error.fail_to_unmarshal_response", err)
			return nil, err
		}
		answer := geminiResp.Candidates[0].Content.Parts[0].Text
		response := constructChatCompletionStreamReponse(answer_id, answer)
		data, _ := json.Marshal(response)
		fmt.Fprint(w, string(data))
		return &models.LLMAnswer{
			Answer:   answer,
			AnswerId: answer_id,
		}, nil
	}

	// Handle streaming response
	setSSEHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithErrorMessage(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return nil, err
	}

	var answer string
	var headerData = []byte("data: ")
	ioreader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()

	count := 0
	for {
		count++
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		if chatSession.Debug {
			log.Printf("%s", line)
		}
		if err != nil {
			// Create an instance of ErrorResponse
			if errors.Is(err, io.EOF) {
				log.Printf("End of stream reached: %+v, %+v", err, line)
				return &models.LLMAnswer{
					Answer:   answer,
					AnswerId: answer_id,
				}, nil
			} else {
				log.Printf("Error while reading response: %+v, %+v", err, line)
				return nil, err
			}
		}
		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)
		if len(line) > 0 {
			answer = gemini.ParseRespLine(line, answer)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}
	return &models.LLMAnswer{
		AnswerId: answer_id,
		Answer:   answer,
	}, nil
}

// Claude ChatModel implementation
type ClaudeChatModel struct {
	h *ChatHandler
}

func (m *ClaudeChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.chatStreamClaude(w, chatSession, chat_compeletion_messages, chatUuid, regenerate)
}

// Test ChatModel implementation
type TestChatModel struct {
	h *ChatHandler
}

func (m *TestChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.chatStreamTest(w, chatSession, chat_compeletion_messages, chatUuid, regenerate)
}

// Ollama ChatModel implementation
type OllamaChatModel struct {
	h *ChatHandler
}

func (m *OllamaChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.chatOllamStream(w, chatSession, chat_compeletion_messages, chatUuid, regenerate)
}

// Completion ChatModel implementation
type CompletionChatModel struct {
	h *ChatHandler
}

func (m *CompletionChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.CompletionStream(w, chatSession, chat_compeletion_messages, chatUuid, regenerate, stream)
}

// Gemini ChatModel implementation
type GeminiChatModel struct {
	h *ChatHandler
}

func (m *GeminiChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.chatStreamGemini(w, chatSession, chat_compeletion_messages, chatUuid, regenerate, stream)
}

// Custom ChatModel implementation
type CustomChatModel struct {
	h *ChatHandler
}

func (m *CustomChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.customChatStream(w, chatSession, chat_compeletion_messages, chatUuid, regenerate)
}

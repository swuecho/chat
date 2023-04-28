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
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/sqlc_queries"

	"github.com/gorilla/mux"
)

type ChatHandler struct {
	chatService *ChatService
}

func NewChatHandler(chatService *ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) Register(router *mux.Router) {
	router.HandleFunc("/chat", h.chatHandler).Methods(http.MethodPost)
	router.HandleFunc("/chat_stream", h.OpenAIChatCompletionAPIWithStreamHandler).Methods(http.MethodPost)
}

type ChatOptions struct {
	Uuid string
}
type ChatRequest struct {
	Prompt      string
	SessionUuid string
	ChatUuid    string
	Regenerate  bool
	Options     ChatOptions
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

func (h *ChatHandler) chatHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "Invalid request body"})
		return
	}
	defer r.Body.Close()
	ctx := r.Context()
	userIDInt32, err := getUserID(ctx)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	answerMsg, err := h.chatService.Chat(req.SessionUuid, req.ChatUuid, req.Prompt, userIDInt32)

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "Success", "text": answerMsg.Content, "chatUuid": answerMsg.Uuid})
}

// OpenAIChatCompletionAPIWithStreamHandler is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
func (h *ChatHandler) OpenAIChatCompletionAPIWithStreamHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
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
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if req.Regenerate {
		regenerateAnswer(h, w, chatSessionUuid, chatUuid)
	} else {
		genAnswer(h, w, chatSessionUuid, chatUuid, newQuestion, userID)
	}

}

// regenerateAnswer is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
// if there is no prompt yet, it will create a new prompt and use it as request
// otherwise,
//
//	it will create a message, use prompt + get latest N message + newQuestion as request
func genAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string, newQuestion string, userID int32) {
	ctx := context.Background()
	chatSession, err := h.chatService.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	fmt.Printf("chatSession: %+v ", chatSession)
	if err != nil {
		http.Error(w,
			eris.Wrap(err, "fail to get session: ").Error(),
			http.StatusInternalServerError,
		)
		return
	}

	existingPrompt := true

	_, err = h.chatService.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existingPrompt = false
		} else {
			http.Error(w, eris.Wrap(err, "fail to get prompt: ").Error(), http.StatusInternalServerError)
		}
	}

	if existingPrompt {
		_, err := h.chatService.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, userID)
		if err != nil {
			http.Error(w,
				eris.Wrap(err, "fail to create message: ").Error(),
				http.StatusInternalServerError,
			)
		}
	} else {
		chatPrompt, err := h.chatService.CreateChatPromptSimple(chatSessionUuid, newQuestion, userID)
		if err != nil {
			http.Error(w,
				eris.Wrap(err, "fail to create prompt: ").Error(),
				http.StatusInternalServerError,
			)
			return
		}
		log.Printf("%+v\n", chatPrompt)
	}

	msgs, err := h.chatService.getAskMessages(chatSession, chatUuid, false)
	if err != nil {
		RespondWithError(w,
			http.StatusInternalServerError,
			eris.Wrap(err, "fail to collect messages: ").Error(),
			err,
		)
		return
	}

	// calc total tokens
	totalTokens := lo.SumBy(msgs, func(msg Message) int32 {
		return msg.TokenCount()
	})

	// check if total tokens exceed limit
	if totalTokens > chatSession.MaxTokens*2/3 {
		RespondWithError(w, http.StatusRequestEntityTooLarge, "error.token_length_exceed_limit",
			map[string]interface{}{
				"max_tokens":   chatSession.MaxTokens,
				"total_tokens": totalTokens,
			})
		return
	}

	chatStreamFn := h.chooseChatStreamFn(chatSession, msgs)

	answerText, answerID, shouldReturn := chatStreamFn(w, chatSession, msgs, chatUuid, false)

	if shouldReturn {
		return
	}
	if !isTest(msgs) {
		h.chatService.logChat(chatSession, msgs, answerText)
	}

	if _, err := h.chatService.CreateChatMessageSimple(ctx, chatSessionUuid, answerID, "assistant", answerText, userID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "failed to create message").Error(), nil)
		return
	}
}

func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string) {
	ctx := context.Background()
	chatSession, err := h.chatService.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, eris.Wrap(err, "fail to get chat session").Error(), err)
		return
	}

	msgs, err := h.chatService.getAskMessages(chatSession, chatUuid, true)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Get chat message error", err)
		return
	}

	// calc total tokens
	totalTokens := lo.SumBy(msgs, func(msg Message) int32 {
		return msg.TokenCount()
	})

	if totalTokens > chatSession.MaxTokens*2/3 {
		RespondWithError(w, http.StatusRequestEntityTooLarge, "error.token_length_exceed_limit",
			map[string]interface{}{
				"max_tokens":   chatSession.MaxTokens,
				"total_tokens": totalTokens,
			})
		return
	}

	// Determine whether the chat is a test or not
	chatStreamFn := h.chooseChatStreamFn(chatSession, msgs)

	answerText, _, shouldReturn := chatStreamFn(w, chatSession, msgs, chatUuid, true)
	if shouldReturn {
		return
	}

	h.chatService.logChat(chatSession, msgs, answerText)

	// Delete previous message and create new one
	if err := h.chatService.UpdateChatMessageContent(ctx, chatUuid, answerText); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to update message: ").Error(), nil)
	}
}

func (h *ChatHandler) chooseChatStreamFn(chat_session sqlc_queries.ChatSession, msgs []Message) func(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	model := chat_session.Model
	isTestChat := isTest(msgs)
	isClaude := strings.HasPrefix(model, "claude")
	isChatGPT := strings.HasPrefix(model, "gpt")
	completionModel := mapset.NewSet[string]()
	completionModel.Add(openai.GPT3TextDavinci003)
	completionModel.Add(openai.GPT3TextDavinci002)
	isCompletion := completionModel.Contains(model)

	chatStreamFn := h.customChatStream
	if isClaude {
		chatStreamFn = h.chatStreamClaude
	} else if isTestChat {
		chatStreamFn = h.chatStreamTest
	} else if isChatGPT {
		chatStreamFn = h.chatStream
	} else if isCompletion {
		chatStreamFn = h.CompletionStream
	}
	return chatStreamFn
}

func isTest(msgs []Message) bool {
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
	chatModel, err := h.chatService.q.ChatModelByName(context.Background(), model)
	log.Printf("%+v", chatModel)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Failed to get model by name").Error(), err)
		return true
	}
	if !chatModel.EnablePerModeRatelimit {
		return false
	}
	ctx := context.Background()
	rate, err := h.chatService.q.RateLimiteByUserAndSessionUUID(ctx,
		sqlc_queries.RateLimiteByUserAndSessionUUIDParams{
			Uuid:   chatSessionUuid,
			UserID: userID,
		})
	log.Printf("%+v", rate)
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "error.fail_to_get_rate_limit", err)
		return true
	}

	// get last model usage in 10min
	usage10Min, err := h.chatService.q.GetChatMessagesCountByUserAndModel(ctx,
		sqlc_queries.GetChatMessagesCountByUserAndModelParams{
			UserID: userID,
			Model:  rate.ChatModelName,
		})

	log.Printf("%+v", usage10Min)

	if int32(usage10Min) > rate.RateLimit {
		RespondWithError(w, http.StatusTooManyRequests, fmt.Sprintf("error.%s_over_limit", rate.ChatModelName), err)
		return true
	}
	return false
}

func (h *ChatHandler) chatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	// check per chat_model limit

	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return "", "", true
	}

	chatModel, err := h.chatService.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
	}

	baseUrl, err := getModelBaseUrl(chatModel.Url)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get base url").Error(), err)
		return "", "", true
	}
	log.Println(baseUrl)

	token := os.Getenv(chatModel.ApiAuthKey)
	config := openai.DefaultConfig(token)
	config.BaseURL = baseUrl
	client := openai.NewClientWithConfig(config)
	// handler proxy
	configOpenAIProxy(config)

	openai_req := NewChatCompletionRequest(chatSession, chat_compeletion_messages)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stream, err := client.CreateChatCompletionStream(ctx, openai_req)

	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return "", "", true
	}
	defer stream.Close()

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
	}

	var answer string
	var answer_id string
	textBuffer := newTextBuffer(int(chatSession.N), "", "")
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
				req_j, _ := json.Marshal(openai_req)
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
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Stream error: %v", err), nil)
			return "", "", true
		}
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Delta.Content
		textBuffer.appendByIndex(textIdx, delta)
		// log.Println(delta)
		if chatSession.Debug {
			log.Printf("%s", delta)
		}
		answer = textBuffer.String("\n")
		if answer_id == "" {
			answer_id = strings.TrimPrefix(response.ID, "chatcmpl-")
		}
		if strings.HasSuffix(delta, "\n") || len(answer) < 200 {
			response.Choices[0].Delta.Content = answer
			data, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}
	return answer, answer_id, false
}

func (h *ChatHandler) CompletionStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	// check per chat_model limit

	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return "", "", true
	}

	chatModel, err := h.chatService.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
	}

	baseUrl, err := getModelBaseUrl(chatModel.Url)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get base url").Error(), err)
		return "", "", true
	}

	token := os.Getenv(chatModel.ApiAuthKey)
	config := openai.DefaultConfig(token)
	config.BaseURL = baseUrl
	client := openai.NewClientWithConfig(config)
	// handler proxy
	configOpenAIProxy(config)

	// latest message contents
	prompt := chat_compeletion_messages[len(chat_compeletion_messages)-1].Content
	N := int(chatSession.N)
	req := openai.CompletionRequest{
		Model:       chatSession.Model,
		MaxTokens:   int(chatSession.MaxTokens),
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
		N:           N,
		Prompt:      prompt,
		Stream:      true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	stream, err := client.CreateCompletionStream(ctx, req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return "", "", true
	}
	defer stream.Close()

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
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
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Stream error: %v", err), nil)
			return "", "", true
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

		if strings.HasSuffix(delta, "\n") || len(answer) < 200 {
			response := constructChatCompletionStreamReponse(answer_id, answer)
			data, _ := json.Marshal(response)
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}
	return answer, answer_id, false
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

func (h *ChatHandler) chatStreamClaude(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	claudeRateLimiteToken <- struct{}{}
	// Release the API token
	defer func() { <-claudeRateLimiteToken }()
	// set the api key
	chatModel, err := h.chatService.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
	}

	// OPENAI_API_KEY

	// create a new strings.Builder
	// iterate through the messages and format them
	// print the user's question
	// convert assistant's response to json format
	prompt := formatClaudePrompt(chat_compeletion_messages)
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
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_make_request", err)
		return "", "", true
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
		Timeout: 2 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return "", "", true
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
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
			return "", "", true
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
			answer_id = uuid.NewString()
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

	return answer, answer_id, false
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

func (h *ChatHandler) customChatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	// set the api key
	chat_model, err := h.chatService.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
	}
	apiKey := os.Getenv(chat_model.ApiAuthKey)
	// set the url
	url := chat_model.Url

	// create a new strings.Builder
	// iterate through the messages and format them
	// print the user's question
	// convert assistant's response to json format
	prompt := formatClaudePrompt(chat_compeletion_messages)
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
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "post to claude api").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
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
			return "", "", true
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
			answer_id = uuid.NewString()
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

	return answer, answer_id, false
}

func (h *ChatHandler) chatStreamTest(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	//message := Message{Role: "assitant", Content:}
	answer_id := chatUuid
	if !regenerate {
		answer_id = uuid.NewString()
	}
	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)

	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
	}
	answer := "Hi, I am a chatbot. I can help you to find the best answer for your question. Please ask me a question."
	resp := constructChatCompletionStreamReponse(answer_id, answer)
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()

	if chatSession.Debug {
		openai_req := NewChatCompletionRequest(chatSession, chat_compeletion_messages)

		req_j, _ := json.Marshal(openai_req)
		answer = answer + "\n" + string(req_j)
		req_as_resp := constructChatCompletionStreamReponse(answer_id, answer)
		data, _ := json.Marshal(req_as_resp)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
	return answer, answer_id, false
}

func NewChatCompletionRequest(chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message) openai.ChatCompletionRequest {
	openai_message := messagesToOpenAIMesages(chat_compeletion_messages)
	openai_req := openai.ChatCompletionRequest{
		Model:       chatSession.Model,
		Messages:    openai_message,
		MaxTokens:   int(chatSession.MaxTokens),
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
		N:           int(chatSession.N),
		Stream:      true,
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

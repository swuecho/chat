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
	"strconv"
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
	service *ChatService
}

func NewChatHandler(sqlc_q *sqlc_queries.Queries) *ChatHandler {
	// create a new ChatService instance
	chatService := NewChatService(sqlc_q)
	return &ChatHandler{
		service: chatService,
	}
}

func (h *ChatHandler) Register(router *mux.Router) {
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
	chatSession, err := h.service.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	fmt.Printf("chatSession: %+v ", chatSession)
	if err != nil {
		http.Error(w,
			eris.Wrap(err, "fail to get session: ").Error(),
			http.StatusInternalServerError,
		)
		return
	}

	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		http.Error(w,
			eris.Wrap(err, "fail to get model: ").Error(),
			http.StatusInternalServerError,
		)
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
		_, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, userID, baseURL, chatSession.SummarizeMode)
		if err != nil {
			http.Error(w,
				eris.Wrap(err, "fail to create message: ").Error(),
				http.StatusInternalServerError,
			)
		}
	} else {
		chatPrompt, err := h.service.CreateChatPromptSimple(chatSessionUuid, newQuestion, userID)
		if err != nil {
			http.Error(w,
				eris.Wrap(err, "fail to create prompt: ").Error(),
				http.StatusInternalServerError,
			)
			return
		}
		log.Printf("%+v\n", chatPrompt)
	}

	msgs, err := h.service.getAskMessages(chatSession, chatUuid, false)
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
		h.service.logChat(chatSession, msgs, answerText)
	}

	if _, err := h.service.CreateChatMessageSimple(ctx, chatSessionUuid, answerID, "assistant", answerText, userID, baseURL, chatSession.SummarizeMode); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "failed to create message").Error(), nil)
		return
	}
}

func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string) {
	ctx := context.Background()
	chatSession, err := h.service.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, eris.Wrap(err, "fail to get chat session").Error(), err)
		return
	}

	msgs, err := h.service.getAskMessages(chatSession, chatUuid, true)
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

	h.service.logChat(chatSession, msgs, answerText)

	// Delete previous message and create new one
	if err := h.service.UpdateChatMessageContent(ctx, chatUuid, answerText); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to update message: ").Error(), nil)
	}
}

func (h *ChatHandler) chooseChatStreamFn(chat_session sqlc_queries.ChatSession, msgs []Message) func(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	model := chat_session.Model
	isTestChat := isTest(msgs)
	isClaude := strings.HasPrefix(model, "claude")
	isClaude3 := false
	if strings.HasPrefix(model, "claude-3") {
		isClaude = false
		isClaude3 = true
	}
	isChatGPT := strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "deepseek") || strings.HasPrefix(model, "yi")
	isOllama := strings.HasPrefix(model, "ollama-")
	isGemini := strings.HasPrefix(model, "gemini")

	completionModel := mapset.NewSet[string]()

	completionModel.Add(openai.GPT3TextDavinci003)
	completionModel.Add(openai.GPT3TextDavinci002)
	isCompletion := completionModel.Contains(model)

	chatStreamFn := h.customChatStream
	if isClaude {
		chatStreamFn = h.chatStreamClaude
	} else if isClaude3 {
		chatStreamFn = h.chatStreamClaude3
	} else if isTestChat {
		chatStreamFn = h.chatStreamTest
	} else if isChatGPT {
		chatStreamFn = h.chatStream
	} else if isOllama {
		chatStreamFn = h.chatOllamStram
	} else if isCompletion {
		chatStreamFn = h.CompletionStream
	} else if isGemini {
		chatStreamFn = h.chatStreamGemini
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
	chatModel, err := h.service.q.ChatModelByName(context.Background(), model)
	log.Printf("%+v", chatModel)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Failed to get model by name").Error(), err)
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
		RespondWithError(w, http.StatusUnauthorized, "error.fail_to_get_rate_limit", err)
		return true
	}

	// get last model usage in 10min
	usage10Min, err := h.service.q.GetChatMessagesCountByUserAndModel(ctx,
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

	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
	}

	config, err := genOpenAIConfig(chatModel)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "gen open ai config").Error(), err)
		return "", "", true
	}

	client := openai.NewClientWithConfig(config)

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
		perWordStreamLimitStr := os.Getenv("PER_WORD_STREAM_LIMIT")

		if perWordStreamLimitStr == "" {
			perWordStreamLimitStr = "200"
		}

		perWordStreamLimit, err := strconv.Atoi(perWordStreamLimitStr)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("per word stream limit error: %v", err), nil)
			return "", "", true
		}

		if strings.HasSuffix(delta, "\n") || len(answer) < perWordStreamLimit {
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

	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
	}

	config, err := genOpenAIConfig(chatModel)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "gen open ai config").Error(), err)
		return "", "", true
	}

	client := openai.NewClientWithConfig(config)
	// latest message contents
	prompt := chat_compeletion_messages[len(chat_compeletion_messages)-1].Content

	//totalInputToken := chat_compeletion_messages[len(chat_compeletion_messages)-1].TokenCount()
	// max - input = max possible output
	//maxOutputToken := int(chatSession.MaxTokens - totalInputToken) - 500

	N := int(chatSession.N)
	req := openai.CompletionRequest{
		Model: chatSession.Model,
		// MaxTokens:   maxOutputToken,
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
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
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
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
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

type Delta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ContentBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta Delta  `json:"delta"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type StartBlock struct {
	Type         string       `json:"type"`
	Index        int          `json:"index"`
	ContentBlock ContentBlock `json:"content_block"`
}

// claude-3-opus-20240229
// claude-3-sonnet-20240229
// claude-3-haiku-20240307
func (h *ChatHandler) chatStreamClaude3(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	claudeRateLimiteToken <- struct{}{}
	log.Printf("%+v", chatSession)
	// Release the API token
	defer func() { <-claudeRateLimiteToken }()
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	log.Printf("%+v", chatModel)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
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
		messages = messagesToOpenAIMesages(chat_compeletion_messages[1:])
	} else {
		// only system message, return and do nothing
		RespondWithError(w, http.StatusInternalServerError, "error.claude_system_message_notice", err)
		return "", "", true
	}
	// create the json data
	jsonData := map[string]interface{}{
		"system":      chat_compeletion_messages[0].Content,
		"model":       chatSession.Model,
		"messages":    messages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"top_p":       chatSession.TopP,
		"stream":      true,
	}
	log.Printf("%+v", jsonData)

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
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
			return "", "", true
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("[DONE]")) {
			// stream.isFinished = true
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			break
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
			log.Println(string(line))
			RespondWithError(w, http.StatusInternalServerError, string(line), nil)
			return "", "", true
		}
		if answer_id == "" {
			answer_id = uuid.NewString()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_start\"")) {
			var response StartBlock
			_ = json.Unmarshal(line, &response)
			answer = response.ContentBlock.Text
			if len(answer) < 200 || len(answer)%2 == 0 {
				data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_delta\"")) {
			var response ContentBlockDelta
			_ = json.Unmarshal(line, &response)
			answer = response.Delta.Text
			if len(answer) < 200 || len(answer)%2 == 0 {
				data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
		}
	}

	return answer, answer_id, false
}

type OllamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Done               bool      `json:"done"`
	Message            Message   `json:"message"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int64     `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int64     `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

func (h *ChatHandler) chatOllamStram(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return "", "", true
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
			return "", "", true
		}
		var streamResp OllamaResponse
		err = json.Unmarshal(line, &streamResp)
		if err != nil {
			return "", "", true
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
			answer_id = uuid.NewString()
		}

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
	chat_model, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
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
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
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
	//totalInputToken := lo.SumBy(chat_compeletion_messages, func(m Message) int32 {
	//	return m.TokenCount()
	//})
	// max - input = max possible output
	//maxOutputToken := int(chatSession.MaxTokens - totalInputToken) - 500 // offset
	openai_req := openai.ChatCompletionRequest{
		Model:    chatSession.Model,
		Messages: openai_message,
		//MaxTokens:   maxOutputToken,
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

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

// curl https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=$API_KEY \
//     -H 'Content-Type: application/json' \
//     -X POST \
//     -d '{
//       "contents": [{
//         "parts":[{
//           "text": "Write a story about a magic backpack."}]}]}' 2> /dev/null



func (h *ChatHandler) chatStreamGemini(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []Message, chatUuid string, regenerate bool) (string, string, bool) {
	payloadBytes, err := GenGemminPayload(chat_compeletion_messages)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error generating gemmi payload").Error(), err)
		return "", "", true
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:streamGenerateContent?alt=sse&key=$GEMINI_API_KEY", chatSession.Model)
	url = os.ExpandEnv(url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		// handle err
		fmt.Println("Error while creating request: ", err)
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "create request to gemini api").Error(), err)
	}
	req.Header.Set("Content-Type", "application/json")

	// create the http client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while sending request: ", err)
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "post to gemini api").Error(), err)
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
	}

	var answer string
	answer_id := chatUuid
	if !regenerate {
		answer_id = uuid.NewString()
	}

	var headerData = []byte("data: ")

	// loop over the response body and print data
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
			} else {
				// 2024/05/20 15:56:12 http: superfluous response.WriteHeader call from github.com/gorilla/handlers.(*responseLogger).WriteHeader (handlers.go:61)
				fmt.Printf("Error while reading response: %+v", err)
				return "", "", true
			}
		}
		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)
		if len(line) > 0 {
			answer = ParseRespLine(line, answer)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}
	return answer, answer_id, false

}


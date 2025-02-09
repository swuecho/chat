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
	"github.com/rotisserie/eris"
	openai "github.com/sashabaranov/go-openai"

	"github.com/gorilla/mux"

	claude "github.com/swuecho/chat_backend/llm/claude"
	"github.com/swuecho/chat_backend/llm/gemini"
	llm_openai "github.com/swuecho/chat_backend/llm/openai"
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
		RespondWithError(w, http.StatusBadRequest, "Invalid request format", err)
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
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
		return
	}

	fmt.Printf("userID: %d", userID)

	chatSnapshot, err := h.service.q.ChatSnapshotByUserIdAndUuid(ctx, sqlc_queries.ChatSnapshotByUserIdAndUuidParams{
		UserID: userID,
		Uuid:   snapshotUuid,
	})
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, eris.Wrap(err, "fail to get chat snapshot").Error(), err)
		return
	}

	fmt.Printf("chatSnapshot: %+v", chatSnapshot)

	var session sqlc_queries.ChatSession
	err = json.Unmarshal(chatSnapshot.Session, &session)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to deserialize chat session").Error(), err)
		return
	}
	var simpleChatMessages []SimpleChatMessage
	err = json.Unmarshal(chatSnapshot.Conversation, &simpleChatMessages)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to deserialize conversation").Error(), err)
		return
	}

	genBotAnswer(h, w, session, simpleChatMessages, newQuestion, userID, req.Stream)

}

// ChatCompletionHandler is an HTTP handler that sends the stream to the client as Server-Sent Events (SSE)
func (h *ChatHandler) ChatCompletionHandler(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		RespondWithError(w, http.StatusBadRequest, "Invalid request format", err)
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
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", err)
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
		if newQuestion != "" {
			_, err := h.service.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, "", chatSession.Model, userID, baseURL, chatSession.SummarizeMode)
			if err != nil {
				http.Error(w,
					eris.Wrap(err, "fail to create message: ").Error(),
					http.StatusInternalServerError,
				)
			}
		} else {
			log.Println("no new question, regenerate answer")
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

	chatStreamFn := h.chooseChatStreamFn(chatSession, msgs)
	LLMAnswer, err := chatStreamFn(w, chatSession, msgs, chatUuid, false, streamOutput)
	if err != nil {
		log.Printf("Error generating answer: %v", err)
		return
	}
	if !isTest(msgs) {
		log.Printf("LLMAnswer: %+v", LLMAnswer)
		h.service.logChat(chatSession, msgs, LLMAnswer.ReasoningContent+LLMAnswer.Answer)
	}

	if _, err := h.service.CreateChatMessageSimple(ctx, chatSessionUuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, chatSession.Model, userID, baseURL, chatSession.SummarizeMode); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "failed to create message").Error(), nil)
		return
	}
}

func genBotAnswer(h *ChatHandler, w http.ResponseWriter, session sqlc_queries.ChatSession, simpleChatMessages []SimpleChatMessage, newQuestion string, userID int32, streamOutput bool) {
	chatModel, err := h.service.q.ChatModelByName(context.Background(), session.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return
	}

	baseURL, _ := getModelBaseUrl(chatModel.Url)

	messages := simpleChatMessagesToMessages(simpleChatMessages)
	messages = append(messages, models.Message{
		Role:    "user",
		Content: newQuestion,
	})
	chatStreamFn := h.chooseChatStreamFn(session, messages)

	LLMAnswer, err := chatStreamFn(w, session, messages, "", false, streamOutput)
	if err != nil {
		log.Printf("Error generating answer: %v", err)
		return
	}

	if !isTest(messages) {
		h.service.logChat(session, messages, LLMAnswer.Answer)
	}

	ctx := context.Background()
	if _, err := h.service.CreateChatMessageSimple(ctx, session.Uuid, LLMAnswer.AnswerId, "assistant", LLMAnswer.Answer, LLMAnswer.ReasoningContent, session.Model, userID, baseURL, session.SummarizeMode); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "failed to create message").Error(), nil)
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
		RespondWithError(w, http.StatusBadRequest, eris.Wrap(err, "fail to get chat session").Error(), err)
		return
	}

	msgs, err := h.service.getAskMessages(chatSession, chatUuid, true)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Get chat message error", err)
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
	chatStreamFn := h.chooseChatStreamFn(chatSession, msgs)

	LLMAnswer, err := chatStreamFn(w, chatSession, msgs, chatUuid, true, stream)
	if err != nil {
		log.Printf("Error regenerating answer: %v", err)
		return
	}

	h.service.logChat(chatSession, msgs, LLMAnswer.Answer)

	// Delete previous message and create new one
	if err := h.service.UpdateChatMessageContent(ctx, chatUuid, LLMAnswer.Answer); err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to update message: ").Error(), nil)
	}
}

func (h *ChatHandler) chooseChatStreamFn(chat_session sqlc_queries.ChatSession, msgs []models.Message) func(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
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

	chatStreamFn := h.chatStream
	if isClaude {
		chatStreamFn = h.chatStreamClaude
	} else if isClaude3 {
		chatStreamFn = h.chatStreamClaude3
	} else if isTestChat {
		chatStreamFn = h.chatStreamTest
	} else if isOllama {
		chatStreamFn = h.chatOllamStream
	} else if isCompletion {
		chatStreamFn = h.CompletionStream
	} else if isGemini {
		chatStreamFn = h.chatStreamGemini
	} else if isCustom {
		chatStreamFn = h.customChatStream
	}
	return chatStreamFn
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

	if err != nil {
		// no rows
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(w, http.StatusInternalServerError, "error.fail_to_get_rate_limit", err)
			return true
		}
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_get_rate_limit", err)
		return true
	}

	log.Printf("%+v", usage10Min)

	if int32(usage10Min) > rate.RateLimit {
		RespondWithError(w, http.StatusTooManyRequests, fmt.Sprintf("error.%s_over_limit", rate.ChatModelName), err)
		return true
	}
	return false
}

func getPerWordStreamLimit() int {
	perWordStreamLimitStr := os.Getenv("PER_WORD_STREAM_LIMIT")

	if perWordStreamLimitStr == "" {
		perWordStreamLimitStr = "200"
	}

	perWordStreamLimit, err := strconv.Atoi(perWordStreamLimitStr)
	if err != nil {
		log.Printf("get per word stream limit: %v", eris.Wrap(err, "get per word stream limit").Error())
		return 200
	}

	return perWordStreamLimit
}

func (h *ChatHandler) chatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	// check per chat_model limit

	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return nil, eris.New("exceed per mode rate limit")
	}

	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return nil, err
	}

	config, err := genOpenAIConfig(chatModel)
	log.Printf("%+v", config.String()) 
	// print all config details
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "gen open ai config").Error(), err)
		return nil, err
	}

	client := openai.NewClientWithConfig(config)

	chatFiles, err := h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting chat files").Error(), err)
		return nil, err
	}

	openai_req := NewChatCompletionRequest(chatSession, chat_compeletion_messages, chatFiles, streamOutput)
	if len(openai_req.Messages) <= 1 {
		err := eris.New("system message notice")
		RespondWithError(w, http.StatusInternalServerError, "error.system_message_notice", err)
		return nil, err
	}
	log.Printf("%+v", openai_req)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !streamOutput {
		completion, err := client.CreateChatCompletion(ctx, openai_req)
		if err != nil {
			log.Printf("fail to do request: %+v", err)
			RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
			return nil, err
		}
		log.Printf("completion: %+v", completion)
		data, _ := json.Marshal(completion)
		fmt.Fprint(w, string(data))
		return &models.LLMAnswer{Answer: completion.Choices[0].Message.Content, AnswerId: completion.ID}, nil
	}
	log.Print("before request")
	stream, err := client.CreateChatCompletionStream(ctx, openai_req)

	if err != nil {
		log.Printf("fail to do request: %+v", err)
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}
	defer stream.Close()

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return nil, eris.New("Streaming unsupported!")
	}

	var answer string
	var answer_id string
	var hasReason bool
	bufferLen := int(chatSession.N)
	if bufferLen == 0 {
		log.Println("chatSession.N is 0")
		bufferLen += 1
	}
	textBuffer := newTextBuffer(bufferLen, "", "")
	reasonBuffer := newTextBuffer(bufferLen, "<think>\n\n", "\n\n</think>\n\n")
	if regenerate {
		answer_id = chatUuid
	}
	for {
		rawLine, err := stream.RecvRaw()
		if err != nil {
			log.Printf("stream error: %+v", err)
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
				// no reason in the answer (so do not disrupt the context)
				llmAnswer := models.LLMAnswer{Answer: textBuffer.String("\n"), AnswerId: answer_id}
				if hasReason {
					llmAnswer.ReasoningContent = reasonBuffer.String("\n")
				}
				return &llmAnswer, nil
			} else {
				log.Printf("%v", err)
				RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Stream error: %v", err), nil)
				return nil, err
			}
		}
		response := llm_openai.ChatCompletionStreamResponse{}
		err = json.Unmarshal(rawLine, &response)
		if err != nil {
			log.Printf("Could not unmarshal response: %v\n", err)
			continue
		}
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Delta
		textBuffer.appendByIndex(textIdx, delta.Content)
		if len(delta.ReasoningContent) > 0 {
			hasReason = true
			reasonBuffer.appendByIndex(textIdx, delta.ReasoningContent)
		}

		if chatSession.Debug {
			log.Printf("%+v", response)
			log.Printf("%+v", delta)
		}
		if hasReason {
			answer = reasonBuffer.String("\n") + textBuffer.String("\n")
		} else {
			answer = textBuffer.String("\n")
		}
		if answer_id == "" {
			answer_id = strings.TrimPrefix(response.ID, "chatcmpl-")
		}
		perWordStreamLimit := getPerWordStreamLimit()

		if strings.HasSuffix(answer, "\n") || len(answer) < perWordStreamLimit {
			if len(answer) == 0 {
				log.Printf("%s", "no content in answer")
			} else {
				constructedResponse := constructChatCompletionStreamReponse(answer_id, answer)
				data, _ := json.Marshal(constructedResponse)
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
		}
	}
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
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
		return nil, err
	}

	config, err := genOpenAIConfig(chatModel)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "gen open ai config").Error(), err)
		return nil, err
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	stream, err := client.CreateCompletionStream(ctx, req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}
	defer stream.Close()

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Stream error: %v", err), nil)
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

func (h *ChatHandler) chatStreamClaude(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	claudeRateLimiteToken <- struct{}{}
	// Release the API token
	defer func() { <-claudeRateLimiteToken }()
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_make_request", err)
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
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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

// claude-3-opus-20240229
// claude-3-sonnet-20240229
// claude-3-haiku-20240307
func (h *ChatHandler) chatStreamClaude3(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
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
		return nil, err
	}
	chatFiles, err := h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting chat files").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, "error.system_message_notice", err)
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
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_make_request", err)
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
	} else {
		// set the streaming flag
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
	}

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%+v", err)
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}

	if !stream {
		// Unmarshal directly from resp.Body
		var message claude.Response
		if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
			RespondWithError(w, http.StatusInternalServerError, "error.fail_to_unmarshal_response", err)
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

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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
			RespondWithError(w, http.StatusInternalServerError, string(line), nil)
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

func (h *ChatHandler) chatOllamStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_make_request", err)
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
		RespondWithError(w, http.StatusInternalServerError, "error.fail_to_do_request", err)
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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

func (h *ChatHandler) customChatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	// set the api key
	chat_model, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "get chat model").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "post to claude api").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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

func (h *ChatHandler) chatStreamTest(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	//message := Message{Role: "assitant", Content:}
	chatFiles, err := h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting chat files").Error(), err)
		return nil, err
	}

	answer_id := chatUuid
	if !regenerate {
		answer_id = NewUUID()
	}
	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)

	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error getting chat files").Error(), err)
		return nil, err
	}
	payloadBytes, err := gemini.GenGemminPayload(chat_compeletion_messages, chatFiles)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "Error generating gemmi payload").Error(), err)
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
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "create request to gemini api").Error(), err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while sending request: ", err)
		RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "post to gemini api").Error(), err)
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
			RespondWithError(w, http.StatusInternalServerError, "error.fail_to_read_response", err)
			return nil, err
		}
		// body to GeminiResponse
		var geminiResp gemini.ResponseBody
		err = json.Unmarshal(body, &geminiResp)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "error.fail_to_unmarshal_response", err)
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
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
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

package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	uuid "github.com/iris-contrib/go.uuid"
	"github.com/rotisserie/eris"
	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chatgpt_backend/sqlc_queries"

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

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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
	userIDStr, ok := ctx.Value(userContextKey).(string)
	if !ok {
		RespondWithError(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	userIDInt, _ := strconv.Atoi(userIDStr)
	answerMsg, err := h.chatService.Chat(req.SessionUuid, req.ChatUuid, req.Prompt, int32(userIDInt))

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
		fmt.Fprintf(w, "Invalid request body: %v", err)
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
		// TODO: get context
		// get the 10 chatMessage right before the current chatMessage, order by created_at
		// Send the response as JSON
		// Update the chatMessage content with chatUuid with new answer
		regenerateAnswer(h, w, chatSessionUuid, chatUuid)
	} else {
		// insert ChatMessage into database
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

	if isTest(msgs) {
		answerText, answerID, shouldReturn := chatStreamTest(w, chatSession, msgs)
		if shouldReturn {
			return
		}
		_, err := h.chatService.CreateChatMessageSimple(ctx, chatSessionUuid, answerID, "assistant", answerText, userID)
		if err != nil {
			http.Error(w,
				eris.Wrap(err, "fail to create message: ").Error(),
				http.StatusInternalServerError,
			)
		}
	} else {
		answerText, answerID, shouldReturn := chatStream(w, chatSession, msgs)
		if shouldReturn {
			return
		}
		_, err = h.chatService.CreateChatMessageSimple(ctx, chatSessionUuid, answerID, "assistant", answerText, userID)
		if err != nil {
			RespondWithError(w,
				http.StatusInternalServerError,
				eris.Wrap(err, "fail to create message: ").Error(),
				nil,
			)
			return
		}
	}
}

func regenerateAnswer(h *ChatHandler, w http.ResponseWriter, chatSessionUuid string, chatUuid string) {
	ctx := context.Background()
	chat_session, err := h.chatService.q.GetChatSessionByUUID(ctx, chatSessionUuid)
	if err != nil {
		http.Error(w, "Error: '"+err.Error()+"'", http.StatusBadRequest)
		return
	}

	chatCompletionMessages, err := h.chatService.getAskMessages(chat_session, chatUuid, true)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Get chat message error", err)
		return
	}
	// Determine whether the chat is a test or not
	isTestChat := isTest(chatCompletionMessages)

	if isTestChat {
		answerText, answerID, shouldReturn := chatStreamTest(w, chat_session, chatCompletionMessages)
		if shouldReturn {
			return
		}
		// Delete previous message and create new one
		err := h.chatService.DeleteAndCreateChatMessage(chatSessionUuid, chatUuid, chat_session.UserID, answerID, answerText)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to create message: ").Error(), nil)
		}
	} else {
		answerText, answerID, shouldReturn := chatStream(w, chat_session, chatCompletionMessages)
		if shouldReturn {
			return
		}

		// Delete previous message and create new one
		err := h.chatService.DeleteAndCreateChatMessage(chatSessionUuid, chatUuid, chat_session.UserID, answerID, answerText)

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, eris.Wrap(err, "fail to create message: ").Error(), nil)
		}
	}
}

func isTest(msgs []openai.ChatCompletionMessage) bool {
	lastMsgs := msgs[len(msgs)-1]
	promptMsg := msgs[0]
	return promptMsg.Content == "test_demo_bestqa" || lastMsgs.Content == "test_demo_bestqa"
}

func chatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []openai.ChatCompletionMessage) (string, string, bool) {
	client := openai.NewClient(appConfig.OPENAI.API_KEY)

	openai_req := NewChatCompletionRequest(chatSession, chat_compeletion_messages)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	stream, err := client.CreateChatCompletionStream(ctx, openai_req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("CompletionStream error: %v", err), nil)
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
		delta := response.Choices[0].Delta.Content
		// log.Println(delta)
		if chatSession.Debug {
			log.Printf("%s", delta)
		}
		answer += delta
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

func chatStreamTest(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []openai.ChatCompletionMessage) (string, string, bool) {
	//message := Message{Role: "assitant", Content:}
	uuid, _ := uuid.NewV4()
	answer_id := uuid.String()
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

func NewChatCompletionRequest(chatSession sqlc_queries.ChatSession, chat_compeletion_messages []openai.ChatCompletionMessage) openai.ChatCompletionRequest {
	openai_req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    chat_compeletion_messages,
		MaxTokens:   int(chatSession.MaxTokens),
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
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

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

	uuid "github.com/iris-contrib/go.uuid"
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
	ctx := r.Context()
	userID, err := getUserID(ctx)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	log.Printf("chatSessionUuid: %s", chatSessionUuid)
	log.Printf("chatUuid: %s", chatUuid)
	log.Printf("newQuestion: %s", newQuestion)
	if req.Regenerate {
		// TODO: get context
		// get the 10 chatMessage right before the current chatMessage, order by created_at
		chat_session, err := h.chatService.q.GetChatSessionByUUID(ctx, chatSessionUuid)
		if err != nil {
			http.Error(w, "Error: '"+err.Error()+"'", http.StatusBadRequest)
			return
		}

		// Send the response as JSON
		chatCompletionMessages, err := h.chatService.getAskMessages(chat_session, chatUuid, true)
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Get chat message error", err)
			return
		}

		// Set up SSE headers
		if chat_session.Debug {
			log.Printf("%+v\n", chatCompletionMessages)
		}
		answerText, _, shouldReturn := chat_stream(w, chat_session, chatCompletionMessages)
		if shouldReturn {
			return
		}
		// Update the chatMessage content with chatUuid with new answer
		err = h.chatService.UpdateChatMessageContent(ctx, chatUuid, answerText)

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Update chat message error", err)
			return
		}
		return
	}

	chatSession, err := h.chatService.q.GetChatSessionByUUID(ctx, chatSessionUuid)

	if err != nil {
		http.Error(w, fmt.Errorf("fail to create or update session: %w", err).Error(), http.StatusInternalServerError)
	}

	existingPrompt := true

	_, err = h.chatService.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existingPrompt = false
		} else {
			http.Error(w, fmt.Errorf("fail to get prompt: %w", err).Error(), http.StatusInternalServerError)
		}
	}

	if existingPrompt {
		_, err := h.chatService.CreateChatMessageSimple(ctx, chatSession.Uuid, chatUuid, "user", newQuestion, userID)

		if err != nil {
			http.Error(w, fmt.Errorf("fail to create message: %w", err).Error(), http.StatusInternalServerError)
		}
	} else {
		chatPrompt, err := h.chatService.CreateChatPromptSimple(chatSessionUuid, newQuestion, userID)
		if err != nil {
			http.Error(w, fmt.Errorf("fail to create prompt: %w", err).Error(), http.StatusInternalServerError)
		}
		log.Printf("%+v\n", chatPrompt)
	}

	msgs, err := h.chatService.getAskMessages(chatSession, chatUuid, false)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("fail to collect messages: %w", err).Error(), err)
		return
	}

	if existingPrompt {
		msgs = append(msgs, NewUserMessage(newQuestion))
	}
	if len(msgs) == 0 {
		http.Error(w, "No messages found", http.StatusNotFound)
	}
	if msgs[0].Content == "test_demo_bestqa" || msgs[len(msgs)-1].Content == "test_demo_bestqa" {
		answerText, answerID, shouldReturn := test_replay(w)
		if shouldReturn {
			return
		}
		m, err := h.chatService.CreateChatMessageSimple(ctx, chatSessionUuid, answerID, "assistant", answerText, userID)

		log.Println(m)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, fmt.Errorf("fail to create message: %w", err).Error(), http.StatusInternalServerError)
		}
	} else {

		// Set up SSE headers
		answerText, answerID, shouldReturn := chat_stream(w, chatSession, msgs)
		if shouldReturn {
			return
		}
		// insert ChatMessage into database
		_, err := h.chatService.CreateChatMessageSimple(ctx, chatSessionUuid, answerID, "assistant", answerText, userID)

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("fail to create message: %w", err).Error(), nil)
		}
	}

}

func chat_stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []openai.ChatCompletionMessage) (string, string, bool) {
	apiKey := appConfig.OPENAI.API_KEY

	client := openai.NewClient(apiKey)

	openai_req := openai.ChatCompletionRequest{
		Model:       openai.GPT3Dot5Turbo,
		Messages:    chat_compeletion_messages,
		MaxTokens:   int(chatSession.MaxTokens),
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
		// PresencePenalty:  presencePenalty,
		// FrequencyPenalty: frequencyPenalty,
		// N:                n,
		Stream: true,
	}
	ctx := context.Background()
	stream, err := client.CreateChatCompletionStream(ctx, openai_req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("CompletionStream error: %v", err), nil)
		return "", "", true
	}
	defer stream.Close()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

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

func test_replay(w http.ResponseWriter) (string, string, bool) {
	//message := Message{Role: "assitant", Content:}
	uuid, _ := uuid.NewV4()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)

	if !ok {
		RespondWithError(w, http.StatusInternalServerError, "Streaming unsupported!", nil)
		return "", "", true
	}
	answer := "Hi, I am a chatbot. I can help you to find the best answer for your question. Please ask me a question."
	answer_id := uuid.String()
	resp := constructChatCompletionStreamReponse(answer_id, answer)
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()
	return answer, answer_id, false
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

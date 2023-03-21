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
	"github.com/samber/lo"
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
	Message      Message     `json:"message"`
	FinishReason interface{} `json:"finish_reason"`
	Index        int         `json:"index"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenaiChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

func NewUserMessage(content string) Message {
	return Message{Role: "user", Content: content}
}

func (h *ChatHandler) chatHandler(w http.ResponseWriter, r *http.Request) {
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
	log.Printf("Received prompt: %s\n", newQuestion)
	ctx := r.Context()
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", http.StatusBadRequest)
		return
	}
	answer_msg, err := h.chatService.Chat(chatSessionUuid, chatUuid, newQuestion, int32(userIDInt))
	if err != nil {
		fmt.Fprintf(w, "problem in chat: %v", err)
		return
	}

	// Send the response as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "Success", "text": answer_msg.Content, "chatUuid": answer_msg.Uuid})
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
	userIDStr := ctx.Value(userContextKey).(string)
	userIDInt, err := strconv.Atoi(userIDStr)
	userID := int32(userIDInt)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Error: '"+userIDStr+"' is not a valid user ID. Please enter a valid user ID.", nil)
		return
	}

	log.Printf("chatSessionUuid: %s", chatSessionUuid)
	log.Printf("chatUuid: %s", chatUuid)
	log.Printf("newQuestion: %s", newQuestion)
	if req.Regenerate {
		// TODO: get context
		// get the 10 chatMessage right before the current chatMessage, order by created_at
		chat_prompts, err := h.chatService.q.GetChatPromptsBySessionUUID(ctx, chatSessionUuid)
		if err != nil {
			http.Error(w, "Error: '"+err.Error()+"'", http.StatusBadRequest)
			return
		}
		chat_session, err := h.chatService.q.GetChatSessionByUUID(ctx, chatSessionUuid)
		if err != nil {
			http.Error(w, "Error: '"+err.Error()+"'", http.StatusBadRequest)
			return
		}
		lastN := chat_session.MaxLength
		if chat_session.MaxLength == 0 {
			lastN = 10
		}
		msgs, err := h.chatService.q.GetLastNChatMessages(ctx,
			sqlc_queries.GetLastNChatMessagesParams{
				Uuid:  chatUuid,
				Limit: lastN,
			})
		if err != nil {
			http.Error(w, "Error: '"+err.Error()+"'", http.StatusBadRequest)
			return
		}
		ChatCompletionMessagesFromPrompt := lo.Map(chat_prompts, func(m sqlc_queries.ChatPrompt, _ int) openai.ChatCompletionMessage {
			return openai.ChatCompletionMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		})
		chatCompletionMessages := lo.FilterMap(msgs, func(m sqlc_queries.ChatMessage, _ int) (openai.ChatCompletionMessage, bool) {
			if m.Role == "user" {
				return openai.ChatCompletionMessage{
					Role:    m.Role,
					Content: m.Content,
				}, true
			} else {
				return openai.ChatCompletionMessage{}, false
			}
		})
		// Send the response as JSON
		chatCompletionMessages = append(ChatCompletionMessagesFromPrompt, chatCompletionMessages...)

		// Set up SSE headers
		answerText, _, shouldReturn := chat_stream(ctx, chat_session, chatCompletionMessages, w)
		if shouldReturn {
			return
		}
		// Update the chatMessage content with chatUuid with new answer
		err = h.chatService.q.UpdateChatMessageContent(ctx,
			sqlc_queries.UpdateChatMessageContentParams{
				Uuid:    chatUuid,
				Content: answerText,
			})

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, "Update chat message error", err)
			return
		}
		return
	}

	////

	// no session exists
	//
	// if no session chat_created, create new chat_session with $uuid
	// create a new prompt with topic = $uuid, role = "system", content= req.Prompt

	// if session avaiable,
	// GetChatPromptBySessionID and create Message from Prompt
	// GetLatestMessagesBySessionID  and create Messsage(s) from messages

	// Check if the chat session exists

	// no session exists
	// create session and prompt

	chatSession, err := h.chatService.q.CreateOrUpdateChatSessionByUUID(ctx, sqlc_queries.CreateOrUpdateChatSessionByUUIDParams{
		Uuid:   chatSessionUuid,
		UserID: userID,
		Topic:  firstN(newQuestion, 30),
	})

	if err != nil {
		http.Error(w, fmt.Errorf("fail to create or update session: %w", err).Error(), http.StatusInternalServerError)
	}

	log.Println(chatSession)

	existingPrompt := true

	log.Println(chatSessionUuid)
	_, err = h.chatService.q.GetOneChatPromptBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			existingPrompt = false
		} else {
			http.Error(w, fmt.Errorf("fail to get prompt: %w", err).Error(), http.StatusInternalServerError)
		}
	}

	if existingPrompt {
		_, err := h.chatService.q.CreateChatMessage(ctx,
			sqlc_queries.CreateChatMessageParams{
				ChatSessionUuid: chatSession.Uuid,
				Uuid:            chatUuid,
				Role:            "user",
				Content:         newQuestion,
				Raw:             json.RawMessage([]byte("{}")),
				UserID:          userID,
				CreatedBy:       userID,
				UpdatedBy:       userID,
			})

		if err != nil {
			http.Error(w, fmt.Errorf("fail to create message: %w", err).Error(), http.StatusInternalServerError)
		}
	} else {
		uuidVar, _ := uuid.NewV4()
		chatPrompt, err := h.chatService.q.CreateChatPrompt(ctx,
			sqlc_queries.CreateChatPromptParams{
				Uuid:            uuidVar.String(),
				ChatSessionUuid: chatSessionUuid,
				Role:            "system",
				Content:         newQuestion,
				UserID:          userID,
				CreatedBy:       userID,
				UpdatedBy:       userID,
			})
		if err != nil {
			http.Error(w, fmt.Errorf("fail to create prompt: %w", err).Error(), http.StatusInternalServerError)
		}
		log.Printf("%+v\n", chatPrompt)
	}

	chat_prompts, err := h.chatService.q.GetChatPromptsBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		http.Error(w, fmt.Errorf("fail to get prompt: %w", err).Error(), http.StatusInternalServerError)
	}

	chat_massages, err := h.chatService.q.GetLatestMessagesBySessionUUID(ctx,
		sqlc_queries.GetLatestMessagesBySessionUUIDParams{ChatSessionUuid: chatSession.Uuid, Limit: 5})

	if err != nil {
		http.Error(w, fmt.Errorf("fail to get messages: %w", err).Error(), http.StatusInternalServerError)
	}
	chat_prompt_msgs := lo.Map(chat_prompts, func(m sqlc_queries.ChatPrompt, _ int) Message {
		return Message{Role: m.Role, Content: m.Content}
	})
	chat_message_msgs := lo.Map(chat_massages, func(m sqlc_queries.ChatMessage, _ int) Message {
		return Message{Role: m.Role, Content: m.Content}
	})
	msgs := append(chat_prompt_msgs, chat_message_msgs...)

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
		// insert ChatMessage into database
		chatMessageParams := sqlc_queries.CreateChatMessageParams{
			ChatSessionUuid: chatSessionUuid,
			Uuid:            answerID,
			Role:            "assistant",
			Content:         answerText,
			UserID:          int32(userIDInt),
			CreatedBy:       int32(userIDInt),
			UpdatedBy:       int32(userIDInt),
			Raw:             json.RawMessage([]byte("{}")),
		}
		log.Println(chatMessageParams)

		m, err := h.chatService.q.CreateChatMessage(ctx, chatMessageParams)

		log.Println(m)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, fmt.Errorf("fail to create message: %w", err).Error(), http.StatusInternalServerError)
		}
	} else {

		// Send the response as JSON
		chatCompletionMessages := lo.Map(msgs, func(m Message, _ int) openai.ChatCompletionMessage {
			return openai.ChatCompletionMessage{
				Role:    m.Role,
				Content: m.Content,
			}
		})

		// Set up SSE headers
		answerText, answerID, shouldReturn := chat_stream(ctx, chatSession, chatCompletionMessages, w)
		if shouldReturn {
			return
		}
		// insert ChatMessage into database
		chatMessage := sqlc_queries.CreateChatMessageParams{
			Uuid:            answerID,
			ChatSessionUuid: chatSessionUuid,
			Role:            "assistant",
			UserID:          int32(userIDInt),
			Content:         answerText,
			CreatedBy:       int32(userIDInt),
			UpdatedBy:       int32(userIDInt),
			Raw:             json.RawMessage([]byte("{}")),
		}

		_, err := h.chatService.q.CreateChatMessage(ctx, chatMessage)

		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Errorf("fail to create message: %w", err).Error(), nil)
		}
	}

}

func chat_stream(ctx context.Context, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []openai.ChatCompletionMessage, w http.ResponseWriter) (string, string, bool) {
	apiKey := OPENAI_API_KEY

	client := openai.NewClient(apiKey)
	// temperature := float32(0.8)
	// topP := float32(1)
	// presencePenalty := float32(0)
	// frequencyPenalty := float32(0)
	// n := 1

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
			break
		}
		if err != nil {
			RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Stream error: %v", err), nil)
			return "", "", true
		}
		delta := response.Choices[0].Delta.Content
		// log.Println(delta)
		fmt.Printf("%q\n", delta)
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

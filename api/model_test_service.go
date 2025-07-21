package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// TestChatModel implements ChatModel interface for testing purposes
type TestChatModel struct {
	h *ChatHandler
}

// Stream implements the ChatModel interface for test scenarios
func (m *TestChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.chatStreamTest(w, chatSession, chat_completion_messages, chatUuid, regenerate)
}

// chatStreamTest handles test chat streaming with mock responses
func (m *TestChatModel) chatStreamTest(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to get chat files", err.Error()))
		return nil, err
	}

	answer_id := chatUuid
	if !regenerate {
		answer_id = NewUUID()
	}
	
	flusher, err := setupSSEStream(w)
	if err != nil {
		RespondWithAPIError(w, createAPIError(APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		}, "", err.Error()))
		return nil, err
	}
	
	answer := "Hi, I am a chatbot. I can help you to find the best answer for your question. Please ask me a question."
	resp := constructChatCompletionStreamResponse(answer_id, answer)
	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()

	if chatSession.Debug {
		openai_req := NewChatCompletionRequest(chatSession, chat_completion_messages, chatFiles, false)
		req_j, _ := json.Marshal(openai_req)
		answer = answer + "\n" + string(req_j)
		req_as_resp := constructChatCompletionStreamResponse(answer_id, answer)
		data, _ := json.Marshal(req_as_resp)
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
	
	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}


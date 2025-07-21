package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	claude "github.com/swuecho/chat_backend/llm/claude"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// CustomModelResponse represents the response structure for custom models
type CustomModelResponse struct {
	Completion string `json:"completion"`
	Stop       string `json:"stop"`
	StopReason string `json:"stop_reason"`
	Truncated  bool   `json:"truncated"`
	LogID      string `json:"log_id"`
	Model      string `json:"model"`
	Exception  any    `json:"exception"`
}

// CustomChatModel implements ChatModel interface for custom model providers
type CustomChatModel struct {
	h *ChatHandler
}

// Stream implements the ChatModel interface for custom model scenarios
func (m *CustomChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.customChatStream(w, chatSession, chat_completion_messages, chatUuid, regenerate)
}

// customChatStream handles streaming for custom model providers
func (m *CustomChatModel) customChatStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// Get chat model configuration
	chat_model, err := m.h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}

	// Get API key from environment
	apiKey := os.Getenv(chat_model.ApiAuthKey)
	url := chat_model.Url

	// Format messages for the custom model
	prompt := claude.FormatClaudePrompt(chat_completion_messages)
	
	// Create request payload
	jsonData := map[string]any{
		"prompt":               prompt,
		"model":                chatSession.Model,
		"max_tokens_to_sample": chatSession.MaxTokens,
		"temperature":          chatSession.Temperature,
		"stop_sequences":       []string{"\n\nHuman:"},
		"stream":               true,
	}

	// Marshal request data
	jsonValue, _ := json.Marshal(jsonData)
	
	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrChatRequestFailed, "Failed to create custom model request", err.Error()))
		return nil, err
	}

	// Set authentication header if configured
	authHeaderName := chat_model.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrChatRequestFailed, "Failed to send custom model request", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	// Setup streaming response
	ioreader := bufio.NewReader(resp.Body)
	flusher, err := setupSSEStream(w)
	if err != nil {
		RespondWithAPIError(w, createAPIError(APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		}, "", err.Error()))
		return nil, err
	}

	var answer string
	var answer_id string
	var lastFlushLength int

	if regenerate {
		answer_id = chatUuid
	}

	headerData := []byte("data: ")
	count := 0
	
	// Process streaming response
	for {
		count++
		// Prevent infinite loop
		if count > 10000 {
			break
		}
		
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break
			}
			return nil, err
		}
		
		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("[DONE]")) {
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

		// Determine when to flush the response
		shouldFlush := strings.Contains(answer, "\n") ||
			len(answer) < 200 ||
			(len(answer)-lastFlushLength) >= 500

		if shouldFlush {
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			lastFlushLength = len(answer)
		}
	}

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}
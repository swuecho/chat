package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	// Get request context for cancellation support
	ctx := m.h.GetRequestContext()
	return m.customChatStream(ctx, w, chatSession, chat_completion_messages, chatUuid, regenerate)
}

// customChatStream handles streaming for custom model providers
func (m *CustomChatModel) customChatStream(ctx context.Context, w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// Get chat model configuration
	chat_model, err := GetChatModel(m.h.service.q, chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrResourceNotFound(""), "chat model: "+chatSession.Model, ""))
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

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
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
	SetStreamingHeaders(req)

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
	answer_id = GenerateAnswerID(chatUuid, regenerate)

	headerData := []byte("data: ")
	count := 0

	// Process streaming response
	for {
		// Check if client disconnected or context was cancelled
		select {
		case <-ctx.Done():
			log.Printf("Custom model stream cancelled by client: %v", ctx.Err())
			// Return current accumulated content when cancelled
			return &models.LLMAnswer{Answer: answer, AnswerId: answer_id}, nil
		default:
		}

		count++
		// Prevent infinite loop
		if count > MaxStreamingLoopIterations {
			break
		}

		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println(ErrorEndOfStream)
				break
			}
			return nil, err
		}

		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("[DONE]")) {
			fmt.Println(ErrorDoneBreak)
			err := FlushResponse(w, flusher, StreamingResponse{
				AnswerID: answer_id,
				Content:  answer,
				IsFinal:  true,
			})
			if err != nil {
				log.Printf("Failed to flush final response: %v", err)
			}
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
			len(answer) < SmallAnswerThreshold ||
			(len(answer)-lastFlushLength) >= FlushCharacterThreshold

		if shouldFlush {
			err := FlushResponse(w, flusher, StreamingResponse{
				AnswerID: answer_id,
				Content:  answer,
				IsFinal:  false,
			})
			if err != nil {
				log.Printf("Failed to flush response: %v", err)
			}
			lastFlushLength = len(answer)
		}
	}

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

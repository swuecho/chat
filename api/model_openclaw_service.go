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

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// OpenClawChatModel implements ChatModel interface for OpenClaw gateway
// This allows the chat web UI to communicate with OpenClaw agents similar to
// how Telegram or Discord integrations work - messages go to OpenClaw,
// and responses stream back to the web UI.
type OpenClawChatModel struct {
	h *ChatHandler
}

// Stream implements the ChatModel interface for OpenClaw
func (m *OpenClawChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	ctx := m.h.GetRequestContext()
	return m.openclawChatStream(ctx, w, chatSession, chat_completion_messages, chatUuid, regenerate)
}

// openclawChatStream handles streaming for OpenClaw gateway
// Uses OpenClaw's OpenAI-compatible API endpoint for seamless integration
func (m *OpenClawChatModel) openclawChatStream(ctx context.Context, w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// Get OpenClaw gateway URL from environment or use default
	gatewayURL := os.Getenv("OPENCLAW_GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}
	apiKey := os.Getenv("OPENCLAW_API_KEY")

	// Convert messages to OpenAI format
	openaiMessages := make([]map[string]string, len(chat_completion_messages))
	for i, msg := range chat_completion_messages {
		openaiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// Build request to OpenClaw's OpenAI-compatible endpoint
	url := fmt.Sprintf("%s/v1/chat/completions", gatewayURL)
	
	jsonData := map[string]any{
		"model":    chatSession.Model,
		"messages": openaiMessages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"stream":      true,
		// Pass session context for continuity
		"session": chatSession.Uuid,
	}

	jsonValue, _ := json.Marshal(jsonData)

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrChatRequestFailed, "Failed to create OpenClaw request", err.Error()))
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrChatRequestFailed, "Failed to send OpenClaw request", err.Error()))
		return nil, err
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("OpenClaw API error: status=%d, body=%s", resp.StatusCode, string(body))
		RespondWithAPIError(w, createAPIError(ErrChatRequestFailed, fmt.Sprintf("OpenClaw API returned status %d", resp.StatusCode), string(body)))
		return nil, fmt.Errorf("openclaw API error: status %d", resp.StatusCode)
	}

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
			log.Printf("OpenClaw stream cancelled by client: %v", ctx.Err())
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
			break
		}

		if answer_id == "" {
			answer_id = NewUUID()
		}

		// Parse OpenAI-style streaming response
		var response struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int    `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(line, &response); err != nil {
			log.Printf("Failed to unmarshal OpenClaw response: %v, line: %s", err, string(line))
			continue
		}

		// Extract content from delta
		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			answer += content
		}

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

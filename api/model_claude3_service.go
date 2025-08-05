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
	"time"

	openai "github.com/sashabaranov/go-openai"
	claude "github.com/swuecho/chat_backend/llm/claude"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ClaudeResponse represents the response structure from Claude API
type ClaudeResponse struct {
	Completion string `json:"completion"`
	Stop       string `json:"stop"`
	StopReason string `json:"stop_reason"`
	Truncated  bool   `json:"truncated"`
	LogID      string `json:"log_id"`
	Model      string `json:"model"`
	Exception  any    `json:"exception"`
}

// Claude3 ChatModel implementation
type Claude3ChatModel struct {
	h *ChatHandler
}

func (m *Claude3ChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	// Get chat model configuration
	chatModel, err := GetChatModel(m.h.service.q, chatSession.Model)
	if err != nil {
		return nil, err
	}

	// Get chat files if any
	chatFiles, err := GetChatFiles(m.h.chatfileService.q, chatSession.Uuid)
	if err != nil {
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
		return nil, ErrSystemMessageError
	}
	// Prepare request payload
	jsonData := map[string]any{
		"system":      chat_compeletion_messages[0].Content,
		"model":       chatSession.Model,
		"messages":    messages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"top_p":       chatSession.TopP,
		"stream":      stream,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, ErrValidationInvalidInputGeneric.WithDetail("failed to marshal request payload").WithDebugInfo(err.Error())
	}

	// Get request context for cancellation support
	ctx := m.h.GetRequestContext()
	
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", chatModel.Url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, ErrClaudeRequestFailed.WithDetail("failed to create HTTP request").WithDebugInfo(err.Error())
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)

	if apiKey == "" {
		return nil, ErrAuthInvalidCredentials.WithDetail(fmt.Sprintf("missing API key for model %s", chatSession.Model))
	}

	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	if !stream {
		req.Header.Set("Accept", "application/json")
		client := http.Client{
			Timeout: 5 * time.Minute,
		}

		llmAnswer, err := doGenerateClaude3(ctx, client, req)
		if err != nil {
			return nil, ErrClaudeRequestFailed.WithDetail("failed to generate response").WithDebugInfo(err.Error())
		}

		answerResponse := constructChatCompletionStreamResponse(llmAnswer.AnswerId, llmAnswer.Answer)
		data, err := json.Marshal(answerResponse)
		if err != nil {
			return nil, ErrInternalUnexpected.WithDetail("failed to marshal response").WithDebugInfo(err.Error())
		}

		if _, err := fmt.Fprint(w, string(data)); err != nil {
			return nil, ErrClaudeResponseFaild.WithDetail("failed to write response").WithDebugInfo(err.Error())
		}

		return llmAnswer, nil
	}

	// Handle streaming response
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	llmAnswer, err := m.h.chatStreamClaude3(ctx, w, req, chatUuid, regenerate)
	if err != nil {
		return nil, ErrClaudeStreamFailed.WithDetail("failed to stream response").WithDebugInfo(err.Error())
	}
	return llmAnswer, nil
}

func doGenerateClaude3(ctx context.Context, client http.Client, req *http.Request) (*models.LLMAnswer, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrClaudeRequestFailed.WithMessage("Failed to process Claude request").WithDebugInfo(err.Error())
	}
	// Unmarshal directly from resp.Body
	var message claude.Response
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, ErrClaudeInvalidResponse.WithMessage("Failed to unmarshal Claude response").WithDebugInfo(err.Error())
	}
	defer resp.Body.Close()
	uuid := message.ID
	firstMessage := message.Content[0].Text

	return &models.LLMAnswer{
		AnswerId: uuid,
		Answer:   firstMessage,
	}, nil
}

// claude-3-opus-20240229
// claude-3-sonnet-20240229
// claude-3-haiku-20240307
func (h *ChatHandler) chatStreamClaude3(ctx context.Context, w http.ResponseWriter, req *http.Request, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrClaudeRequestFailed.WithMessage("Failed to process Claude streaming request").WithDebugInfo(err.Error())
	}

	// Use smaller buffer for more responsive streaming
	ioreader := bufio.NewReaderSize(resp.Body, 1024)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	flusher, err := setupSSEStream(w)
	if err != nil {
		return nil, APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		}
	}

	// Flush immediately to establish connection
	flusher.Flush()

	var answer string
	answer_id := GenerateAnswerID(chatUuid, regenerate)

	var headerData = []byte("data: ")
	count := 0
	for {
		// Check if client disconnected or context was cancelled
		select {
		case <-ctx.Done():
			log.Printf("Claude stream cancelled by client: %v", ctx.Err())
			// Return current accumulated content when cancelled
			return &models.LLMAnswer{Answer: answer, AnswerId: answer_id}, nil
		default:
		}

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
					err := FlushResponse(w, flusher, StreamingResponse{
						AnswerID: NewUUID(),
						Content:  string(line),
						IsFinal:  true,
					})
					if err != nil {
						log.Printf("Failed to flush error response: %v", err)
					}
				}
				fmt.Println("End of stream reached")
				return nil, err
			}
			return nil, err
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("event: message_stop")) {
			// stream.isFinished = true
			// No need to send full content at the end since we're sending deltas
			break
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
			log.Println(string(line))
			return nil, ErrClaudeStreamFailed.WithMessage("Error in Claude API response").WithDebugInfo(string(line))
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_start\"")) {
			answer = claude.AnswerFromBlockStart(line)
			err := FlushResponse(w, flusher, StreamingResponse{
				AnswerID: answer_id,
				Content:  answer,
				IsFinal:  false,
			})
			if err != nil {
				log.Printf("Failed to flush content block start: %v", err)
			}
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_delta\"")) {
			delta := claude.AnswerFromBlockDelta(line)
			answer += delta // Still accumulate for final answer storage
			// Send only the delta content
			err := FlushResponse(w, flusher, StreamingResponse{
				AnswerID: answer_id,
				Content:  delta,
				IsFinal:  false,
			})
			if err != nil {
				log.Printf("Failed to flush content block delta: %v", err)
			}
		}
		// Flush after every iteration to ensure immediate delivery
		// This prevents data from being held in buffers
		if count%3 == 0 {
			flusher.Flush()
		}
	}
	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

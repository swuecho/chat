package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/dto"
	claude "github.com/swuecho/chat_backend/llm/claude"
	"github.com/swuecho/chat_backend/models"
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
	h Handler
}

// NewClaude3ChatModel creates a new Claude3ChatModel.
func NewClaude3ChatModel(h Handler) *Claude3ChatModel {
	return &Claude3ChatModel{h: h}
}

func (m *Claude3ChatModel) Stream(ctx context.Context, input Request) (<-chan StreamChunk, error) {
	chatSession, chatCompletionMessages := input.Session, input.Messages
	chatUuid, regenerate, stream := input.ChatUUID, input.Regenerate, input.Stream
	chatModel, chatFiles := input.Model, input.Files

	var claudeMessages []models.Message
	if len(chatCompletionMessages) > 1 {
		claudeMessages = chatCompletionMessages[1:]
		if len(claudeMessages) > 0 && claudeMessages[0].Role == "assistant" {
			claudeMessages = claudeMessages[1:]
		}
	} else {
		return nil, dto.ErrSystemMessageError
	}

	messages := messagesToOpenAIMesages(claudeMessages, chatFiles)

	jsonData := map[string]any{
		"system":      chatCompletionMessages[0].Content,
		"model":       chatSession.Model,
		"messages":    messages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"top_p":       chatSession.TopP,
		"stream":      stream,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, dto.ErrValidationInvalidInputGeneric.WithDetail("failed to marshal request payload").WithDebugInfo(err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, "POST", chatModel.URL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, classifiedFailure("claude", "create request", domain.ProviderFailureConfiguration, false, err)
	}

	apiKey := os.Getenv(chatModel.APIAuthKey)
	if apiKey == "" {
		return nil, classifiedFailure("claude", "configure", domain.ProviderFailureConfiguration, false, fmt.Errorf("missing API key for model %s", chatSession.Model))
	}

	authHeaderName := chatModel.APIAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	ch := make(chan StreamChunk, 10)

	if !stream {
		go func() {
			defer close(ch)
			req.Header.Set("Accept", "application/json")
			client := http.Client{Timeout: 5 * time.Minute}
			llmAnswer, err := doGenerateClaude3(ctx, client, req)
			if err != nil {
				ch <- StreamChunk{Err: err}
				return
			}
			ch <- StreamChunk{
				ID:   llmAnswer.AnswerId,
				Done: true,
				FinalAnswer: &models.LLMAnswer{
					Answer:   llmAnswer.Answer,
					AnswerId: llmAnswer.AnswerId,
				},
			}
		}()
		return ch, nil
	}

	go func() {
		defer close(ch)
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		chatStreamClaude3(ctx, ch, req, chatUuid, regenerate)
	}()
	return ch, nil
}

func doGenerateClaude3(ctx context.Context, client http.Client, req *http.Request) (*models.LLMAnswer, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, normalizeFailure("claude", "complete", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, domain.NewProviderHTTPFailure("claude", "complete", resp.StatusCode, fmt.Errorf("unexpected HTTP status %s", resp.Status))
	}
	var message claude.Response
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, classifiedFailure("claude", "decode response", domain.ProviderFailureInvalidResponse, false, err)
	}
	if len(message.Content) == 0 {
		return nil, classifiedFailure("claude", "decode response", domain.ProviderFailureInvalidResponse, false, errors.New("response contains no content"))
	}
	firstMessage := message.Content[0].Text

	return &models.LLMAnswer{
		AnswerId: message.ID,
		Answer:   firstMessage,
	}, nil
}

func chatStreamClaude3(ctx context.Context, ch chan<- StreamChunk, req *http.Request, chatUuid string, regenerate bool) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		ch <- StreamChunk{Err: normalizeFailure("claude", "open stream", err)}
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		ch <- StreamChunk{Err: domain.NewProviderHTTPFailure("claude", "open stream", resp.StatusCode, fmt.Errorf("unexpected HTTP status %s", resp.Status))}
		return
	}

	ioreader := bufio.NewReaderSize(resp.Body, 1024)
	defer resp.Body.Close()

	var answer string
	answerID := generateAnswerID(chatUuid, regenerate)
	var headerData = []byte("data: ")
	count := 0

	for {
		select {
		case <-ctx.Done():
			slog.Info("Claude stream cancelled by client", "error", ctx.Err())
			ch <- StreamChunk{Done: true, FinalAnswer: &models.LLMAnswer{Answer: answer, AnswerId: answerID}}
			return
		default:
		}

		count++
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
					ch <- StreamChunk{
						ID:          newUUID(),
						Done:        true,
						FinalAnswer: &models.LLMAnswer{Answer: string(line), AnswerId: answerID},
					}
				}
				fmt.Println("End of stream reached")
				break
			}
			ch <- StreamChunk{Err: normalizeFailure("claude", "read stream", err)}
			return
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("event: message_stop")) {
			break
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
			ch <- StreamChunk{Err: classifiedFailure("claude", "read stream", domain.ProviderFailureInvalidResponse, false, errors.New(string(line)))}
			return
		}
		if answerID == "" {
			answerID = newUUID()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_start\"")) {
			delta := claude.AnswerFromBlockStart(line)
			answer = delta
			if len(delta) > 0 {
				ch <- StreamChunk{ID: answerID, Content: delta}
			}
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_delta\"")) {
			delta := claude.AnswerFromBlockDelta(line)
			answer += delta
			if len(delta) > 0 {
				ch <- StreamChunk{ID: answerID, Content: delta}
			}
		}
	}

	ch <- StreamChunk{
		ID:   answerID,
		Done: true,
		FinalAnswer: &models.LLMAnswer{
			Answer:   answer,
			AnswerId: answerID,
		},
	}
}

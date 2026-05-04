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

	"github.com/swuecho/chat_backend/dto"
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
	h Handler
}

// NewCustomChatModel creates a new CustomChatModel.
func NewCustomChatModel(h Handler) *CustomChatModel {
	return &CustomChatModel{h: h}
}

func (m *CustomChatModel) Stream(ctx context.Context, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool, stream bool) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	go func() {
		defer close(ch)
		m.customChatStream(ctx, ch, chatSession, chatCompletionMessages, chatUuid, regenerate)
	}()
	return ch, nil
}

func (m *CustomChatModel) customChatStream(ctx context.Context, ch chan<- StreamChunk, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool) {
	chatModel, err := GetChatModel(ctx, m.h.Queries(), chatSession.Model)
	if err != nil {
		ch <- StreamChunk{Err: err}
		return
	}

	apiKey := os.Getenv(chatModel.ApiAuthKey)
	url := chatModel.Url

	prompt := claude.FormatClaudePrompt(chatCompletionMessages)

	jsonData := map[string]any{
		"prompt":               prompt,
		"model":                chatSession.Model,
		"max_tokens_to_sample": chatSession.MaxTokens,
		"temperature":          chatSession.Temperature,
		"stop_sequences":       []string{"\n\nHuman:"},
		"stream":               true,
	}

	jsonValue, _ := json.Marshal(jsonData)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrChatRequestFailed.WithMessage("Failed to create custom model request").WithDebugInfo(err.Error())}
		return
	}

	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	SetStreamingHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrChatRequestFailed.WithMessage("Failed to send custom model request").WithDebugInfo(err.Error())}
		return
	}
	defer resp.Body.Close()

	ioreader := bufio.NewReader(resp.Body)

	var answer string
	answerID := generateAnswerID(chatUuid, regenerate)
	var lastFlushLength int
	headerData := []byte("data: ")
	count := 0

	for {
		select {
		case <-ctx.Done():
			slog.Info("Custom model stream cancelled by client", "error", ctx.Err())
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
				fmt.Println("End of stream reached")
				break
			}
			ch <- StreamChunk{Err: err}
			return
		}

		if !bytes.HasPrefix(line, headerData) {
			continue
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("[DONE]")) {
			fmt.Println("DONE break")
			break
		}

		if answerID == "" {
			answerID = NewUUID()
		}

		var response CustomModelResponse
		_ = json.Unmarshal(line, &response)
		answer = response.Completion

		shouldFlush := len(answer) < 200 || (len(answer)-lastFlushLength) >= 500
		if shouldFlush {
			delta := answer[lastFlushLength:]
			if len(delta) > 0 {
				ch <- StreamChunk{ID: answerID, Content: delta}
			}
			lastFlushLength = len(answer)
		}
	}

	// Send remaining content
	if len(answer) > lastFlushLength {
		ch <- StreamChunk{ID: answerID, Content: answer[lastFlushLength:]}
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

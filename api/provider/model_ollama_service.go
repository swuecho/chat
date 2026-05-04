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
	"strings"
	"time"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// OllamaResponse represents the response structure from Ollama API
type OllamaResponse struct {
	Model              string         `json:"model"`
	CreatedAt          time.Time      `json:"created_at"`
	Done               bool           `json:"done"`
	Message            models.Message `json:"message"`
	TotalDuration      int64          `json:"total_duration"`
	LoadDuration       int64          `json:"load_duration"`
	PromptEvalCount    int            `json:"prompt_eval_count"`
	PromptEvalDuration int64          `json:"prompt_eval_duration"`
	EvalCount          int            `json:"eval_count"`
	EvalDuration       int64          `json:"eval_duration"`
}

// Ollama ChatModel implementation
type OllamaChatModel struct {
	h Handler
}

// NewOllamaChatModel creates a new OllamaChatModel.
func NewOllamaChatModel(h Handler) *OllamaChatModel {
	return &OllamaChatModel{h: h}
}

func (m *OllamaChatModel) Stream(ctx context.Context, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool, stream bool) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	go func() {
		defer close(ch)
		chatOllamStream(ctx, ch, m.h, chatSession, chatCompletionMessages, chatUuid, regenerate)
	}()
	return ch, nil
}

func chatOllamStream(ctx context.Context, ch chan<- StreamChunk, h Handler, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool) {
	chatModel, err := GetChatModel(ctx, h.Queries(), chatSession.Model)
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrResourceNotFound("chat model: " + chatSession.Model)}
		return
	}

	jsonData := map[string]any{
		"model":    strings.Replace(chatSession.Model, "ollama-", "", 1),
		"messages": chatCompletionMessages,
	}
	jsonValue, _ := json.Marshal(jsonData)

	req, err := http.NewRequestWithContext(ctx, "POST", chatModel.Url, bytes.NewBuffer(jsonValue))
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrInternalUnexpected.WithMessage("Failed to make request").WithDebugInfo(err.Error())}
		return
	}

	apiKey := os.Getenv(chatModel.ApiAuthKey)
	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Access-Control-Allow-Origin", "*")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrInternalUnexpected.WithMessage("Failed to create chat completion stream").WithDebugInfo(err.Error())}
		return
	}

	ioreader := bufio.NewReader(resp.Body)
	defer resp.Body.Close()

	var answer string
	answerID := generateAnswerID(chatUuid, regenerate)

	count := 0
	for {
		select {
		case <-ctx.Done():
			slog.Info("Ollama stream cancelled by client", "error", ctx.Err())
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
		var streamResp OllamaResponse
		if err := json.Unmarshal(line, &streamResp); err != nil {
			ch <- StreamChunk{Err: err}
			return
		}
		delta := strings.ReplaceAll(streamResp.Message.Content, "<0x0A>", "\n")
		answer += delta

		if streamResp.Done {
			fmt.Println("DONE break")
			break
		}
		if answerID == "" {
			answerID = NewUUID()
		}

		if len(delta) > 0 {
			ch <- StreamChunk{ID: answerID, Content: delta}
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

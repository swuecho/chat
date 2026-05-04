package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/llm/gemini"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// GeminiClient handles communication with the Gemini API
type GeminiClient struct {
	client *http.Client
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient() *GeminiClient {
	return &GeminiClient{
		client: &http.Client{Timeout: 5 * time.Minute},
	}
}

// Gemini ChatModel implementation
type GeminiChatModel struct {
	h      Handler
	client *GeminiClient
}

func NewGeminiChatModel(h Handler) *GeminiChatModel {
	return &GeminiChatModel{
		h:      h,
		client: NewGeminiClient(),
	}
}

func (m *GeminiChatModel) Stream(ctx context.Context, chatSession sqlc_queries.ChatSession, messages []models.Message, chatUuid string, regenerate bool, stream bool) (<-chan StreamChunk, error) {
	answerID := generateAnswerID(chatUuid, regenerate)

	chatFiles, err := GetChatFiles(ctx, m.h.Queries(), chatSession.Uuid)
	if err != nil {
		return nil, err
	}

	payloadBytes, err := gemini.GenGemminPayload(messages, chatFiles)
	if err != nil {
		return nil, dto.ErrInternalUnexpected.WithMessage("Failed to generate Gemini payload").WithDebugInfo(err.Error())
	}

	url := gemini.BuildAPIURL(chatSession.Model, stream)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, dto.ErrInternalUnexpected.WithMessage("Failed to create Gemini API request").WithDebugInfo(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	ch := make(chan StreamChunk, 10)

	if stream {
		go func() {
			defer close(ch)
			m.handleStreamResponse(ctx, ch, req, answerID)
		}()
		return ch, nil
	}

	go func() {
		defer close(ch)
		llmAnswer, err := gemini.HandleRegularResponse(*m.client.client, req)
		if err != nil {
			ch <- StreamChunk{Err: err}
			return
		}
		if llmAnswer == nil {
			ch <- StreamChunk{Err: dto.ErrInternalUnexpected.WithMessage("Empty response from Gemini")}
			return
		}
		ch <- StreamChunk{
			ID:   answerID,
			Done: true,
			FinalAnswer: &models.LLMAnswer{
				Answer:   llmAnswer.Answer,
				AnswerId: answerID,
			},
		}
	}()
	return ch, nil
}

func GenerateChatTitle(ctx context.Context, model, chatText string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", dto.ErrInternalUnexpected.WithMessage("GEMINI_API_KEY environment variable not set")
	}

	if strings.TrimSpace(chatText) == "" {
		return "", dto.ErrValidationInvalidInput("chat text cannot be empty")
	}

	messages := []models.Message{
		{
			Role:    "user",
			Content: `Generate a short title (3-6 words) for this conversation. Output ONLY the title text, no quotes, no markdown, no prefixes like "Title:". Example: "Python list comprehension guide"`,
		},
		{
			Role:    "user",
			Content: chatText,
		},
	}

	payloadBytes, err := gemini.GenGemminPayload(messages, nil)
	if err != nil {
		return "", dto.ErrInternalUnexpected.WithMessage("Failed to generate Gemini payload").WithDebugInfo(err.Error())
	}

	url := gemini.BuildAPIURL(model, false)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", dto.ErrInternalUnexpected.WithMessage("Failed to create Gemini API request").WithDebugInfo(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	answer, err := gemini.HandleRegularResponse(http.Client{Timeout: 1 * time.Minute}, req)
	if err != nil {
		return "", dto.ErrInternalUnexpected.WithMessage("Failed to handle Gemini response").WithDebugInfo(err.Error())
	}

	if answer == nil || answer.Answer == "" {
		return "", dto.ErrInternalUnexpected.WithMessage("Empty response from Gemini")
	}

	title := strings.TrimSpace(answer.Answer)
	title = strings.Trim(title, `"`)
	title = strings.Trim(title, `*`)
	title = strings.Trim(title, `#`)
	title = strings.TrimPrefix(title, "Title:")
	title = strings.TrimPrefix(title, "title:")
	title = strings.TrimPrefix(title, "Title: ")
	title = strings.TrimPrefix(title, "title: ")
	title = strings.TrimSpace(title)
	for strings.HasPrefix(title, "#") || strings.HasPrefix(title, "-") || strings.HasPrefix(title, "*") {
		title = strings.TrimLeft(title, "#-* ")
		title = strings.TrimSpace(title)
	}
	if title == "" {
		return "", dto.ErrInternalUnexpected.WithMessage("Invalid title generated")
	}

	return FirstN(title, 100), nil
}

func (m *GeminiChatModel) handleStreamResponse(ctx context.Context, ch chan<- StreamChunk, req *http.Request, answerID string) {
	resp, err := m.client.client.Do(req)
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrInternalUnexpected.WithMessage("Failed to send Gemini API request").WithDebugInfo(err.Error())}
		return
	}
	defer resp.Body.Close()

	var answer string
	slog.Info("gemini response", "statusCode", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		slog.Info("gemini error body", "body", string(errorBody))
		var apiError gemini.GoogleApiError
		if json.Unmarshal(errorBody, &apiError) == nil && apiError.Error.Message != "" {
			slog.Warn("API returned non-200 status", "statusCode", resp.StatusCode, "statusText", http.StatusText(resp.StatusCode), "error", &apiError)
		} else {
			slog.Warn("API returned non-200 status", "statusCode", resp.StatusCode, "statusText", http.StatusText(resp.StatusCode), "body", string(errorBody))
		}
		ch <- StreamChunk{Err: dto.APIError{
			HTTPCode: apiError.Error.Code,
			Code:     apiError.Error.Status,
			Message:  apiError.Error.Message,
		}}
		return
	}
	ioreader := bufio.NewReader(resp.Body)
	headerData := []byte("data: ")

	for count := 0; count < 10000; count++ {
		select {
		case <-ctx.Done():
			slog.Info("Gemini stream cancelled by client", "error", ctx.Err())
			ch <- StreamChunk{Done: true, FinalAnswer: &models.LLMAnswer{Answer: answer, AnswerId: answerID}}
			return
		default:
		}

		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				ch <- StreamChunk{
					ID:          answerID,
					Done:        true,
					FinalAnswer: &models.LLMAnswer{Answer: answer, AnswerId: answerID},
				}
				return
			}
			ch <- StreamChunk{Err: dto.ErrInternalUnexpected.WithMessage("Error reading stream").WithDebugInfo(err.Error())}
			return
		}

		if !bytes.HasPrefix(line, headerData) {
			continue
		}

		line = bytes.TrimPrefix(line, headerData)
		if len(line) > 0 {
			delta := gemini.ParseRespLineDelta(line)
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
			AnswerId: answerID,
			Answer:   answer,
		},
	}
}

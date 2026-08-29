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
	"strings"
	"time"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/llm/gemini"
	"github.com/swuecho/chat_backend/models"
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

func (m *GeminiChatModel) Stream(ctx context.Context, input Request) (<-chan StreamChunk, error) {
	chatSession, messages := input.Session, input.Messages
	chatUuid, regenerate, stream := input.ChatUUID, input.Regenerate, input.Stream
	answerID := generateAnswerID(chatUuid, regenerate)
	chatFiles := input.Files

	geminiFiles := make([]gemini.File, 0, len(chatFiles))
	for _, file := range chatFiles {
		geminiFiles = append(geminiFiles, gemini.File{Name: file.Name, Data: file.Data, MIMEType: file.MIMEType})
	}
	payloadBytes, err := gemini.GenGemminPayload(messages, geminiFiles)
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

func GenerateChatTitle(ctx context.Context, model ModelConfig, chatText string) (string, error) {
	if strings.TrimSpace(chatText) == "" {
		return "", dto.ErrValidationInvalidInput("chat text cannot be empty")
	}

	messages := []models.Message{
		{
			Role:    "system",
			Content: `Generate a short title (3-6 words) for this conversation. Output ONLY the title text, no quotes, no markdown, no prefixes like "Title:". Example: "Python list comprehension guide"`,
		},
		{
			Role:    "user",
			Content: chatText,
		},
	}

	h := newTitleGenerationHandler()
	var titleModel ChatModel
	switch model.APIType {
	case "claude":
		titleModel = NewClaude3ChatModel(h)
	case "gemini":
		titleModel = NewGeminiChatModel(h)
	case "ollama":
		titleModel = NewOllamaChatModel(h)
	case "custom":
		titleModel = NewCustomChatModel(h)
	default:
		if strings.HasSuffix(strings.TrimSuffix(model.URL, "/"), "/completions") &&
			!strings.HasSuffix(strings.TrimSuffix(model.URL, "/"), "/chat/completions") {
			titleModel = NewCompletionChatModel(h)
		} else {
			titleModel = NewOpenAIChatModel(h)
		}
	}

	stream, err := titleModel.Stream(ctx, Request{
		Session: Session{Model: model.Name, MaxTokens: 64, Temperature: 0.2, TopP: 1, N: 1},
		Model:   model, Messages: messages, ChatUUID: "title-generation",
	})
	if err != nil {
		return "", err
	}

	var answer string
	for chunk := range stream {
		if chunk.Err != nil {
			return "", chunk.Err
		}
		if chunk.FinalAnswer != nil {
			answer = chunk.FinalAnswer.Answer
		} else if chunk.Content != "" {
			answer += chunk.Content
		}
	}
	if strings.TrimSpace(answer) == "" {
		return "", dto.ErrInternalUnexpected.WithMessage("Empty response from title generation model")
	}

	title := strings.TrimSpace(answer)
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

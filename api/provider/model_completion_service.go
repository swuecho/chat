package provider

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// CompletionChatModel implements ChatModel interface for OpenAI completion models
type CompletionChatModel struct {
	h Handler
}

// NewCompletionChatModel creates a new CompletionChatModel.
func NewCompletionChatModel(h Handler) *CompletionChatModel {
	return &CompletionChatModel{h: h}
}

func (m *CompletionChatModel) Stream(ctx context.Context, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool, stream bool) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	go func() {
		defer close(ch)
		m.completionStream(ctx, ch, chatSession, chatCompletionMessages, chatUuid, regenerate)
	}()
	return ch, nil
}

func (m *CompletionChatModel) completionStream(ctx context.Context, ch chan<- StreamChunk, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool) {
	m.h.Config().RateLimiter.Wait(ctx)

	if err := m.h.CheckModelAccess(ctx, chatSession.Uuid, chatSession.Model, chatSession.UserID); err != nil {
		ch <- StreamChunk{Err: err}
		return
	}

	chatModel, err := GetChatModel(ctx, m.h.Queries(), chatSession.Model)
	if err != nil {
		ch <- StreamChunk{Err: err}
		return
	}

	config, err := GenOpenAIConfig(*chatModel, m.h.Config())
	if err != nil {
		ch <- StreamChunk{Err: err}
		return
	}

	client := openai.NewClientWithConfig(config)
	prompt := chatCompletionMessages[len(chatCompletionMessages)-1].Content

	N := chatSession.N
	req := openai.CompletionRequest{
		Model:       chatSession.Model,
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
		N:           int(N),
		Prompt:      prompt,
		Stream:      true,
	}

	ctx, cancel := context.WithTimeout(ctx, dto.DefaultRequestTimeout)
	defer cancel()

	stream, err := client.CreateCompletionStream(ctx, req)
	if err != nil {
		ch <- StreamChunk{Err: dto.ErrInternalUnexpected.WithMessage("Failed to create completion stream").WithDebugInfo(err.Error())}
		return
	}
	defer stream.Close()

	var answer string
	answerID := generateAnswerID(chatUuid, regenerate)
	TextBuffer := NewTextBuffer(N, "```\n"+prompt, "\n```\n")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Completion stream cancelled by client", "error", ctx.Err())
			ch <- StreamChunk{Done: true, FinalAnswer: &models.LLMAnswer{Answer: answer, AnswerId: answerID}}
			return
		default:
		}

		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			if len(answer) > 0 {
				ch <- StreamChunk{ID: answerID, Content: answer, Done: true, FinalAnswer: &models.LLMAnswer{AnswerId: answerID, Answer: answer}}
				return
			}
			break
		}

		if err != nil {
			ch <- StreamChunk{Err: dto.ErrChatStreamFailed.WithMessage("Stream error occurred").WithDebugInfo(err.Error())}
			return
		}

		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Text
		TextBuffer.AppendByIndex(textIdx, delta)

		if chatSession.Debug {
			slog.Info("completion chunk", "index", textIdx, "delta", delta)
		}

		if answerID == "" {
			answerID = response.ID
		}

		answer = TextBuffer.String("\n\n")

		perWordStreamLimit := GetPerWordStreamLimit()
		if strings.HasSuffix(delta, "\n") || len(answer) < perWordStreamLimit {
			if len(answer) > 0 {
				ch <- StreamChunk{ID: answerID, Content: answer}
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

package provider

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
)

// CompletionChatModel implements ChatModel interface for OpenAI completion models
type CompletionChatModel struct {
	h Handler
}

// NewCompletionChatModel creates a new CompletionChatModel.
func NewCompletionChatModel(h Handler) *CompletionChatModel {
	return &CompletionChatModel{h: h}
}

func (m *CompletionChatModel) Stream(ctx context.Context, input Request) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 10)
	go func() {
		defer close(ch)
		m.completionStream(ctx, ch, input)
	}()
	return ch, nil
}

func (m *CompletionChatModel) completionStream(ctx context.Context, ch chan<- StreamChunk, input Request) {
	chatSession, chatCompletionMessages := input.Session, input.Messages
	chatUuid, regenerate, chatModel := input.ChatUUID, input.Regenerate, input.Model
	if limiter := m.h.Config().RateLimiter; limiter != nil {
		if err := limiter.Wait(ctx); err != nil {
			emitChunk(ctx, ch, StreamChunk{Err: normalizeFailure("openai-completion", "wait for local rate limit", err)})
			return
		}
	}

	if err := m.h.CheckModelAccess(ctx, chatSession.UUID, chatSession.Model, chatSession.UserID); err != nil {
		emitChunk(ctx, ch, StreamChunk{Err: err})
		return
	}

	config, err := GenOpenAIConfig(chatModel, m.h.Config())
	if err != nil {
		emitChunk(ctx, ch, StreamChunk{Err: classifiedFailure("openai-completion", "configure", domain.ProviderFailureConfiguration, false, err)})
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
		emitChunk(ctx, ch, StreamChunk{Err: normalizeFailure("openai-completion", "open stream", err)})
		return
	}
	defer stream.Close()

	var answer string
	answerID := generateAnswerID(chatUuid, regenerate, input.NewID)
	TextBuffer := NewTextBuffer(N, "```\n"+prompt, "\n```\n")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Completion stream cancelled by client", "error", ctx.Err())
			return
		default:
		}

		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			if len(answer) > 0 {
				emitChunk(ctx, ch, StreamChunk{ID: answerID, Content: answer, Done: true, FinalAnswer: &models.LLMAnswer{AnswerId: answerID, Answer: answer}})
				return
			}
			break
		}

		if err != nil {
			emitChunk(ctx, ch, StreamChunk{Err: normalizeFailure("openai-completion", "read stream", err)})
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
				if !emitChunk(ctx, ch, StreamChunk{ID: answerID, Content: answer}) {
					return
				}
			}
		}
	}

	emitChunk(ctx, ch, StreamChunk{
		ID:   answerID,
		Done: true,
		FinalAnswer: &models.LLMAnswer{
			AnswerId: answerID,
			Answer:   answer,
		},
	})
}

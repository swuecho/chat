package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/dto"
	llm_openai "github.com/swuecho/chat_backend/llm/openai"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// OpenAI ChatModel implementation
type OpenAIChatModel struct {
	h Handler
}

// NewOpenAIChatModel creates a new OpenAIChatModel.
func NewOpenAIChatModel(h Handler) *OpenAIChatModel {
	return &OpenAIChatModel{h: h}
}

func (m *OpenAIChatModel) Stream(ctx context.Context, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (<-chan StreamChunk, error) {
	m.h.Config().RateLimiter.Wait(ctx)

	if err := m.h.CheckModelAccess(ctx, chatSession.Uuid, chatSession.Model, chatSession.UserID); err != nil {
		return nil, err
	}

	chatModel, err := GetChatModel(ctx, m.h.Queries(), chatSession.Model)
	if err != nil {
		return nil, err
	}

	config, err := GenOpenAIConfig(*chatModel, m.h.Config())
	if err != nil {
		return nil, dto.ErrOpenAIConfigFailed.WithMessage("Failed to generate OpenAI config").WithDebugInfo(err.Error())
	}

	chatFiles, err := GetChatFiles(ctx, m.h.Queries(), chatSession.Uuid)
	if err != nil {
		return nil, err
	}

	openaiReq := NewChatCompletionRequest(chatSession, chatCompletionMessages, chatFiles, streamOutput)
	openaiReq.Model = NormalizeOpenAIModelName(*chatModel, openaiReq.Model)
	if len(openaiReq.Messages) <= 1 {
		return nil, dto.ErrSystemMessageError
	}
	slog.Info("OpenAI request prepared", "model", openaiReq.Model, "messageCount", len(openaiReq.Messages), "temperature", openaiReq.Temperature)
	client := openai.NewClientWithConfig(config)

	ch := make(chan StreamChunk, 10)
	go func() {
		defer close(ch)
		if streamOutput {
			doChatStream(ctx, ch, client, openaiReq, chatSession.N, chatUuid, regenerate, chatModel.Url, config.BaseURL)
		} else {
			handleRegularResponse(ctx, ch, client, openaiReq, chatModel.Url, config.BaseURL)
		}
	}()
	return ch, nil
}

func handleRegularResponse(ctx context.Context, ch chan<- StreamChunk, client *openai.Client, req openai.ChatCompletionRequest, configuredURL string, baseURL string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	completion, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Info("OpenAI request failed", "model", req.Model, "configuredURL", configuredURL, "baseURL", baseURL, "error", err)
		ch <- StreamChunk{Err: dto.ErrOpenAIRequestFailed.WithMessage("Failed to create chat completion").WithDebugInfo(err.Error())}
		return
	}

	ch <- StreamChunk{
		ID:      completion.ID,
		Content: completion.Choices[0].Message.Content,
		Done:    true,
		FinalAnswer: &models.LLMAnswer{
			Answer:   completion.Choices[0].Message.Content,
			AnswerId: completion.ID,
		},
	}
}

// doChatStream handles streaming chat completion responses from OpenAI.
// It sends chunks on the provided channel and closes it when done.
func doChatStream(ctx context.Context, ch chan<- StreamChunk, client *openai.Client, req openai.ChatCompletionRequest, bufferLen int32, chatUuid string, regenerate bool, configuredURL string, baseURL string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	slog.Info("Creating OpenAI stream")
	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		slog.Info("OpenAI stream setup failed", "model", req.Model, "configuredURL", configuredURL, "baseURL", baseURL, "error", err)
		ch <- StreamChunk{Err: dto.ErrOpenAIStreamFailed.WithMessage("Failed to create chat completion stream").WithDebugInfo(err.Error())}
		return
	}
	defer func() {
		if err := stream.Close(); err != nil {
			slog.Error("error closing OpenAI stream", "error", err)
		}
	}()

	var answerID string
	var hasReason bool
	var reasonTagOpened bool
	var reasonTagClosed bool

	if bufferLen == 0 {
		slog.Info("Buffer length is 0, setting to 1")
		bufferLen = 1
	}

	TextBuffer := NewTextBuffer(bufferLen, "", "")
	reasonBuffer := NewTextBuffer(bufferLen, "<think>\n\n", "\n\n</think>\n\n")
	answerID = generateAnswerID(chatUuid, regenerate)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stream cancelled by client", "error", ctx.Err())
			llmAnswer := models.LLMAnswer{Answer: TextBuffer.String("\n"), AnswerId: answerID}
			if hasReason {
				llmAnswer.ReasoningContent = reasonBuffer.String("\n")
			}
			ch <- StreamChunk{Done: true, FinalAnswer: &llmAnswer}
			return
		default:
		}

		rawLine, err := stream.RecvRaw()
		if err != nil {
			slog.Info("OpenAI stream receive error", "model", req.Model, "configuredURL", configuredURL, "baseURL", baseURL, "error", err)
			if errors.Is(err, io.EOF) {
				if TextBuffer.String("\n") == "" && reasonBuffer.String("\n") == "" {
					errMsg := fmt.Sprintf("stream closed without content; verify configured URL %q resolves to a valid OpenAI-compatible base URL %q and that model %q is valid", configuredURL, baseURL, req.Model)
					slog.Info(errMsg)
					ch <- StreamChunk{Err: dto.ErrOpenAIStreamFailed.WithMessage("Stream closed without content").WithDebugInfo(errMsg)}
					return
				}
				llmAnswer := models.LLMAnswer{Answer: TextBuffer.String("\n"), AnswerId: answerID}
				if hasReason {
					llmAnswer.ReasoningContent = reasonBuffer.String("\n")
				}
				ch <- StreamChunk{Done: true, FinalAnswer: &llmAnswer}
				return
			}
			slog.Info("Stream error", "error", err)
			ch <- StreamChunk{Err: dto.ErrOpenAIStreamFailed.WithMessage("Stream error occurred").WithDebugInfo(err.Error())}
			return
		}

		response := llm_openai.ChatCompletionStreamResponse{}
		if err := json.Unmarshal(rawLine, &response); err != nil {
			slog.Info("Could not unmarshal response", "error", err)
			continue
		}

		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Delta

		TextBuffer.AppendByIndex(textIdx, delta.Content)
		if len(delta.ReasoningContent) > 0 {
			hasReason = true
			reasonBuffer.AppendByIndex(textIdx, delta.ReasoningContent)
		}

		if answerID == "" {
			answerID = strings.TrimPrefix(response.ID, "chatcmpl-")
		}

		if len(delta.Content) > 0 || len(delta.ReasoningContent) > 0 {
			deltaToSend := processDelta(delta, &reasonTagOpened, &reasonTagClosed, hasReason)
			if len(deltaToSend) > 0 {
				ch <- StreamChunk{ID: answerID, Content: deltaToSend}
			}
		}
	}
}

// processDelta handles the logic for processing delta content with thinking tags
func processDelta(delta llm_openai.ChatCompletionStreamChoiceDelta, reasonTagOpened *bool, reasonTagClosed *bool, hasReason bool) string {
	var deltaToSend string

	if len(delta.ReasoningContent) > 0 {
		if !*reasonTagOpened {
			deltaToSend = "<think>" + delta.ReasoningContent
			*reasonTagOpened = true
		} else {
			deltaToSend = delta.ReasoningContent
		}
	} else if hasReason && !*reasonTagClosed {
		deltaToSend = "</think>" + delta.Content
		*reasonTagClosed = true
	} else {
		deltaToSend = delta.Content
	}

	return deltaToSend
}

// NewUserMessage creates a new OpenAI user message
func NewUserMessage(content string) openai.ChatCompletionMessage {
	return openai.ChatCompletionMessage{Role: "user", Content: content}
}

// NewChatCompletionRequest creates an OpenAI chat completion request from session and messages
func NewChatCompletionRequest(chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatFiles []sqlc_queries.ChatFile, streamOutput bool) openai.ChatCompletionRequest {
	openaiMessages := messagesToOpenAIMesages(chatCompletionMessages, chatFiles)

	for _, m := range openaiMessages {
		b, _ := m.MarshalJSON()
		slog.Info("openai message", "msg", string(b))
	}

	slog.Info("openai messages", "messages", openaiMessages)
	topP := float32(chatSession.TopP) - 0.01
	if topP <= 0 {
		topP = 0.01
	}
	openaiReq := openai.ChatCompletionRequest{
		Model:       chatSession.Model,
		Messages:    openaiMessages,
		Temperature: float32(chatSession.Temperature),
		TopP:        topP,
		N:           int(chatSession.N),
		Stream:      streamOutput,
	}
	return openaiReq
}

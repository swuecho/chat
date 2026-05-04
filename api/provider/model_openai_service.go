package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	llm_openai "github.com/swuecho/chat_backend/llm/openai"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/dto"
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

func (m *OpenAIChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	m.h.Config().RateLimiter.Wait(m.h.RequestContext())

	exceedPerModeRateLimitOrError := m.h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return nil, fmt.Errorf("exceed per mode rate limit")
	}

	chatModel, err := GetChatModel(m.h.RequestContext(), m.h.Queries(), chatSession.Model)
	if err != nil {
		return nil, err
	}

	config, err := GenOpenAIConfig(*chatModel, m.h.Config())
	slog.Info("%+v", config.String())
	// print all config details
	if err != nil {
		return nil, dto.ErrOpenAIConfigFailed.WithMessage("Failed to generate OpenAI config").WithDebugInfo(err.Error())
	}

	chatFiles, err := GetChatFiles(m.h.RequestContext(), m.h.Queries(), chatSession.Uuid)
	if err != nil {
		return nil, err
	}

	openaiReq := NewChatCompletionRequest(chatSession, chatCompletionMessages, chatFiles, streamOutput)
	openaiReq.Model = NormalizeOpenAIModelName(*chatModel, openaiReq.Model)
	if len(openaiReq.Messages) <= 1 {
		return nil, dto.ErrSystemMessageError
	}
	slog.Info("OpenAI request prepared - Model: %s, MessageCount: %d, Temperature: %.2f",
		openaiReq.Model, len(openaiReq.Messages), openaiReq.Temperature)
	client := openai.NewClientWithConfig(config)
	if streamOutput {
		return doChatStream(w, client, openaiReq, chatSession.N, chatUuid, regenerate, m.h, chatModel.Url, config.BaseURL)
	} else {
		return handleRegularResponse(m.h.RequestContext(), w, client, openaiReq, chatModel.Url, config.BaseURL)
	}

}

func handleRegularResponse(ctx context.Context, w http.ResponseWriter, client *openai.Client, req openai.ChatCompletionRequest, configuredURL string, baseURL string) (*models.LLMAnswer, error) {
	// check per chat_model limit
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	completion, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Info("OpenAI request failed - Model: %s, ConfiguredURL: %s, BaseURL: %s, Error: %+v", req.Model, configuredURL, baseURL, err)
		return nil, dto.ErrOpenAIRequestFailed.WithMessage("Failed to create chat completion").WithDebugInfo(err.Error())
	}
	slog.Info("completion: %+v", completion)
	data, _ := json.Marshal(completion)
	fmt.Fprint(w, string(data))
	return &models.LLMAnswer{Answer: completion.Choices[0].Message.Content, AnswerId: completion.ID}, nil
}

// doChatStream handles streaming chat completion responses from OpenAI
// It properly manages thinking tags for models that support reasoning content
func doChatStream(w http.ResponseWriter, client *openai.Client, req openai.ChatCompletionRequest, bufferLen int32, chatUuid string, regenerate bool, handler Handler, configuredURL string, baseURL string) (*models.LLMAnswer, error) {
	// Use request context with timeout
	ctx, cancel := context.WithTimeout(handler.RequestContext(), 5*time.Minute)
	defer cancel()

	slog.Info("Creating OpenAI stream")
	stream, err := client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		slog.Info("OpenAI stream setup failed - Model: %s, ConfiguredURL: %s, BaseURL: %s, Error: %+v", req.Model, configuredURL, baseURL, err)
		return nil, dto.ErrOpenAIStreamFailed.WithMessage("Failed to create chat completion stream").WithDebugInfo(err.Error())
	}
	defer func() {
		if err := stream.Close(); err != nil {
			slog.Error("error: closing OpenAI stream: %v", err)
		}
	}()

	// Setup Server-Sent Events (SSE) streaming
	flusher, err := SetupSSEStream(w)
	if err != nil {
		return nil, dto.APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		}
	}

	// Initialize streaming state
	var answer_id string

	var hasReason bool       // Whether we've detected any reasoning content
	var reasonTagOpened bool // Whether we've sent the opening <think> tag
	var reasonTagClosed bool // Whether we've sent the closing </think> tag

	// Ensure minimum buffer length
	if bufferLen == 0 {
		slog.Info("Buffer length is 0, setting to 1")
		bufferLen = 1
	}

	// Initialize buffers for accumulating content
	TextBuffer := NewTextBuffer(bufferLen, "", "")
	reasonBuffer := NewTextBuffer(bufferLen, "<think>\n\n", "\n\n</think>\n\n")
	answer_id = generateAnswerID(chatUuid, regenerate)
	// Main streaming loop
	for {
		// Check if client disconnected or context was cancelled
		select {
		case <-ctx.Done():
			slog.Info("Stream cancelled by client: %v", ctx.Err())
			// Return current accumulated content when cancelled
			llmAnswer := models.LLMAnswer{Answer: TextBuffer.String("\n"), AnswerId: answer_id}
			if hasReason {
				llmAnswer.ReasoningContent = reasonBuffer.String("\n")
			}
			return &llmAnswer, nil
		default:
		}

		rawLine, err := stream.RecvRaw()
		if err != nil {
			slog.Info("OpenAI stream receive error - Model: %s, ConfiguredURL: %s, BaseURL: %s, Error: %+v", req.Model, configuredURL, baseURL, err)
			if errors.Is(err, io.EOF) {
				if TextBuffer.String("\n") == "" && reasonBuffer.String("\n") == "" {
					errMsg := fmt.Sprintf("stream closed without content; verify configured URL %q resolves to a valid OpenAI-compatible base URL %q and that model %q is valid", configuredURL, baseURL, req.Model)
					slog.Info(errMsg)
					return nil, dto.ErrOpenAIStreamFailed.WithMessage("Stream closed without content").WithDebugInfo(errMsg)
				}
				// Stream ended successfully - return accumulated content
				llmAnswer := models.LLMAnswer{Answer: TextBuffer.String("\n"), AnswerId: answer_id}
				if hasReason {
					llmAnswer.ReasoningContent = reasonBuffer.String("\n")
				}
				return &llmAnswer, nil
			} else {
				slog.Info("Stream error: %v", err)
				return nil, dto.ErrOpenAIStreamFailed.WithMessage("Stream error occurred").WithDebugInfo(err.Error())
			}
		}
		// Parse the streaming response
		response := llm_openai.ChatCompletionStreamResponse{}
		err = json.Unmarshal(rawLine, &response)
		if err != nil {
			slog.Info("Could not unmarshal response: %v\n", err)
			continue
		}

		// Extract delta content from the response
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Delta

		// Accumulate content in buffers (for final answer construction)
		TextBuffer.AppendByIndex(textIdx, delta.Content)
		if len(delta.ReasoningContent) > 0 {
			hasReason = true
			reasonBuffer.AppendByIndex(textIdx, delta.ReasoningContent)
		}

		// Set answer ID from response if not already set
		if answer_id == "" {
			answer_id = strings.TrimPrefix(response.ID, "chatcmpl-")
		}

		// Process and send delta content
		if len(delta.Content) > 0 || len(delta.ReasoningContent) > 0 {
			deltaToSend := processDelta(delta, &reasonTagOpened, &reasonTagClosed, hasReason)
			if len(deltaToSend) > 0 {
				slog.Info("delta: %s", deltaToSend)
				err := FlushResponse(w, flusher, StreamingResponse{
					AnswerID: answer_id,
					Content:  deltaToSend,
					IsFinal:  false,
				})
				if err != nil {
					slog.Info("Failed to flush response: %v", err)
				}
			}
		}
	}
}

// processDelta handles the logic for processing delta content with thinking tags
func processDelta(delta llm_openai.ChatCompletionStreamChoiceDelta, reasonTagOpened *bool, reasonTagClosed *bool, hasReason bool) string {
	var deltaToSend string

	if len(delta.ReasoningContent) > 0 {
		// Handle reasoning content
		if !*reasonTagOpened {
			// First time seeing reasoning content, add opening tag
			deltaToSend = "<think>" + delta.ReasoningContent
			*reasonTagOpened = true
		} else {
			// Continue reasoning content
			deltaToSend = delta.ReasoningContent
		}
	} else if hasReason && !*reasonTagClosed {
		// We had reasoning content before and now we have regular content for the first time
		// Close the think tag first, then send the content
		deltaToSend = "</think>" + delta.Content
		*reasonTagClosed = true
	} else {
		// Regular content without reasoning
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
		slog.Info("messages: %+v\n", string(b))
	}

	slog.Info("messages: %+v\n", openaiMessages)
	// Ensure TopP is always greater than 0 to prevent API validation errors
	topP := float32(chatSession.TopP) - 0.01
	if topP <= 0 {
		topP = 0.01 // Minimum valid value
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

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	openai "github.com/sashabaranov/go-openai"
	llm_openai "github.com/swuecho/chat_backend/llm/openai"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// OpenAI ChatModel implementation
type OpenAIChatModel struct {
	h *ChatHandler
}

func (m *OpenAIChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chatCompletionMessages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := m.h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return nil, eris.New("exceed per mode rate limit")
	}

	chatModel, err := GetChatModel(m.h.service.q, chatSession.Model)
	if err != nil {
		return nil, err
	}

	config, err := genOpenAIConfig(*chatModel)
	log.Printf("%+v", config.String())
	// print all config details
	if err != nil {
		return nil, ErrOpenAIConfigFailed.WithMessage("Failed to generate OpenAI config").WithDebugInfo(err.Error())
	}

	chatFiles, err := GetChatFiles(m.h.chatfileService.q, chatSession.Uuid)
	if err != nil {
		return nil, err
	}

	openaiReq := NewChatCompletionRequest(chatSession, chatCompletionMessages, chatFiles, streamOutput)
	if len(openaiReq.Messages) <= 1 {
		return nil, ErrSystemMessageError
	}
	log.Printf("OpenAI request prepared - Model: %s, MessageCount: %d, Temperature: %.2f",
		openaiReq.Model, len(openaiReq.Messages), openaiReq.Temperature)
	client := openai.NewClientWithConfig(config)
	if streamOutput {
		return doChatStream(w, client, openaiReq, chatSession.N, chatUuid, regenerate, m.h)
	} else {
		return handleRegularResponse(w, client, openaiReq)
	}

}

func handleRegularResponse(w http.ResponseWriter, client *openai.Client, req openai.ChatCompletionRequest) (*models.LLMAnswer, error) {
	// check per chat_model limit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	completion, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Printf("fail to do request: %+v", err)
		return nil, ErrOpenAIRequestFailed.WithMessage("Failed to create chat completion").WithDebugInfo(err.Error())
	}
	log.Printf("completion: %+v", completion)
	data, _ := json.Marshal(completion)
	fmt.Fprint(w, string(data))
	return &models.LLMAnswer{Answer: completion.Choices[0].Message.Content, AnswerId: completion.ID}, nil
}

// doChatStream handles streaming chat completion responses from OpenAI
// It properly manages thinking tags for models that support reasoning content
func doChatStream(w http.ResponseWriter, client *openai.Client, req openai.ChatCompletionRequest, bufferLen int32, chatUuid string, regenerate bool, handler *ChatHandler) (*models.LLMAnswer, error) {
	// Use request context with timeout, but prioritize client cancellation
	baseCtx := context.Background()
	if handler != nil {
		baseCtx = handler.GetRequestContext()
	}
	ctx, cancel := context.WithTimeout(baseCtx, 5*time.Minute)
	defer cancel()

	log.Print("Creating OpenAI stream")
	stream, err := client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		log.Printf("fail to do request: %+v", err)
		return nil, ErrOpenAIStreamFailed.WithMessage("Failed to create chat completion stream").WithDebugInfo(err.Error())
	}
	defer func() {
		if err := stream.Close(); err != nil {
			log.Printf("Error closing OpenAI stream: %v", err)
		}
	}()

	// Setup Server-Sent Events (SSE) streaming
	flusher, err := setupSSEStream(w)
	if err != nil {
		return nil, APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		}
	}

	// Initialize streaming state
	var answer_id string
	
	var hasReason bool           // Whether we've detected any reasoning content
	var reasonTagOpened bool     // Whether we've sent the opening <think> tag
	var reasonTagClosed bool     // Whether we've sent the closing </think> tag
	
	// Ensure minimum buffer length
	if bufferLen == 0 {
		log.Println("Buffer length is 0, setting to 1")
		bufferLen = 1
	}
	
	// Initialize buffers for accumulating content
	textBuffer := newTextBuffer(bufferLen, "", "")
	reasonBuffer := newTextBuffer(bufferLen, "<think>\n\n", "\n\n</think>\n\n")
	answer_id = GenerateAnswerID(chatUuid, regenerate)
	// Main streaming loop
	for {
		// Check if client disconnected or context was cancelled
		select {
		case <-ctx.Done():
			log.Printf("Stream cancelled by client: %v", ctx.Err())
			// Return current accumulated content when cancelled
			llmAnswer := models.LLMAnswer{Answer: textBuffer.String("\n"), AnswerId: answer_id}
			if hasReason {
				llmAnswer.ReasoningContent = reasonBuffer.String("\n")
			}
			return &llmAnswer, nil
		default:
		}

		rawLine, err := stream.RecvRaw()
		if err != nil {
			log.Printf("stream error: %+v", err)
			if errors.Is(err, io.EOF) {
				// Stream ended successfully - return accumulated content
				llmAnswer := models.LLMAnswer{Answer: textBuffer.String("\n"), AnswerId: answer_id}
				if hasReason {
					llmAnswer.ReasoningContent = reasonBuffer.String("\n")
				}
				return &llmAnswer, nil
			} else {
				log.Printf("Stream error: %v", err)
				return nil, ErrOpenAIStreamFailed.WithMessage("Stream error occurred").WithDebugInfo(err.Error())
			}
		}
		// Parse the streaming response
		response := llm_openai.ChatCompletionStreamResponse{}
		err = json.Unmarshal(rawLine, &response)
		if err != nil {
			log.Printf("Could not unmarshal response: %v\n", err)
			continue
		}
		
		// Extract delta content from the response
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Delta
		
		// Accumulate content in buffers (for final answer construction)
		textBuffer.appendByIndex(textIdx, delta.Content)
		if len(delta.ReasoningContent) > 0 {
			hasReason = true
			reasonBuffer.appendByIndex(textIdx, delta.ReasoningContent)
		}

		// Set answer ID from response if not already set
		if answer_id == "" {
			answer_id = strings.TrimPrefix(response.ID, "chatcmpl-")
		}

		// Process and send delta content
		if len(delta.Content) > 0 || len(delta.ReasoningContent) > 0 {
			deltaToSend := processDelta(delta, &reasonTagOpened, &reasonTagClosed, hasReason)
			if len(deltaToSend) > 0 {
				log.Printf("delta: %s", deltaToSend)
				err := FlushResponse(w, flusher, StreamingResponse{
					AnswerID: answer_id,
					Content:  deltaToSend,
					IsFinal:  false,
				})
				if err != nil {
					log.Printf("Failed to flush response: %v", err)
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
		log.Printf("messages: %+v\n", string(b))
	}

	log.Printf("messages: %+v\n", openaiMessages)
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

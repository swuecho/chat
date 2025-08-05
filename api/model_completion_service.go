package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/rotisserie/eris"
	openai "github.com/sashabaranov/go-openai"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// CompletionChatModel implements ChatModel interface for OpenAI completion models
type CompletionChatModel struct {
	h *ChatHandler
}

// Stream implements the ChatModel interface for completion model scenarios
func (m *CompletionChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	// Get request context for cancellation support
	ctx := m.h.GetRequestContext()
	return m.completionStream(ctx, w, chatSession, chat_completion_messages, chatUuid, regenerate, stream)
}

// completionStream handles streaming for OpenAI completion models
func (m *CompletionChatModel) completionStream(ctx context.Context, w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_completion_messages []models.Message, chatUuid string, regenerate bool, _ bool) (*models.LLMAnswer, error) {
	// Check per chat_model rate limit
	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := m.h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return nil, eris.New("exceed per mode rate limit")
	}

	// Get chat model configuration
	chatModel, err := GetChatModel(m.h.service.q, chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrResourceNotFound(""), "chat model "+chatSession.Model, ""))
		return nil, err
	}

	// Generate OpenAI client configuration
	config, err := genOpenAIConfig(*chatModel)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to generate OpenAI configuration", err.Error()))
		return nil, err
	}

	client := openai.NewClientWithConfig(config)

	// Get the latest message content as prompt
	prompt := chat_completion_messages[len(chat_completion_messages)-1].Content

	// Create completion request
	N := chatSession.N
	req := openai.CompletionRequest{
		Model:       chatSession.Model,
		Temperature: float32(chatSession.Temperature),
		TopP:        float32(chatSession.TopP),
		N:           int(N),
		Prompt:      prompt,
		Stream:      true,
	}

	// Create completion stream with timeout
	ctx, cancel := context.WithTimeout(context.Background(), DefaultRequestTimeout)
	defer cancel()

	stream, err := client.CreateCompletionStream(ctx, req)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Failed to create completion stream", err.Error()))
		return nil, err
	}
	defer stream.Close()

	// Setup SSE streaming
	flusher, err := setupSSEStream(w)
	if err != nil {
		RespondWithAPIError(w, createAPIError(ErrInternalUnexpected, "Streaming unsupported by client", err.Error()))
		return nil, err
	}

	var answer string
	answer_id := GenerateAnswerID(chatUuid, regenerate)
	textBuffer := newTextBuffer(N, "```\n"+prompt, "\n```\n")

	// Process streaming response
	for {
		// Check if client disconnected or context was cancelled
		select {
		case <-ctx.Done():
			log.Printf("Completion stream cancelled by client: %v", ctx.Err())
			// Return current accumulated content when cancelled
			return &models.LLMAnswer{Answer: answer, AnswerId: answer_id}, nil
		default:
		}

		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// Send the final message
			if len(answer) > 0 {
				err := FlushResponse(w, flusher, StreamingResponse{
					AnswerID: answer_id,
					Content:  answer,
					IsFinal:  true,
				})
				if err != nil {
					log.Printf("Failed to flush final response: %v", err)
				}
			}

			// Include debug information if enabled
			if chatSession.Debug {
				req_j, _ := json.Marshal(req)
				log.Println(string(req_j))
				answer = answer + "\n" + string(req_j)
				err := FlushResponse(w, flusher, StreamingResponse{
					AnswerID: answer_id,
					Content:  answer,
					IsFinal:  true,
				})
				if err != nil {
					log.Printf("Failed to flush debug response: %v", err)
				}
			}
			break
		}

		if err != nil {
			RespondWithAPIError(w, ErrChatStreamFailed.WithMessage("Stream error occurred").WithDebugInfo(err.Error()))
			return nil, err
		}

		// Process response chunk
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Text
		textBuffer.appendByIndex(textIdx, delta)

		if chatSession.Debug {
			log.Printf("%d: %s", textIdx, delta)
		}

		if answer_id == "" {
			answer_id = response.ID
		}

		// Concatenate all string builders into a single string
		answer = textBuffer.String("\n\n")

		// Determine when to flush the response
		perWordStreamLimit := getPerWordStreamLimit()
		if strings.HasSuffix(delta, "\n") || len(answer) < perWordStreamLimit {
			if len(answer) == 0 {
				log.Print(ErrorNoContent)
			} else {
				err := FlushResponse(w, flusher, StreamingResponse{
					AnswerID: answer_id,
					Content:  answer,
					IsFinal:  false,
				})
				if err != nil {
					log.Printf("Failed to flush response: %v", err)
				}
			}
		}
	}

	return &models.LLMAnswer{AnswerId: answer_id, Answer: answer}, nil
}

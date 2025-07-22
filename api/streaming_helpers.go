// Package main provides streaming utilities for chat responses.
// This file contains common streaming functionality to reduce code duplication
// across different model service implementations.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// constructChatCompletionStreamResponse creates an OpenAI chat completion stream response
func constructChatCompletionStreamResponse(answerID string, content string) openai.ChatCompletionStreamResponse {
	resp := openai.ChatCompletionStreamResponse{
		ID: answerID,
		Choices: []openai.ChatCompletionStreamChoice{
			{
				Index: 0,
				Delta: openai.ChatCompletionStreamChoiceDelta{
					Content: content,
				},
			},
		},
	}
	return resp
}

// StreamingResponse represents a common streaming response structure
type StreamingResponse struct {
	AnswerID string
	Content  string
	IsFinal  bool
}

// FlushResponse sends a streaming response to the client
func FlushResponse(w http.ResponseWriter, flusher http.Flusher, response StreamingResponse) error {
	if response.Content == "" && !response.IsFinal {
		return nil // Skip empty non-final responses
	}

	streamResponse := constructChatCompletionStreamResponse(response.AnswerID, response.Content)
	data, err := json.Marshal(streamResponse)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "data: %v\n\n", string(data))
	flusher.Flush()
	return nil
}

// ShouldFlushContent determines when to flush content based on common rules
func ShouldFlushContent(content string, lastFlushLength int, isSmallContent bool) bool {
	return strings.Contains(content, "\n") ||
		(isSmallContent && len(content) < SmallAnswerThreshold) ||
		(len(content)-lastFlushLength) >= FlushCharacterThreshold
}

// SetStreamingHeaders sets common headers for streaming responses
func SetStreamingHeaders(req *http.Request) {
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Set("Accept", AcceptEventStream)
	req.Header.Set("Cache-Control", CacheControlNoCache)
	req.Header.Set("Connection", ConnectionKeepAlive)
}

// GenerateAnswerID creates a new answer ID if not provided in regenerate mode
func GenerateAnswerID(chatUuid string, regenerate bool) string {
	if regenerate {
		return chatUuid
	}
	return NewUUID()
}

// GetChatModel retrieves a chat model by name with consistent error handling
func GetChatModel(queries *sqlc_queries.Queries, modelName string) (*sqlc_queries.ChatModel, error) {
	chatModel, err := queries.ChatModelByName(context.Background(), modelName)
	if err != nil {
		return nil, ErrResourceNotFound("chat model: " + modelName)
	}
	return &chatModel, nil
}

// GetChatFiles retrieves chat files for a session with consistent error handling
func GetChatFiles(queries *sqlc_queries.Queries, sessionUUID string) ([]sqlc_queries.ChatFile, error) {
	chatFiles, err := queries.ListChatFilesWithContentBySessionUUID(context.Background(), sessionUUID)
	if err != nil {
		return nil, ErrInternalUnexpected.WithMessage("Failed to get chat files").WithDebugInfo(err.Error())
	}
	return chatFiles, nil
}

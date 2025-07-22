// Package main provides streaming utilities for chat responses.
// This file contains common streaming functionality to reduce code duplication
// across different model service implementations.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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
		   (len(content) - lastFlushLength) >= FlushCharacterThreshold
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
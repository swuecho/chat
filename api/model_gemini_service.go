package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/swuecho/chat_backend/llm/gemini"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

// curl https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=$API_KEY \
//     -H 'Content-Type: application/json' \
//     -X POST \
//     -d '{
//       "contents": [{
//         "parts":[{
//           "text": "Write a story about a magic backpack."}]}]}' 2> /dev/null

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
	h      *ChatHandler
	client *GeminiClient
}

func NewGeminiChatModel(h *ChatHandler) *GeminiChatModel {
	return &GeminiChatModel{
		h:      h,
		client: NewGeminiClient(),
	}
}

func (m *GeminiChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	answerID := chatUuid
	if !regenerate {
		answerID = NewUUID()
	}

	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to get chat files").WithDebugInfo(err.Error())
	}

	payloadBytes, err := gemini.GenGemminPayload(messages, chatFiles)
	if err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to generate Gemini payload").WithDebugInfo(err.Error())
	}

	url := m.buildAPIURL(chatSession.Model, stream)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to create Gemini API request").WithDebugInfo(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	if stream {
		return m.handleStreamResponse(w, req, answerID)
	}
	return m.handleRegularResponse(w, req, answerID)
}

func (m *GeminiChatModel) buildAPIURL(model string, stream bool) string {
	endpoint := "generateContent"
	if stream {
		endpoint = "streamGenerateContent?alt=sse"
	}
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:%s&key=$GEMINI_API_KEY", model, endpoint)
	return os.ExpandEnv(url)
}

func (m *GeminiChatModel) handleRegularResponse(w http.ResponseWriter, req *http.Request, answerID string) (*models.LLMAnswer, error) {
	resp, err := m.client.client.Do(req)
	if err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to send Gemini API request").WithDebugInfo(err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to read Gemini response").WithDebugInfo(err.Error())
	}

	var geminiResp gemini.ResponseBody
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to parse Gemini response").WithDebugInfo(err.Error())
	}

	answer := geminiResp.Candidates[0].Content.Parts[0].Text
	response := constructChatCompletionStreamReponse(answerID, answer)
	data, _ := json.Marshal(response)
	fmt.Fprint(w, string(data))

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answerID,
	}, nil
}

func (m *GeminiChatModel) handleStreamResponse(w http.ResponseWriter, req *http.Request, answerID string) (*models.LLMAnswer, error) {
	resp, err := m.client.client.Do(req)
	if err != nil {
		return nil, ErrInternalUnexpected.WithDetail("Failed to send Gemini API request").WithDebugInfo(err.Error())
	}
	defer resp.Body.Close()

	setSSEHeader(w)
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		}
	}

	var answer string
	ioreader := bufio.NewReader(resp.Body)
	headerData := []byte("data: ")

	for count := 0; count < 10000; count++ {
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return &models.LLMAnswer{
					Answer:   answer,
					AnswerId: answerID,
				}, nil
			}
			return nil, ErrInternalUnexpected.WithDetail("Error reading stream").WithDebugInfo(err.Error())
		}

		if !bytes.HasPrefix(line, headerData) {
			continue
		}

		line = bytes.TrimPrefix(line, headerData)
		if len(line) > 0 {
			answer = gemini.ParseRespLine(line, answer)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answerID, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}

	return &models.LLMAnswer{
		AnswerId: answerID,
		Answer:   answer,
	}, nil
}

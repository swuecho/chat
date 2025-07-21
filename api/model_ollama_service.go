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
	"strings"
	"time"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Ollama ChatModel implementation
type OllamaChatModel struct {
	h *ChatHandler
}

func (m *OllamaChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	return m.h.chatOllamStream(w, chatSession, chat_compeletion_messages, chatUuid, regenerate)
}

func (h *ChatHandler) chatOllamStream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// set the api key
	chatModel, err := h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}
	jsonData := map[string]interface{}{
		"model":    strings.Replace(chatSession.Model, "ollama-", "", 1),
		"messages": chat_compeletion_messages,
	}
	// convert data to json format
	jsonValue, _ := json.Marshal(jsonData)
	// create the request
	req, err := http.NewRequest("POST", chatModel.Url, bytes.NewBuffer(jsonValue))

	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to make request").WithDebugInfo(err.Error()))
		return nil, err
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)
	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")

	// set the streaming flag
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Access-Control-Allow-Origin", "*")

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithMessage("Failed to create chat completion stream").WithDebugInfo(err.Error()))
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		RespondWithAPIError(w, APIError{
			HTTPCode: http.StatusInternalServerError,
			Code:     "STREAM_UNSUPPORTED",
			Message:  "Streaming unsupported by client",
		})
		return nil, err
	}

	var answer string
	var answer_id string

	if regenerate {
		answer_id = chatUuid
	}

	count := 0
	for {
		count++
		// prevent infinite loop
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("End of stream reached")
				break // Exit loop if end of stream
			}
			return nil, err
		}
		var streamResp OllamaResponse
		err = json.Unmarshal(line, &streamResp)
		if err != nil {
			return nil, err
		}
		delta := strings.ReplaceAll(streamResp.Message.Content, "<0x0A>", "\n")
		answer += delta // Still accumulate for final answer storage
		
		if streamResp.Done {
			// stream.isFinished = true
			fmt.Println("DONE break")
			// No need to send full content at the end since we're sending deltas
			break
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}

		// Send delta content immediately when available
		if len(delta) > 0 {
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, delta))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}

	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

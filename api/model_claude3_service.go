package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
	claude "github.com/swuecho/chat_backend/llm/claude"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Claude3 ChatModel implementation
type Claude3ChatModel struct {
	h *ChatHandler
}

func (m *Claude3ChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error) {
	ctx := context.Background()

	// Get chat model configuration
	chatModel, err := m.h.service.q.ChatModelByName(ctx, chatSession.Model)
	if err != nil {
		return nil, ErrChatModelNotFound.WithDetail(err.Error())
	}

	// Get chat files if any
	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(ctx, chatSession.Uuid)
	if err != nil {
		return nil, ErrDatabaseQuery.WithDetail(err.Error())
	}

	// create a new strings.Builder
	// iterate through the messages and format them
	// print the user's question
	// convert assistant's response to json format
	//     "messages": [
	//	{"role": "user", "content": "Hello, world"}
	//	]
	// first message is user instead of system
	var messages []openai.ChatCompletionMessage
	if len(chat_compeletion_messages) > 1 {
		// first message used as system message
		// messages start with second message
		// drop the first assistant message if it is an assistant message
		claude_messages := chat_compeletion_messages[1:]

		if len(claude_messages) > 0 && claude_messages[0].Role == "assistant" {
			claude_messages = claude_messages[1:]
		}
		messages = messagesToOpenAIMesages(claude_messages, chatFiles)
	} else {
		// only system message, return and do nothing
		return nil, ErrSystemMessageError
	}
	// Prepare request payload
	jsonData := map[string]interface{}{
		"system":      chat_compeletion_messages[0].Content,
		"model":       chatSession.Model,
		"messages":    messages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"top_p":       chatSession.TopP,
		"stream":      stream,
	}

	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return nil, ErrValidationInvalidInputGeneric.WithDetail("failed to marshal request payload").WithDebugInfo(err.Error())
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", chatModel.Url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, ErrClaudeRequestFailed.WithDetail("failed to create HTTP request").WithDebugInfo(err.Error())
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)

	if apiKey == "" {
		return nil, ErrAuthInvalidCredentials.WithDetail(fmt.Sprintf("missing API key for model %s", chatSession.Model))
	}

	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	if !stream {
		req.Header.Set("Accept", "application/json")
		client := http.Client{
			Timeout: 5 * time.Minute,
		}

		llmAnswer, err := doGenerateClaude3(client, req)
		if err != nil {
			return nil, ErrClaudeRequestFailed.WithDetail("failed to generate response").WithDebugInfo(err.Error())
		}

		answerResponse := constructChatCompletionStreamReponse(llmAnswer.AnswerId, llmAnswer.Answer)
		data, err := json.Marshal(answerResponse)
		if err != nil {
			return nil, ErrInternalUnexpected.WithDetail("failed to marshal response").WithDebugInfo(err.Error())
		}

		if _, err := fmt.Fprint(w, string(data)); err != nil {
			return nil, ErrClaudeResponseFaild.WithDetail("failed to write response").WithDebugInfo(err.Error())
		}

		return llmAnswer, nil
	}

	// Handle streaming response
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	llmAnswer, err := m.h.chatStreamClaude3(w, req, chatUuid, regenerate)
	if err != nil {
		return nil, ErrClaudeStreamFailed.WithDetail("failed to stream response").WithDebugInfo(err.Error())
	}
	return llmAnswer, nil
}

func doGenerateClaude3(client http.Client, req *http.Request) (*models.LLMAnswer, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrClaudeRequestFailed.WithMessage("Failed to process Claude request").WithDebugInfo(err.Error())
	}
	// Unmarshal directly from resp.Body
	var message claude.Response
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		return nil, ErrClaudeInvalidResponse.WithMessage("Failed to unmarshal Claude response").WithDebugInfo(err.Error())
	}
	defer resp.Body.Close()
	uuid := message.ID
	firstMessage := message.Content[0].Text

	return &models.LLMAnswer{
		AnswerId: uuid,
		Answer:   firstMessage,
	}, nil
}

// claude-3-opus-20240229
// claude-3-sonnet-20240229
// claude-3-haiku-20240307
func (h *ChatHandler) chatStreamClaude3(w http.ResponseWriter, req *http.Request, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, ErrClaudeRequestFailed.WithMessage("Failed to process Claude streaming request").WithDebugInfo(err.Error())
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

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
	var answer_id string

	if regenerate {
		answer_id = chatUuid
	}

	var headerData = []byte("data: ")
	count := 0
	for {
		count++
		// prevent infinite loop
		if count > 10000 {
			break
		}
		line, err := ioreader.ReadBytes('\n')
		log.Printf("%+v", string(line))
		if err != nil {
			if errors.Is(err, io.EOF) {
				if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
					log.Println(string(line))
					data, _ := json.Marshal(constructChatCompletionStreamReponse(NewUUID(), string(line)))
					fmt.Fprintf(w, "data: %v\n\n", string(data))
					flusher.Flush()
				}
				fmt.Println("End of stream reached")
				return nil, err
			}
			return nil, err
		}
		line = bytes.TrimPrefix(line, headerData)

		if bytes.HasPrefix(line, []byte("event: message_stop")) {
			// stream.isFinished = true
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
			break
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"error\"")) {
			log.Println(string(line))
			return nil, ErrClaudeStreamFailed.WithMessage("Error in Claude API response").WithDebugInfo(string(line))
		}
		if answer_id == "" {
			answer_id = NewUUID()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_start\"")) {
			answer = claude.AnswerFromBlockStart(line)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
		if bytes.HasPrefix(line, []byte("{\"type\":\"content_block_delta\"")) {
			answer += claude.AnswerFromBlockDelta(line)
			data, _ := json.Marshal(constructChatCompletionStreamReponse(answer_id, answer))
			fmt.Fprintf(w, "data: %v\n\n", string(data))
			flusher.Flush()
		}
	}
	return &models.LLMAnswer{
		Answer:   answer,
		AnswerId: answer_id,
	}, nil
}

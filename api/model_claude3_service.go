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
	// Obtain the API token (buffer 1, send to channel will block if there is a token in the buffer)
	log.Printf("%+v", chatSession)
	// Release the API token
	// set the api key
	chatModel, err := m.h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	log.Printf("%+v", chatModel)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat model: "+chatSession.Model))
		return nil, err
	}
	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		RespondWithAPIError(w, ErrResourceNotFound("chat files "+chatSession.Uuid))
		return nil, err
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
		RespondWithAPIError(w, ErrSystemMessageError)
		return nil, err
	}
	// create the json data
	jsonData := map[string]interface{}{
		"system":      chat_compeletion_messages[0].Content,
		"model":       chatSession.Model,
		"messages":    messages,
		"max_tokens":  chatSession.MaxTokens,
		"temperature": chatSession.Temperature,
		"top_p":       chatSession.TopP,
		"stream":      stream,
	}
	log.Printf("%+v", jsonData)

	// convert data to json format
	jsonValue, _ := json.Marshal(jsonData)
	log.Printf("%+v", string(jsonValue))
	// create the request
	req, err := http.NewRequest("POST", chatModel.Url, bytes.NewBuffer(jsonValue))

	if err != nil {
		log.Printf("%+v", err)
		RespondWithAPIError(w, ErrChatStreamFailed.WithDebugInfo(err.Error()))
		return nil, err
	}

	// add headers to the request
	apiKey := os.Getenv(chatModel.ApiAuthKey)
	authHeaderName := chatModel.ApiAuthHeader
	if authHeaderName != "" {
		req.Header.Set(authHeaderName, apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	if !stream {
		req.Header.Set("Accept", "application/json")
		return doGenerateClaude3(w, req)
	} else {
		// set the streaming flag
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
		return m.h.chatStreamClaude3(w, req, chatUuid, regenerate)
	}
}

func doGenerateClaude3(w http.ResponseWriter, req *http.Request) (*models.LLMAnswer, error) {

	// create the http client and send the request
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%+v", err)
		RespondWithAPIError(w, ErrChatStreamFailed.WithDebugInfo(err.Error()))
		return nil, err
	}

	// Unmarshal directly from resp.Body
	var message claude.Response
	if err := json.NewDecoder(resp.Body).Decode(&message); err != nil {
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to unmarshal response").WithDebugInfo(err.Error()))
		return nil, err
	}
	defer resp.Body.Close()
	uuid := message.ID
	firstMessage := message.Content[0].Text
	answer := constructChatCompletionStreamReponse(uuid, firstMessage)
	data, _ := json.Marshal(answer)
	fmt.Fprint(w, string(data))
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
		log.Printf("%+v", err)
		RespondWithAPIError(w, ErrInternalUnexpected.WithDetail("Failed to process request").WithDebugInfo(err.Error()))
		return nil, err
	}

	ioreader := bufio.NewReader(resp.Body)

	// read the response body
	defer resp.Body.Close()
	// loop over the response body and print data

	setSSEHeader(w)

	flusher, ok := w.(http.Flusher)
	if !ok {
		apiErr := ErrInternalUnexpected
		apiErr.Detail = "Streaming unsupported by the client"
		RespondWithAPIError(w, apiErr)
		return nil, err
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
			RespondWithAPIError(w, ErrChatStreamFailed.WithDetail("Error in Claude API response").WithDebugInfo(string(line)))
			return nil, err
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

type OllamaResponse struct {
	Model              string         `json:"model"`
	CreatedAt          time.Time      `json:"created_at"`
	Done               bool           `json:"done"`
	Message            models.Message `json:"message"`
	TotalDuration      int64          `json:"total_duration"`
	LoadDuration       int64          `json:"load_duration"`
	PromptEvalCount    int            `json:"prompt_eval_count"`
	PromptEvalDuration int64          `json:"prompt_eval_duration"`
	EvalCount          int            `json:"eval_count"`
	EvalDuration       int64          `json:"eval_duration"`
}

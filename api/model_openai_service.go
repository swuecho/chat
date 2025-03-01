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

func (m *OpenAIChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chatMessages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	ctx := context.Background()
	
	// Rate limiting and access control
	openAIRateLimiter.Wait(ctx)
	if m.h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID) {
		return nil, fmt.Errorf("rate limit exceeded for model %s", chatSession.Model)
	}

	// Get chat model configuration
	chatModel, err := m.h.service.q.ChatModelByName(ctx, chatSession.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat model %s: %w", chatSession.Model, err)
	}

	config, err := genOpenAIConfig(chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate OpenAI config: %w", err)
	}

	// Get chat files if any
	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(ctx, chatSession.Uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat files: %w", err)
	}

	// Create OpenAI request
	openaiReq := NewChatCompletionRequest(chatSession, chatMessages, chatFiles, streamOutput)
	if len(openaiReq.Messages) <= 1 {
		return nil, fmt.Errorf("insufficient messages for completion")
	}

	client := openai.NewClientWithConfig(config)
	
	if streamOutput {
		return doChatStream(w, client, openaiReq, chatSession.N, chatUuid, regenerate)
	}
	return doGenerate(w, client, openaiReq)
}

func doGenerate(w http.ResponseWriter, client *openai.Client, req openai.ChatCompletionRequest) (*models.LLMAnswer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	completion, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(completion.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	// Write response
	data, err := json.Marshal(completion)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal completion: %w", err)
	}
	
	if _, err := fmt.Fprint(w, string(data)); err != nil {
		return nil, fmt.Errorf("failed to write response: %w", err)
	}

	return &models.LLMAnswer{
		Answer:   completion.Choices[0].Message.Content,
		AnswerId: completion.ID,
	}, nil
}

func doChatStream(w http.ResponseWriter, client *openai.Client, req openai.ChatCompletionRequest, bufferLen int32, chatUuid string, regenerate bool) (*models.LLMAnswer, error) {
	// check per chat_model limit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Print("before request")
	stream, err := client.CreateChatCompletionStream(ctx, req)

	if err != nil {
		log.Printf("fail to do request: %+v", err)
		return nil, ErrOpenAIStreamFailed.WithMessage("Failed to create chat completion stream").WithDebugInfo(err.Error())
	}
	defer stream.Close()

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
	var hasReason bool
	if bufferLen == 0 {
		log.Println("chatSession.N is 0")
		bufferLen += 1
	}
	textBuffer := newTextBuffer(bufferLen, "", "")
	reasonBuffer := newTextBuffer(bufferLen, "<think>\n\n", "\n\n</think>\n\n")
	if regenerate {
		answer_id = chatUuid
	}
	for {
		rawLine, err := stream.RecvRaw()
		if err != nil {
			log.Printf("stream error: %+v", err)
			if errors.Is(err, io.EOF) {
				// send the last message
				if len(answer) > 0 {
					final_resp := constructChatCompletionStreamReponse(answer_id, answer)
					data, _ := json.Marshal(final_resp)
					fmt.Fprintf(w, "data: %v\n\n", string(data))
					flusher.Flush()
				}

				// no reason in the answer (so do not disrupt the context)
				llmAnswer := models.LLMAnswer{Answer: textBuffer.String("\n"), AnswerId: answer_id}
				if hasReason {
					llmAnswer.ReasoningContent = reasonBuffer.String("\n")
				}
				return &llmAnswer, nil
			} else {
				log.Printf("%v", err)
				return nil, ErrOpenAIStreamFailed.WithMessage("Stream error occurred").WithDebugInfo(err.Error())
			}
		}
		response := llm_openai.ChatCompletionStreamResponse{}
		err = json.Unmarshal(rawLine, &response)
		if err != nil {
			log.Printf("Could not unmarshal response: %v\n", err)
			continue
		}
		textIdx := response.Choices[0].Index
		delta := response.Choices[0].Delta
		textBuffer.appendByIndex(textIdx, delta.Content)
		if len(delta.ReasoningContent) > 0 {
			hasReason = true
			reasonBuffer.appendByIndex(textIdx, delta.ReasoningContent)
		}

		if hasReason {
			answer = reasonBuffer.String("\n") + textBuffer.String("\n")
		} else {
			answer = textBuffer.String("\n")
		}
		if answer_id == "" {
			answer_id = strings.TrimPrefix(response.ID, "chatcmpl-")
		}
		perWordStreamLimit := getPerWordStreamLimit()

		if strings.HasSuffix(answer, "\n") || len(answer) < perWordStreamLimit {
			if len(answer) == 0 {
				log.Printf("%s", "no content in answer")
			} else {
				constructedResponse := constructChatCompletionStreamReponse(answer_id, answer)
				data, _ := json.Marshal(constructedResponse)
				fmt.Fprintf(w, "data: %v\n\n", string(data))
				flusher.Flush()
			}
		}
	}
}

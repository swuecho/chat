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

func (m *OpenAIChatModel) Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, streamOutput bool) (*models.LLMAnswer, error) {
	openAIRateLimiter.Wait(context.Background())

	exceedPerModeRateLimitOrError := m.h.CheckModelAccess(w, chatSession.Uuid, chatSession.Model, chatSession.UserID)
	if exceedPerModeRateLimitOrError {
		return nil, eris.New("exceed per mode rate limit")
	}

	chatModel, err := m.h.service.q.ChatModelByName(context.Background(), chatSession.Model)
	if err != nil {
		return nil, ErrResourceNotFound("chat model: " + chatSession.Model)
	}

	config, err := genOpenAIConfig(chatModel)
	log.Printf("%+v", config.String())
	// print all config details
	if err != nil {
		return nil, ErrOpenAIConfigFailed.WithMessage("Failed to generate OpenAI config").WithDebugInfo(err.Error())
	}

	chatFiles, err := m.h.chatfileService.q.ListChatFilesWithContentBySessionUUID(context.Background(), chatSession.Uuid)
	if err != nil {
		return nil, ErrInternalUnexpected.WithMessage("Failed to get chat files").WithDebugInfo(err.Error())
	}

	openai_req := NewChatCompletionRequest(chatSession, chat_compeletion_messages, chatFiles, streamOutput)
	if len(openai_req.Messages) <= 1 {
		return nil, ErrSystemMessageError
	}
	log.Printf("%+v", openai_req)
	client := openai.NewClientWithConfig(config)
	if streamOutput {
		return doChatStream(w, client, openai_req, chatSession.N, chatUuid, regenerate)
	} else {
		return handleRegularResponse(w, client, openai_req)
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

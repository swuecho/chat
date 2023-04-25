package main

import (
	"strings"

	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
)

func messagesToOpenAIMesages(messages []Message) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m Message, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	})
	return open_ai_msgs
}


// in adminn panel the config is full url https://api.openai.com/v1/chat/completions
func getModelBaseUrl(model_url string) string {
	var baseUrl string
	if chat_index := strings.Index(model_url, "/chat/"); chat_index != -1 {
		baseUrl = model_url[:chat_index]
	} else {
		baseUrl = model_url
	}
	return baseUrl
}
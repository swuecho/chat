package main

import (
	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
)

func messagesToOpenAIMesages(messages []Message) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m Message, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
	})
	return open_ai_msgs
}

// func sqlChatMessagesToOpenAIMesages(messages []sqlc_queries.ChatMessage) []openai.ChatCompletionMessage {
// 	open_ai_msgs := lo.Map(messages, func(m sqlc_queries.ChatMessage, _ int) openai.ChatCompletionMessage {
// 		return openai.ChatCompletionMessage{Role: m.Role, Content: m.Content}
// 	})
// 	return open_ai_msgs
// }

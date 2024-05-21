package main

import (
	"fmt"
	"strings"
)

func formatClaudePrompt(chat_compeletion_messages []Message) string {
	var sb strings.Builder

	for _, message := range chat_compeletion_messages {

		if message.Role != "assistant" {
			sb.WriteString(fmt.Sprintf("\n\nHuman: %s\n\nAssistant: ", message.Content))
		} else {

			sb.WriteString(fmt.Sprintf("%s\n", message.Content))
		}
	}
	prompt := sb.String()
	return prompt
}

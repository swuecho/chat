package main

import (
	"fmt"
	"strings"
)

func formatMinstralPrompt(chat_compeletion_messages []Message) string {
	var sb strings.Builder

	for _, message := range chat_compeletion_messages {

		if message.Role != "assistant" {
			sb.WriteString(fmt.Sprintf("<s>[INST] %s[/INST]\n", message.Content))
		} else {
			sb.WriteString(fmt.Sprintf("%s </s> \n", message.Content))
		}
	}
	prompt := sb.String()
	print(prompt)
	return prompt
}

// ### Prompt Format
// ### System:
// {system}
// ### User:
// {usr}
// ### Assistant:
func formatNeuralChatPrompt(chat_compeletion_messages []Message) string {
	var sb strings.Builder

	for _, message := range chat_compeletion_messages {
		if message.Role == "system" {
			sb.WriteString(fmt.Sprintf("### System:\n%s\n", message.Content))
		} else if message.Role == "user" {
			sb.WriteString(fmt.Sprintf("### User:\n%s\n", message.Content))
		} else {
			sb.WriteString(fmt.Sprintf("### Assistant:\n%s\n", message.Content))
		}
	}
	sb.WriteString("### Assitant:\n")
	prompt := sb.String()
	print(prompt)
	return prompt
}

package claude

import (
	"encoding/json"
	"fmt"
	"strings"

	models "github.com/swuecho/chat_backend/models"
)

type Delta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ContentBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta Delta  `json:"delta"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type StartBlock struct {
	Type         string       `json:"type"`
	Index        int          `json:"index"`
	ContentBlock ContentBlock `json:"content_block"`
}

func AnswerFromBlockDelta(line []byte) string {
	var response ContentBlockDelta
	_ = json.Unmarshal(line, &response)
	return response.Delta.Text
}

func AnswerFromBlockStart(line []byte) string {
	var response StartBlock
	_ = json.Unmarshal(line, &response)
	return response.ContentBlock.Text
}

func FormatClaudePrompt(chat_compeletion_messages []models.Message) string {
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

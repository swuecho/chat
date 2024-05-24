package gemini 

import (
	"encoding/json"
	"fmt"
	"log"
	models "github.com/swuecho/chat_backend/models"
)

type Part struct {
	Text string `json:"text"`
}

type GeminiMessage struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type GeminPayload struct {
	Contents []GeminiMessage `json:"contents"`
}

type Content struct {
	Parts []struct {
		Text string `json:"text"`
	} `json:"parts"`
	Role string `json:"role"`
}

type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type Candidate struct {
	Content       Content        `json:"content"`
	FinishReason  string         `json:"finishReason"`
	Index         int            `json:"index"`
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

type PromptFeedback struct {
	SafetyRatings []SafetyRating `json:"safetyRatings"`
}

type ResponseBody struct {
	Candidates     []Candidate    `json:"candidates"`
	PromptFeedback PromptFeedback `json:"promptFeedback"`
}

func ParseRespLine(line []byte, answer string) string {
	var resp ResponseBody
	if err := json.Unmarshal(line, &resp); err != nil {
		fmt.Println("Failed to parse request body:", err)
	}

	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			answer += part.Text
		}

	}
	return answer
}

func GenGemminPayload(chat_compeletion_messages []models.Message) ([]byte, error) {
	payload := GeminPayload{
		Contents: make([]GeminiMessage, len(chat_compeletion_messages)),
	}
	for i, message := range chat_compeletion_messages {
		geminiMessage := GeminiMessage{
			Role: message.Role,
			Parts: []Part{
				{Text: message.Content},
			},
		}
		if message.Role == "assistant" {
			geminiMessage.Role = "model"
		} else if message.Role == "system" {
			geminiMessage.Role = "user"
		}
		payload.Contents[i] = geminiMessage
	}
	payloadBytes, err := json.Marshal(payload)
	log.Printf("%s\n", string(payloadBytes))
	if err != nil {
		fmt.Println("Error marshalling payload:", err)
		// handle err
		return nil, err
	}
	return payloadBytes, nil
}

package sqlc_queries

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
)

type SimpleChatMessage struct {
	Uuid               string     `json:"uuid"`
	DateTime           string     `json:"dateTime"`
	Text               string     `json:"text"`
	Model              string     `json:"model"`
	Inversion          bool       `json:"inversion"`
	Error              bool       `json:"error"`
	Loading            bool       `json:"loading"`
	IsPin              bool       `json:"isPin"`
	IsPrompt           bool       `json:"isPrompt"`
	Artifacts          []Artifact `json:"artifacts,omitempty"`
	SuggestedQuestions []string   `json:"suggestedQuestions,omitempty"`
}

type Artifact struct {
	UUID     string `json:"uuid"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Language string `json:"language,omitempty"`
}

func (q *Queries) GetChatHistoryBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]SimpleChatMessage, error) {

	chat_prompts, err := q.GetChatPromptsBySessionUUID(ctx, uuid)
	if err != nil {
		return nil, eris.Wrap(err, "fail to get prompt: ")
	}

	simple_prompts := lo.Map(chat_prompts, func(prompt ChatPrompt, idx int) SimpleChatMessage {
		return SimpleChatMessage{
			Uuid:      prompt.Uuid,
			DateTime:  prompt.UpdatedAt.Format(time.RFC3339),
			Text:      prompt.Content,
			Inversion: idx%2 == 0,
			Error:     false,
			Loading:   false,
			IsPin:     false,
			IsPrompt:  true,
		}
	})

	messages, err := q.GetChatMessagesBySessionUUID(ctx,
		GetChatMessagesBySessionUUIDParams{
			Uuid:   uuid,
			Offset: pageNum - 1,
			Limit:  pageSize,
		})
	if err != nil {
		return nil, eris.Wrap(err, "fail to get message: ")
	}

	simple_msgs := lo.Map(messages, func(message ChatMessage, _ int) SimpleChatMessage {
		text := message.Content
		// prepend reason content
		if len(message.ReasoningContent) > 0 {
			text = message.ReasoningContent + message.Content
		}

		// Extract artifacts from database
		var artifacts []Artifact
		if message.Artifacts != nil {
			err := json.Unmarshal(message.Artifacts, &artifacts)
			if err != nil {
				// Log error but don't fail the request
				artifacts = []Artifact{}
			}
		}

		// Extract suggested questions from database
		var suggestedQuestions []string
		if message.SuggestedQuestions != nil {
			err := json.Unmarshal(message.SuggestedQuestions, &suggestedQuestions)
			if err != nil {
				// Log error but don't fail the request
				suggestedQuestions = []string{}
			}
		}

		return SimpleChatMessage{
			Uuid:               message.Uuid,
			DateTime:           message.UpdatedAt.Format(time.RFC3339),
			Text:               text,
			Model:              message.Model,
			Inversion:          message.Role == "user",
			Error:              false,
			Loading:            false,
			IsPin:              message.IsPin,
			Artifacts:          artifacts,
			SuggestedQuestions: suggestedQuestions,
		}
	})

	msgs := append(simple_prompts, simple_msgs...)
	return msgs, nil
}

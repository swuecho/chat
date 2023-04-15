package sqlc_queries

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
)

func (user *AuthUser) Role() string {
	role := "user"
	if user.IsSuperuser {
		role = "admin"
	}
	return role
}

func (m *ChatMessage) Authenticate(q Queries, userID int32) (bool, error) {
	messageID := m.ID
	ctx := context.Background()
	v, e := q.HasChatMessagePermission(ctx, HasChatMessagePermissionParams{messageID, userID})
	return v, e
}

func (s *ChatSession) Authenticate(q Queries, userID int32) (bool, error) {
	sessionID := s.ID
	ctx := context.Background()
	v, e := q.HasChatSessionPermission(ctx, HasChatSessionPermissionParams{sessionID, userID})
	return v, e
}

func (p *ChatPrompt) Authenticate(q Queries, userID int32) (bool, error) {
	sessionID := p.ID
	ctx := context.Background()
	v, e := q.HasChatPromptPermission(ctx, HasChatPromptPermissionParams{sessionID, userID})
	return v, e
}

// Create a RawMessage from ChatSession
func (cs *ChatSession) ToRawMessage() *json.RawMessage {
	// Marshal ChatSession struct to json.RawMessage
	chatSessionJSON, err := json.Marshal(cs)
	if err != nil {
		return nil
	}
	var rawMessage json.RawMessage = chatSessionJSON
	return &rawMessage
}

type MessageWithRoleAndContent interface {
	GetRole() string
	GetContent() string
}

func (m ChatMessage) GetRole() string {
	return m.Role
}

func (m ChatMessage) GetContent() string {
	return m.Content
}

func (m ChatPrompt) GetRole() string {
	return m.Role
}

func (m ChatPrompt) GetContent() string {
	return m.Content
}

func SqlChatsToOpenAIMesages(messages []MessageWithRoleAndContent) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m MessageWithRoleAndContent, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.GetRole(), Content: m.GetContent()}
	})
	return open_ai_msgs
}

func SqlChatsToOpenAIMessagesGenerics[T MessageWithRoleAndContent](messages []T) []openai.ChatCompletionMessage {
	open_ai_msgs := lo.Map(messages, func(m T, _ int) openai.ChatCompletionMessage {
		return openai.ChatCompletionMessage{Role: m.GetRole(), Content: m.GetContent()}
	})
	return open_ai_msgs
}

// TODO: How to write generics function without create new interface?

// // SumIntsOrFloats sums the values of map m. It supports both floats and integers
// // as map values.
// func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
// 	var s V
// 	for _, v := range m {
// 		s += v
// 	}
// 	return s
// }

// func ConvertToMessages[T ChatPrompt | ChatMessage](input []T) []openai.ChatCompletionMessage {
// 	// Define an empty slice to hold the converted messages
// 	output := make([]openai.ChatCompletionMessage, 0)

// 	// Loop over the input slice and convert each element to a Message
// 	for _, obj := range input {
// 		output = append(output, openai.ChatCompletionMessage{
// 			Role:    obj.Role,
// 			Content: obj.Content,
// 		})
// 	}
// 	return output
// }

// """
// type ChatMessage struct {
// 	ID              int32
// 	Uuid            string
// 	ChatSessionUuid string
// 	Role            string
// 	Content         string

// }

// type ChatPrompt struct {
// 	ID              int32
// 	Uuid            string
// 	ChatSessionUuid string
// 	Role            string
// 	Content         string
// 	Score           float64
// }

// type Message struct {
// 	Role string
// 	Content string
// }

// """

// please write a generic method that convert a list of ChatPrompt or ChatMessage to Message in golang

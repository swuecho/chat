package sqlc_queries

import (
	"context"
	"encoding/json"
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

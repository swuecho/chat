package sqlc_queries

import (
	"encoding/json"
)

func (user *AuthUser) Role() string {
	role := "user"
	if user.IsSuperuser {
		role = "admin"
	}
	return role
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

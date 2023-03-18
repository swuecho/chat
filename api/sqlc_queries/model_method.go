package sqlc_queries

import "context"

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

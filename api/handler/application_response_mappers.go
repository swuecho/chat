package handler

import "github.com/swuecho/chat_backend/svc"

func snapshotResponse(s svc.ChatSnapshot) map[string]any {
	return map[string]any{"id": s.ID, "typ": s.Type, "uuid": s.UUID, "userId": s.UserID,
		"title": s.Title, "summary": s.Summary, "model": s.Model, "tags": s.Tags,
		"session": s.Session, "conversation": s.Conversation, "createdAt": s.CreatedAt, "text": s.Text}
}

func snapshotSummaryResponse(s svc.ChatSnapshotSummary) map[string]any {
	return map[string]any{"uuid": s.UUID, "title": s.Title, "summary": s.Summary, "tags": s.Tags, "createdAt": s.CreatedAt, "typ": s.Type}
}

func snapshotSearchResponse(s svc.ChatSnapshotSearchResult) map[string]any {
	return map[string]any{"uuid": s.UUID, "title": s.Title, "rank": s.Rank}
}

func historyMessageResponse(m svc.SessionHistoryMessage) map[string]any {
	response := map[string]any{"uuid": m.UUID, "dateTime": m.DateTime, "text": m.Text, "model": m.Model,
		"inversion": m.Inversion, "error": m.Error, "loading": m.Loading, "isPin": m.IsPin,
		"isPrompt": m.IsPrompt}
	if len(m.Artifacts) > 0 {
		response["artifacts"] = m.Artifacts
	}
	if len(m.SuggestedQuestions) > 0 {
		response["suggestedQuestions"] = m.SuggestedQuestions
	}
	return response
}

func adminMessageResponse(m svc.AdminSessionMessage) map[string]any {
	return map[string]any{"id": m.ID, "uuid": m.UUID, "role": m.Role, "content": m.Content,
		"reasoningContent": m.ReasoningContent, "model": m.Model, "tokenCount": m.TokenCount,
		"userId": m.UserID, "createdAt": m.CreatedAt, "updatedAt": m.UpdatedAt}
}

func promptResponse(p svc.ChatPrompt) map[string]any {
	return map[string]any{"id": p.ID, "uuid": p.Uuid, "chatSessionUuid": p.ChatSessionUuid,
		"role": p.Role, "content": p.Content, "score": p.Score, "userId": p.UserID,
		"createdAt": p.CreatedAt, "updatedAt": p.UpdatedAt, "createdBy": p.CreatedBy,
		"updatedBy": p.UpdatedBy, "isDeleted": p.IsDeleted, "tokenCount": p.TokenCount}
}

func promptResponses(prompts []svc.ChatPrompt) []map[string]any {
	result := make([]map[string]any, 0, len(prompts))
	for _, prompt := range prompts {
		result = append(result, promptResponse(prompt))
	}
	return result
}

func messageResponse(m svc.ChatMessage) map[string]any {
	return map[string]any{"id": m.ID, "uuid": m.Uuid, "chatSessionUuid": m.ChatSessionUuid,
		"role": m.Role, "content": m.Content, "reasoningContent": m.ReasoningContent,
		"model": m.Model, "llmSummary": m.LlmSummary, "score": m.Score, "userId": m.UserID,
		"createdAt": m.CreatedAt, "updatedAt": m.UpdatedAt, "createdBy": m.CreatedBy,
		"updatedBy": m.UpdatedBy, "isDeleted": m.IsDeleted, "isPin": m.IsPin,
		"tokenCount": m.TokenCount, "raw": m.Raw, "artifacts": m.Artifacts,
		"suggestedQuestions": m.SuggestedQuestions}
}

func messageResponses(messages []svc.ChatMessage) []map[string]any {
	result := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		result = append(result, messageResponse(message))
	}
	return result
}

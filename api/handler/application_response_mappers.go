package handler

import (
	"encoding/json"
	"time"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/svc"
)

type snapshotHTTPResponse struct {
	ID           int32           `json:"id"`
	Type         string          `json:"typ"`
	UUID         string          `json:"uuid"`
	UserID       int32           `json:"userId"`
	Title        string          `json:"title"`
	Summary      string          `json:"summary"`
	Model        string          `json:"model"`
	Tags         json.RawMessage `json:"tags"`
	Session      json.RawMessage `json:"session"`
	Conversation json.RawMessage `json:"conversation"`
	CreatedAt    time.Time       `json:"createdAt"`
	Text         string          `json:"text"`
}

type snapshotSummaryHTTPResponse struct {
	UUID      string          `json:"uuid"`
	Title     string          `json:"title"`
	Summary   string          `json:"summary"`
	Tags      json.RawMessage `json:"tags"`
	CreatedAt time.Time       `json:"createdAt"`
	Type      string          `json:"typ"`
}
type snapshotSearchHTTPResponse struct {
	UUID  string  `json:"uuid"`
	Title string  `json:"title"`
	Rank  float32 `json:"rank"`
}
type historyMessageHTTPResponse struct {
	UUID               string            `json:"uuid"`
	DateTime           string            `json:"dateTime"`
	Text               string            `json:"text"`
	Model              string            `json:"model"`
	Inversion          bool              `json:"inversion"`
	Error              bool              `json:"error"`
	Loading            bool              `json:"loading"`
	IsPin              bool              `json:"isPin"`
	IsPrompt           bool              `json:"isPrompt"`
	Artifacts          []domain.Artifact `json:"artifacts,omitempty"`
	SuggestedQuestions []string          `json:"suggestedQuestions,omitempty"`
}
type adminMessageHTTPResponse struct {
	ID               int32     `json:"id"`
	UUID             string    `json:"uuid"`
	Role             string    `json:"role"`
	Content          string    `json:"content"`
	ReasoningContent string    `json:"reasoningContent"`
	Model            string    `json:"model"`
	TokenCount       int32     `json:"tokenCount"`
	UserID           int32     `json:"userId"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
type promptHTTPResponse struct {
	ID              int32     `json:"id"`
	UUID            string    `json:"uuid"`
	ChatSessionUUID string    `json:"chatSessionUuid"`
	Role            string    `json:"role"`
	Content         string    `json:"content"`
	Score           float64   `json:"score"`
	UserID          int32     `json:"userId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	CreatedBy       int32     `json:"createdBy"`
	UpdatedBy       int32     `json:"updatedBy"`
	IsDeleted       bool      `json:"isDeleted"`
	TokenCount      int32     `json:"tokenCount"`
}
type messageHTTPResponse struct {
	ID                 int32           `json:"id"`
	UUID               string          `json:"uuid"`
	ChatSessionUUID    string          `json:"chatSessionUuid"`
	Role               string          `json:"role"`
	Content            string          `json:"content"`
	ReasoningContent   string          `json:"reasoningContent"`
	Model              string          `json:"model"`
	LLMSummary         string          `json:"llmSummary"`
	Score              float64         `json:"score"`
	UserID             int32           `json:"userId"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	CreatedBy          int32           `json:"createdBy"`
	UpdatedBy          int32           `json:"updatedBy"`
	IsDeleted          bool            `json:"isDeleted"`
	IsPin              bool            `json:"isPin"`
	TokenCount         int32           `json:"tokenCount"`
	Raw                json.RawMessage `json:"raw"`
	Artifacts          json.RawMessage `json:"artifacts"`
	SuggestedQuestions json.RawMessage `json:"suggestedQuestions"`
}

func snapshotResponse(s svc.ChatSnapshot) snapshotHTTPResponse {
	return snapshotHTTPResponse{ID: s.ID, Type: s.Type, UUID: s.UUID, UserID: s.UserID, Title: s.Title, Summary: s.Summary, Model: s.Model, Tags: s.Tags, Session: s.Session, Conversation: s.Conversation, CreatedAt: s.CreatedAt, Text: s.Text}
}

func snapshotSummaryResponse(s svc.ChatSnapshotSummary) snapshotSummaryHTTPResponse {
	return snapshotSummaryHTTPResponse{UUID: s.UUID, Title: s.Title, Summary: s.Summary, Tags: s.Tags, CreatedAt: s.CreatedAt, Type: s.Type}
}

func snapshotSearchResponse(s svc.ChatSnapshotSearchResult) snapshotSearchHTTPResponse {
	return snapshotSearchHTTPResponse{UUID: s.UUID, Title: s.Title, Rank: s.Rank}
}

func historyMessageResponse(m svc.SessionHistoryMessage) historyMessageHTTPResponse {
	return historyMessageHTTPResponse{UUID: m.UUID, DateTime: m.DateTime, Text: m.Text, Model: m.Model, Inversion: m.Inversion, Error: m.Error, Loading: m.Loading, IsPin: m.IsPin, IsPrompt: m.IsPrompt, Artifacts: m.Artifacts, SuggestedQuestions: m.SuggestedQuestions}
}

func adminMessageResponse(m svc.AdminSessionMessage) adminMessageHTTPResponse {
	return adminMessageHTTPResponse{ID: m.ID, UUID: m.UUID, Role: m.Role, Content: m.Content, ReasoningContent: m.ReasoningContent, Model: m.Model, TokenCount: m.TokenCount, UserID: m.UserID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func promptResponse(p svc.ChatPrompt) promptHTTPResponse {
	return promptHTTPResponse{ID: p.ID, UUID: p.UUID, ChatSessionUUID: p.ChatSessionUUID, Role: p.Role, Content: p.Content, Score: p.Score, UserID: p.UserID, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt, CreatedBy: p.CreatedBy, UpdatedBy: p.UpdatedBy, IsDeleted: p.IsDeleted, TokenCount: p.TokenCount}
}

func promptResponses(prompts []svc.ChatPrompt) []promptHTTPResponse {
	result := make([]promptHTTPResponse, 0, len(prompts))
	for _, prompt := range prompts {
		result = append(result, promptResponse(prompt))
	}
	return result
}

func messageResponse(m svc.ChatMessage) messageHTTPResponse {
	return messageHTTPResponse{ID: m.ID, UUID: m.UUID, ChatSessionUUID: m.ChatSessionUUID, Role: m.Role, Content: m.Content, ReasoningContent: m.ReasoningContent, Model: m.Model, LLMSummary: m.LLMSummary, Score: m.Score, UserID: m.UserID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt, CreatedBy: m.CreatedBy, UpdatedBy: m.UpdatedBy, IsDeleted: m.IsDeleted, IsPin: m.IsPin, TokenCount: m.TokenCount, Raw: m.Raw, Artifacts: m.Artifacts, SuggestedQuestions: m.SuggestedQuestions}
}

func messageResponses(messages []svc.ChatMessage) []messageHTTPResponse {
	result := make([]messageHTTPResponse, 0, len(messages))
	for _, message := range messages {
		result = append(result, messageResponse(message))
	}
	return result
}

// Package dto provides shared data transfer objects used across the application.
package dto

import (
	"fmt"
	"time"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/validation"
)

// --- Request types ---

type ConversationRequest struct {
	UUID            string `json:"uuid,omitempty"`
	ConversationID  string `json:"conversationId,omitempty"`
	ParentMessageID string `json:"parentMessageId,omitempty"`
}

type RequestOption struct {
	Prompt  string              `json:"prompt,omitempty"`
	Options ConversationRequest `json:"options,omitempty"`
}

type ChatRequest struct {
	Prompt      string `json:"prompt"`
	SessionUuid string `json:"sessionUuid"`
	ChatUuid    string `json:"chatUuid"`
	Regenerate  bool   `json:"regenerate"`
	Stream      bool   `json:"stream,omitempty"`
}

type BotRequest struct {
	Message      string `json:"message"`
	SnapshotUuid string `json:"snapshot_uuid"`
	Stream       bool   `json:"stream"`
}

type UpdateChatSessionRequest struct {
	Uuid            string  `json:"uuid"`
	Topic           string  `json:"topic"`
	MaxLength       int32   `json:"maxLength"`
	Temperature     float64 `json:"temperature"`
	Model           string  `json:"model"`
	TopP            float64 `json:"topP"`
	N               int32   `json:"n"`
	MaxTokens       int32   `json:"maxTokens"`
	Debug           bool    `json:"debug"`
	SummarizeMode   bool    `json:"summarizeMode"`
	ArtifactEnabled bool    `json:"artifactEnabled"`
	ExploreMode     bool    `json:"exploreMode"`
	WorkspaceUUID   string  `json:"workspaceUuid,omitempty"`
}

func (r *UpdateChatSessionRequest) Validate() error {
	if err := validation.UUID("uuid", r.Uuid, true); err != nil {
		return err
	}
	if err := validation.UUID("workspaceUuid", r.WorkspaceUUID, false); err != nil {
		return err
	}
	if err := validation.Topic(r.Topic, false); err != nil {
		return err
	}
	if err := validation.ModelName("model", r.Model, true); err != nil {
		return err
	}
	if r.MaxLength != 0 && (r.MaxLength < 1 || r.MaxLength > validation.MaxTokenCount) {
		return fmt.Errorf("maxLength must be between 1 and %d", validation.MaxTokenCount)
	}
	if r.Temperature < 0 || r.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	if r.TopP < 0 || r.TopP > 1 {
		return fmt.Errorf("topP must be between 0 and 1")
	}
	if r.N < 1 || r.N > 128 {
		return fmt.Errorf("n must be between 1 and 128")
	}
	return validation.TokenCount("maxTokens", r.MaxTokens, false)
}

// --- Response types ---

type TokenResult struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

// Artifact is retained as a transport alias for compatibility. The canonical
// application type lives in domain.
type Artifact = domain.Artifact

type SimpleChatMessage struct {
	Uuid      string     `json:"uuid"`
	DateTime  string     `json:"dateTime"`
	Text      string     `json:"text"`
	Inversion bool       `json:"inversion"`
	Error     bool       `json:"error"`
	Loading   bool       `json:"loading"`
	IsPin     bool       `json:"isPin"`
	IsPrompt  bool       `json:"isPrompt"`
	Artifacts []Artifact `json:"artifacts,omitempty"`
}

func (msg SimpleChatMessage) GetRole() string {
	if msg.Inversion {
		return "user"
	}
	return "assistant"
}

type SimpleChatSession struct {
	Uuid            string  `json:"uuid"`
	IsEdit          bool    `json:"isEdit"`
	Title           string  `json:"title"`
	MaxLength       int     `json:"maxLength"`
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"topP"`
	N               int32   `json:"n"`
	MaxTokens       int32   `json:"maxTokens"`
	Debug           bool    `json:"debug"`
	Model           string  `json:"model"`
	SummarizeMode   bool    `json:"summarizeMode"`
	ArtifactEnabled bool    `json:"artifactEnabled"`
	WorkspaceUuid   string  `json:"workspaceUuid"`
}

type ChatMessageResponse struct {
	Uuid            string     `json:"uuid"`
	ChatSessionUuid string     `json:"chatSessionUuid"`
	Role            string     `json:"role"`
	Content         string     `json:"content"`
	Score           float64    `json:"score"`
	UserID          int32      `json:"userId"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	CreatedBy       int32      `json:"createdBy"`
	UpdatedBy       int32      `json:"updatedBy"`
	Artifacts       []Artifact `json:"artifacts,omitempty"`
}

type ChatSessionResponse struct {
	Uuid            string    `json:"uuid"`
	Topic           string    `json:"topic"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	MaxLength       int32     `json:"maxLength"`
	ArtifactEnabled bool      `json:"artifactEnabled"`
}

type Pagination struct {
	Page  int32         `json:"page"`
	Size  int32         `json:"size"`
	Data  []interface{} `json:"data"`
	Total int64         `json:"total"`
}

func (p *Pagination) Offset() int32 {
	return (p.Page - 1) * p.Size
}

// --- Workspace types ---

type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	IsDefault   bool   `json:"isDefault"`
}

type UpdateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

type UpdateWorkspaceOrderRequest struct {
	OrderPosition int32 `json:"orderPosition"`
}

type WorkspaceResponse struct {
	Uuid          string `json:"uuid"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	Icon          string `json:"icon"`
	IsDefault     bool   `json:"isDefault"`
	OrderPosition int32  `json:"orderPosition"`
	SessionCount  int64  `json:"sessionCount,omitempty"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type CreateSessionInWorkspaceRequest struct {
	Topic               string `json:"topic"`
	Model               string `json:"model"`
	DefaultSystemPrompt string `json:"defaultSystemPrompt"`
}

func (r *CreateSessionInWorkspaceRequest) Validate() error {
	if err := validation.Topic(r.Topic, false); err != nil {
		return err
	}
	return validation.ModelName("model", r.Model, true)
}

// --- Chat instruction response ---

type ChatInstructionResponse struct {
	ArtifactInstruction string `json:"artifactInstruction"`
}

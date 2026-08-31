package handler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/swuecho/chat_backend/svc"
	"github.com/swuecho/chat_backend/validation"
)

type createChatModelRequest struct {
	// ID, UserID, and IsTitleModel are accepted for compatibility with clients
	// that submit a previously returned model object. The path and authenticated
	// identity remain authoritative and these values are intentionally ignored.
	ID                      int32  `json:"id"`
	Name                    string `json:"name"`
	Label                   string `json:"label"`
	IsDefault               bool   `json:"isDefault"`
	URL                     string `json:"url"`
	APIAuthHeader           string `json:"apiAuthHeader"`
	APIAuthKey              string `json:"apiAuthKey"`
	EnablePerModelRateLimit bool   `json:"enablePerModeRatelimit"`
	MaxToken                int32  `json:"maxToken"`
	DefaultToken            int32  `json:"defaultToken"`
	OrderNumber             int32  `json:"orderNumber"`
	HTTPTimeout             int32  `json:"httpTimeOut"`
	APIType                 string `json:"apiType"`
	IsEnable                bool   `json:"isEnable"`
	UserID                  int32  `json:"userId"`
	IsTitleModel            bool   `json:"isTitleModel"`
}

type createAuthUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Username    string `json:"username"`
	IsStaff     bool   `json:"isStaff"`
	IsSuperuser bool   `json:"isSuperuser"`
}

type updateAuthUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type chatPromptRequest struct {
	UUID            string  `json:"uuid"`
	ChatSessionUUID string  `json:"chatSessionUuid"`
	Role            string  `json:"role"`
	Content         string  `json:"content"`
	TokenCount      int32   `json:"tokenCount"`
	Score           float64 `json:"score"`
}

type chatMessageRequest struct {
	UUID               string          `json:"uuid"`
	ChatSessionUUID    string          `json:"chatSessionUuid"`
	Role               string          `json:"role"`
	Content            string          `json:"content"`
	ReasoningContent   string          `json:"reasoningContent"`
	Model              string          `json:"model"`
	LLMSummary         string          `json:"llmSummary"`
	TokenCount         int32           `json:"tokenCount"`
	UserID             int32           `json:"userId"`
	CreatedBy          int32           `json:"createdBy"`
	UpdatedBy          int32           `json:"updatedBy"`
	Score              float64         `json:"score"`
	Raw                json.RawMessage `json:"raw"`
	Artifacts          json.RawMessage `json:"artifacts"`
	SuggestedQuestions json.RawMessage `json:"suggestedQuestions"`
}

type chatModelPrivilegeRequest struct {
	UserEmail     string `json:"userEmail"`
	ChatModelName string `json:"chatModelName"`
	RateLimit     int32  `json:"rateLimit"`
}

type createBotAnswerHistoryRequest struct {
	BotUUID    string `json:"botUuid"`
	Prompt     string `json:"prompt"`
	Answer     string `json:"answer"`
	Model      string `json:"model"`
	TokensUsed int32  `json:"tokensUsed"`
}

type createChatSessionRequest struct {
	UUID                string `json:"uuid" jsonschema:"required,format=uuid"`
	Topic               string `json:"topic" jsonschema:"maxLength=200"`
	Model               string `json:"model" jsonschema:"required,minLength=1,maxLength=200"`
	DefaultSystemPrompt string `json:"defaultSystemPrompt"`
}

func (r *createChatSessionRequest) Validate() error {
	if err := validation.UUID("uuid", r.UUID, true); err != nil {
		return err
	}
	if err := validation.Topic(r.Topic, false); err != nil {
		return err
	}
	return validation.ModelName("model", r.Model, true)
}

type createWorkspaceRequest struct {
	Name        string `json:"name" jsonschema:"required,minLength=1,maxLength=200"`
	Description string `json:"description,omitempty" jsonschema:"maxLength=2000"`
	Color       string `json:"color,omitempty" jsonschema:"maxLength=100"`
	Icon        string `json:"icon,omitempty" jsonschema:"maxLength=100"`
	IsDefault   bool   `json:"isDefault,omitempty"`
}

func (r *createWorkspaceRequest) Validate() error {
	return validateWorkspaceName(r.Name)
}

type updateWorkspaceRequest struct {
	Name        string `json:"name" jsonschema:"required,minLength=1,maxLength=200"`
	Description string `json:"description,omitempty" jsonschema:"maxLength=2000"`
	Color       string `json:"color,omitempty" jsonschema:"maxLength=100"`
	Icon        string `json:"icon,omitempty" jsonschema:"maxLength=100"`
}

func (r *updateWorkspaceRequest) Validate() error {
	return validateWorkspaceName(r.Name)
}

type updateWorkspaceOrderRequest struct {
	OrderPosition int32 `json:"orderPosition" jsonschema:"minimum=0"`
}

func (r *updateWorkspaceOrderRequest) Validate() error {
	if r.OrderPosition < 0 {
		return fmt.Errorf("orderPosition must not be negative")
	}
	return nil
}

type createSessionInWorkspaceRequest struct {
	Topic               string `json:"topic,omitempty" jsonschema:"maxLength=200"`
	Model               string `json:"model" jsonschema:"required,minLength=1,maxLength=200"`
	DefaultSystemPrompt string `json:"defaultSystemPrompt,omitempty"`
}

func (r *createSessionInWorkspaceRequest) Validate() error {
	if err := validation.Topic(r.Topic, false); err != nil {
		return err
	}
	return validation.ModelName("model", r.Model, true)
}

func validateWorkspaceName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if utf8.RuneCountInString(name) > 200 {
		return fmt.Errorf("name must be at most 200 characters")
	}
	return nil
}

type updateSessionTopicRequest struct {
	Topic string `json:"topic" jsonschema:"required,minLength=1,maxLength=200"`
}

func (r *updateSessionTopicRequest) Validate() error {
	return validation.Topic(r.Topic, true)
}

type updateSessionMaxLengthRequest struct {
	MaxLength int32 `json:"maxLength" jsonschema:"minimum=1,maximum=1000000"`
}

type setTitleModelRequest struct {
	ModelID int32 `json:"modelId"`
}

type updateBotAnswerHistoryRequest struct {
	Answer     string `json:"answer"`
	TokensUsed int32  `json:"tokensUsed"`
}

type createCommentRequest struct {
	Content string `json:"content"`
}

func (r *createCommentRequest) Validate() error {
	if r.Content == "" {
		return fmt.Errorf("content is required")
	}
	return nil
}

type activeSessionRequest struct {
	ChatSessionUUID string `json:"chatSessionUuid"`
}

type createAPIKeyRequest struct {
	Name              string `json:"name"`
	ExpiresAt         string `json:"expiresAt"`
	RequestsPerMinute int32  `json:"requestsPerMinute"`
}

type updateBotSettingsRequest struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Model   string `json:"model"`
}

func (r *updateBotSettingsRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	r.Summary = strings.TrimSpace(r.Summary)
	r.Model = strings.TrimSpace(r.Model)
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	return validation.ModelName("model", r.Model, true)
}

type updateBotModelRequest struct {
	Model string `json:"model"`
}

func (r *updateBotModelRequest) Validate() error {
	r.Model = strings.TrimSpace(r.Model)
	return validation.ModelName("model", r.Model, true)
}

type updateSnapshotMetadataRequest struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

type pageRequest struct {
	Page int32 `json:"page"`
	Size int32 `json:"size"`
}

func (r *pageRequest) Validate() error {
	if r.Page < 1 {
		return fmt.Errorf("page must be positive")
	}
	if r.Size < 1 || r.Size > validation.MaxPageSize {
		return fmt.Errorf("size must be between 1 and %d", validation.MaxPageSize)
	}
	return nil
}

func (r *createAPIKeyRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	if len(r.Name) < 1 || len(r.Name) > 100 {
		return fmt.Errorf("key name must be between 1 and 100 characters")
	}
	if r.RequestsPerMinute == 0 {
		r.RequestsPerMinute = 60
	}
	if r.RequestsPerMinute < 1 || r.RequestsPerMinute > 10_000 {
		return fmt.Errorf("requestsPerMinute must be between 1 and 10000")
	}
	if r.ExpiresAt != "" {
		expires, err := time.Parse(time.RFC3339, r.ExpiresAt)
		if err != nil || !expires.After(time.Now()) {
			return fmt.Errorf("expiresAt must be a future RFC3339 timestamp")
		}
	}
	return nil
}

func (r createAPIKeyRequest) expiration() *time.Time {
	if r.ExpiresAt == "" {
		return nil
	}
	expires, _ := time.Parse(time.RFC3339, r.ExpiresAt)
	return &expires
}

func (r *activeSessionRequest) Validate() error {
	return validation.UUID("chatSessionUuid", r.ChatSessionUUID, true)
}

func (r *updateBotAnswerHistoryRequest) Validate() error {
	return validation.TokenCount("tokensUsed", r.TokensUsed, true)
}

func (r *setTitleModelRequest) Validate() error {
	if r.ModelID <= 0 {
		return fmt.Errorf("modelId must be a positive integer")
	}
	return nil
}

func (r *updateSessionMaxLengthRequest) Validate() error {
	return validation.TokenCount("maxLength", r.MaxLength, false)
}

func (r *createChatModelRequest) Validate() error {
	if err := validation.ModelName("name", r.Name, true); err != nil {
		return err
	}
	if err := validation.TokenCount("maxToken", r.MaxToken, true); err != nil {
		return err
	}
	if err := validation.TokenCount("defaultToken", r.DefaultToken, true); err != nil {
		return err
	}
	if r.DefaultToken > 0 && r.MaxToken > 0 && r.DefaultToken > r.MaxToken {
		return fmt.Errorf("defaultToken must not exceed maxToken")
	}
	return nil
}

func (r *chatPromptRequest) Validate() error {
	if err := validation.UUID("uuid", r.UUID, false); err != nil {
		return err
	}
	if err := validation.UUID("chatSessionUuid", r.ChatSessionUUID, false); err != nil {
		return err
	}
	return validation.TokenCount("tokenCount", r.TokenCount, true)
}

func (r *chatMessageRequest) Validate() error {
	if err := validation.UUID("uuid", r.UUID, true); err != nil {
		return err
	}
	if err := validation.UUID("chatSessionUuid", r.ChatSessionUUID, true); err != nil {
		return err
	}
	if err := validation.ModelName("model", r.Model, false); err != nil {
		return err
	}
	return validation.TokenCount("tokenCount", r.TokenCount, true)
}

func (r *chatModelPrivilegeRequest) Validate() error {
	if err := validation.ModelName("chatModelName", r.ChatModelName, true); err != nil {
		return err
	}
	if r.RateLimit < 1 || r.RateLimit > validation.MaxTokenCount {
		return fmt.Errorf("rateLimit must be between 1 and %d", validation.MaxTokenCount)
	}
	return nil
}

func (r *createBotAnswerHistoryRequest) Validate() error {
	if err := validation.UUID("botUuid", r.BotUUID, true); err != nil {
		return err
	}
	if err := validation.ModelName("model", r.Model, true); err != nil {
		return err
	}
	return validation.TokenCount("tokensUsed", r.TokensUsed, true)
}

func (r createChatModelRequest) createInput(userID int32, apiType string) svc.CreateChatModelInput {
	return svc.CreateChatModelInput{Name: r.Name, Label: r.Label, IsDefault: r.IsDefault, Url: r.URL,
		ApiAuthHeader: r.APIAuthHeader, ApiAuthKey: r.APIAuthKey, UserID: userID,
		EnablePerModeRatelimit: r.EnablePerModelRateLimit, MaxToken: 4096, DefaultToken: 2048,
		OrderNumber: r.OrderNumber, HttpTimeOut: 120, ApiType: apiType}
}

func (r createChatModelRequest) updateInput(id, userID int32, apiType string) svc.UpdateChatModelInput {
	return svc.UpdateChatModelInput{ID: id, Name: r.Name, Label: r.Label, IsDefault: r.IsDefault, Url: r.URL,
		ApiAuthHeader: r.APIAuthHeader, ApiAuthKey: r.APIAuthKey, UserID: userID,
		EnablePerModeRatelimit: r.EnablePerModelRateLimit, MaxToken: r.MaxToken, DefaultToken: r.DefaultToken,
		OrderNumber: r.OrderNumber, HttpTimeOut: r.HTTPTimeout, IsEnable: r.IsEnable, ApiType: apiType}
}

func (r createAuthUserRequest) input() svc.CreateAuthUserInput {
	return svc.CreateAuthUserInput{Email: r.Email, Password: r.Password, FirstName: r.FirstName,
		LastName: r.LastName, Username: r.Username, IsStaff: r.IsStaff, IsSuperuser: r.IsSuperuser}
}

func (r updateAuthUserRequest) selfInput(id int32) svc.UpdateAuthUserInput {
	return svc.UpdateAuthUserInput{ID: id, FirstName: r.FirstName, LastName: r.LastName}
}

func (r updateAuthUserRequest) emailInput() svc.UpdateAuthUserByEmailInput {
	return svc.UpdateAuthUserByEmailInput{Email: r.Email, FirstName: r.FirstName, LastName: r.LastName}
}

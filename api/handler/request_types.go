package handler

import (
	"encoding/json"
	"fmt"

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
	UUID                string `json:"uuid"`
	Topic               string `json:"topic"`
	Model               string `json:"model"`
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

type updateSessionTopicRequest struct {
	Topic string `json:"topic"`
}

func (r *updateSessionTopicRequest) Validate() error {
	return validation.Topic(r.Topic, true)
}

type updateSessionMaxLengthRequest struct {
	MaxLength int32 `json:"maxLength"`
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

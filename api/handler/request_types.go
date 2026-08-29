package handler

import (
	"encoding/json"

	"github.com/swuecho/chat_backend/svc"
)

type createChatModelRequest struct {
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

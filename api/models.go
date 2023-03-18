package main

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type TokenResult struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

type ConversationRequest struct {
	UUID            string `json:"uuid,omitempty"`
	ConversationID  string `json:"conversationId,omitempty"`
	ParentMessageID string `json:"parentMessageId,omitempty"`
}

type RequestOption struct {
	Prompt  string              `json:"prompt,omitempty"`
	Options ConversationRequest `json:"options,omitempty"`
}

type SimpleChatMessage struct {
	Uuid                string              `json:"uuid"`
	DateTime            string              `json:"dateTime"`
	Text                string              `json:"text"`
	Inversion           bool                `json:"inversion"`
	Error               bool                `json:"error"`
	Loading             bool                `json:"loading"`
	ConversationOptions ConversationRequest `json:"conversationOptions,omitempty"`
	RequestOptions      RequestOption       `json:"requestOptions,omitempty"`
	IsPrompt            bool                `json:"isPrompt"`
}

type SimpleChatSession struct {
	Uuid   string `json:"uuid"`
	IsEdit bool   `json:"isEdit"`
	Title  string `json:"title"`
}

package main

import (
	"log"
	"time"
)

type ErrorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Message struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	tokenCount int32
}



func (m Message) TokenCount() int32 {
	if m.tokenCount != 0 {
		return m.tokenCount
	} else {
		tokenCount, err := getTokenCount(m.Content)
		if err != nil {
			log.Println(err)
		}
		return int32(tokenCount) + 1
	}
}

type TokenResult struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

type MultiMessage struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Content string `json:"content"`
	Type    string `json:"type"`
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
	Uuid      string `json:"uuid"`
	DateTime  string `json:"dateTime"`
	Text      string `json:"text"`
	Inversion bool   `json:"inversion"`
	Error     bool   `json:"error"`
	Loading   bool   `json:"loading"`
	IsPin     bool   `json:"isPin"`
	IsPrompt  bool   `json:"isPrompt"`
}

func (msg SimpleChatMessage) GetRole() string {
	var role string
	if msg.Inversion {
		role = "user"
	} else {
		role = "assistant"
	}
	return role

}

type SimpleChatSession struct {
	Uuid          string  `json:"uuid"`
	IsEdit        bool    `json:"isEdit"`
	Title         string  `json:"title"`
	MaxLength     int     `json:"maxLength"`
	Temperature   float64 `json:"temperature"`
	TopP          float64 `json:"topP"`
	N             int32   `json:"n"`
	MaxTokens     int32   `json:"maxTokens"`
	Debug         bool    `json:"debug"`
	Model         string  `json:"model"`
	SummarizeMode bool    `json:"summarizeMode"`
}

type ChatMessageResponse struct {
	Uuid            string    `json:"uuid"`
	ChatSessionUuid string    `json:"chatSessionUuid"`
	Role            string    `json:"role"`
	Content         string    `json:"content"`
	Score           float64   `json:"score"`
	UserID          int32     `json:"userId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	CreatedBy       int32     `json:"createdBy"`
	UpdatedBy       int32     `json:"updatedBy"`
}

type ChatSessionResponse struct {
	Uuid      string    `json:"uuid"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	MaxLength int32     `json:"maxLength"`
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

package dto

import (
	openai "github.com/sashabaranov/go-openai"
)

// ChatRequest represents a chat request from the client
type ChatRequest struct {
	Prompt      string `json:"prompt"`
	SessionUuid string `json:"session_uuid"`
	ChatUuid    string `json:"chat_uuid"`
	Regenerate  bool   `json:"regenerate"`
	Stream      bool   `json:"stream,omitempty"`
}

// ChatCompletionResponse represents a chat completion response
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   Usage  `json:"usage"`
	Choices []Choice `json:"choices"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice represents a chat completion choice
type Choice struct {
	Message      openai.ChatCompletionMessage `json:"message"`
	FinishReason any                          `json:"finish_reason"`
	Index        int                          `json:"index"`
}

// OpenaiChatRequest represents an OpenAI-compatible chat request
type OpenaiChatRequest struct {
	Model    string                         `json:"model"`
	Messages []openai.ChatCompletionMessage `json:"messages"`
}

// BotRequest represents a chatbot request
type BotRequest struct {
	Message      string `json:"message"`
	SnapshotUuid string `json:"snapshot_uuid"`
	Stream       bool   `json:"stream"`
}

// ChatInstructionResponse represents instructions for chat
type ChatInstructionResponse struct {
	ArtifactInstruction string `json:"artifactInstruction"`
	ToolInstruction     string `json:"toolInstruction"`
}

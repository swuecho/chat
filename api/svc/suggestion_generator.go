package svc

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/llm/gemini"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type SuggestionGenerator interface {
	GenerateSuggestions(context.Context, string, []models.Message) []string
}

type LLMSuggestionGenerator struct {
	q                      *sqlc_queries.Queries
	openAIKey, openAIProxy string
}

func NewLLMSuggestionGenerator(q *sqlc_queries.Queries, openAIKey, openAIProxy string) *LLMSuggestionGenerator {
	return &LLMSuggestionGenerator{q: q, openAIKey: openAIKey, openAIProxy: openAIProxy}
}

// generateSuggestedQuestions generates follow-up questions based on the conversation context
func (s *LLMSuggestionGenerator) GenerateSuggestions(ctx context.Context, content string, messages []models.Message) []string {
	// Create a simplified prompt to generate follow-up questions
	prompt := `Based on the following conversation, generate 3 thoughtful follow-up questions that would help explore the topic further. Return only the questions, one per line, without numbering or bullet points.

Conversation context:
`

	// Add the last few messages for context (limit to avoid token overflow)
	contextMessages := messages
	if len(messages) > 6 {
		contextMessages = messages[len(messages)-6:]
	}

	for _, msg := range contextMessages {
		prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
	}

	prompt += fmt.Sprintf("assistant: %s\n\nGenerate 3 follow-up questions:", content)

	// Use the preferred models (deepseek-chat or gemini-2.0-flash) to generate suggestions
	questions := s.callLLMForSuggestions(ctx, prompt)

	// Parse the response into individual questions
	lines := strings.Split(strings.TrimSpace(questions), "\n")
	var result []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && len(result) < 3 {
			// Clean up any numbering or bullet points that might remain
			line = strings.TrimPrefix(line, "1. ")
			line = strings.TrimPrefix(line, "2. ")
			line = strings.TrimPrefix(line, "3. ")
			line = strings.TrimPrefix(line, "- ")
			line = strings.TrimPrefix(line, "• ")
			result = append(result, line)
		}
	}

	return result
}

// callLLMForSuggestions makes a simple API call to generate suggested questions
func (s *LLMSuggestionGenerator) callLLMForSuggestions(ctx context.Context, prompt string) string {
	// Get all models and find preferred models for suggestions
	allModels, err := s.q.ListChatModels(ctx)
	if err != nil {
		slog.Warn("Failed to list models for suggestions", "error", err)
		return ""
	}

	// Filter for enabled models and prioritize deepseek-chat or gemini-2.0-flash
	var selectedModel sqlc_queries.ChatModel
	var foundPreferred bool

	// First pass: look for preferred models
	for _, model := range allModels {
		if !model.IsEnable {
			continue
		}
		modelNameLower := strings.ToLower(model.Name)
		if strings.Contains(modelNameLower, "deepseek-chat") || strings.Contains(modelNameLower, "gemini-2.0-flash") {
			selectedModel = model
			foundPreferred = true
			break
		}
	}

	// Second pass: fallback to any gemini or openai model if preferred not found
	if !foundPreferred {
		for _, model := range allModels {
			if !model.IsEnable {
				continue
			}
			apiType := strings.ToLower(model.ApiType)
			modelName := strings.ToLower(model.Name)

			// Prefer gemini models, then openai
			if apiType == "gemini" || (apiType == "openai" && strings.Contains(modelName, "gpt")) {
				selectedModel = model
				break
			}
		}
	}

	if selectedModel.ID == 0 {
		slog.Warn("No suitable models available for suggestions")
		return ""
	}

	// Use different API calls based on model type
	apiType := strings.ToLower(selectedModel.ApiType)
	modelName := strings.ToLower(selectedModel.Name)

	if apiType == "gemini" || strings.Contains(modelName, "gemini") {
		return s.callGeminiForSuggestions(ctx, selectedModel, prompt)
	} else if strings.Contains(modelName, "deepseek") || apiType == "openai" {
		return s.callOpenAICompatibleForSuggestions(ctx, selectedModel, prompt)
	}

	slog.Warn("Unsupported model type for suggestions", "apiType", selectedModel.ApiType)
	return ""
}

// callGeminiForSuggestions makes a Gemini API call for suggestions
func (s *LLMSuggestionGenerator) callGeminiForSuggestions(ctx context.Context, model sqlc_queries.ChatModel, prompt string) string {
	// Validate API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		slog.Warn("GEMINI_API_KEY environment variable not set")
		return ""
	}

	// Create messages for Gemini
	messages := []models.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Generate Gemini payload
	payloadBytes, err := gemini.GenGemminPayload(messages, nil)
	if err != nil {
		slog.Warn("Failed to generate Gemini payload for suggestions", "error", err)
		return ""
	}

	// Build URL
	url := gemini.BuildAPIURL(model.Name, false)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Warn("Failed to create Gemini request for suggestions", "error", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the API call with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	answer, err := gemini.HandleRegularResponse(http.Client{Timeout: 30 * time.Second}, req)
	if err != nil {
		slog.Warn("Failed to get Gemini response for suggestions", "error", err)
		return ""
	}

	if answer == nil || answer.Answer == "" {
		slog.Warn("Empty response from Gemini for suggestions")
		return ""
	}

	return answer.Answer
}

// callOpenAICompatibleForSuggestions makes an OpenAI-compatible API call for suggestions (including deepseek)
func (s *LLMSuggestionGenerator) callOpenAICompatibleForSuggestions(ctx context.Context, model sqlc_queries.ChatModel, prompt string) string {
	// Generate OpenAI client configuration
	config, err := provider.GenOpenAIConfig(providerModel(model), provider.Config{OpenAIKey: s.openAIKey, OpenAIProxy: s.openAIProxy})
	if err != nil {
		slog.Warn("Failed to generate OpenAI configuration for suggestions", "error", err)
		return ""
	}

	client := openai.NewClientWithConfig(config)

	// Create a simple chat completion request for generating suggestions
	req := openai.ChatCompletionRequest{
		Model:       model.Name,
		Temperature: 0.7,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 200, // Keep suggestions concise
	}

	// Make the API call with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		slog.Warn("Failed to generate suggested questions", "model", model.Name, "error", err)
		return ""
	}

	if len(resp.Choices) == 0 {
		slog.Warn("No response choices for suggested questions", "model", model.Name)
		return ""
	}

	return resp.Choices[0].Message.Content
}

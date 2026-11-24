package svc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type SuggestionGenerator interface {
	GenerateSuggestions(context.Context, string, []domain.Message) []string
}

type LLMSuggestionGenerator struct {
	q           *sqlc_queries.Queries
	openAIProxy string
}

func NewLLMSuggestionGenerator(q *sqlc_queries.Queries, openAIProxy string) *LLMSuggestionGenerator {
	return &LLMSuggestionGenerator{q: q, openAIProxy: openAIProxy}
}

// generateSuggestedQuestions generates follow-up questions based on the conversation context
func (s *LLMSuggestionGenerator) GenerateSuggestions(ctx context.Context, content string, messages []domain.Message) []string {
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

// callLLMForSuggestions picks a model and asks it for follow-up questions.
// Failures are logged and yield no suggestions; they never fail the chat.
func (s *LLMSuggestionGenerator) callLLMForSuggestions(ctx context.Context, prompt string) string {
	allModels, err := s.q.ListChatModels(ctx)
	if err != nil {
		slog.Warn("Failed to list models for suggestions", "error", err)
		return ""
	}
	model, ok := selectSuggestionModel(allModels)
	if !ok {
		slog.Warn("No enabled models available for suggestions")
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, suggestionTimeout)
	defer cancel()

	answer, err := provider.Complete(ctx, providerModel(model), []domain.Message{{Role: "user", Content: prompt}}, provider.CompleteOptions{
		MaxTokens:   200, // Keep suggestions concise
		Temperature: 0.7,
		ChatUUID:    "suggestions",
		OpenAIProxy: s.openAIProxy,
	})
	if err != nil {
		slog.Warn("Failed to generate suggested questions", "model", model.Name, "error", err)
		return ""
	}
	return answer
}

const suggestionTimeout = 30 * time.Second

// selectSuggestionModel prefers fast, cheap models, then any Gemini or GPT
// model, then any enabled model.
func selectSuggestionModel(models []sqlc_queries.ChatModel) (sqlc_queries.ChatModel, bool) {
	preferences := []func(sqlc_queries.ChatModel) bool{
		func(m sqlc_queries.ChatModel) bool {
			name := strings.ToLower(m.Name)
			return strings.Contains(name, "deepseek-chat") || strings.Contains(name, "gemini-2.0-flash")
		},
		func(m sqlc_queries.ChatModel) bool {
			apiType := strings.ToLower(m.ApiType)
			return apiType == "gemini" || (apiType == "openai" && strings.Contains(strings.ToLower(m.Name), "gpt"))
		},
		func(sqlc_queries.ChatModel) bool { return true },
	}
	for _, matches := range preferences {
		for _, model := range models {
			if model.IsEnable && matches(model) {
				return model, true
			}
		}
	}
	return sqlc_queries.ChatModel{}, false
}

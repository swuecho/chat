package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "embed"
	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"
	"github.com/swuecho/chat_backend/llm/gemini"
	models "github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatService struct {
	q *sqlc_queries.Queries
}

//go:embed artifact_instruction.txt
var artifactInstructionText string

//go:embed tool_instruction.txt
var toolInstructionText string

// NewChatService creates a new ChatService with database queries.
func NewChatService(q *sqlc_queries.Queries) *ChatService {
	return &ChatService{q: q}
}

// loadArtifactInstruction loads the artifact instruction from file.
// Returns the instruction content or an error if the file cannot be read.
func loadArtifactInstruction() (string, error) {
	if artifactInstructionText == "" {
		return "", eris.New("artifact instruction text is empty")
	}
	return artifactInstructionText, nil
}

// loadToolInstruction loads the tool-use instruction from file.
// Returns the instruction content or an error if the file cannot be read.
func loadToolInstruction() (string, error) {
	if toolInstructionText == "" {
		return "", eris.New("tool instruction text is empty")
	}
	return toolInstructionText, nil
}

func appendInstructionToSystemMessage(msgs []models.Message, instruction string) {
	if instruction == "" || len(msgs) == 0 {
		return
	}

	systemMsgFound := false
	for i, msg := range msgs {
		if msg.Role == "system" {
			msgs[i].Content = msg.Content + "\n" + instruction
			msgs[i].SetTokenCount(int32(len(msgs[i].Content) / TokenEstimateRatio))
			systemMsgFound = true
			break
		}
	}

	if !systemMsgFound {
		msgs[0].Content = msgs[0].Content + "\n" + instruction
		msgs[0].SetTokenCount(int32(len(msgs[0].Content) / TokenEstimateRatio))
	}
}

// getAskMessages retrieves and processes chat messages for LLM requests.
// It combines prompts and messages, applies length limits, and adds artifact instructions (unless explore mode is enabled).
// Parameters:
//   - chatSession: The chat session containing configuration
//   - chatUuid: UUID for message identification (used in regenerate mode)
//   - regenerate: If true, excludes the target message from history
//
// Returns combined message array or error.
func (s *ChatService) getAskMessages(chatSession sqlc_queries.ChatSession, chatUuid string, regenerate bool) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*RequestTimeoutSeconds)
	defer cancel()

	chatSessionUuid := chatSession.Uuid

	lastN := chatSession.MaxLength
	if chatSession.MaxLength == 0 {
		lastN = DefaultMaxLength
	}

	chat_prompts, err := s.q.GetChatPromptsBySessionUUID(ctx, chatSessionUuid)

	if err != nil {
		return nil, eris.Wrap(err, "fail to get prompt: ")
	}

	var chatMessages []sqlc_queries.ChatMessage
	if regenerate {
		chatMessages, err = s.q.GetLastNChatMessages(ctx,
			sqlc_queries.GetLastNChatMessagesParams{
				ChatSessionUuid: chatSessionUuid,
				Uuid:            chatUuid,
				Limit:           lastN,
			})

	} else {
		chatMessages, err = s.q.GetLatestMessagesBySessionUUID(ctx,
			sqlc_queries.GetLatestMessagesBySessionUUIDParams{ChatSessionUuid: chatSession.Uuid, Limit: lastN})
	}

	if err != nil {
		return nil, eris.Wrap(err, "fail to get messages: ")
	}
	chatPromptMsgs := lo.Map(chat_prompts, func(m sqlc_queries.ChatPrompt, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})
	chatMessageMsgs := lo.Map(chatMessages, func(m sqlc_queries.ChatMessage, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})
	msgs := append(chatPromptMsgs, chatMessageMsgs...)

	// Add artifact instruction to system messages only if artifact mode is enabled
	if chatSession.ArtifactEnabled {
		artifactInstruction, err := loadArtifactInstruction()
		if err != nil {
			log.Printf("Warning: Failed to load artifact instruction: %v", err)
			artifactInstruction = "" // Use empty string if file can't be loaded
		}

		appendInstructionToSystemMessage(msgs, artifactInstruction)
	}

	if chatSession.CodeRunnerEnabled {
		toolInstruction, err := loadToolInstruction()
		if err != nil {
			log.Printf("Warning: Failed to load tool instruction: %v", err)
			toolInstruction = ""
		}

		appendInstructionToSystemMessage(msgs, toolInstruction)
	}

	return msgs, nil
}

// CreateChatPromptSimple creates a new chat prompt for a session.
// This is typically used to start a new conversation with a system message.
func (s *ChatService) CreateChatPromptSimple(ctx context.Context, chatSessionUuid string, newQuestion string, userID int32) (sqlc_queries.ChatPrompt, error) {
	tokenCount, err := getTokenCount(newQuestion)
	if err != nil {
		log.Printf("Warning: Failed to get token count for prompt: %v", err)
		tokenCount = len(newQuestion) / TokenEstimateRatio // Fallback estimate
	}
	chatPrompt, err := s.q.CreateChatPrompt(ctx,
		sqlc_queries.CreateChatPromptParams{
			Uuid:            NewUUID(),
			ChatSessionUuid: chatSessionUuid,
			Role:            "system",
			Content:         newQuestion,
			UserID:          userID,
			CreatedBy:       userID,
			UpdatedBy:       userID,
			TokenCount:      int32(tokenCount),
		})
	return chatPrompt, err
}

// CreateChatMessageSimple creates a new chat message with optional summarization and artifact extraction.
// Handles token counting, content summarization for long messages, and artifact parsing.
// Parameters:
//   - ctx: Request context for cancellation
//   - sessionUuid, uuid: Message and session identifiers
//   - role: Message role (user/assistant/system)
//   - content, reasoningContent: Message content and reasoning (if any)
//   - model: LLM model name
//   - userId: User ID for ownership
//   - baseURL: API base URL for summarization
//   - is_summarize_mode: Whether to enable automatic summarization
//
// Returns created message or error.
func (s *ChatService) CreateChatMessageSimple(ctx context.Context, sessionUuid, uuid, role, content, reasoningContent, model string, userId int32, baseURL string, is_summarize_mode bool) (sqlc_queries.ChatMessage, error) {
	numTokens, err := getTokenCount(content)
	if err != nil {
		log.Printf("Warning: Failed to get token count: %v", err)
		numTokens = len(content) / TokenEstimateRatio // Fallback estimate
	}

	summary := ""

	if is_summarize_mode && numTokens > SummarizeThreshold {
		log.Println("summarizing")
		summary = llm_summarize_with_timeout(baseURL, content)
		log.Println("summarizing: " + summary)
	}

	// Extract artifacts from content
	artifacts := extractArtifacts(content)
	artifactsJSON, err := json.Marshal(artifacts)
	if err != nil {
		log.Printf("Warning: Failed to marshal artifacts: %v", err)
		artifactsJSON = json.RawMessage([]byte("[]"))
	}

	chatMessage := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:    sessionUuid,
		Uuid:               uuid,
		Role:               role,
		Content:            content,
		ReasoningContent:   reasoningContent,
		Model:              model,
		UserID:             userId,
		CreatedBy:          userId,
		UpdatedBy:          userId,
		LlmSummary:         summary,
		TokenCount:         int32(numTokens),
		Raw:                json.RawMessage([]byte("{}")),
		Artifacts:          artifactsJSON,
		SuggestedQuestions: json.RawMessage([]byte("[]")),
	}
	message, err := s.q.CreateChatMessage(ctx, chatMessage)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
}

// CreateChatMessageWithSuggestedQuestions creates a chat message with optional suggested questions for explore mode
func (s *ChatService) CreateChatMessageWithSuggestedQuestions(ctx context.Context, sessionUuid, uuid, role, content, reasoningContent, model string, userId int32, baseURL string, is_summarize_mode, exploreMode bool, messages []models.Message) (sqlc_queries.ChatMessage, error) {
	numTokens, err := getTokenCount(content)
	if err != nil {
		log.Printf("Warning: Failed to get token count: %v", err)
		numTokens = len(content) / TokenEstimateRatio // Fallback estimate
	}

	summary := ""
	if is_summarize_mode && numTokens > SummarizeThreshold {
		log.Println("summarizing")
		summary = llm_summarize_with_timeout(baseURL, content)
		log.Println("summarizing: " + summary)
	}

	// Extract artifacts from content
	artifacts := extractArtifacts(content)
	artifactsJSON, err := json.Marshal(artifacts)
	if err != nil {
		log.Printf("Warning: Failed to marshal artifacts: %v", err)
		artifactsJSON = json.RawMessage([]byte("[]"))
	}

	// Generate suggested questions if explore mode is enabled and role is assistant
	suggestedQuestions := json.RawMessage([]byte("[]"))
	if exploreMode && role == "assistant" && messages != nil {
		questions := s.generateSuggestedQuestions(content, messages)
		if questionsJSON, err := json.Marshal(questions); err == nil {
			suggestedQuestions = questionsJSON
		} else {
			log.Printf("Warning: Failed to marshal suggested questions: %v", err)
		}
	}

	chatMessage := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid:    sessionUuid,
		Uuid:               uuid,
		Role:               role,
		Content:            content,
		ReasoningContent:   reasoningContent,
		Model:              model,
		UserID:             userId,
		CreatedBy:          userId,
		UpdatedBy:          userId,
		LlmSummary:         summary,
		TokenCount:         int32(numTokens),
		Raw:                json.RawMessage([]byte("{}")),
		Artifacts:          artifactsJSON,
		SuggestedQuestions: suggestedQuestions,
	}
	message, err := s.q.CreateChatMessage(ctx, chatMessage)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
}

// generateSuggestedQuestions generates follow-up questions based on the conversation context
func (s *ChatService) generateSuggestedQuestions(content string, messages []models.Message) []string {
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
	questions := s.callLLMForSuggestions(prompt)

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
func (s *ChatService) callLLMForSuggestions(prompt string) string {
	ctx := context.Background()

	// Get all models and find preferred models for suggestions
	allModels, err := s.q.ListChatModels(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list models for suggestions: %v", err)
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
		log.Printf("Warning: No suitable models available for suggestions")
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

	log.Printf("Warning: Unsupported model type for suggestions: %s", selectedModel.ApiType)
	return ""
}

// callGeminiForSuggestions makes a Gemini API call for suggestions
func (s *ChatService) callGeminiForSuggestions(ctx context.Context, model sqlc_queries.ChatModel, prompt string) string {
	// Validate API key
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Printf("Warning: GEMINI_API_KEY environment variable not set")
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
		log.Printf("Warning: Failed to generate Gemini payload for suggestions: %v", err)
		return ""
	}

	// Build URL
	url := gemini.BuildAPIURL(model.Name, false)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Warning: Failed to create Gemini request for suggestions: %v", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	// Make the API call with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	answer, err := gemini.HandleRegularResponse(http.Client{Timeout: 30 * time.Second}, req)
	if err != nil {
		log.Printf("Warning: Failed to get Gemini response for suggestions: %v", err)
		return ""
	}

	if answer == nil || answer.Answer == "" {
		log.Printf("Warning: Empty response from Gemini for suggestions")
		return ""
	}

	return answer.Answer
}

// callOpenAICompatibleForSuggestions makes an OpenAI-compatible API call for suggestions (including deepseek)
func (s *ChatService) callOpenAICompatibleForSuggestions(ctx context.Context, model sqlc_queries.ChatModel, prompt string) string {
	// Generate OpenAI client configuration
	config, err := genOpenAIConfig(model)
	if err != nil {
		log.Printf("Warning: Failed to generate OpenAI configuration for suggestions: %v", err)
		return ""
	}

	client := openai.NewClientWithConfig(config)

	// Create a simple chat completion request for generating suggestions
	req := openai.ChatCompletionRequest{
		Model:       model.Name,
		Temperature: 0.7, // Slightly creative but not too random
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
		log.Printf("Warning: Failed to generate suggested questions with %s: %v", model.Name, err)
		return ""
	}

	if len(resp.Choices) == 0 {
		log.Printf("Warning: No response choices returned for suggested questions from %s", model.Name)
		return ""
	}

	return resp.Choices[0].Message.Content
}

// UpdateChatMessageContent updates the content of an existing chat message.
// Recalculates token count for the updated content.
func (s *ChatService) UpdateChatMessageContent(ctx context.Context, uuid, content string) error {
	// encode
	// num_tokens
	num_tokens, err := getTokenCount(content)
	if err != nil {
		log.Printf("Warning: Failed to get token count for update: %v", err)
		num_tokens = len(content) / TokenEstimateRatio // Fallback estimate
	}

	err = s.q.UpdateChatMessageContent(ctx, sqlc_queries.UpdateChatMessageContentParams{
		Uuid:       uuid,
		Content:    content,
		TokenCount: int32(num_tokens),
	})
	return err
}

// UpdateChatMessageSuggestions updates the suggested questions for a chat message
func (s *ChatService) UpdateChatMessageSuggestions(ctx context.Context, uuid string, suggestedQuestions json.RawMessage) error {
	_, err := s.q.UpdateChatMessageSuggestions(ctx, sqlc_queries.UpdateChatMessageSuggestionsParams{
		Uuid:               uuid,
		SuggestedQuestions: suggestedQuestions,
	})
	return err
}

// logChat creates a chat log entry for analytics and debugging.
// Logs the session, messages, and LLM response for audit purposes.
func (s *ChatService) logChat(chatSession sqlc_queries.ChatSession, msgs []models.Message, answerText string) {
	// log chat
	sessionRaw := chatSession.ToRawMessage()
	if sessionRaw == nil {
		log.Println("failed to marshal chat session")
		return
	}
	question, err := json.Marshal(msgs)
	if err != nil {
		log.Printf("Warning: Failed to marshal chat messages: %v", err)
		return // Skip logging if marshalling fails
	}
	answerRaw, err := json.Marshal(answerText)
	if err != nil {
		log.Printf("Warning: Failed to marshal answer: %v", err)
		return // Skip logging if marshalling fails
	}

	s.q.CreateChatLog(context.Background(), sqlc_queries.CreateChatLogParams{
		Session:  *sessionRaw,
		Question: question,
		Answer:   answerRaw,
	})
}

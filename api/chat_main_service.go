package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/rotisserie/eris"
	"github.com/samber/lo"
	models "github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatService struct {
	q *sqlc_queries.Queries
}

// NewChatService creates a new ChatService with database queries.
func NewChatService(q *sqlc_queries.Queries) *ChatService {
	return &ChatService{q: q}
}

// loadArtifactInstruction loads the artifact instruction from file.
// Returns the instruction content or an error if the file cannot be read.
func loadArtifactInstruction() (string, error) {
	content, err := os.ReadFile("artifact_instruction.txt")
	if err != nil {
		return "", eris.Wrap(err, "failed to read artifact instruction file")
	}
	return string(content), nil
}

// getAskMessages retrieves and processes chat messages for LLM requests.
// It combines prompts and messages, applies length limits, and adds artifact instructions.
// Parameters:
//   - chatSession: The chat session containing configuration
//   - chatUuid: UUID for message identification (used in regenerate mode)
//   - regenerate: If true, excludes the target message from history
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

	// Add artifact instruction to system messages
	artifactInstruction, err := loadArtifactInstruction()
	if err != nil {
		log.Printf("Warning: Failed to load artifact instruction: %v", err)
		artifactInstruction = "" // Use empty string if file can't be loaded
	}

	// Append artifact instruction to system messages or add as new system message
	systemMsgFound := false
	for i, msg := range msgs {
		if msg.Role == "system" {
			msgs[i].Content = msg.Content + "\n" + artifactInstruction
			msgs[i].SetTokenCount(int32(len(msgs[i].Content) / TokenEstimateRatio)) // Rough token estimate
			systemMsgFound = true
			break
		}
	}

	if !systemMsgFound {
		// append to the first message
		msgs[0].Content = msgs[0].Content + "\n" + artifactInstruction
		msgs[0].SetTokenCount(int32(len(msgs[0].Content) / TokenEstimateRatio)) // Rough token estimate
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
		ChatSessionUuid:  sessionUuid,
		Uuid:             uuid,
		Role:             role,
		Content:          content,
		ReasoningContent: reasoningContent,
		Model:            model,
		UserID:           userId,
		CreatedBy:        userId,
		UpdatedBy:        userId,
		LlmSummary:       summary,
		TokenCount:       int32(numTokens),
		Raw:              json.RawMessage([]byte("{}")),
		Artifacts:        artifactsJSON,
	}
	message, err := s.q.CreateChatMessage(ctx, chatMessage)
	if err != nil {
		return sqlc_queries.ChatMessage{}, eris.Wrap(err, "failed to create message ")
	}
	return message, nil
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

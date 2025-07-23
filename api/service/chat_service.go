package service

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/swuecho/chat_backend/models"
	pkgerrors "github.com/swuecho/chat_backend/pkg/errors"
	"github.com/swuecho/chat_backend/repository"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type chatService struct {
	repos repository.CoreRepositoryManager
}

func NewChatService(repos repository.CoreRepositoryManager) ChatService {
	return &chatService{repos: repos}
}

func (s *chatService) GetChatSession(ctx context.Context, uuid string) (*sqlc_queries.ChatSession, error) {
	session, err := s.repos.ChatSession().GetByUUID(ctx, uuid)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &session, nil
}

func (s *chatService) CreateChatSession(ctx context.Context, userID int32, topic string, model string) (*sqlc_queries.ChatSession, error) {
	// Generate UUID for session
	sessionUUID := generateUUID()
	
	params := sqlc_queries.CreateChatSessionParams{
		UserID:    userID,
		Topic:     topic,
		MaxLength: 4000, // Default max length
		Uuid:      sessionUUID,
		Model:     model,
	}
	
	session, err := s.repos.ChatSession().Create(ctx, params)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	
	return &session, nil
}

func (s *chatService) GetUserChatSessions(ctx context.Context, userID int32) ([]sqlc_queries.ChatSession, error) {
	sessions, err := s.repos.ChatSession().GetByUserID(ctx, userID)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return sessions, nil
}

func (s *chatService) UpdateChatSession(ctx context.Context, uuid string, updates ChatSessionUpdate) (*sqlc_queries.ChatSession, error) {
	// For now, we'll implement topic updates only since that's what the SQLC methods support
	// Other updates like temperature, maxLength, topP would need additional SQLC methods
	
	if updates.Topic != nil {
		// Use the topic-specific update method
		params := sqlc_queries.UpdateChatSessionTopicByUUIDParams{
			Uuid:  uuid,
			Topic: *updates.Topic,
		}
		
		session, err := s.repos.ChatSession().UpdateTopicByUUID(ctx, params)
		if err != nil {
			return nil, pkgerrors.FromDatabaseError(err)
		}
		return &session, nil
	}
	
	// If no supported updates are provided, return the current session
	session, err := s.repos.ChatSession().GetByUUID(ctx, uuid)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	
	return &session, nil
}

func (s *chatService) DeleteChatSession(ctx context.Context, uuid string, userID int32) error {
	// First validate ownership
	if err := s.ValidateChatSession(ctx, uuid, userID); err != nil {
		return err
	}
	
	// Delete the session
	err := s.repos.ChatSession().Delete(ctx, uuid)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}
	
	return nil
}

func (s *chatService) GetAskMessages(ctx context.Context, chatSession sqlc_queries.ChatSession, chatUuid string, regenerate bool) ([]models.Message, error) {
	// Request timeout
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	chatSessionUuid := chatSession.Uuid
	lastN := chatSession.MaxLength
	if chatSession.MaxLength == 0 {
		lastN = 4000 // Default max length
	}

	// Get chat prompts first (system messages)
	var chatPromptMsgs []models.Message
	if prompts, err := s.getChatPrompts(ctx, chatSessionUuid); err != nil {
		log.Printf("Warning: Failed to get chat prompts: %v", err)
	} else {
		chatPromptMsgs = lo.Map(prompts, func(m sqlc_queries.ChatPrompt, _ int) models.Message {
			msg := models.Message{Role: m.Role, Content: m.Content}
			msg.SetTokenCount(int32(m.TokenCount))
			return msg
		})
	}

	// Retrieve chat messages
	var chatMessages []sqlc_queries.ChatMessage
	var err error
	
	if regenerate && chatUuid != "" {
		// For regenerate mode, get messages excluding the one being regenerated
		chatMessages, err = s.getMessagesExcluding(ctx, chatSessionUuid, chatUuid, lastN)
	} else {
		// Normal mode - get latest messages
		chatMessages, err = s.repos.ChatMessage().GetBySessionUUID(ctx, chatSessionUuid, lastN)
	}
	
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	// Convert to models.Message format
	chatMessageMsgs := lo.Map(chatMessages, func(m sqlc_queries.ChatMessage, _ int) models.Message {
		msg := models.Message{Role: m.Role, Content: m.Content}
		msg.SetTokenCount(int32(m.TokenCount))
		return msg
	})

	// Combine prompts and messages
	msgs := append(chatPromptMsgs, chatMessageMsgs...)

	// Add artifact instruction to system messages
	artifactInstruction, err := s.loadArtifactInstruction()
	if err != nil {
		log.Printf("Warning: Failed to load artifact instruction: %v", err)
		artifactInstruction = ""
	}

	// Enhance system messages with artifact instruction
	msgs = s.addArtifactInstruction(msgs, artifactInstruction)

	return msgs, nil
}

func (s *chatService) ProcessChatRequest(ctx context.Context, req ChatRequest) (*models.LLMAnswer, error) {
	// Get chat session
	session, err := s.GetChatSession(ctx, req.SessionUUID)
	if err != nil {
		return nil, err // Already properly formatted by GetChatSession
	}

	// Validate user access
	if err := s.ValidateChatSession(ctx, req.SessionUUID, req.UserID); err != nil {
		return nil, err
	}

	// Get messages for LLM
	_, err = s.GetAskMessages(ctx, *session, req.ChatUUID, req.Regenerate)
	if err != nil {
		return nil, err // Already properly formatted by GetAskMessages
	}

	// This would need to be implemented with model provider factory
	// For now, return placeholder
	return &models.LLMAnswer{
		Answer:   "Service layer implementation in progress",
		AnswerId: req.ChatUUID,
	}, nil
}

func (s *chatService) ValidateChatSession(ctx context.Context, sessionUUID string, userID int32) error {
	session, err := s.repos.ChatSession().GetByUUID(ctx, sessionUUID)
	if err != nil {
		return pkgerrors.NotFound("chat session")
	}
	
	if session.UserID != userID {
		return pkgerrors.ErrForbidden.WithDetail("unauthorized access to chat session")
	}
	
	return nil
}

// CreateChatMessage creates a new chat message in the session
func (s *chatService) CreateChatMessage(ctx context.Context, sessionUUID, messageUUID, role, content, model string, userID int32) (*sqlc_queries.ChatMessage, error) {
	// Validate session ownership first
	if err := s.ValidateChatSession(ctx, sessionUUID, userID); err != nil {
		return nil, err
	}

	// Calculate token count
	tokenCount := s.estimateTokenCount(content)

	params := sqlc_queries.CreateChatMessageParams{
		ChatSessionUuid: sessionUUID,
		Uuid:            messageUUID,
		Role:            role,
		Content:         content,
		Model:           model,
		UserID:          userID,
		CreatedBy:       userID,
		UpdatedBy:       userID,
		TokenCount:      tokenCount,
		// Other fields would be set based on requirements
	}

	message, err := s.repos.ChatMessage().Create(ctx, params)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	return &message, nil
}

// UpdateChatMessage updates the content of an existing chat message
func (s *chatService) UpdateChatMessage(ctx context.Context, messageUUID, content string, userID int32) (*sqlc_queries.ChatMessage, error) {
	// Get the message first to validate ownership through session
	message, err := s.repos.ChatMessage().GetByUUID(ctx, messageUUID)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	// Validate session ownership
	if err := s.ValidateChatSession(ctx, message.ChatSessionUuid, userID); err != nil {
		return nil, err
	}

	// Update the message content using UpdateByUUID
	params := sqlc_queries.UpdateChatMessageByUUIDParams{
		Uuid:       messageUUID,
		Content:    content,
		TokenCount: s.estimateTokenCount(content),
		IsPin:      message.IsPin, // Preserve existing pin status
		Artifacts:  message.Artifacts, // Preserve existing artifacts
	}

	updatedMessage, err := s.repos.ChatMessage().UpdateByUUID(ctx, params)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	
	return &updatedMessage, nil
}

// Helper methods
func (s *chatService) loadArtifactInstruction() (string, error) {
	content, err := os.ReadFile("artifact_instruction.txt")
	if err != nil {
		// This is not a critical error, log but don't fail
		log.Printf("Warning: Failed to read artifact instruction file: %v", err)
		return "", nil
	}
	return string(content), nil
}

func (s *chatService) getChatPrompts(ctx context.Context, sessionUUID string) ([]sqlc_queries.ChatPrompt, error) {
	return s.repos.ChatPrompt().GetBySessionUUID(ctx, sessionUUID)
}

func (s *chatService) getMessagesExcluding(ctx context.Context, sessionUUID, excludeUUID string, limit int32) ([]sqlc_queries.ChatMessage, error) {
	// Get all messages and filter out the excluded one
	// This is a simplified implementation - ideally this would be done at the database level
	messages, err := s.repos.ChatMessage().GetBySessionUUID(ctx, sessionUUID, limit+1) // Get one extra in case we need to exclude
	if err != nil {
		return nil, err
	}
	
	// Filter out the message with excludeUUID
	filtered := make([]sqlc_queries.ChatMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Uuid != excludeUUID {
			filtered = append(filtered, msg)
		}
	}
	
	// Limit to requested number
	if len(filtered) > int(limit) {
		filtered = filtered[:limit]
	}
	
	return filtered, nil
}

func (s *chatService) addArtifactInstruction(msgs []models.Message, artifactInstruction string) []models.Message {
	if artifactInstruction == "" {
		return msgs
	}

	// Find system message and append instruction
	systemMsgFound := false
	for i, msg := range msgs {
		if msg.Role == "system" {
			msgs[i].Content = msg.Content + "\n" + artifactInstruction
			msgs[i].SetTokenCount(s.estimateTokenCount(msgs[i].Content))
			systemMsgFound = true
			break
		}
	}

	// If no system message found, add one at the beginning
	if !systemMsgFound {
		systemMsg := models.Message{
			Role:    "system",
			Content: artifactInstruction,
		}
		systemMsg.SetTokenCount(s.estimateTokenCount(artifactInstruction))
		msgs = append([]models.Message{systemMsg}, msgs...)
	}

	return msgs
}

func (s *chatService) estimateTokenCount(content string) int32 {
	// Rough estimation: 1 token per 4 characters
	// This should be replaced with proper token counting in production
	return int32(len(content) / 4)
}

// generateUUID creates a proper UUID for sessions
func generateUUID() string {
	return uuid.New().String()
}

// Additional message management methods

func (s *chatService) GetChatMessageByID(ctx context.Context, id int32) (*sqlc_queries.ChatMessage, error) {
	message, err := s.repos.ChatMessage().GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &message, nil
}

func (s *chatService) GetChatMessageByUUID(ctx context.Context, uuid string) (*sqlc_queries.ChatMessage, error) {
	message, err := s.repos.ChatMessage().GetByUUID(ctx, uuid)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &message, nil
}

func (s *chatService) UpdateChatMessageByUUID(ctx context.Context, params sqlc_queries.UpdateChatMessageByUUIDParams, userID int32) (*sqlc_queries.ChatMessage, error) {
	// Get the message first to validate ownership through session
	message, err := s.repos.ChatMessage().GetByUUID(ctx, params.Uuid)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	// Validate session ownership
	if err := s.ValidateChatSession(ctx, message.ChatSessionUuid, userID); err != nil {
		return nil, err
	}

	// Update the message
	updatedMessage, err := s.repos.ChatMessage().UpdateByUUID(ctx, params)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	
	return &updatedMessage, nil
}

func (s *chatService) DeleteChatMessage(ctx context.Context, id int32, userID int32) error {
	// Get the message first to validate ownership through session
	message, err := s.repos.ChatMessage().GetByID(ctx, id)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}

	// Validate session ownership
	if err := s.ValidateChatSession(ctx, message.ChatSessionUuid, userID); err != nil {
		return err
	}

	// Delete the message
	err = s.repos.ChatMessage().Delete(ctx, id)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}
	
	return nil
}

func (s *chatService) DeleteChatMessageByUUID(ctx context.Context, uuid string, userID int32) error {
	// Get the message first to validate ownership through session
	message, err := s.repos.ChatMessage().GetByUUID(ctx, uuid)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}

	// Validate session ownership
	if err := s.ValidateChatSession(ctx, message.ChatSessionUuid, userID); err != nil {
		return err
	}

	// Delete the message
	err = s.repos.ChatMessage().DeleteByUUID(ctx, uuid)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}
	
	return nil
}

func (s *chatService) GetAllChatMessages(ctx context.Context) ([]sqlc_queries.ChatMessage, error) {
	messages, err := s.repos.ChatMessage().GetAll(ctx)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return messages, nil
}

func (s *chatService) GetLatestMessagesBySessionID(ctx context.Context, chatSessionUuid string, limit int32) ([]sqlc_queries.ChatMessage, error) {
	messages, err := s.repos.ChatMessage().GetLatestBySessionUUID(ctx, chatSessionUuid, limit)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return messages, nil
}

func (s *chatService) GetFirstMessageBySessionUUID(ctx context.Context, chatSessionUuid string) (*sqlc_queries.ChatMessage, error) {
	message, err := s.repos.ChatMessage().GetFirstBySessionUUID(ctx, chatSessionUuid)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &message, nil
}

func (s *chatService) GetChatMessagesBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]sqlc_queries.ChatMessage, error) {
	// Use the existing repository method with pagination
	messages, err := s.repos.ChatMessage().GetBySessionUUID(ctx, uuid, pageSize)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	
	// Simple pagination logic - skip messages based on pageNum
	// This is a simplified implementation
	offset := (pageNum - 1) * pageSize
	if offset >= int32(len(messages)) {
		return []sqlc_queries.ChatMessage{}, nil
	}
	
	end := offset + pageSize
	if end > int32(len(messages)) {
		end = int32(len(messages))
	}
	
	return messages[offset:end], nil
}

func (s *chatService) DeleteChatMessagesBySessionUUID(ctx context.Context, uuid string, userID int32) error {
	// Validate session ownership
	if err := s.ValidateChatSession(ctx, uuid, userID); err != nil {
		return err
	}

	// Get all messages for the session and delete them
	messages, err := s.repos.ChatMessage().GetBySessionUUID(ctx, uuid, 10000) // Large limit to get all
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}

	// Delete each message
	for _, message := range messages {
		if err := s.repos.ChatMessage().Delete(ctx, message.ID); err != nil {
			return pkgerrors.FromDatabaseError(err)
		}
	}
	
	return nil
}

func (s *chatService) GetChatMessagesCount(ctx context.Context, userID int32) (int32, error) {
	// This would require a specific repository method for counting messages by user
	// For now, implement by getting user sessions and counting their messages
	sessions, err := s.repos.ChatSession().GetByUserID(ctx, userID)
	if err != nil {
		return 0, pkgerrors.FromDatabaseError(err)
	}
	
	var totalCount int32
	for _, session := range sessions {
		messages, err := s.repos.ChatMessage().GetBySessionUUID(ctx, session.Uuid, 10000) // Large limit
		if err != nil {
			continue // Skip errors for now
		}
		totalCount += int32(len(messages))
	}
	
	return totalCount, nil
}
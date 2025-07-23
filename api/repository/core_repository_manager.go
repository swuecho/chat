package repository

import "github.com/swuecho/chat_backend/sqlc_queries"

type coreRepositoryManager struct {
	chatSession ChatSessionRepository
	chatMessage ChatMessageRepository
	chatModel   ChatModelRepository
	chatPrompt  ChatPromptRepository
}

// NewCoreRepositoryManager creates a new core repository manager with essential repositories
func NewCoreRepositoryManager(queries *sqlc_queries.Queries) CoreRepositoryManager {
	return &coreRepositoryManager{
		chatSession: NewChatSessionRepository(queries),
		chatMessage: NewChatMessageRepository(queries),
		chatModel:   NewChatModelRepository(queries),
		chatPrompt:  NewChatPromptRepository(queries),
	}
}

func (rm *coreRepositoryManager) ChatSession() ChatSessionRepository {
	return rm.chatSession
}

func (rm *coreRepositoryManager) ChatMessage() ChatMessageRepository {
	return rm.chatMessage
}

func (rm *coreRepositoryManager) ChatModel() ChatModelRepository {
	return rm.chatModel
}

func (rm *coreRepositoryManager) ChatPrompt() ChatPromptRepository {
	return rm.chatPrompt
}
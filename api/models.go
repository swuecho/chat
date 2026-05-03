package main

import (
	"net/http"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// Re-export commonly used DTOs for backward compatibility within package main.
type (
	TokenResult                    = dto.TokenResult
	ConversationRequest            = dto.ConversationRequest
	RequestOption                  = dto.RequestOption
	Artifact                       = dto.Artifact
	SimpleChatMessage              = dto.SimpleChatMessage
	SimpleChatSession              = dto.SimpleChatSession
	ChatMessageResponse            = dto.ChatMessageResponse
	ChatSessionResponse            = dto.ChatSessionResponse
	Pagination                     = dto.Pagination
	UpdateChatSessionRequest       = dto.UpdateChatSessionRequest
	CreateWorkspaceRequest         = dto.CreateWorkspaceRequest
	UpdateWorkspaceRequest         = dto.UpdateWorkspaceRequest
	UpdateWorkspaceOrderRequest    = dto.UpdateWorkspaceOrderRequest
	WorkspaceResponse              = dto.WorkspaceResponse
	CreateSessionInWorkspaceRequest = dto.CreateSessionInWorkspaceRequest
	ChatInstructionResponse        = dto.ChatInstructionResponse
)

// ChatModel interface - used by all LLM providers.
type ChatModel interface {
	Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chatMessages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error)
}

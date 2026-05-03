// Package main — Core interfaces and type aliases for backward compatibility.
package main

import (
	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/provider"
)

// ChatModel is an alias for provider.ChatModel.
type ChatModel = provider.ChatModel

// --- DTO type aliases for backward compatibility ---

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

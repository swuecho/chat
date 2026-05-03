// Package main — Service shims that delegate to the svc package.
// These exist for backward compatibility; new code should use svc.* directly.
package main

import (
	"github.com/swuecho/chat_backend/sqlc_queries"
	"github.com/swuecho/chat_backend/svc"
)

// Service type aliases point to the svc package implementations.
type (
	ChatSessionService           = svc.ChatSessionService
	ChatWorkspaceService         = svc.ChatWorkspaceService
	ChatMessageService           = svc.ChatMessageService
	ChatPromptService            = svc.ChatPromptService
	ChatSnapshotService          = svc.ChatSnapshotService
	ChatFileService              = svc.ChatFileService
	ChatService                  = svc.ChatService
	ChatCommentService           = svc.ChatCommentService
	AuthUserService              = svc.AuthUserService
	BotAnswerHistoryService      = svc.BotAnswerHistoryService
	UserActiveChatSessionService = svc.UserActiveChatSessionService
	JWTSecretService             = svc.JWTSecretService
)

// Constructor wrappers delegate to svc package.
var (
	NewChatSessionService           = svc.NewChatSessionService
	NewChatWorkspaceService         = svc.NewChatWorkspaceService
	NewChatMessageService           = svc.NewChatMessageService
	NewChatPromptService            = svc.NewChatPromptService
	NewChatSnapshotService          = svc.NewChatSnapshotService
	NewChatFileService              = svc.NewChatFileService
	NewChatService                  = svc.NewChatService
	NewChatCommentService           = svc.NewChatCommentService
	NewAuthUserService              = svc.NewAuthUserService
	NewBotAnswerHistoryService      = svc.NewBotAnswerHistoryService
	NewUserActiveChatSessionService = svc.NewUserActiveChatSessionService
	NewJWTSecretService             = svc.NewJWTSecretService
)

// Re-export types defined in svc that are used by handlers.
type (
	SessionHistoryInfo = svc.SessionHistoryInfo
	UserAnalysisData   = svc.UserAnalysisData
	UserAnalysisInfo   = svc.UserAnalysisInfo
	ModelUsageInfo     = svc.ModelUsageInfo
	ActivityInfo       = svc.ActivityInfo
	AutoMigrateLegacySessionsResult = svc.AutoMigrateLegacySessionsResult
)

// Ensure sqlc_queries import is used.
var _ = (*sqlc_queries.Queries)(nil)

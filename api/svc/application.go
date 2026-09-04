package svc

import "github.com/swuecho/chat_backend/sqlc_queries"

// ApplicationServices is the composition container for shared application
// services. It owns one instance of each service used by HTTP adapters.
type ApplicationServices struct {
	ChatModels      *ChatModelService
	APIKeys         *APIKeyService
	AuthUsers       *AuthUserService
	AdminSessions   *SessionAdminQueryService
	Prompts         *ChatPromptService
	Sessions        *ChatSessionService
	ActiveSessions  *UserActiveChatSessionService
	Workspaces      *ChatWorkspaceService
	Messages        *ChatMessageService
	Artifacts       *ArtifactService
	Conversations   *SessionConversationService
	Snapshots       *ChatSnapshotService
	Chat            *ChatService
	RateLimits      *SessionRateLimitService
	SnapshotQueries *SessionSnapshotQueryService
	RuntimeModels   *SessionModelService
	BotHistory      *SessionBotHistoryService
	ModelPrivileges *ChatModelPrivilegeService
	Files           *ChatFileService
	Comments        *ChatCommentService
	BotHistoryCRUD  *BotAnswerHistoryService
	ChatUseCases    *ChatUseCaseFactory
}

type ChatUseCaseFactory struct {
	chat          *ChatService
	sessions      *ChatSessionService
	conversations *SessionConversationService
	botHistory    *SessionBotHistoryService
}

func (f *ChatUseCaseFactory) Complete(models ModelSelector, chunks AnswerChunkSink, policy ChatLogPolicy) *CompleteChatUseCase {
	return NewCompleteChatUseCase(f.chat, f.sessions, f.conversations, models, chunks, policy, f.chat.AuditLogger())
}
func (f *ChatUseCaseFactory) Bot(models ModelSelector, chunks AnswerChunkSink, policy ChatLogPolicy) *GenerateBotAnswerUseCase {
	return NewGenerateBotAnswerUseCase(f.chat, models, chunks, policy, f.botHistory, f.chat.AuditLogger())
}
func (f *ChatUseCaseFactory) Regenerate(models ModelSelector, chunks AnswerChunkSink, policy ChatLogPolicy) *RegenerateAnswerUseCase {
	return NewRegenerateAnswerUseCase(f.chat, f.sessions, models, chunks, policy, f.chat.SuggestionGenerator(), f.chat.AuditLogger())
}

func NewApplicationServices(q *sqlc_queries.Queries, openAIKey, openAIProxy, jwtSecret string, rateLimit int32) *ApplicationServices {
	app := &ApplicationServices{
		ChatModels: NewChatModelService(q), APIKeys: NewAPIKeyService(q),
		AuthUsers: NewAuthUserService(q, jwtSecret, rateLimit), AdminSessions: NewSessionAdminQueryService(q),
		Prompts: NewChatPromptService(q), Sessions: NewChatSessionService(q),
		ActiveSessions: NewUserActiveChatSessionService(q), Workspaces: NewChatWorkspaceService(q),
		Messages: NewChatMessageService(q), Artifacts: NewArtifactService(q), Conversations: NewSessionConversationService(q),
		Snapshots: NewChatSnapshotService(q), Chat: NewChatService(q, openAIKey, openAIProxy),
		RateLimits: NewSessionRateLimitService(q), SnapshotQueries: NewSessionSnapshotQueryService(q),
		RuntimeModels: NewSessionModelService(q), BotHistory: NewSessionBotHistoryService(q),
		ModelPrivileges: NewChatModelPrivilegeService(q), Files: NewChatFileService(q),
		Comments: NewChatCommentService(q), BotHistoryCRUD: NewBotAnswerHistoryService(q),
	}
	app.ChatUseCases = &ChatUseCaseFactory{chat: app.Chat, sessions: app.Sessions, conversations: app.Conversations, botHistory: app.BotHistory}
	return app
}

package service

import (
	"context"
	"net/http"

	"github.com/swuecho/chat_backend/models"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatService handles chat business logic
type ChatService interface {
	// GetChatSession retrieves a chat session by UUID
	GetChatSession(ctx context.Context, uuid string) (*sqlc_queries.ChatSession, error)
	
	// GetUserChatSessions retrieves all chat sessions for a user
	GetUserChatSessions(ctx context.Context, userID int32) ([]sqlc_queries.ChatSession, error)
	
	// CreateChatSession creates a new chat session
	CreateChatSession(ctx context.Context, userID int32, topic string, model string) (*sqlc_queries.ChatSession, error)
	
	// UpdateChatSession updates a chat session with provided fields
	UpdateChatSession(ctx context.Context, uuid string, updates ChatSessionUpdate) (*sqlc_queries.ChatSession, error)
	
	// DeleteChatSession deletes a chat session
	DeleteChatSession(ctx context.Context, uuid string, userID int32) error
	
	// GetAskMessages retrieves and processes chat messages for LLM requests
	GetAskMessages(ctx context.Context, chatSession sqlc_queries.ChatSession, chatUuid string, regenerate bool) ([]models.Message, error)
	
	// ProcessChatRequest handles a chat request and returns an LLM answer
	ProcessChatRequest(ctx context.Context, req ChatRequest) (*models.LLMAnswer, error)
	
	// ValidateChatSession validates if a user can access a chat session
	ValidateChatSession(ctx context.Context, sessionUUID string, userID int32) error
	
	// Message operations
	GetChatMessageByID(ctx context.Context, id int32) (*sqlc_queries.ChatMessage, error)
	GetChatMessageByUUID(ctx context.Context, uuid string) (*sqlc_queries.ChatMessage, error)
	CreateChatMessage(ctx context.Context, sessionUUID, messageUUID, role, content, model string, userID int32) (*sqlc_queries.ChatMessage, error)
	UpdateChatMessage(ctx context.Context, messageUUID, content string, userID int32) (*sqlc_queries.ChatMessage, error)
	UpdateChatMessageByUUID(ctx context.Context, params sqlc_queries.UpdateChatMessageByUUIDParams, userID int32) (*sqlc_queries.ChatMessage, error)
	DeleteChatMessage(ctx context.Context, id int32, userID int32) error
	DeleteChatMessageByUUID(ctx context.Context, uuid string, userID int32) error
	GetAllChatMessages(ctx context.Context) ([]sqlc_queries.ChatMessage, error)
	GetLatestMessagesBySessionID(ctx context.Context, chatSessionUuid string, limit int32) ([]sqlc_queries.ChatMessage, error)
	GetFirstMessageBySessionUUID(ctx context.Context, chatSessionUuid string) (*sqlc_queries.ChatMessage, error)
	GetChatMessagesBySessionUUID(ctx context.Context, uuid string, pageNum, pageSize int32) ([]sqlc_queries.ChatMessage, error)
	DeleteChatMessagesBySessionUUID(ctx context.Context, uuid string, userID int32) error
	GetChatMessagesCount(ctx context.Context, userID int32) (int32, error)
}

// AuthService handles authentication business logic
type AuthService interface {
	// Login authenticates a user and returns tokens
	Login(ctx context.Context, username string, password string) (*LoginResult, error)
	
	// ValidateToken validates a JWT token and returns user info
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	
	// RefreshToken refreshes an access token
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResult, error)
}

// ModelService handles model provider business logic
type ModelService interface {
	// GetAvailableModels returns models available to a user
	GetAvailableModels(ctx context.Context, userID int32) ([]sqlc_queries.ChatModel, error)
	
	// GetSystemModels returns all system models with usage statistics
	GetSystemModelsWithUsage(ctx context.Context, timePeriod string) ([]ChatModelWithUsage, error)
	
	// GetModelByID retrieves a specific model by ID
	GetModelByID(ctx context.Context, modelID int32) (*sqlc_queries.ChatModel, error)
	
	// GetModelByName retrieves a specific model by name
	GetModelByName(ctx context.Context, name string) (*sqlc_queries.ChatModel, error)
	
	// GetDefaultModel returns the default chat model
	GetDefaultModel(ctx context.Context) (*sqlc_queries.ChatModel, error)
	
	// CreateModel creates a new chat model
	CreateModel(ctx context.Context, params ChatModelCreateRequest) (*sqlc_queries.ChatModel, error)
	
	// UpdateModel updates an existing model
	UpdateModel(ctx context.Context, modelID int32, updates ChatModelUpdateRequest) (*sqlc_queries.ChatModel, error)
	
	// DeleteModel deletes a model
	DeleteModel(ctx context.Context, modelID int32, userID int32) error
	
	// CreateModelInstance creates a ChatModel instance for processing
	CreateModelInstance(ctx context.Context, modelName string) (ChatModel, error)
}

// FileService handles file upload and management business logic
type FileService interface {
	// UploadFile uploads a file for a chat session
	UploadFile(ctx context.Context, sessionUUID string, userID int32, filename string, data []byte, mimeType string) (*sqlc_queries.ChatFile, error)
	
	// GetFiles retrieves files for a chat session
	GetFiles(ctx context.Context, sessionUUID string, userID int32) ([]sqlc_queries.ChatFile, error)
	
	// DeleteFile deletes a file
	DeleteFile(ctx context.Context, fileID int32, userID int32) error
}

// Data structures for service operations

// ChatSessionUpdate contains fields that can be updated in a chat session
type ChatSessionUpdate struct {
	Topic       *string  `json:"topic,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	MaxLength   *int32   `json:"max_length,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
}

// ChatModelWithUsage represents a chat model with usage statistics
type ChatModelWithUsage struct {
	sqlc_queries.ChatModel
	LastUsageTime string `json:"lastUsageTime,omitempty"`
	MessageCount  int64  `json:"messageCount"`
}

// ChatModelCreateRequest represents parameters for creating a new model
type ChatModelCreateRequest struct {
	Name                   string `json:"name"`
	Label                  string `json:"label"`
	IsDefault              bool   `json:"is_default"`
	Url                    string `json:"url"`
	ApiAuthHeader          string `json:"api_auth_header"`
	ApiAuthKey             string `json:"api_auth_key"`
	UserID                 int32  `json:"user_id"`
	EnablePerModeRatelimit bool   `json:"enable_per_mode_ratelimit"`
	MaxToken               int32  `json:"max_token"`
	DefaultToken           int32  `json:"default_token"`
}

// ChatModelUpdateRequest represents parameters for updating a model
type ChatModelUpdateRequest struct {
	Name                   *string `json:"name,omitempty"`
	Label                  *string `json:"label,omitempty"`
	IsDefault              *bool   `json:"is_default,omitempty"`
	Url                    *string `json:"url,omitempty"`
	ApiAuthHeader          *string `json:"api_auth_header,omitempty"`
	ApiAuthKey             *string `json:"api_auth_key,omitempty"`
	EnablePerModeRatelimit *bool   `json:"enable_per_mode_ratelimit,omitempty"`
	MaxToken               *int32  `json:"max_token,omitempty"`
	DefaultToken           *int32  `json:"default_token,omitempty"`
}

// Request/Response types
type ChatRequest struct {
	SessionUUID string
	ChatUUID    string
	Prompt      string
	Regenerate  bool
	Stream      bool
	UserID      int32
}

type LoginResult struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	User         *sqlc_queries.AuthUser `json:"user"`
}

type TokenClaims struct {
	UserID   int32  `json:"userId"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}

// ChatModel interface (from models.go but duplicated here to avoid circular dependencies)
type ChatModel interface {
	Stream(w http.ResponseWriter, chatSession sqlc_queries.ChatSession, chat_compeletion_messages []models.Message, chatUuid string, regenerate bool, stream bool) (*models.LLMAnswer, error)
}

// ServiceManager aggregates all services
type ServiceManager interface {
	Chat() ChatService
	Auth() AuthService
	Model() ModelService
	File() FileService
}
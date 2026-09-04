package handler

import (
	"time"

	"github.com/swuecho/chat_backend/dto"
	"github.com/swuecho/chat_backend/svc"
)

type chatModelHTTPResponse struct {
	ID                      int32  `json:"id"`
	Name                    string `json:"name"`
	Label                   string `json:"label"`
	IsDefault               bool   `json:"isDefault"`
	URL                     string `json:"url"`
	APIAuthHeader           string `json:"apiAuthHeader"`
	APIAuthKey              string `json:"apiAuthKey"`
	UserID                  int32  `json:"userId"`
	EnablePerModelRateLimit bool   `json:"enablePerModeRatelimit"`
	MaxToken                *int32 `json:"maxToken,omitempty"`
	DefaultToken            *int32 `json:"defaultToken,omitempty"`
	OrderNumber             int32  `json:"orderNumber"`
	HTTPTimeout             int32  `json:"httpTimeOut"`
	IsEnable                bool   `json:"isEnable"`
	APIType                 string `json:"apiType"`
	IsTitleModel            bool   `json:"isTitleModel"`
}

type authUserHTTPResponse struct {
	ID          int32     `json:"id"`
	LastLogin   time.Time `json:"lastLogin"`
	IsSuperuser bool      `json:"isSuperuser"`
	Username    string    `json:"username"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Email       string    `json:"email"`
	IsStaff     bool      `json:"isStaff"`
	IsActive    bool      `json:"isActive"`
	DateJoined  time.Time `json:"dateJoined"`
}

type updatedUserHTTPResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type chatModelWithUsageHTTPResponse struct {
	chatModelHTTPResponse
	LastUsageTime time.Time `json:"lastUsageTime,omitempty"`
	MessageCount  int64     `json:"messageCount"`
}

type uuidHTTPResponse struct {
	UUID string `json:"uuid"`
}

type sessionCreatedHTTPResponse struct {
	SessionUUID string `json:"sessionUuid" jsonschema:"format=uuid"`
}

type countHTTPResponse struct {
	Count int64 `json:"count"`
}

type rateHTTPResponse struct {
	Rate int32 `json:"rate"`
}

type messageHTTPStatusResponse struct {
	Message string `json:"message"`
}

type fileUploadHTTPResponse struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Type string `json:"type"`
	Size string `json:"size"`
}

type fileMetaHTTPResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type paginatedHTTPResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalPages int64 `json:"totalPages"`
	TotalCount int64 `json:"totalCount"`
}

func newPaginatedHTTPResponse[T any](items []T, totalPages, totalCount int64) paginatedHTTPResponse[T] {
	return paginatedHTTPResponse[T]{Items: items, TotalPages: totalPages, TotalCount: totalCount}
}

type snapshotPageHTTPResponse struct {
	Data     []snapshotSummaryHTTPResponse `json:"data"`
	Page     int32                         `json:"page"`
	PageSize int32                         `json:"page_size"`
	Total    int64                         `json:"total"`
}

type snapshotListHTTPResponse struct {
	Items  []snapshotSummaryHTTPResponse `json:"items"`
	Total  int64                         `json:"total"`
	Limit  int32                         `json:"limit"`
	Offset int32                         `json:"offset"`
}

type userStatsPageHTTPResponse struct {
	Items  []UserStat `json:"items"`
	Total  int64      `json:"total"`
	Limit  int32      `json:"limit"`
	Offset int32      `json:"offset"`
}

type sessionHistoryPageHTTPResponse struct {
	Items  []svc.SessionHistoryInfo `json:"items"`
	Total  int64                    `json:"total"`
	Limit  int32                    `json:"limit"`
	Offset int32                    `json:"offset"`
}

type migrationHTTPResponse struct {
	HasLegacySessions bool                   `json:"hasLegacySessions"`
	MigratedSessions  int                    `json:"migratedSessions"`
	DefaultWorkspace  *dto.WorkspaceResponse `json:"defaultWorkspace,omitempty"`
}

type workspaceSessionCreatedHTTPResponse struct {
	UUID            string `json:"uuid"`
	Topic           string `json:"topic"`
	Model           string `json:"model"`
	ArtifactEnabled bool   `json:"artifactEnabled"`
	WorkspaceUUID   string `json:"workspaceUuid"`
	CreatedAt       string `json:"createdAt"`
}

type workspaceSessionHTTPResponse struct {
	UUID            string  `json:"uuid"`
	Title           string  `json:"title"`
	IsEdit          bool    `json:"isEdit"`
	Model           string  `json:"model"`
	WorkspaceUUID   string  `json:"workspaceUuid"`
	MaxLength       int32   `json:"maxLength"`
	Temperature     float64 `json:"temperature"`
	MaxTokens       int32   `json:"maxTokens"`
	TopP            float64 `json:"topP"`
	N               int32   `json:"n"`
	SummarizeMode   bool    `json:"summarizeMode"`
	ExploreMode     bool    `json:"exploreMode"`
	ArtifactEnabled bool    `json:"artifactEnabled"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type chatSessionHTTPResponse struct {
	ID              int32     `json:"id"`
	UserID          int32     `json:"userId"`
	UUID            string    `json:"uuid"`
	Topic           string    `json:"topic"`
	Active          bool      `json:"active"`
	Model           string    `json:"model"`
	MaxLength       int32     `json:"maxLength"`
	Temperature     float64   `json:"temperature"`
	TopP            float64   `json:"topP"`
	MaxTokens       int32     `json:"maxTokens"`
	N               int32     `json:"n"`
	SummarizeMode   bool      `json:"summarizeMode"`
	WorkspaceID     *int32    `json:"workspaceId"`
	ArtifactEnabled bool      `json:"artifactEnabled"`
	ExploreMode     bool      `json:"exploreMode"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func chatSessionResponse(session svc.ChatSession) chatSessionHTTPResponse {
	return chatSessionHTTPResponse{ID: session.ID, UserID: session.UserID, UUID: session.UUID,
		Topic: session.Topic, Active: session.Active, Model: session.Model, MaxLength: session.MaxLength,
		Temperature: session.Temperature, TopP: session.TopP, MaxTokens: session.MaxTokens, N: session.N,
		SummarizeMode: session.SummarizeMode, WorkspaceID: session.WorkspaceID,
		ArtifactEnabled: session.ArtifactEnabled, ExploreMode: session.ExploreMode,
		CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt}
}

type suggestionsHTTPResponse struct {
	NewSuggestions []string `json:"newSuggestions"`
	AllSuggestions []string `json:"allSuggestions"`
}

type gatewayRequestDetailHTTPResponse struct {
	ID                     int64                         `json:"id"`
	RequestUUID            string                        `json:"requestUuid" jsonschema:"format=uuid"`
	RequestedModel         string                        `json:"requestedModel"`
	Provider               string                        `json:"provider"`
	Status                 string                        `json:"status"`
	Stream                 bool                          `json:"stream"`
	PromptTokens           int32                         `json:"promptTokens"`
	CompletionTokens       int32                         `json:"completionTokens"`
	TotalTokens            int32                         `json:"totalTokens"`
	LatencyMs              int64                         `json:"latencyMs"`
	ProviderRequestID      string                        `json:"providerRequestId"`
	ErrorCode              string                        `json:"errorCode"`
	RequestBytes           int64                         `json:"requestBytes"`
	ResponseBytes          int64                         `json:"responseBytes"`
	RequestSHA256          string                        `json:"requestSha256"`
	ResponseSHA256         string                        `json:"responseSha256"`
	RequestTruncated       bool                          `json:"requestTruncated"`
	ResponseTruncated      bool                          `json:"responseTruncated"`
	RequestClassification  gatewayRequestClassification  `json:"requestClassification"`
	ResponseClassification gatewayResponseClassification `json:"responseClassification"`
	CreatedAt              time.Time                     `json:"createdAt"`
	CompletedAt            *time.Time                    `json:"completedAt" jsonschema:"nullable"`
	RetentionUntil         time.Time                     `json:"retentionUntil"`
	RequestCapture         capturedSample                `json:"requestCapture"`
	ResponseCapture        capturedSample                `json:"responseCapture"`
}

type gatewayRequestSummaryHTTPResponse struct {
	ID                int64      `json:"id"`
	RequestUUID       string     `json:"requestUuid" jsonschema:"format=uuid"`
	RequestedModel    string     `json:"requestedModel"`
	Provider          string     `json:"provider"`
	Status            string     `json:"status"`
	Stream            bool       `json:"stream"`
	PromptTokens      int32      `json:"promptTokens"`
	CompletionTokens  int32      `json:"completionTokens"`
	TotalTokens       int32      `json:"totalTokens"`
	LatencyMs         int64      `json:"latencyMs"`
	RequestBytes      int64      `json:"requestBytes"`
	ResponseBytes     int64      `json:"responseBytes"`
	RequestTruncated  bool       `json:"requestTruncated"`
	ResponseTruncated bool       `json:"responseTruncated"`
	CreatedAt         time.Time  `json:"createdAt"`
	CompletedAt       *time.Time `json:"completedAt" jsonschema:"nullable"`
	RetentionUntil    time.Time  `json:"retentionUntil"`
	ErrorCode         string     `json:"errorCode"`
}

type openAIErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   *string `json:"param"`
	Code    string  `json:"code"`
}

type openAIErrorResponse struct {
	Error openAIErrorDetail `json:"error"`
}

type gatewayModelHTTPResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

type gatewayModelListHTTPResponse struct {
	Object string                     `json:"object"`
	Data   []gatewayModelHTTPResponse `json:"data"`
}

type gatewayRequestClassification struct {
	Format            string         `json:"format"`
	MessageCount      int            `json:"message_count,omitempty"`
	Roles             map[string]int `json:"roles,omitempty"`
	Multimodal        bool           `json:"multimodal,omitempty"`
	HasTools          bool           `json:"has_tools"`
	HasResponseFormat bool           `json:"has_response_format"`
}

type gatewayResponseClassification struct {
	Stream          bool   `json:"stream"`
	StatusCode      int    `json:"status_code"`
	ContentType     string `json:"content_type"`
	ContentEncoding string `json:"content_encoding"`
	Successful      bool   `json:"successful"`
}

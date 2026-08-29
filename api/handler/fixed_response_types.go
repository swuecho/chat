package handler

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/swuecho/chat_backend/dto"
)

type uuidHTTPResponse struct {
	UUID string `json:"uuid"`
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
	Debug           bool    `json:"debug"`
	SummarizeMode   bool    `json:"summarizeMode"`
	ExploreMode     bool    `json:"exploreMode"`
	ArtifactEnabled bool    `json:"artifactEnabled"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type suggestionsHTTPResponse struct {
	NewSuggestions []string `json:"newSuggestions"`
	AllSuggestions []string `json:"allSuggestions"`
}

type suggestedQuestionsDelta struct {
	Content            string   `json:"content"`
	SuggestedQuestions []string `json:"suggestedQuestions"`
}

type suggestedQuestionsChoice struct {
	Index        int                     `json:"index"`
	Delta        suggestedQuestionsDelta `json:"delta"`
	FinishReason *string                 `json:"finish_reason"`
}

type suggestedQuestionsChunk struct {
	ID      string                     `json:"id"`
	Object  string                     `json:"object"`
	Choices []suggestedQuestionsChoice `json:"choices"`
}

type gatewayRequestDetailHTTPResponse struct {
	ID                     int64           `json:"id"`
	RequestUUID            uuid.UUID       `json:"requestUuid"`
	RequestedModel         string          `json:"requestedModel"`
	Provider               string          `json:"provider"`
	Status                 string          `json:"status"`
	Stream                 bool            `json:"stream"`
	PromptTokens           int32           `json:"promptTokens"`
	CompletionTokens       int32           `json:"completionTokens"`
	TotalTokens            int32           `json:"totalTokens"`
	LatencyMs              int64           `json:"latencyMs"`
	ProviderRequestID      string          `json:"providerRequestId"`
	ErrorCode              string          `json:"errorCode"`
	RequestBytes           int64           `json:"requestBytes"`
	ResponseBytes          int64           `json:"responseBytes"`
	RequestSHA256          string          `json:"requestSha256"`
	ResponseSHA256         string          `json:"responseSha256"`
	RequestTruncated       bool            `json:"requestTruncated"`
	ResponseTruncated      bool            `json:"responseTruncated"`
	RequestClassification  json.RawMessage `json:"requestClassification"`
	ResponseClassification json.RawMessage `json:"responseClassification"`
	CreatedAt              time.Time       `json:"createdAt"`
	CompletedAt            *time.Time      `json:"completedAt"`
	RetentionUntil         time.Time       `json:"retentionUntil"`
	RequestCapture         capturedSample  `json:"requestCapture"`
	ResponseCapture        capturedSample  `json:"responseCapture"`
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

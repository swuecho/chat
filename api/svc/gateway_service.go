package svc

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type GatewayService struct{ q *sqlc_queries.Queries }

func NewGatewayService(q *sqlc_queries.Queries) *GatewayService { return &GatewayService{q: q} }

type GatewayCredential struct {
	ID                int64
	UserID            int32
	Status            string
	RequestsPerMinute int32
	ExpiresAt         *time.Time
}

type GatewayModel struct {
	ID            int32
	Name          string
	Url           string
	ApiAuthHeader string
	ApiAuthKey    string
	HttpTimeOut   int32
	ApiType       string
	IsEnable      bool
}

type CreateGatewayRequestInput struct {
	APIKeyID              int64
	UserID                int32
	ChatModelID           int32
	RequestedModel        string
	Provider              string
	Stream                bool
	RequestBytes          int64
	RequestSHA256         string
	RequestSample         []byte
	RequestTruncated      bool
	RequestClassification json.RawMessage
	RetentionUntil        time.Time
}

type CompleteGatewayRequestInput struct {
	ID                     int64
	Status                 string
	PromptTokens           int32
	CompletionTokens       int32
	TotalTokens            int32
	LatencyMs              int64
	ProviderRequestID      string
	ErrorCode              string
	ResponseBytes          int64
	ResponseSHA256         string
	ResponseSample         []byte
	ResponseTruncated      bool
	ResponseClassification json.RawMessage
}

func (s *GatewayService) CredentialByHash(ctx context.Context, hash string) (GatewayCredential, error) {
	key, err := s.q.VirtualAdminAPIKeyByHash(ctx, hash)
	if err != nil {
		return GatewayCredential{}, err
	}
	var expiresAt *time.Time
	if key.ExpiresAt.Valid {
		expiresAt = &key.ExpiresAt.Time
	}
	return GatewayCredential{ID: key.ID, UserID: key.UserID, Status: key.Status, RequestsPerMinute: key.RequestsPerMinute, ExpiresAt: expiresAt}, nil
}

func (s *GatewayService) CountRecentRequests(ctx context.Context, keyID int64) (int64, error) {
	return s.q.CountRecentGatewayRequests(ctx, keyID)
}

func (s *GatewayService) TouchKey(ctx context.Context, keyID int64) error {
	return s.q.TouchVirtualAPIKey(ctx, keyID)
}

func (s *GatewayService) ListModelNames(ctx context.Context) ([]string, error) {
	models, err := s.q.ListGatewayModels(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(models))
	for i, model := range models {
		names[i] = model.Name
	}
	return names, nil
}

func (s *GatewayService) ModelByName(ctx context.Context, name string) (GatewayModel, error) {
	model, err := s.q.ChatModelByName(ctx, name)
	if err != nil {
		return GatewayModel{}, err
	}
	return GatewayModel{ID: model.ID, Name: model.Name, Url: model.Url, ApiAuthHeader: model.ApiAuthHeader, ApiAuthKey: model.ApiAuthKey, HttpTimeOut: model.HttpTimeOut, ApiType: model.ApiType, IsEnable: model.IsEnable}, nil
}

func (s *GatewayService) CreateRequest(ctx context.Context, input CreateGatewayRequestInput) (int64, error) {
	record, err := s.q.CreateGatewayRequest(ctx, sqlc_queries.CreateGatewayRequestParams{
		RequestUuid: uuid.New(), ApiKeyID: input.APIKeyID, UserID: input.UserID,
		ChatModelID: sql.NullInt32{Int32: input.ChatModelID, Valid: true}, RequestedModel: input.RequestedModel,
		Provider: input.Provider, Stream: input.Stream, RequestBytes: input.RequestBytes,
		RequestSha256: input.RequestSHA256, RequestSample: input.RequestSample, RequestTruncated: input.RequestTruncated,
		RequestClassification: input.RequestClassification, RetentionUntil: input.RetentionUntil,
	})
	return record.ID, err
}

func (s *GatewayService) CompleteRequest(ctx context.Context, input CompleteGatewayRequestInput) error {
	return s.q.CompleteGatewayRequest(ctx, sqlc_queries.CompleteGatewayRequestParams{
		ID: input.ID, Status: input.Status, PromptTokens: input.PromptTokens, CompletionTokens: input.CompletionTokens,
		TotalTokens: input.TotalTokens, LatencyMs: input.LatencyMs, ProviderRequestID: input.ProviderRequestID,
		ErrorCode: input.ErrorCode, ResponseBytes: input.ResponseBytes, ResponseSha256: input.ResponseSHA256,
		ResponseSample: input.ResponseSample, ResponseTruncated: input.ResponseTruncated,
		ResponseClassification: input.ResponseClassification,
	})
}

func (s *GatewayService) PurgeExpiredSamples(ctx context.Context) error {
	_, err := s.q.PurgeExpiredGatewaySamples(ctx)
	return err
}

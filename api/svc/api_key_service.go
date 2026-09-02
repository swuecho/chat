package svc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type APIKeyService struct{ q *sqlc_queries.Queries }

func NewAPIKeyService(q *sqlc_queries.Queries) *APIKeyService { return &APIKeyService{q: q} }

type APIKeyView struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	KeyPrefix         string     `json:"keyPrefix"`
	Status            string     `json:"status"`
	RequestsPerMinute int32      `json:"requestsPerMinute"`
	ExpiresAt         *time.Time `json:"expiresAt" jsonschema:"nullable"`
	LastUsedAt        *time.Time `json:"lastUsedAt" jsonschema:"nullable"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type CreatedAPIKey struct {
	APIKeyView
	Key string `json:"key"`
}

type APIKeyUsage struct {
	RequestedModel                                            string `json:"requestedModel"`
	RequestCount, PromptTokens, CompletionTokens, TotalTokens int64
	LastUsedAt                                                any `json:"lastUsedAt"`
}

type GatewayRequestView struct {
	ID                                                          int64
	RequestUUID                                                 string
	RequestedModel, Provider, Status                            string
	Stream                                                      bool
	PromptTokens, CompletionTokens, TotalTokens                 int32
	LatencyMs, RequestBytes, ResponseBytes                      int64
	ProviderRequestID, ErrorCode, RequestSHA256, ResponseSHA256 string
	RequestSample, ResponseSample                               []byte
	RequestTruncated, ResponseTruncated                         bool
	RequestClassification, ResponseClassification               json.RawMessage
	CreatedAt                                                   time.Time
	CompletedAt                                                 *time.Time
	RetentionUntil                                              time.Time
}

func apiKeyView(key sqlc_queries.VirtualApiKey) APIKeyView {
	var expiresAt, lastUsedAt *time.Time
	if key.ExpiresAt.Valid {
		expiresAt = &key.ExpiresAt.Time
	}
	if key.LastUsedAt.Valid {
		lastUsedAt = &key.LastUsedAt.Time
	}
	return APIKeyView{key.ID, key.Name, key.KeyPrefix, key.Status, key.RequestsPerMinute, expiresAt, lastUsedAt, key.CreatedAt}
}

func (s *APIKeyService) List(ctx context.Context, userID int32) ([]APIKeyView, error) {
	keys, err := s.q.ListVirtualAPIKeysByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	views := make([]APIKeyView, len(keys))
	for i, key := range keys {
		views[i] = apiKeyView(key)
	}
	return views, nil
}

func (s *APIKeyService) Create(ctx context.Context, userID int32, name string, expiresAt *time.Time, requestsPerMinute int32) (CreatedAPIKey, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return CreatedAPIKey{}, err
	}
	plaintext := "sk-chat-" + base64.RawURLEncoding.EncodeToString(secret)
	sum := sha256.Sum256([]byte(plaintext))
	prefix := plaintext
	if len(prefix) > 18 {
		prefix = prefix[:18]
	}
	expires := sql.NullTime{}
	if expiresAt != nil {
		expires = sql.NullTime{Time: *expiresAt, Valid: true}
	}
	key, err := s.q.CreateVirtualAPIKey(ctx, sqlc_queries.CreateVirtualAPIKeyParams{
		UserID: userID, Name: name, KeyPrefix: prefix, KeyHash: hex.EncodeToString(sum[:]), ExpiresAt: expires, RequestsPerMinute: requestsPerMinute,
	})
	if err != nil {
		return CreatedAPIKey{}, err
	}
	return CreatedAPIKey{APIKeyView: apiKeyView(key), Key: plaintext}, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, id int64, userID int32) (int64, error) {
	return s.q.RevokeVirtualAPIKey(ctx, sqlc_queries.RevokeVirtualAPIKeyParams{ID: id, UserID: userID})
}

func (s *APIKeyService) Usage(ctx context.Context, id int64, userID int32) ([]APIKeyUsage, error) {
	if _, err := s.q.VirtualAPIKeyByIDAndUser(ctx, sqlc_queries.VirtualAPIKeyByIDAndUserParams{ID: id, UserID: userID}); err != nil {
		return nil, err
	}
	rows, err := s.q.GatewayUsageByKey(ctx, sqlc_queries.GatewayUsageByKeyParams{ApiKeyID: id, UserID: userID})
	if err != nil {
		return nil, err
	}
	result := make([]APIKeyUsage, len(rows))
	for i, row := range rows {
		result[i] = APIKeyUsage{RequestedModel: row.RequestedModel, RequestCount: row.RequestCount, PromptTokens: row.PromptTokens, CompletionTokens: row.CompletionTokens, TotalTokens: row.TotalTokens, LastUsedAt: row.LastUsedAt}
	}
	return result, nil
}

type APIKeyRequestsQuery struct {
	KeyID  int64
	UserID int32
	Limit  int32
}

func (s *APIKeyService) Requests(ctx context.Context, query APIKeyRequestsQuery) ([]GatewayRequestView, error) {
	if _, err := s.q.VirtualAPIKeyByIDAndUser(ctx, sqlc_queries.VirtualAPIKeyByIDAndUserParams{ID: query.KeyID, UserID: query.UserID}); err != nil {
		return nil, err
	}
	rows, err := s.q.ListGatewayRequestsByKey(ctx, sqlc_queries.ListGatewayRequestsByKeyParams{ApiKeyID: query.KeyID, UserID: query.UserID, Limit: query.Limit})
	if err != nil {
		return nil, err
	}
	result := make([]GatewayRequestView, len(rows))
	for i, row := range rows {
		var completed *time.Time
		if row.CompletedAt.Valid {
			completed = &row.CompletedAt.Time
		}
		result[i] = GatewayRequestView{ID: row.ID, RequestUUID: row.RequestUuid.String(), RequestedModel: row.RequestedModel, Provider: row.Provider, Status: row.Status, Stream: row.Stream, PromptTokens: row.PromptTokens, CompletionTokens: row.CompletionTokens, TotalTokens: row.TotalTokens, LatencyMs: row.LatencyMs, RequestBytes: row.RequestBytes, ResponseBytes: row.ResponseBytes, RequestTruncated: row.RequestTruncated, ResponseTruncated: row.ResponseTruncated, CreatedAt: row.CreatedAt, CompletedAt: completed, RetentionUntil: row.RetentionUntil, ErrorCode: row.ErrorCode}
	}
	return result, nil
}

func (s *APIKeyService) RequestDetail(ctx context.Context, requestID, keyID int64, userID int32) (GatewayRequestView, error) {
	r, err := s.q.GatewayRequestByIDAndUser(ctx, sqlc_queries.GatewayRequestByIDAndUserParams{ID: requestID, ApiKeyID: keyID, UserID: userID})
	var completed *time.Time
	if r.CompletedAt.Valid {
		completed = &r.CompletedAt.Time
	}
	return GatewayRequestView{ID: r.ID, RequestUUID: r.RequestUuid.String(), RequestedModel: r.RequestedModel, Provider: r.Provider, Status: r.Status, Stream: r.Stream, PromptTokens: r.PromptTokens, CompletionTokens: r.CompletionTokens, TotalTokens: r.TotalTokens, LatencyMs: r.LatencyMs, ProviderRequestID: r.ProviderRequestID, ErrorCode: r.ErrorCode, RequestBytes: r.RequestBytes, ResponseBytes: r.ResponseBytes, RequestSHA256: r.RequestSha256, ResponseSHA256: r.ResponseSha256, RequestSample: r.RequestSample, ResponseSample: r.ResponseSample, RequestTruncated: r.RequestTruncated, ResponseTruncated: r.ResponseTruncated, RequestClassification: r.RequestClassification, ResponseClassification: r.ResponseClassification, CreatedAt: r.CreatedAt, CompletedAt: completed, RetentionUntil: r.RetentionUntil}, err
}

package svc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
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
	ExpiresAt         *time.Time `json:"expiresAt"`
	LastUsedAt        *time.Time `json:"lastUsedAt"`
	CreatedAt         time.Time  `json:"createdAt"`
}

type CreatedAPIKey struct {
	APIKeyView
	Key string `json:"key"`
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

func (s *APIKeyService) Usage(ctx context.Context, id int64, userID int32) ([]sqlc_queries.GatewayUsageByKeyRow, error) {
	if _, err := s.q.VirtualAPIKeyByIDAndUser(ctx, sqlc_queries.VirtualAPIKeyByIDAndUserParams{ID: id, UserID: userID}); err != nil {
		return nil, err
	}
	return s.q.GatewayUsageByKey(ctx, sqlc_queries.GatewayUsageByKeyParams{ApiKeyID: id, UserID: userID})
}

func (s *APIKeyService) Requests(ctx context.Context, keyID int64, userID int32, limit int32) ([]sqlc_queries.ListGatewayRequestsByKeyRow, error) {
	if _, err := s.q.VirtualAPIKeyByIDAndUser(ctx, sqlc_queries.VirtualAPIKeyByIDAndUserParams{ID: keyID, UserID: userID}); err != nil {
		return nil, err
	}
	return s.q.ListGatewayRequestsByKey(ctx, sqlc_queries.ListGatewayRequestsByKeyParams{ApiKeyID: keyID, UserID: userID, Limit: limit})
}

func (s *APIKeyService) RequestDetail(ctx context.Context, requestID, keyID int64, userID int32) (sqlc_queries.GatewayRequest, error) {
	return s.q.GatewayRequestByIDAndUser(ctx, sqlc_queries.GatewayRequestByIDAndUserParams{ID: requestID, ApiKeyID: keyID, UserID: userID})
}

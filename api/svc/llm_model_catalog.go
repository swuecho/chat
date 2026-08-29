package svc

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

const modelCatalogTTL = 30 * time.Second

// llmModelCatalog keeps the small model configuration set in memory. Refresh
// replaces the complete immutable snapshot under a lock, so concurrent chat
// requests never observe a partially updated catalog.
type llmModelCatalog struct {
	q        *sqlc_queries.Queries
	mu       sync.RWMutex
	models   map[string]provider.ModelConfig
	loadedAt time.Time
}

func newLLMModelCatalog(q *sqlc_queries.Queries) *llmModelCatalog {
	return &llmModelCatalog{q: q, models: make(map[string]provider.ModelConfig)}
}

func (c *llmModelCatalog) get(ctx context.Context, name string) (provider.ModelConfig, error) {
	c.mu.RLock()
	model, found := c.models[name]
	fresh := time.Since(c.loadedAt) < modelCatalogTTL
	c.mu.RUnlock()
	if found && fresh {
		return model, nil
	}
	if err := c.refresh(ctx); err != nil {
		return provider.ModelConfig{}, err
	}
	c.mu.RLock()
	model, found = c.models[name]
	c.mu.RUnlock()
	if !found {
		return provider.ModelConfig{}, sql.ErrNoRows
	}
	return model, nil
}

func (c *llmModelCatalog) refresh(ctx context.Context) error {
	rows, err := c.q.ListChatModels(ctx)
	if err != nil {
		return err
	}
	models := make(map[string]provider.ModelConfig, len(rows))
	for _, row := range rows {
		models[row.Name] = providerModel(row)
	}
	c.mu.Lock()
	c.models, c.loadedAt = models, time.Now()
	c.mu.Unlock()
	return nil
}

func providerSession(session sqlc_queries.ChatSession) provider.Session {
	return provider.Session{UUID: session.Uuid, UserID: session.UserID, Model: session.Model,
		MaxTokens: session.MaxTokens, Temperature: session.Temperature,
		TopP: session.TopP, N: session.N, Debug: session.Debug}
}

func providerModel(model sqlc_queries.ChatModel) provider.ModelConfig {
	return provider.ModelConfig{Name: model.Name, URL: model.Url,
		APIAuthHeader: model.ApiAuthHeader, APIAuthKey: model.ApiAuthKey,
		APIType: model.ApiType, EnablePerModelRateLimit: model.EnablePerModeRatelimit}
}

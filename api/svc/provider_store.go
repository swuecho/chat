package svc

import (
	"context"

	"github.com/swuecho/chat_backend/provider"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// llmProviderStore is the persistence adapter for the provider package. Keeping
// these mappings here prevents generated sqlc records from crossing the LLM
// provider boundary.
type llmProviderStore struct{ q *sqlc_queries.Queries }

func newLLMProviderStore(q *sqlc_queries.Queries) provider.QueryStore {
	return &llmProviderStore{q: q}
}

func (s *llmProviderStore) ChatModelByName(ctx context.Context, name string) (provider.ModelConfig, error) {
	m, err := s.q.ChatModelByName(ctx, name)
	if err != nil {
		return provider.ModelConfig{}, err
	}
	return provider.ModelConfig{
		Name: m.Name, URL: m.Url, APIAuthHeader: m.ApiAuthHeader,
		APIAuthKey: m.ApiAuthKey, APIType: m.ApiType,
		EnablePerModelRateLimit: m.EnablePerModeRatelimit,
	}, nil
}

func (s *llmProviderStore) ListChatFilesWithContentBySessionUUID(ctx context.Context, sessionUUID string) ([]provider.File, error) {
	files, err := s.q.ListChatFilesWithContentBySessionUUID(ctx, sessionUUID)
	if err != nil {
		return nil, err
	}
	result := make([]provider.File, 0, len(files))
	for _, file := range files {
		result = append(result, provider.File{Name: file.Name, Data: file.Data, MIMEType: file.MimeType})
	}
	return result, nil
}

func providerSession(session sqlc_queries.ChatSession) provider.Session {
	return provider.Session{
		UUID: session.Uuid, UserID: session.UserID, Model: session.Model,
		MaxTokens: session.MaxTokens, Temperature: session.Temperature,
		TopP: session.TopP, N: session.N, Debug: session.Debug,
	}
}

func providerModel(model sqlc_queries.ChatModel) provider.ModelConfig {
	return provider.ModelConfig{
		Name: model.Name, URL: model.Url, APIAuthHeader: model.ApiAuthHeader,
		APIAuthKey: model.ApiAuthKey, APIType: model.ApiType,
		EnablePerModelRateLimit: model.EnablePerModeRatelimit,
	}
}

package service

import (
	"context"
	"time"

	"github.com/samber/lo"
	pkgerrors "github.com/swuecho/chat_backend/pkg/errors"
	"github.com/swuecho/chat_backend/repository"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type modelService struct {
	repos repository.CoreRepositoryManager
}

func NewModelService(repos repository.CoreRepositoryManager) ModelService {
	return &modelService{repos: repos}
}

func (s *modelService) GetAvailableModels(ctx context.Context, userID int32) ([]sqlc_queries.ChatModel, error) {
	models, err := s.repos.ChatModel().GetAll(ctx)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return models, nil
}

func (s *modelService) GetSystemModelsWithUsage(ctx context.Context, timePeriod string) ([]ChatModelWithUsage, error) {
	// Get system models
	models, err := s.repos.ChatModel().GetSystemModels(ctx)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	// Get usage statistics
	usageStats, err := s.repos.ChatModel().GetModelUsageStats(ctx, timePeriod)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	// Create usage map
	usageMap := make(map[string]sqlc_queries.GetLatestUsageTimeOfModelRow)
	for _, usage := range usageStats {
		usageMap[usage.Model] = usage
	}

	// Combine models with usage statistics
	modelsWithUsage := lo.Map(models, func(model sqlc_queries.ChatModel, _ int) ChatModelWithUsage {
		usage := usageMap[model.Name]
		lastUsageTime := ""
		if !usage.LatestMessageTime.IsZero() {
			lastUsageTime = usage.LatestMessageTime.Format(time.RFC3339)
		}
		
		return ChatModelWithUsage{
			ChatModel:     model,
			LastUsageTime: lastUsageTime,
			MessageCount:  usage.MessageCount,
		}
	})

	return modelsWithUsage, nil
}

func (s *modelService) GetModelByID(ctx context.Context, modelID int32) (*sqlc_queries.ChatModel, error) {
	model, err := s.repos.ChatModel().GetByID(ctx, modelID)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &model, nil
}

func (s *modelService) GetModelByName(ctx context.Context, name string) (*sqlc_queries.ChatModel, error) {
	model, err := s.repos.ChatModel().GetByName(ctx, name)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &model, nil
}

func (s *modelService) GetDefaultModel(ctx context.Context) (*sqlc_queries.ChatModel, error) {
	model, err := s.repos.ChatModel().GetDefaultModel(ctx)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}
	return &model, nil
}

func (s *modelService) CreateModel(ctx context.Context, params ChatModelCreateRequest) (*sqlc_queries.ChatModel, error) {
	// Validate input
	if params.Name == "" {
		return nil, pkgerrors.ValidationFailed("name", "Model name is required")
	}

	// Create SQLC parameters
	sqlcParams := sqlc_queries.CreateChatModelParams{
		Name:                   params.Name,
		Label:                  params.Label,
		IsDefault:              params.IsDefault,
		Url:                    params.Url,
		ApiAuthHeader:          params.ApiAuthHeader,
		ApiAuthKey:             params.ApiAuthKey,
		UserID:                 params.UserID,
		EnablePerModeRatelimit: params.EnablePerModeRatelimit,
		MaxToken:               params.MaxToken,
		DefaultToken:           params.DefaultToken,
	}

	model, err := s.repos.ChatModel().Create(ctx, sqlcParams)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	return &model, nil
}

func (s *modelService) UpdateModel(ctx context.Context, modelID int32, updates ChatModelUpdateRequest) (*sqlc_queries.ChatModel, error) {
	// First get existing model
	existing, err := s.repos.ChatModel().GetByID(ctx, modelID)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	// Build update parameters with current values as defaults
	params := sqlc_queries.UpdateChatModelParams{
		ID:                     modelID,
		Name:                   existing.Name,
		Label:                  existing.Label,
		IsDefault:              existing.IsDefault,
		Url:                    existing.Url,
		ApiAuthHeader:          existing.ApiAuthHeader,
		ApiAuthKey:             existing.ApiAuthKey,
		UserID:                 existing.UserID,
		EnablePerModeRatelimit: existing.EnablePerModeRatelimit,
		MaxToken:               existing.MaxToken,
	}

	// Apply updates if provided
	if updates.Name != nil {
		params.Name = *updates.Name
	}
	if updates.Label != nil {
		params.Label = *updates.Label
	}
	if updates.IsDefault != nil {
		params.IsDefault = *updates.IsDefault
	}
	if updates.Url != nil {
		params.Url = *updates.Url
	}
	if updates.ApiAuthHeader != nil {
		params.ApiAuthHeader = *updates.ApiAuthHeader
	}
	if updates.ApiAuthKey != nil {
		params.ApiAuthKey = *updates.ApiAuthKey
	}
	if updates.EnablePerModeRatelimit != nil {
		params.EnablePerModeRatelimit = *updates.EnablePerModeRatelimit
	}
	if updates.MaxToken != nil {
		params.MaxToken = *updates.MaxToken
	}

	model, err := s.repos.ChatModel().Update(ctx, params)
	if err != nil {
		return nil, pkgerrors.FromDatabaseError(err)
	}

	return &model, nil
}

func (s *modelService) DeleteModel(ctx context.Context, modelID int32, userID int32) error {
	// Verify model exists first
	_, err := s.repos.ChatModel().GetByID(ctx, modelID)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}

	// TODO: Add authorization check to ensure user can delete this model
	// For now, we'll allow deletion but this should be enhanced with proper auth

	err = s.repos.ChatModel().Delete(ctx, modelID)
	if err != nil {
		return pkgerrors.FromDatabaseError(err)
	}

	return nil
}

func (s *modelService) CreateModelInstance(ctx context.Context, modelName string) (ChatModel, error) {
	// TODO: Implement model factory pattern here
	// This would create appropriate model instances (OpenAI, Claude, etc.)
	return nil, pkgerrors.ErrInternalServer.WithDetail("Model factory pattern not yet implemented")
}
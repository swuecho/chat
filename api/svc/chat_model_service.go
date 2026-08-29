package svc

import (
	"context"
	"time"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelService struct{ q *sqlc_queries.Queries }

func NewChatModelService(q *sqlc_queries.Queries) *ChatModelService { return &ChatModelService{q: q} }

type CreateChatModelInput struct {
	Name                   string `json:"name"`
	Label                  string `json:"label"`
	IsDefault              bool   `json:"isDefault"`
	Url                    string `json:"url"`
	ApiAuthHeader          string `json:"apiAuthHeader"`
	ApiAuthKey             string `json:"apiAuthKey"`
	UserID                 int32
	EnablePerModeRatelimit bool   `json:"enablePerModeRatelimit"`
	MaxToken               int32  `json:"maxToken"`
	DefaultToken           int32  `json:"defaultToken"`
	OrderNumber            int32  `json:"orderNumber"`
	HttpTimeOut            int32  `json:"httpTimeOut"`
	ApiType                string `json:"apiType"`
}

type UpdateChatModelInput struct {
	ID                     int32
	Name                   string `json:"name"`
	Label                  string `json:"label"`
	IsDefault              bool   `json:"isDefault"`
	Url                    string `json:"url"`
	ApiAuthHeader          string `json:"apiAuthHeader"`
	ApiAuthKey             string `json:"apiAuthKey"`
	UserID                 int32
	EnablePerModeRatelimit bool   `json:"enablePerModeRatelimit"`
	MaxToken               int32  `json:"maxToken"`
	DefaultToken           int32  `json:"defaultToken"`
	OrderNumber            int32  `json:"orderNumber"`
	HttpTimeOut            int32  `json:"httpTimeOut"`
	IsEnable               bool   `json:"isEnable"`
	ApiType                string `json:"apiType"`
}

type ChatModelWithUsage struct {
	sqlc_queries.ChatModel
	LastUsageTime time.Time `json:"lastUsageTime,omitempty"`
	MessageCount  int64     `json:"messageCount"`
}

func (s *ChatModelService) ListSystemWithUsage(ctx context.Context, interval string) ([]ChatModelWithUsage, error) {
	models, err := s.q.ListSystemChatModels(ctx)
	if err != nil {
		return nil, err
	}
	usageRows, err := s.q.GetLatestUsageTimeOfModel(ctx, interval)
	if err != nil {
		return nil, err
	}
	usageByModel := make(map[string]sqlc_queries.GetLatestUsageTimeOfModelRow, len(usageRows))
	for _, usage := range usageRows {
		usageByModel[usage.Model] = usage
	}
	result := make([]ChatModelWithUsage, len(models))
	for i, model := range models {
		usage := usageByModel[model.Name]
		result[i] = ChatModelWithUsage{ChatModel: model, LastUsageTime: usage.LatestMessageTime, MessageCount: usage.MessageCount}
	}
	return result, nil
}

func (s *ChatModelService) ByID(ctx context.Context, id int32) (sqlc_queries.ChatModel, error) {
	return s.q.ChatModelByID(ctx, id)
}

func (s *ChatModelService) Create(ctx context.Context, input CreateChatModelInput) (sqlc_queries.ChatModel, error) {
	return s.q.CreateChatModel(ctx, sqlc_queries.CreateChatModelParams(input))
}

func (s *ChatModelService) Update(ctx context.Context, input UpdateChatModelInput) (sqlc_queries.ChatModel, error) {
	return s.q.UpdateChatModel(ctx, sqlc_queries.UpdateChatModelParams(input))
}

func (s *ChatModelService) Delete(ctx context.Context, id, userID int32) error {
	return s.q.DeleteChatModel(ctx, sqlc_queries.DeleteChatModelParams{ID: id, UserID: userID})
}

func (s *ChatModelService) Default(ctx context.Context) (sqlc_queries.ChatModel, error) {
	return s.q.GetDefaultChatModel(ctx)
}

func (s *ChatModelService) TitleModel(ctx context.Context) (sqlc_queries.ChatModel, error) {
	return s.q.GetTitleChatModel(ctx)
}

func (s *ChatModelService) SetTitleModel(ctx context.Context, modelID, userID int32) (sqlc_queries.ChatModel, error) {
	return s.q.SetTitleChatModel(ctx, sqlc_queries.SetTitleChatModelParams{ModelID: modelID, UserID: userID})
}

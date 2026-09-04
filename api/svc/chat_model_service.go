package svc

import (
	"context"
	"database/sql"
	"time"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatModelService struct{ q *sqlc_queries.Queries }

func NewChatModelService(q *sqlc_queries.Queries) *ChatModelService { return &ChatModelService{q: q} }

type CreateChatModelInput struct {
	Name                   string
	Label                  string
	IsDefault              bool
	Url                    string
	ApiAuthHeader          string
	ApiAuthKey             string
	UserID                 int32
	EnablePerModeRatelimit bool
	OrderNumber            int32
	HttpTimeOut            int32
	ApiType                string
}

type UpdateChatModelInput struct {
	ID                     int32
	Name                   string
	Label                  string
	IsDefault              bool
	Url                    string
	ApiAuthHeader          string
	ApiAuthKey             string
	UserID                 int32
	EnablePerModeRatelimit bool
	MaxToken               *int32
	DefaultToken           *int32
	OrderNumber            int32
	HttpTimeOut            int32
	IsEnable               bool
	ApiType                string
}

type ChatModel struct {
	ID                                                         int32
	Name, Label, URL, APIAuthHeader, APIAuthKey, APIType       string
	IsDefault, IsEnable, IsTitleModel, EnablePerModelRateLimit bool
	UserID, OrderNumber, HTTPTimeout                           int32
	MaxToken, DefaultToken                                     *int32
}

func chatModelFromRecord(m sqlc_queries.ChatModel) ChatModel {
	return ChatModel{ID: m.ID, Name: m.Name, Label: m.Label, URL: m.Url, APIAuthHeader: m.ApiAuthHeader,
		APIAuthKey: m.ApiAuthKey, APIType: m.ApiType, IsDefault: m.IsDefault, IsEnable: m.IsEnable,
		IsTitleModel: m.IsTitleModel, EnablePerModelRateLimit: m.EnablePerModeRatelimit, UserID: m.UserID,
		MaxToken: nullableInt32(m.MaxToken), DefaultToken: nullableInt32(m.DefaultToken), OrderNumber: m.OrderNumber, HTTPTimeout: m.HttpTimeOut}
}

func nullableInt32(value sql.NullInt32) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}

type ChatModelWithUsage struct {
	ChatModel
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
		result[i] = ChatModelWithUsage{ChatModel: chatModelFromRecord(model), LastUsageTime: usage.LatestMessageTime, MessageCount: usage.MessageCount}
	}
	return result, nil
}

func (s *ChatModelService) ByID(ctx context.Context, id int32) (ChatModel, error) {
	m, err := s.q.ChatModelByID(ctx, id)
	return chatModelFromRecord(m), err
}

func (s *ChatModelService) Create(ctx context.Context, input CreateChatModelInput) (ChatModel, error) {
	m, err := s.q.CreateChatModel(ctx, sqlc_queries.CreateChatModelParams(input))
	return chatModelFromRecord(m), err
}

func (s *ChatModelService) Update(ctx context.Context, input UpdateChatModelInput) (ChatModel, error) {
	params := sqlc_queries.UpdateChatModelParams{
		ID: input.ID, Name: input.Name, Label: input.Label, IsDefault: input.IsDefault, Url: input.Url,
		ApiAuthHeader: input.ApiAuthHeader, ApiAuthKey: input.ApiAuthKey, UserID: input.UserID,
		EnablePerModeRatelimit: input.EnablePerModeRatelimit, OrderNumber: input.OrderNumber,
		HttpTimeOut: input.HttpTimeOut, IsEnable: input.IsEnable, ApiType: input.ApiType,
	}
	if input.MaxToken != nil {
		params.MaxToken = sql.NullInt32{Int32: *input.MaxToken, Valid: true}
	}
	if input.DefaultToken != nil {
		params.DefaultToken = sql.NullInt32{Int32: *input.DefaultToken, Valid: true}
	}
	m, err := s.q.UpdateChatModel(ctx, params)
	return chatModelFromRecord(m), err
}

func (s *ChatModelService) Delete(ctx context.Context, id, userID int32) error {
	return s.q.DeleteChatModel(ctx, sqlc_queries.DeleteChatModelParams{ID: id, UserID: userID})
}

func (s *ChatModelService) Default(ctx context.Context) (ChatModel, error) {
	m, err := s.q.GetDefaultChatModel(ctx)
	return chatModelFromRecord(m), err
}

func (s *ChatModelService) TitleModel(ctx context.Context) (ChatModel, error) {
	m, err := s.q.GetTitleChatModel(ctx)
	return chatModelFromRecord(m), err
}

func (s *ChatModelService) SetTitleModel(ctx context.Context, modelID, userID int32) (ChatModel, error) {
	m, err := s.q.SetTitleChatModel(ctx, sqlc_queries.SetTitleChatModelParams{ModelID: modelID, UserID: userID})
	return chatModelFromRecord(m), err
}

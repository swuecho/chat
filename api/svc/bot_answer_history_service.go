package svc

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type BotAnswerHistoryService struct {
	q *sqlc_queries.Queries
}

type CreateBotAnswerHistoryInput struct {
	BotUUID    string
	UserID     int32
	Prompt     string
	Answer     string
	Model      string
	TokensUsed int32
}

type BotAnswerHistoryPageQuery struct {
	BotUUID string
	Page    PageWindow
}

type UserAnswerHistoryPageQuery struct {
	UserID int32
	Page   PageWindow
}
type LatestBotAnswerHistoryQuery struct {
	BotUUID string
	Limit   int32
}

func createBotAnswerHistoryParams(input CreateBotAnswerHistoryInput) sqlc_queries.CreateBotAnswerHistoryParams {
	return sqlc_queries.CreateBotAnswerHistoryParams{BotUuid: input.BotUUID, UserID: input.UserID,
		Prompt: input.Prompt, Answer: input.Answer, Model: input.Model, TokensUsed: input.TokensUsed}
}

// NewBotAnswerHistoryService creates a new BotAnswerHistoryService
func NewBotAnswerHistoryService(q *sqlc_queries.Queries) *BotAnswerHistoryService {
	return &BotAnswerHistoryService{q: q}
}

// CreateBotAnswerHistory creates a new bot answer history entry
func (s *BotAnswerHistoryService) CreateBotAnswerHistory(ctx context.Context, input CreateBotAnswerHistoryInput) (sqlc_queries.BotAnswerHistory, error) {
	history, err := s.q.CreateBotAnswerHistory(ctx, createBotAnswerHistoryParams(input))
	if err != nil {
		return sqlc_queries.BotAnswerHistory{}, eris.Wrap(err, "failed to create bot answer history")
	}
	return history, nil
}

// GetBotAnswerHistoryByID gets a bot answer history entry by ID
func (s *BotAnswerHistoryService) GetBotAnswerHistoryByID(ctx context.Context, id int32) (sqlc_queries.GetBotAnswerHistoryByIDRow, error) {
	history, err := s.q.GetBotAnswerHistoryByID(ctx, id)
	if err != nil {
		return sqlc_queries.GetBotAnswerHistoryByIDRow{}, eris.Wrap(err, "failed to get bot answer history by ID")
	}
	return history, nil
}

// GetBotAnswerHistoryByBotUUID gets paginated bot answer history for a specific bot
func (s *BotAnswerHistoryService) GetBotAnswerHistoryByBotUUID(ctx context.Context, query BotAnswerHistoryPageQuery) ([]sqlc_queries.GetBotAnswerHistoryByBotUUIDRow, error) {
	params := sqlc_queries.GetBotAnswerHistoryByBotUUIDParams{
		BotUuid: query.BotUUID,
		Limit:   query.Page.Limit,
		Offset:  query.Page.Offset,
	}
	history, err := s.q.GetBotAnswerHistoryByBotUUID(ctx, params)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get bot answer history by bot UUID")
	}
	return history, nil
}

// GetBotAnswerHistoryByUserID gets paginated bot answer history for a specific user
func (s *BotAnswerHistoryService) GetBotAnswerHistoryByUserID(ctx context.Context, query UserAnswerHistoryPageQuery) ([]sqlc_queries.GetBotAnswerHistoryByUserIDRow, error) {
	params := sqlc_queries.GetBotAnswerHistoryByUserIDParams{
		UserID: query.UserID,
		Limit:  query.Page.Limit,
		Offset: query.Page.Offset,
	}
	history, err := s.q.GetBotAnswerHistoryByUserID(ctx, params)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get bot answer history by user ID")
	}
	return history, nil
}

// UpdateBotAnswerHistory updates an existing bot answer history entry
func (s *BotAnswerHistoryService) UpdateBotAnswerHistory(ctx context.Context, id int32, answer string, tokensUsed int32) (sqlc_queries.BotAnswerHistory, error) {
	params := sqlc_queries.UpdateBotAnswerHistoryParams{
		ID:         id,
		Answer:     answer,
		TokensUsed: tokensUsed,
	}
	history, err := s.q.UpdateBotAnswerHistory(ctx, params)
	if err != nil {
		return sqlc_queries.BotAnswerHistory{}, eris.Wrap(err, "failed to update bot answer history")
	}
	return history, nil
}

// DeleteBotAnswerHistory deletes a bot answer history entry by ID
func (s *BotAnswerHistoryService) DeleteBotAnswerHistory(ctx context.Context, id int32) error {
	err := s.q.DeleteBotAnswerHistory(ctx, id)
	if err != nil {
		return eris.Wrap(err, "failed to delete bot answer history")
	}
	return nil
}

// GetBotAnswerHistoryCountByBotUUID gets the count of history entries for a bot
func (s *BotAnswerHistoryService) GetBotAnswerHistoryCountByBotUUID(ctx context.Context, botUUID string) (int64, error) {
	count, err := s.q.GetBotAnswerHistoryCountByBotUUID(ctx, botUUID)
	if err != nil {
		return 0, eris.Wrap(err, "failed to get bot answer history count by bot UUID")
	}
	return count, nil
}

// GetBotAnswerHistoryCountByUserID gets the count of history entries for a user
func (s *BotAnswerHistoryService) GetBotAnswerHistoryCountByUserID(ctx context.Context, userID int32) (int64, error) {
	count, err := s.q.GetBotAnswerHistoryCountByUserID(ctx, userID)
	if err != nil {
		return 0, eris.Wrap(err, "failed to get bot answer history count by user ID")
	}
	return count, nil
}

// GetLatestBotAnswerHistoryByBotUUID gets the latest history entries for a bot
func (s *BotAnswerHistoryService) GetLatestBotAnswerHistoryByBotUUID(ctx context.Context, query LatestBotAnswerHistoryQuery) ([]sqlc_queries.GetLatestBotAnswerHistoryByBotUUIDRow, error) {
	params := sqlc_queries.GetLatestBotAnswerHistoryByBotUUIDParams{
		BotUuid: query.BotUUID,
		Limit:   query.Limit,
	}
	history, err := s.q.GetLatestBotAnswerHistoryByBotUUID(ctx, params)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get latest bot answer history by bot UUID")
	}
	return history, nil
}

package svc

import (
	"fmt"
	"context"

	"github.com/swuecho/chat_backend/sqlc_queries"
)

type BotAnswerHistoryService struct {
	q *sqlc_queries.Queries
}

// NewBotAnswerHistoryService creates a new BotAnswerHistoryService
func NewBotAnswerHistoryService(q *sqlc_queries.Queries) *BotAnswerHistoryService {
	return &BotAnswerHistoryService{q: q}
}

// Q returns the underlying queries.
func (s *BotAnswerHistoryService) Q() *sqlc_queries.Queries { return s.q }

// CreateBotAnswerHistory creates a new bot answer history entry
func (s *BotAnswerHistoryService) CreateBotAnswerHistory(ctx context.Context, params sqlc_queries.CreateBotAnswerHistoryParams) (sqlc_queries.BotAnswerHistory, error) {
	history, err := s.q.CreateBotAnswerHistory(ctx, params)
	if err != nil {
		return sqlc_queries.BotAnswerHistory{}, fmt.Errorf("failed to create bot answer history: %w", err)
	}
	return history, nil
}

// GetBotAnswerHistoryByID gets a bot answer history entry by ID
func (s *BotAnswerHistoryService) GetBotAnswerHistoryByID(ctx context.Context, id int32) (sqlc_queries.GetBotAnswerHistoryByIDRow, error) {
	history, err := s.q.GetBotAnswerHistoryByID(ctx, id)
	if err != nil {
		return sqlc_queries.GetBotAnswerHistoryByIDRow{}, fmt.Errorf("failed to get bot answer history by ID: %w", err)
	}
	return history, nil
}

// GetBotAnswerHistoryByBotUUID gets paginated bot answer history for a specific bot
func (s *BotAnswerHistoryService) GetBotAnswerHistoryByBotUUID(ctx context.Context, botUUID string, limit, offset int32) ([]sqlc_queries.GetBotAnswerHistoryByBotUUIDRow, error) {
	params := sqlc_queries.GetBotAnswerHistoryByBotUUIDParams{
		BotUuid: botUUID,
		Limit:   limit,
		Offset:  offset,
	}
	history, err := s.q.GetBotAnswerHistoryByBotUUID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot answer history by bot UUID: %w", err)
	}
	return history, nil
}

// GetBotAnswerHistoryByUserID gets paginated bot answer history for a specific user
func (s *BotAnswerHistoryService) GetBotAnswerHistoryByUserID(ctx context.Context, userID, limit, offset int32) ([]sqlc_queries.GetBotAnswerHistoryByUserIDRow, error) {
	params := sqlc_queries.GetBotAnswerHistoryByUserIDParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
	history, err := s.q.GetBotAnswerHistoryByUserID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot answer history by user ID: %w", err)
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
		return sqlc_queries.BotAnswerHistory{}, fmt.Errorf("failed to update bot answer history: %w", err)
	}
	return history, nil
}

// DeleteBotAnswerHistory deletes a bot answer history entry by ID
func (s *BotAnswerHistoryService) DeleteBotAnswerHistory(ctx context.Context, id int32) error {
	err := s.q.DeleteBotAnswerHistory(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete bot answer history: %w", err)
	}
	return nil
}

// GetBotAnswerHistoryCountByBotUUID gets the count of history entries for a bot
func (s *BotAnswerHistoryService) GetBotAnswerHistoryCountByBotUUID(ctx context.Context, botUUID string) (int64, error) {
	count, err := s.q.GetBotAnswerHistoryCountByBotUUID(ctx, botUUID)
	if err != nil {
		return 0, fmt.Errorf("failed to get bot answer history count by bot UUID: %w", err)
	}
	return count, nil
}

// GetBotAnswerHistoryCountByUserID gets the count of history entries for a user
func (s *BotAnswerHistoryService) GetBotAnswerHistoryCountByUserID(ctx context.Context, userID int32) (int64, error) {
	count, err := s.q.GetBotAnswerHistoryCountByUserID(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get bot answer history count by user ID: %w", err)
	}
	return count, nil
}

// GetLatestBotAnswerHistoryByBotUUID gets the latest history entries for a bot
func (s *BotAnswerHistoryService) GetLatestBotAnswerHistoryByBotUUID(ctx context.Context, botUUID string, limit int32) ([]sqlc_queries.GetLatestBotAnswerHistoryByBotUUIDRow, error) {
	params := sqlc_queries.GetLatestBotAnswerHistoryByBotUUIDParams{
		BotUuid: botUUID,
		Limit:   limit,
	}
	history, err := s.q.GetLatestBotAnswerHistoryByBotUUID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest bot answer history by bot UUID: %w", err)
	}
	return history, nil
}

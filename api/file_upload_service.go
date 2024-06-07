package main

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatFileService struct {
	q *sqlc_queries.Queries
}

// ChatFileService creates a new ChatMessageService.
func NewChatFileService(q *sqlc_queries.Queries) *ChatFileService {
	return &ChatFileService{q: q}
}

func (s *ChatFileService) CreateChatUpload(ctx context.Context, params sqlc_queries.CreateChatFileParams) (sqlc_queries.ChatFile, error) {
	upload, err := s.q.CreateChatFile(ctx, params)
	if err != nil {
		return sqlc_queries.ChatFile{}, eris.Wrap(err, "failed to create chat file ")
	}
	return upload, nil
}

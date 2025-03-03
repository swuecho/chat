package main

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatCommentService struct {
	q *sqlc_queries.Queries
}

func NewChatCommentService(q *sqlc_queries.Queries) *ChatCommentService {
	return &ChatCommentService{q: q}
}

// CreateChatComment creates a new chat comment
func (s *ChatCommentService) CreateChatComment(ctx context.Context, params sqlc_queries.CreateChatCommentParams) (sqlc_queries.ChatComment, error) {
	comment, err := s.q.CreateChatComment(ctx, params)
	if err != nil {
		return sqlc_queries.ChatComment{}, eris.Wrap(err, "failed to create comment")
	}
	return comment, nil
}

// GetCommentsBySessionUUID returns comments for a session with author info
func (s *ChatCommentService) GetCommentsBySessionUUID(ctx context.Context, sessionUUID string) ([]sqlc_queries.GetCommentsBySessionUUIDRow, error) {
	comments, err := s.q.GetCommentsBySessionUUID(ctx, sessionUUID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get comments by session UUID")
	}
	return comments, nil
}

// GetCommentsByMessageUUID returns comments for a message with author info
func (s *ChatCommentService) GetCommentsByMessageUUID(ctx context.Context, messageUUID string) ([]sqlc_queries.GetCommentsByMessageUUIDRow, error) {
	comments, err := s.q.GetCommentsByMessageUUID(ctx, messageUUID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get comments by message UUID")
	}
	return comments, nil
}

// CommentWithAuthor represents a comment with author information
type CommentWithAuthor struct {
	UUID           string    `json:"uuid"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
	AuthorUsername string    `json:"authorUsername"`
	AuthorEmail    string    `json:"authorEmail"`
}

// GetCommentsBySession returns comments for a session with author info
func (s *ChatCommentService) GetCommentsBySession(ctx context.Context, sessionUUID string) ([]CommentWithAuthor, error) {
	comments, err := s.q.GetCommentsBySessionUUID(ctx, sessionUUID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get comments by session")
	}

	result := make([]CommentWithAuthor, len(comments))
	for i, c := range comments {
		result[i] = CommentWithAuthor{
			UUID:           c.Uuid,
			Content:        c.Content,
			CreatedAt:      c.CreatedAt,
			AuthorUsername: c.AuthorUsername,
			AuthorEmail:    c.AuthorEmail,
		}
	}
	return result, nil
}

// GetCommentsByMessage returns comments for a message with author info
func (s *ChatCommentService) GetCommentsByMessage(ctx context.Context, messageUUID string) ([]CommentWithAuthor, error) {
	comments, err := s.q.GetCommentsByMessageUUID(ctx, messageUUID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get comments by message")
	}

	result := make([]CommentWithAuthor, len(comments))
	for i, c := range comments {
		result[i] = CommentWithAuthor{
			UUID:           c.Uuid,
			Content:        c.Content,
			CreatedAt:      c.CreatedAt,
			AuthorUsername: c.AuthorUsername,
			AuthorEmail:    c.AuthorEmail,
		}
	}
	return result, nil
}

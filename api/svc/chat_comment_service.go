package svc

import (
	"context"
	"time"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type ChatCommentService struct {
	q *sqlc_queries.Queries
}

type ChatComment sqlc_queries.ChatComment
type SessionComment sqlc_queries.GetCommentsBySessionUUIDRow
type MessageComment sqlc_queries.GetCommentsByMessageUUIDRow

type CreateChatCommentInput struct {
	Uuid            string
	ChatSessionUuid string
	ChatMessageUuid string
	Content         string
	CreatedBy       int32
}

func NewChatCommentService(q *sqlc_queries.Queries) *ChatCommentService {
	return &ChatCommentService{q: q}
}

// CreateChatComment creates a new chat comment
func (s *ChatCommentService) CreateChatComment(ctx context.Context, input CreateChatCommentInput) (ChatComment, error) {
	comment, err := s.q.CreateChatComment(ctx, sqlc_queries.CreateChatCommentParams(input))
	if err != nil {
		return ChatComment{}, eris.Wrap(err, "failed to create comment")
	}
	return ChatComment(comment), nil
}

// GetCommentsBySessionUUID returns comments for a session with author info
func (s *ChatCommentService) GetCommentsBySessionUUID(ctx context.Context, sessionUUID string) ([]SessionComment, error) {
	comments, err := s.q.GetCommentsBySessionUUID(ctx, sessionUUID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get comments by session UUID")
	}
	result := make([]SessionComment, len(comments))
	for i, comment := range comments {
		result[i] = SessionComment(comment)
	}
	return result, nil
}

// GetCommentsByMessageUUID returns comments for a message with author info
func (s *ChatCommentService) GetCommentsByMessageUUID(ctx context.Context, messageUUID string) ([]MessageComment, error) {
	comments, err := s.q.GetCommentsByMessageUUID(ctx, messageUUID)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get comments by message UUID")
	}
	result := make([]MessageComment, len(comments))
	for i, comment := range comments {
		result[i] = MessageComment(comment)
	}
	return result, nil
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

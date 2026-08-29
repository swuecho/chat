package svc

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/swuecho/chat_backend/domain"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// ChatFileService handles operations related to chat file uploads
type ChatFileService struct {
	q *sqlc_queries.Queries
}

// NewChatFileService creates a new ChatFileService instance
func NewChatFileService(q *sqlc_queries.Queries) *ChatFileService {
	return &ChatFileService{q: q}
}

// CreateChatFileInput is the application input accepted from transports.
// SQLC parameter construction remains private to the service.
type CreateChatFileInput struct {
	Name            string
	Data            []byte
	UserID          int32
	ChatSessionUuid string
	MimeType        string
}

// CreateChatUpload handles creating a new chat file upload
func (s *ChatFileService) CreateChatUpload(ctx context.Context, input CreateChatFileInput) (sqlc_queries.ChatFile, error) {
	// Validate input
	if input.ChatSessionUuid == "" {
		return sqlc_queries.ChatFile{}, domain.Invalid("missing session UUID")
	}
	if input.UserID <= 0 {
		return sqlc_queries.ChatFile{}, domain.Invalid("invalid user ID")
	}
	if input.Name == "" {
		return sqlc_queries.ChatFile{}, domain.Invalid("missing file name")
	}
	if len(input.Data) == 0 {
		return sqlc_queries.ChatFile{}, domain.Invalid("empty file data")
	}

	slog.Info("Creating chat file upload", "session", input.ChatSessionUuid, "userID", input.UserID)

	upload, err := s.q.CreateChatFile(ctx, sqlc_queries.CreateChatFileParams(input))
	if err != nil {
		return sqlc_queries.ChatFile{}, domain.Internal("failed to create chat file", err)
	}

	slog.Info("Created chat file upload", "id", upload.ID)
	return upload, nil
}

// GetChatFile retrieves a chat file by ID
func (s *ChatFileService) GetChatFile(ctx context.Context, id int32) (sqlc_queries.GetChatFileByIDRow, error) {
	if id <= 0 {
		return sqlc_queries.GetChatFileByIDRow{}, domain.Invalid("invalid file ID")
	}

	slog.Info("Retrieving chat file", "id", id)

	file, err := s.q.GetChatFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sqlc_queries.GetChatFileByIDRow{}, domain.NotFound("Chat file", err)
		}
		return sqlc_queries.GetChatFileByIDRow{}, domain.Internal("failed to get chat file", err)
	}

	return file, nil
}

// DeleteChatFile deletes a chat file by ID
func (s *ChatFileService) DeleteChatFile(ctx context.Context, id int32) error {
	if id <= 0 {
		return domain.Invalid("invalid file ID")
	}

	slog.Info("Deleting chat file", "id", id)

	_, err := s.q.DeleteChatFile(ctx, id)
	if err != nil {
		return domain.Internal("failed to delete chat file", err)
	}

	return nil
}

// ListChatFilesBySession retrieves chat files for a session
func (s *ChatFileService) ListChatFilesBySession(ctx context.Context, sessionUUID string, userID int32) ([]sqlc_queries.ListChatFilesBySessionUUIDRow, error) {
	if sessionUUID == "" {
		return nil, domain.Invalid("missing session UUID")
	}
	if userID <= 0 {
		return nil, domain.Invalid("invalid user ID")
	}

	slog.Info("Listing chat files", "session", sessionUUID, "userID", userID)

	files, err := s.q.ListChatFilesBySessionUUID(ctx, sqlc_queries.ListChatFilesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
	})
	if err != nil {
		return nil, domain.Internal("failed to list chat files", err)
	}

	return files, nil
}

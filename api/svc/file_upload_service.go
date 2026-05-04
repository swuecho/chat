package svc

import (
	"context"
	"log/slog"

	"github.com/swuecho/chat_backend/dto"
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

// Q returns the underlying queries.
func (s *ChatFileService) Q() *sqlc_queries.Queries { return s.q }

// CreateChatUpload handles creating a new chat file upload
func (s *ChatFileService) CreateChatUpload(ctx context.Context, params sqlc_queries.CreateChatFileParams) (sqlc_queries.ChatFile, error) {
	// Validate input
	if params.ChatSessionUuid == "" {
		return sqlc_queries.ChatFile{}, dto.ErrValidationInvalidInput("missing session UUID")
	}
	if params.UserID <= 0 {
		return sqlc_queries.ChatFile{}, dto.ErrValidationInvalidInput("invalid user ID")
	}
	if params.Name == "" {
		return sqlc_queries.ChatFile{}, dto.ErrValidationInvalidInput("missing file name")
	}
	if len(params.Data) == 0 {
		return sqlc_queries.ChatFile{}, dto.ErrValidationInvalidInput("empty file data")
	}

	slog.Info("Creating chat file upload", "session", params.ChatSessionUuid, "userID", params.UserID)

	upload, err := s.q.CreateChatFile(ctx, params)
	if err != nil {
		return sqlc_queries.ChatFile{}, dto.WrapError(err, "failed to create chat file")
	}

	slog.Info("Created chat file upload", "id", upload.ID)
	return upload, nil
}

// GetChatFile retrieves a chat file by ID
func (s *ChatFileService) GetChatFile(ctx context.Context, id int32) (sqlc_queries.GetChatFileByIDRow, error) {
	if id <= 0 {
		return sqlc_queries.GetChatFileByIDRow{}, dto.ErrValidationInvalidInput("invalid file ID")
	}

	slog.Info("Retrieving chat file", "id", id)

	file, err := s.q.GetChatFileByID(ctx, id)
	if err != nil {
		return sqlc_queries.GetChatFileByIDRow{}, dto.WrapError(err, "failed to get chat file")
	}

	return file, nil
}

// DeleteChatFile deletes a chat file by ID
func (s *ChatFileService) DeleteChatFile(ctx context.Context, id int32) error {
	if id <= 0 {
		return dto.ErrValidationInvalidInput("invalid file ID")
	}

	slog.Info("Deleting chat file", "id", id)

	_, err := s.q.DeleteChatFile(ctx, id)
	if err != nil {
		return dto.WrapError(err, "failed to delete chat file")
	}

	return nil
}

// ListChatFilesBySession retrieves chat files for a session
func (s *ChatFileService) ListChatFilesBySession(ctx context.Context, sessionUUID string, userID int32) ([]sqlc_queries.ListChatFilesBySessionUUIDRow, error) {
	if sessionUUID == "" {
		return nil, dto.ErrValidationInvalidInput("missing session UUID")
	}
	if userID <= 0 {
		return nil, dto.ErrValidationInvalidInput("invalid user ID")
	}

	slog.Info("Listing chat files", "session", sessionUUID, "userID", userID)

	files, err := s.q.ListChatFilesBySessionUUID(ctx, sqlc_queries.ListChatFilesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
	})
	if err != nil {
		return nil, dto.WrapError(err, "failed to list chat files")
	}

	return files, nil
}

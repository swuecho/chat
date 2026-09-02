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

type ChatFile sqlc_queries.ChatFile
type ChatFileDetail sqlc_queries.GetChatFileByIDRow
type ChatFileSummary sqlc_queries.ListChatFilesBySessionUUIDRow

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
func (s *ChatFileService) CreateChatUpload(ctx context.Context, input CreateChatFileInput) (ChatFile, error) {
	// Validate input
	if input.ChatSessionUuid == "" {
		return ChatFile{}, domain.Invalid("missing session UUID")
	}
	if input.UserID <= 0 {
		return ChatFile{}, domain.Invalid("invalid user ID")
	}
	if input.Name == "" {
		return ChatFile{}, domain.Invalid("missing file name")
	}
	if len(input.Data) == 0 {
		return ChatFile{}, domain.Invalid("empty file data")
	}

	slog.Info("Creating chat file upload", "session", input.ChatSessionUuid, "userID", input.UserID)

	upload, err := s.q.CreateChatFile(ctx, sqlc_queries.CreateChatFileParams(input))
	if err != nil {
		return ChatFile{}, domain.Internal("failed to create chat file", err)
	}

	slog.Info("Created chat file upload", "id", upload.ID)
	return ChatFile(upload), nil
}

// GetChatFile retrieves a chat file by ID
func (s *ChatFileService) GetChatFile(ctx context.Context, id int32) (ChatFileDetail, error) {
	if id <= 0 {
		return ChatFileDetail{}, domain.Invalid("invalid file ID")
	}

	slog.Info("Retrieving chat file", "id", id)

	file, err := s.q.GetChatFileByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ChatFileDetail{}, domain.NotFound("Chat file", err)
		}
		return ChatFileDetail{}, domain.Internal("failed to get chat file", err)
	}

	return ChatFileDetail(file), nil
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
func (s *ChatFileService) ListChatFilesBySession(ctx context.Context, sessionUUID string, userID int32) ([]ChatFileSummary, error) {
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

	result := make([]ChatFileSummary, len(files))
	for i, file := range files {
		result[i] = ChatFileSummary(file)
	}
	return result, nil
}

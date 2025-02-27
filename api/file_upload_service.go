package main

import (
	"context"
	"log"

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

// CreateChatUpload handles creating a new chat file upload
func (s *ChatFileService) CreateChatUpload(ctx context.Context, params sqlc_queries.CreateChatFileParams) (sqlc_queries.ChatFile, error) {
	// Validate input
	if params.ChatSessionUuid == "" {
		return sqlc_queries.ChatFile{}, ErrValidationInvalidInput("missing session UUID")
	}
	if params.UserID <= 0 {
		return sqlc_queries.ChatFile{}, ErrValidationInvalidInput("invalid user ID")
	}
	if params.Name == "" {
		return sqlc_queries.ChatFile{}, ErrValidationInvalidInput("missing file name")
	}
	if len(params.Data) == 0 {
		return sqlc_queries.ChatFile{}, ErrValidationInvalidInput("empty file data")
	}

	log.Printf("Creating chat file upload for session %s, user %d", 
		params.ChatSessionUuid, params.UserID)

	upload, err := s.q.CreateChatFile(ctx, params)
	if err != nil {
		return sqlc_queries.ChatFile{}, WrapError(err, "failed to create chat file")
	}

	log.Printf("Created chat file upload ID %d", upload.ID)
	return upload, nil
}

// GetChatFile retrieves a chat file by ID
func (s *ChatFileService) GetChatFile(ctx context.Context, id int32) (sqlc_queries.GetChatFileByIDRow, error) {
	if id <= 0 {
		return sqlc_queries.GetChatFileByIDRow{}, ErrValidationInvalidInput("invalid file ID")
	}

	log.Printf("Retrieving chat file ID %d", id)

	file, err := s.q.GetChatFileByID(ctx, id)
	if err != nil {
		return sqlc_queries.GetChatFileByIDRow{}, WrapError(err, "failed to get chat file")
	}

	return file, nil
}

// DeleteChatFile deletes a chat file by ID
func (s *ChatFileService) DeleteChatFile(ctx context.Context, id int32) error {
	if id <= 0 {
		return ErrValidationInvalidInput("invalid file ID")
	}

	log.Printf("Deleting chat file ID %d", id)

	_, err := s.q.DeleteChatFile(ctx, id)
	if err != nil {
		return WrapError(err, "failed to delete chat file")
	}

	return nil
}

// ListChatFilesBySession retrieves chat files for a session
func (s *ChatFileService) ListChatFilesBySession(ctx context.Context, sessionUUID string, userID int32) ([]sqlc_queries.ChatFile, error) {
	if sessionUUID == "" {
		return nil, ErrValidationInvalidInput("missing session UUID")
	}
	if userID <= 0 {
		return nil, ErrValidationInvalidInput("invalid user ID")
	}

	log.Printf("Listing chat files for session %s, user %d", sessionUUID, userID)

	files, err := s.q.ListChatFilesBySessionUUID(ctx, sqlc_queries.ListChatFilesBySessionUUIDParams{
		ChatSessionUuid: sessionUUID,
		UserID:          userID,
	})
	if err != nil {
		return nil, WrapError(err, "failed to list chat files")
	}

	return files, nil
}

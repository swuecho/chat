package service

import (
	"context"

	"github.com/swuecho/chat_backend/repository"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type fileService struct {
	repos repository.CoreRepositoryManager
}

func NewFileService(repos repository.CoreRepositoryManager) FileService {
	return &fileService{repos: repos}
}

func (s *fileService) UploadFile(ctx context.Context, sessionUUID string, userID int32, filename string, data []byte, mimeType string) (*sqlc_queries.ChatFile, error) {
	// TODO: Implement file upload logic with repository
	return nil, nil
}

func (s *fileService) GetFiles(ctx context.Context, sessionUUID string, userID int32) ([]sqlc_queries.ChatFile, error) {
	// TODO: Implement file retrieval logic with repository
	return nil, nil
}

func (s *fileService) DeleteFile(ctx context.Context, fileID int32, userID int32) error {
	// TODO: Implement file deletion logic with repository
	return nil
}
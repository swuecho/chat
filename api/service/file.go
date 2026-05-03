package service

import (
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// FileService provides file management operations.
type FileService struct {
	q *sqlc_queries.Queries
}

func NewFileService(q *sqlc_queries.Queries) *FileService {
	return &FileService{q: q}
}

func (s *FileService) Q() *sqlc_queries.Queries { return s.q }

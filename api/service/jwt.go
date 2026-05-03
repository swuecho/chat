package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

// JWTSecretService manages JWT secrets.
type JWTSecretService struct {
	q *sqlc_queries.Queries
}

func NewJWTSecretService(q *sqlc_queries.Queries) *JWTSecretService {
	return &JWTSecretService{q: q}
}

// GetOrCreate returns a JWT secret by name, creating one if it doesn't exist.
func (s *JWTSecretService) GetOrCreate(ctx context.Context, name string) (sqlc_queries.JwtSecret, error) {
	secret, err := s.q.GetJwtSecret(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			secretStr, audStr := auth.GenJwtSecretAndAudience()
			return s.q.CreateJwtSecret(ctx, sqlc_queries.CreateJwtSecretParams{
				Name:     name,
				Secret:   secretStr,
				Audience: audStr,
			})
		}
		return sqlc_queries.JwtSecret{}, err
	}
	return secret, nil
}

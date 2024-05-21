package main

// check if jwt_secret and jwt_aud available for 'chat' in database
// if not, create them

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type JWTSecretService struct {
	q *sqlc_queries.Queries
}

// NewJWTSecretService creates a new JWTSecretService.
func NewJWTSecretService(q *sqlc_queries.Queries) *JWTSecretService {
	return &JWTSecretService{q: q}
}

// GetJWTSecret returns a jwt_secret by name.
func (s *JWTSecretService) GetJwtSecret(ctx context.Context, name string) (sqlc_queries.JwtSecret, error) {
	secret, err := s.q.GetJwtSecret(ctx, name)
	if err != nil {
		return sqlc_queries.JwtSecret{}, eris.Wrap(err, "failed to get secret ")
	}
	return secret, nil
}

// GetOrCreateJwtSecret returns a jwt_secret by name.
// if jwt_secret does not exist, create it
func (s *JWTSecretService) GetOrCreateJwtSecret(ctx context.Context, name string) (sqlc_queries.JwtSecret, error) {
	secret, err := s.q.GetJwtSecret(ctx, name)
	if err != nil {
		// no row found, create it
		if errors.Is(err, pgx.ErrNoRows) {
			secret_str, aud_str := auth.GenJwtSecretAndAudience()
			secret, err = s.q.CreateJwtSecret(ctx, sqlc_queries.CreateJwtSecretParams{
				Name:     name,
				Secret:   secret_str,
				Audience: aud_str,
			})
			if err != nil {
				return sqlc_queries.JwtSecret{}, eris.Wrap(err, "failed to create secret ")
			}
		} else {
			return sqlc_queries.JwtSecret{}, eris.Wrap(err, "failed to create secret ")
		}
	}
	return secret, nil
}

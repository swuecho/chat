package svc

// check if jwt_secret and jwt_aud available for 'chat' in database
// if not, create them

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rotisserie/eris"
	"github.com/swuecho/chat_backend/auth"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

type JWTSecretService struct {
	q *sqlc_queries.Queries
}

type JWTSecret sqlc_queries.JwtSecret

// NewJWTSecretService creates a new JWTSecretService.
func NewJWTSecretService(q *sqlc_queries.Queries) *JWTSecretService {
	return &JWTSecretService{q: q}
}

// GetJWTSecret returns a jwt_secret by name.
func (s *JWTSecretService) GetJwtSecret(ctx context.Context, name string) (JWTSecret, error) {
	secret, err := s.q.GetJwtSecret(ctx, name)
	if err != nil {
		return JWTSecret{}, eris.Wrap(err, "failed to get secret ")
	}
	return JWTSecret(secret), nil
}

// GetOrCreateJwtSecret returns a jwt_secret by name.
// if jwt_secret does not exist, create it
func (s *JWTSecretService) GetOrCreateJwtSecret(ctx context.Context, name string) (JWTSecret, error) {
	secret, err := s.q.GetJwtSecret(ctx, name)
	if err != nil {
		// no row found, create it
		if errors.Is(err, sql.ErrNoRows) {
			secret_str, aud_str := auth.GenJwtSecretAndAudience()
			secret, err = s.q.CreateJwtSecret(ctx, sqlc_queries.CreateJwtSecretParams{
				Name:     name,
				Secret:   secret_str,
				Audience: aud_str,
			})
			if err != nil {
				return JWTSecret{}, eris.Wrap(err, "failed to create secret ")
			}
		} else {
			return JWTSecret{}, eris.Wrap(err, "failed to create secret ")
		}
	}
	return JWTSecret(secret), nil
}

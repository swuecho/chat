// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: jwt_secrets.sql

package sqlc_queries

import (
	"context"
)

const createJwtSecret = `-- name: CreateJwtSecret :one
INSERT INTO jwt_secrets (name, secret, audience)
VALUES ($1, $2, $3) RETURNING id, name, secret, audience, lifetime
`

type CreateJwtSecretParams struct {
	Name     string `json:"name"`
	Secret   string `json:"secret"`
	Audience string `json:"audience"`
}

func (q *Queries) CreateJwtSecret(ctx context.Context, arg CreateJwtSecretParams) (JwtSecret, error) {
	row := q.db.QueryRowContext(ctx, createJwtSecret, arg.Name, arg.Secret, arg.Audience)
	var i JwtSecret
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Secret,
		&i.Audience,
		&i.Lifetime,
	)
	return i, err
}

const deleteAllJwtSecrets = `-- name: DeleteAllJwtSecrets :execrows
DELETE FROM jwt_secrets
`

func (q *Queries) DeleteAllJwtSecrets(ctx context.Context) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteAllJwtSecrets)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getJwtSecret = `-- name: GetJwtSecret :one
SELECT id, name, secret, audience, lifetime FROM jwt_secrets WHERE name = $1
`

func (q *Queries) GetJwtSecret(ctx context.Context, name string) (JwtSecret, error) {
	row := q.db.QueryRowContext(ctx, getJwtSecret, name)
	var i JwtSecret
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Secret,
		&i.Audience,
		&i.Lifetime,
	)
	return i, err
}

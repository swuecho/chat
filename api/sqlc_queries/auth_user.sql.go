// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: auth_user.sql

package sqlc_queries

import (
	"context"
)

const createAuthUser = `-- name: CreateAuthUser :one
INSERT INTO auth_user (email, "password", first_name, last_name, username, is_staff, is_superuser)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined
`

type CreateAuthUserParams struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Username    string `json:"username"`
	IsStaff     bool   `json:"isStaff"`
	IsSuperuser bool   `json:"isSuperuser"`
}

func (q *Queries) CreateAuthUser(ctx context.Context, arg CreateAuthUserParams) (AuthUser, error) {
	row := q.db.QueryRow(ctx, createAuthUser,
		arg.Email,
		arg.Password,
		arg.FirstName,
		arg.LastName,
		arg.Username,
		arg.IsStaff,
		arg.IsSuperuser,
	)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.Password,
		&i.LastLogin,
		&i.IsSuperuser,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.IsStaff,
		&i.IsActive,
		&i.DateJoined,
	)
	return i, err
}

const deleteAuthUser = `-- name: DeleteAuthUser :exec
DELETE FROM auth_user WHERE email = $1
`

func (q *Queries) DeleteAuthUser(ctx context.Context, email string) error {
	_, err := q.db.Exec(ctx, deleteAuthUser, email)
	return err
}

const getAllAuthUsers = `-- name: GetAllAuthUsers :many
SELECT id, password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined FROM auth_user ORDER BY id
`

func (q *Queries) GetAllAuthUsers(ctx context.Context) ([]AuthUser, error) {
	rows, err := q.db.Query(ctx, getAllAuthUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuthUser
	for rows.Next() {
		var i AuthUser
		if err := rows.Scan(
			&i.ID,
			&i.Password,
			&i.LastLogin,
			&i.IsSuperuser,
			&i.Username,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.IsStaff,
			&i.IsActive,
			&i.DateJoined,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAuthUserByEmail = `-- name: GetAuthUserByEmail :one
SELECT id, password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined FROM auth_user WHERE email = $1
`

func (q *Queries) GetAuthUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	row := q.db.QueryRow(ctx, getAuthUserByEmail, email)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.Password,
		&i.LastLogin,
		&i.IsSuperuser,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.IsStaff,
		&i.IsActive,
		&i.DateJoined,
	)
	return i, err
}

const getAuthUserByID = `-- name: GetAuthUserByID :one
SELECT id, password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined FROM auth_user WHERE id = $1
`

func (q *Queries) GetAuthUserByID(ctx context.Context, id int32) (AuthUser, error) {
	row := q.db.QueryRow(ctx, getAuthUserByID, id)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.Password,
		&i.LastLogin,
		&i.IsSuperuser,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.IsStaff,
		&i.IsActive,
		&i.DateJoined,
	)
	return i, err
}

const getTotalActiveUserCount = `-- name: GetTotalActiveUserCount :one
SELECT COUNT(*) FROM auth_user WHERE is_active = true
`

func (q *Queries) GetTotalActiveUserCount(ctx context.Context) (int64, error) {
	row := q.db.QueryRow(ctx, getTotalActiveUserCount)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined FROM auth_user WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i AuthUser
	err := row.Scan(
		&i.ID,
		&i.Password,
		&i.LastLogin,
		&i.IsSuperuser,
		&i.Username,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.IsStaff,
		&i.IsActive,
		&i.DateJoined,
	)
	return i, err
}

const getUserStats = `-- name: GetUserStats :many
SELECT 
    auth_user.first_name,
    auth_user.last_name,
    auth_user.email AS user_email,
    COALESCE(user_stats.total_messages, 0) AS total_chat_messages,
    COALESCE(user_stats.total_token_count, 0) AS total_token_count,
    COALESCE(user_stats.total_messages_3_days, 0) AS total_chat_messages_3_days,
    COALESCE(user_stats.total_token_count_3_days, 0) AS total_token_count_3_days,
    COALESCE(auth_user_management.rate_limit, $3::INTEGER) AS rate_limit
FROM auth_user
LEFT JOIN (
    SELECT chat_message_stats.user_id, 
           SUM(total_messages) AS total_messages, 
           SUM(total_token_count) AS total_token_count,
           SUM(CASE WHEN created_at >= NOW() - INTERVAL '3 days' THEN total_messages ELSE 0 END) AS total_messages_3_days,
           SUM(CASE WHEN created_at >= NOW() - INTERVAL '3 days' THEN total_token_count ELSE 0 END) AS total_token_count_3_days
    FROM (
        SELECT user_id, COUNT(*) AS total_messages, SUM(token_count) as total_token_count, MAX(created_at) AS created_at
        FROM chat_message
        GROUP BY user_id, chat_session_uuid
    ) AS chat_message_stats
    GROUP BY chat_message_stats.user_id
) AS user_stats ON auth_user.id = user_stats.user_id
LEFT JOIN auth_user_management ON auth_user.id = auth_user_management.user_id
ORDER BY total_chat_messages DESC, auth_user.id DESC
OFFSET $2
LIMIT $1
`

type GetUserStatsParams struct {
	Limit            int32 `json:"limit"`
	Offset           int32 `json:"offset"`
	DefaultRateLimit int32 `json:"defaultRateLimit"`
}

type GetUserStatsRow struct {
	FirstName              string `json:"firstName"`
	LastName               string `json:"lastName"`
	UserEmail              string `json:"userEmail"`
	TotalChatMessages      int64  `json:"totalChatMessages"`
	TotalTokenCount        int64  `json:"totalTokenCount"`
	TotalChatMessages3Days int64  `json:"totalChatMessages3Days"`
	TotalTokenCount3Days   int64  `json:"totalTokenCount3Days"`
	RateLimit              int32  `json:"rateLimit"`
}

func (q *Queries) GetUserStats(ctx context.Context, arg GetUserStatsParams) ([]GetUserStatsRow, error) {
	rows, err := q.db.Query(ctx, getUserStats, arg.Limit, arg.Offset, arg.DefaultRateLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserStatsRow
	for rows.Next() {
		var i GetUserStatsRow
		if err := rows.Scan(
			&i.FirstName,
			&i.LastName,
			&i.UserEmail,
			&i.TotalChatMessages,
			&i.TotalTokenCount,
			&i.TotalChatMessages3Days,
			&i.TotalTokenCount3Days,
			&i.RateLimit,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listAuthUsers = `-- name: ListAuthUsers :many
SELECT id, password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined FROM auth_user ORDER BY id LIMIT $1 OFFSET $2
`

type ListAuthUsersParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) ListAuthUsers(ctx context.Context, arg ListAuthUsersParams) ([]AuthUser, error) {
	rows, err := q.db.Query(ctx, listAuthUsers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuthUser
	for rows.Next() {
		var i AuthUser
		if err := rows.Scan(
			&i.ID,
			&i.Password,
			&i.LastLogin,
			&i.IsSuperuser,
			&i.Username,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.IsStaff,
			&i.IsActive,
			&i.DateJoined,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateAuthUser = `-- name: UpdateAuthUser :one
UPDATE auth_user SET first_name = $2, last_name= $3, last_login = now() 
WHERE id = $1
RETURNING first_name, last_name, email
`

type UpdateAuthUserParams struct {
	ID        int32  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateAuthUserRow struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func (q *Queries) UpdateAuthUser(ctx context.Context, arg UpdateAuthUserParams) (UpdateAuthUserRow, error) {
	row := q.db.QueryRow(ctx, updateAuthUser, arg.ID, arg.FirstName, arg.LastName)
	var i UpdateAuthUserRow
	err := row.Scan(&i.FirstName, &i.LastName, &i.Email)
	return i, err
}

const updateAuthUserByEmail = `-- name: UpdateAuthUserByEmail :one
UPDATE auth_user SET first_name = $2, last_name= $3, last_login = now() 
WHERE email = $1
RETURNING first_name, last_name, email
`

type UpdateAuthUserByEmailParams struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateAuthUserByEmailRow struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

func (q *Queries) UpdateAuthUserByEmail(ctx context.Context, arg UpdateAuthUserByEmailParams) (UpdateAuthUserByEmailRow, error) {
	row := q.db.QueryRow(ctx, updateAuthUserByEmail, arg.Email, arg.FirstName, arg.LastName)
	var i UpdateAuthUserByEmailRow
	err := row.Scan(&i.FirstName, &i.LastName, &i.Email)
	return i, err
}

const updateAuthUserRateLimitByEmail = `-- name: UpdateAuthUserRateLimitByEmail :one
INSERT INTO auth_user_management (user_id, rate_limit, created_at, updated_at)
VALUES ((SELECT id FROM auth_user WHERE email = $1), $2, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET rate_limit = $2, updated_at = NOW()
RETURNING rate_limit
`

type UpdateAuthUserRateLimitByEmailParams struct {
	Email     string `json:"email"`
	RateLimit int32  `json:"rateLimit"`
}

func (q *Queries) UpdateAuthUserRateLimitByEmail(ctx context.Context, arg UpdateAuthUserRateLimitByEmailParams) (int32, error) {
	row := q.db.QueryRow(ctx, updateAuthUserRateLimitByEmail, arg.Email, arg.RateLimit)
	var rate_limit int32
	err := row.Scan(&rate_limit)
	return rate_limit, err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE auth_user SET "password" = $2 WHERE email = $1
`

type UpdateUserPasswordParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.Exec(ctx, updateUserPassword, arg.Email, arg.Password)
	return err
}

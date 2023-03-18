-- name: GetAllAuthUsers :many
SELECT * FROM auth_user ORDER BY id;

-- name: ListAuthUsers :many
SELECT * FROM auth_user ORDER BY id LIMIT $1 OFFSET $2;

-- name: GetAuthUserByID :one
SELECT * FROM auth_user WHERE id = $1;


-- name: GetAuthUserByEmail :one
SELECT * FROM auth_user WHERE email = $1;

-- name: CreateAuthUser :one
INSERT INTO auth_user (email, "password", first_name, last_name, username, is_staff, is_superuser)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateAuthUser :one
UPDATE auth_user SET "password" = $2, is_superuser = $3, username = $4, first_name = $5, last_name = $6, email = $7, last_login = now() 
WHERE id = $1
RETURNING *;

-- name: DeleteAuthUser :exec
DELETE FROM auth_user WHERE email = $1;

-- name: GetUserByEmail :one
SELECT * FROM auth_user WHERE email = $1;

-- name: UpdateUserPassword :exec
UPDATE auth_user SET "password" = $2 WHERE email = $1;

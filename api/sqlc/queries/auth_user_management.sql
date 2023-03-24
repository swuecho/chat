-- name: GetRateLimit :one
SELECT COALESCE(rate_limit, 100) AS rate_limit
FROM auth_user_management
WHERE user_id = $1;

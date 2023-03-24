-- name: GetRateLimit :one
-- GetRateLimit retrieves the rate limit for a user from the auth_user_management table.
-- If no rate limit is set for the user, it returns the default rate limit of 100.
SELECT COALESCE(rate_limit, 100) AS rate_limit
FROM auth_user_management
WHERE user_id = $1;

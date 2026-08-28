-- name: CreateVirtualAPIKey :one
INSERT INTO virtual_api_key (user_id, name, key_prefix, key_hash, expires_at, requests_per_minute)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListVirtualAPIKeysByUser :many
SELECT * FROM virtual_api_key
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: VirtualAPIKeyByHash :one
SELECT * FROM virtual_api_key
WHERE key_hash = $1;

-- name: VirtualAdminAPIKeyByHash :one
SELECT vak.*
FROM virtual_api_key vak
JOIN auth_user au ON au.id = vak.user_id
WHERE vak.key_hash = $1
  AND au.is_superuser = true
  AND au.is_active = true;

-- name: VirtualAPIKeyByIDAndUser :one
SELECT * FROM virtual_api_key
WHERE id = $1 AND user_id = $2;

-- name: RevokeVirtualAPIKey :execrows
UPDATE virtual_api_key
SET status = 'revoked', revoked_at = NOW()
WHERE id = $1 AND user_id = $2 AND status = 'active';

-- name: TouchVirtualAPIKey :exec
UPDATE virtual_api_key SET last_used_at = NOW() WHERE id = $1;

-- name: CountRecentGatewayRequests :one
SELECT COUNT(*) FROM gateway_request
WHERE api_key_id = $1 AND created_at >= NOW() - INTERVAL '1 minute';

-- name: CreateGatewayRequest :one
INSERT INTO gateway_request (
    request_uuid, api_key_id, user_id, chat_model_id, requested_model, provider, stream,
    request_bytes, request_sha256, request_sample, request_truncated,
    request_classification, retention_until
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: CompleteGatewayRequest :exec
UPDATE gateway_request
SET status = $2,
    prompt_tokens = $3,
    completion_tokens = $4,
    total_tokens = $5,
    latency_ms = $6,
    provider_request_id = $7,
    error_code = $8,
    response_bytes = $9,
    response_sha256 = $10,
    response_sample = $11,
    response_truncated = $12,
    response_classification = $13,
    completed_at = NOW()
WHERE id = $1;

-- name: PurgeExpiredGatewaySamples :execrows
UPDATE gateway_request
SET request_sample = ''::BYTEA, response_sample = ''::BYTEA
WHERE retention_until <= NOW()
  AND (octet_length(request_sample) > 0 OR octet_length(response_sample) > 0);

-- name: GatewayUsageByKey :many
SELECT
    requested_model,
    COUNT(*) AS request_count,
    COALESCE(SUM(prompt_tokens), 0)::BIGINT AS prompt_tokens,
    COALESCE(SUM(completion_tokens), 0)::BIGINT AS completion_tokens,
    COALESCE(SUM(total_tokens), 0)::BIGINT AS total_tokens,
    MAX(created_at) AS last_used_at
FROM gateway_request
WHERE api_key_id = $1 AND user_id = $2
GROUP BY requested_model
ORDER BY request_count DESC, requested_model;

-- name: ListGatewayRequestsByKey :many
SELECT id, request_uuid, requested_model, provider, status, stream,
       prompt_tokens, completion_tokens, total_tokens, latency_ms,
       request_bytes, response_bytes, request_truncated, response_truncated,
       created_at, completed_at, retention_until, error_code
FROM gateway_request
WHERE api_key_id = $1 AND user_id = $2
ORDER BY created_at DESC
LIMIT $3;

-- name: GatewayRequestByIDAndUser :one
SELECT * FROM gateway_request
WHERE id = $1 AND api_key_id = $2 AND user_id = $3;

-- name: ListGatewayModels :many
SELECT * FROM chat_model
WHERE is_enable = true AND api_type = 'openai'
ORDER BY order_number, name;

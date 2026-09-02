-- name: ClaimChatRequest :one
INSERT INTO chat_request (uuid, chat_session_uuid, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (chat_session_uuid, uuid) DO UPDATE
SET status = 'pending',
    error_code = '',
    attempt_count = chat_request.attempt_count + 1,
    updated_at = now()
WHERE chat_request.status IN ('failed', 'canceled', 'cancelled')
RETURNING *;

-- name: GetChatRequest :one
SELECT *
FROM chat_request
WHERE uuid = $1
  AND chat_session_uuid = $2
  AND user_id = $3;

-- name: MarkChatRequestStreaming :execrows
UPDATE chat_request
SET status = 'streaming', updated_at = now()
WHERE uuid = $1
  AND chat_session_uuid = $2
  AND user_id = $3
  AND status = 'pending';

-- name: MarkChatRequestFailed :execrows
UPDATE chat_request
SET status = 'failed', error_code = $4, updated_at = now()
WHERE uuid = $1
  AND chat_session_uuid = $2
  AND user_id = $3
  AND status IN ('pending', 'streaming');

-- name: MarkChatRequestCanceled :execrows
UPDATE chat_request
SET status = 'canceled', error_code = $4, updated_at = now()
WHERE uuid = $1
  AND chat_session_uuid = $2
  AND user_id = $3
  AND status IN ('pending', 'streaming');

-- name: CompleteChatRequestWithMessage :one
WITH eligible_request AS (
    SELECT id
    FROM chat_request
    WHERE chat_request.uuid = sqlc.arg(request_uuid)
      AND chat_request.chat_session_uuid = sqlc.arg(chat_session_uuid)
      AND chat_request.user_id = sqlc.arg(user_id)
      AND chat_request.status IN ('pending', 'streaming')
    FOR UPDATE
),
saved_message AS (
    INSERT INTO chat_message (
        chat_session_uuid, uuid, role, content, reasoning_content, model,
        token_count, score, user_id, created_by, updated_by, llm_summary,
        raw, artifacts, suggested_questions
    )
    SELECT
        sqlc.arg(chat_session_uuid), sqlc.arg(message_uuid), 'assistant',
        sqlc.arg(content), sqlc.arg(reasoning_content), sqlc.arg(model),
        sqlc.arg(token_count), sqlc.arg(score), sqlc.arg(user_id),
        sqlc.arg(created_by), sqlc.arg(updated_by), sqlc.arg(llm_summary),
        sqlc.arg(raw), sqlc.arg(artifacts), sqlc.arg(suggested_questions)
    FROM eligible_request
    ON CONFLICT (chat_session_uuid, uuid) WHERE is_deleted = false DO UPDATE
    SET uuid = EXCLUDED.uuid
    RETURNING *
),
completed_request AS (
    UPDATE chat_request
    SET status = 'completed',
        assistant_uuid = (SELECT saved_message.uuid FROM saved_message),
        error_code = '',
        updated_at = now()
    WHERE chat_request.uuid = sqlc.arg(request_uuid)
      AND chat_request.chat_session_uuid = sqlc.arg(chat_session_uuid)
      AND chat_request.user_id = sqlc.arg(user_id)
      AND chat_request.id IN (SELECT id FROM eligible_request)
    RETURNING id
)
SELECT saved_message.*
FROM saved_message
JOIN completed_request ON true;

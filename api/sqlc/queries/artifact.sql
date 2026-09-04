-- name: ListArtifacts :many
SELECT
    CAST(artifact.value AS text) AS artifact_json,
    message.uuid AS message_uuid,
    session.uuid AS session_uuid,
    session.topic AS session_title,
    message.created_at,
    message.updated_at
FROM chat_message AS message
JOIN chat_session AS session ON session.uuid = message.chat_session_uuid
CROSS JOIN LATERAL jsonb_array_elements(message.artifacts) AS artifact(value)
WHERE message.user_id = sqlc.arg(user_id)
  AND message.is_deleted = false
  AND session.active = true
  AND (sqlc.arg(search)::text = '' OR concat_ws(' ', artifact.value->>'title', artifact.value->>'content', artifact.value->>'type', artifact.value->>'language', session.topic) ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (sqlc.arg(artifact_type)::text = '' OR artifact.value->>'type' = sqlc.arg(artifact_type)::text)
  AND (sqlc.arg(language)::text = '' OR artifact.value->>'language' = sqlc.arg(language)::text)
  AND (sqlc.arg(session_uuid)::text = '' OR session.uuid = sqlc.arg(session_uuid)::text)
ORDER BY message.created_at DESC, artifact.value->>'uuid'
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountArtifacts :one
SELECT count(*)
FROM chat_message AS message
JOIN chat_session AS session ON session.uuid = message.chat_session_uuid
CROSS JOIN LATERAL jsonb_array_elements(message.artifacts) AS artifact(value)
WHERE message.user_id = sqlc.arg(user_id)
  AND message.is_deleted = false
  AND session.active = true
  AND (sqlc.arg(search)::text = '' OR concat_ws(' ', artifact.value->>'title', artifact.value->>'content', artifact.value->>'type', artifact.value->>'language', session.topic) ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (sqlc.arg(artifact_type)::text = '' OR artifact.value->>'type' = sqlc.arg(artifact_type)::text)
  AND (sqlc.arg(language)::text = '' OR artifact.value->>'language' = sqlc.arg(language)::text)
  AND (sqlc.arg(session_uuid)::text = '' OR session.uuid = sqlc.arg(session_uuid)::text);

-- name: UpdateArtifact :one
UPDATE chat_message AS message
SET artifacts = (
    SELECT jsonb_agg(
        CASE WHEN artifact.value->>'uuid' = sqlc.arg(artifact_uuid)::text
            THEN artifact.value || jsonb_build_object('title', sqlc.arg(title)::text, 'content', sqlc.arg(content)::text, 'language', sqlc.arg(language)::text)
            ELSE artifact.value
        END ORDER BY artifact.ordinality
    )
    FROM jsonb_array_elements(message.artifacts) WITH ORDINALITY AS artifact(value, ordinality)
), updated_at = now()
WHERE message.user_id = sqlc.arg(user_id)
  AND message.is_deleted = false
  AND EXISTS (SELECT 1 FROM jsonb_array_elements(message.artifacts) AS artifact WHERE artifact->>'uuid' = sqlc.arg(artifact_uuid)::text)
RETURNING message.uuid;

-- name: DeleteArtifact :one
UPDATE chat_message AS message
SET artifacts = COALESCE((
    SELECT jsonb_agg(artifact.value ORDER BY artifact.ordinality)
    FROM jsonb_array_elements(message.artifacts) WITH ORDINALITY AS artifact(value, ordinality)
    WHERE artifact.value->>'uuid' <> sqlc.arg(artifact_uuid)::text
), '[]'::jsonb), updated_at = now()
WHERE message.user_id = sqlc.arg(user_id)
  AND message.is_deleted = false
  AND EXISTS (SELECT 1 FROM jsonb_array_elements(message.artifacts) AS artifact WHERE artifact->>'uuid' = sqlc.arg(artifact_uuid)::text)
RETURNING message.uuid;

-- name: DuplicateArtifact :one
UPDATE chat_message AS message
SET artifacts = message.artifacts || (
    SELECT (artifact || jsonb_build_object('uuid', sqlc.arg(new_uuid)::text, 'title', artifact->>'title' || ' (Copy)'))
    FROM jsonb_array_elements(message.artifacts) AS artifact
    WHERE artifact->>'uuid' = sqlc.arg(artifact_uuid)::text
    LIMIT 1
), updated_at = now()
WHERE message.user_id = sqlc.arg(user_id)
  AND message.is_deleted = false
  AND EXISTS (SELECT 1 FROM jsonb_array_elements(message.artifacts) AS artifact WHERE artifact->>'uuid' = sqlc.arg(artifact_uuid)::text)
RETURNING message.uuid;

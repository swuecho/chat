you are a golang code assistant. Given table DDL, you will write all queies for sqlc in a crud applicaiton,

please do not send me any generated go code.

### input

CREATE TABLE chat_message (
    id integer PRIMARY KEY,
    chat_session_id integer NOT NULL,
    role character varying(255) NOT NULL,
    content character varying NOT NULL,
    score double precision NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone,
    updated_at timestamp without time zone,
    created_by integer NOT NULL,
    updated_by integer NOT NULL,
    raw jsonb
);

## output

-- name: ListChatMessages :many
SELECT * FROM chat_message ORDER BY id;

-- name: ChatMessagesBySessionID :many
SELECT * FROM chat_message WHERE chat_session_id = $1 ORDER BY id;

-- name: ChatMessageByID :one
SELECT * FROM chat_message WHERE id = $1;

-- name: CreateChatMessage :one
INSERT INTO chat_message (chat_session_id, role, content, model, score, user_id, created_at, updated_at, created_by, updated_by, raw)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateChatMessage :one
UPDATE chat_message SET role = $2, content = $3, score = $4, user_id = $5, updated_at = $6, updated_by = $7, raw = $8
WHERE id = $1
RETURNING *;

-- name: DeleteChatMessage :exec
DELETE FROM chat_message WHERE id = $1;



### input

CREATE TABLE chat_session (
    id integer PRIMARY KEY,
    user_id integer NOT NULL,
    topic character varying(255) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    active boolean default true NOT NULL,
    max_length integer DEFAULT 0 NOT NULL
);


### 
type ChatSessionService struct {
	q *sqlc_queries.Queries
}

// NewChatSessionService creates a new ChatSessionService.
func NewChatSessionService(q *sqlc_queries.Queries) *ChatSessionService {
	return &ChatSessionService{q: q}
}

// CreateChatSession creates a new chat session.
func (s *ChatSessionService) CreateChatSession(ctx context.Context, session_params sqlc_queries.CreateChatSessionParams) (sqlc_queries.ChatSession, error) {
	session, err := s.q.CreateChatSession(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, errors.New("failed to create session")
	}
	return session, nil
}

// GetChatSessionByID returns a chat session by ID.
func (s *ChatSessionService) GetChatSessionByID(ctx context.Context, id int32) (sqlc_queries.ChatSession, error) {
	session, err := s.q.GetChatSessionByID(ctx, id)
	if err != nil {
		return sqlc_queries.ChatSession{}, errors.New("failed to retrieve session")
	}
	return session, nil
}

// UpdateChatSession updates an existing chat session.
func (s *ChatSessionService) UpdateChatSession(ctx context.Context, session_params sqlc_queries.UpdateChatSessionParams) (sqlc_queries.ChatSession, error) {
	session_u, err := s.q.UpdateChatSession(ctx, session_params)
	if err != nil {
		return sqlc_queries.ChatSession{}, errors.New("failed to update session")
	}
	return session_u, nil
}

// DeleteChatSession deletes a chat session by ID.
func (s *ChatSessionService) DeleteChatSession(ctx context.Context, id int32) error {
	err := s.q.DeleteChatSession(ctx, id)
	if err != nil {
		return errors.New("failed to delete session")
	}
	return nil
}

// GetAllChatSessions returns all chat sessions.
func (s *ChatSessionService) GetAllChatSessions(ctx context.Context) ([]sqlc_queries.ChatSession, error) {
	sessions, err := s.q.GetAllChatSessions(ctx)
	if err != nil {
		return nil, errors.New("failed to retrieve sessions")
	}
	return sessions, nil
}





create sql

INSERT INTO auth_user (id, username, email, password, first_name, last_name, is_active, is_staff, is_superuser, date_joined)
VALUES (1, 'echowuhao', 'echowuhao@gmail.com', 
'pbkdf2_sha256$150000$wVq3kpPZc7pJ$+dO5tCzI9Xu9iGkWtL/Ho11DQsoOx2ZB1OVDGOlKyk4=', 'Hao', 'Wu', true, false, false, now());

Note that when generating password hashes using Django or any other library, it is important to use a strong, one-way hashing algorithm with a sufficiently high cost parameter. In this example, the cost factor is set to 150000, which should provide adequate security against brute-force attacks.



DROP FUNCTION IF EXISTS tsvector_immutable(text);
-- why this is necessary?
CREATE FUNCTION tsvector_immutable(text) RETURNS tsvector AS $$
    SELECT to_tsvector($1)
$$ LANGUAGE sql IMMUTABLE;


UPDATE chat_snapshot
SET text = array_to_string(ARRAY(SELECT jsonb_array_elements(conversation)->>'text'), ' ')::text
WHERE text = ''

ALTER TABLE chat_snapshot
ADD COLUMN IF NOT EXISTS text_vector tsvector generated always as	(
    to_tsvector(text)
) stored; 

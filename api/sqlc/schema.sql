CREATE TABLE IF NOT EXISTS jwt_secrets (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    secret TEXT NOT NULL,
    audience TEXT NOT NULL,
    lifetime smallint NOT NULL default 24
);

ALTER TABLE jwt_secrets ADD COLUMN IF NOT EXISTS lifetime smallint NOT NULL default 24;

UPDATE jwt_secrets SET lifetime = 240;

CREATE TABLE IF NOT EXISTS chat_model (
  id SERIAL PRIMARY KEY,  
  -- model name 'claude-v1', 'gpt-3.5-turbo'
  name TEXT UNIQUE  DEFAULT '' NOT NULL,   
  -- model label 'Claude', 'GPT-3.5 Turbo'
  label TEXT  DEFAULT '' NOT NULL,   
  is_default BOOLEAN DEFAULT false NOT NULL,
  url TEXT  DEFAULT '' NOT NULL,  
  api_auth_header TEXT DEFAULT '' NOT NULL,   
  -- env var that contains the api key
  -- for example: OPENAI_API_KEY, which means the api key is stored in an env var called OPENAI_API_KEY
  api_auth_key TEXT DEFAULT '' NOT NULL,
  user_id INTEGER NOT NULL default 1,
  enable_per_mode_ratelimit BOOLEAN DEFAULT false NOT NULL,
  max_token INTEGER NOT NULL default 120,
  default_token INTEGER NOT NULL default 120,
  order_number INTEGER NOT NULL default 1,
  http_time_out INTEGER NOT NULL default 120
);

ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS user_id INTEGER NOT NULL default 1;
ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS enable_per_mode_ratelimit BOOLEAN DEFAULT false NOT NULL;
ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS max_token INTEGER NOT NULL default 4096;
ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS default_token INTEGER NOT NULL default 2048;
ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS order_number INTEGER NOT NULL default 1;
ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS http_time_out INTEGER NOT NULL default 120;
ALTER TABLE chat_model ADD COLUMN IF NOT EXISTS is_enable BOOLEAN DEFAULT true NOT NULL;


INSERT INTO chat_model(name, label, is_default, url, api_auth_header, api_auth_key, max_token, default_token, order_number)
VALUES  ('gpt-3.5-turbo', 'gpt-3.5-turbo(chatgpt)', true, 'https://api.openai.com/v1/chat/completions', 'Authorization', 'OPENAI_API_KEY', 4096, 2048, 1),
        ('claude-2', 'claude', false, 'https://api.anthropic.com/v1/complete', 'x-api-key', 'CLAUDE_API_KEY', 102400, 102400, 3),
        ('claude-v1', 'claude-v1 (claude)', false, 'https://api.anthropic.com/v1/complete', 'x-api-key', 'CLAUDE_API_KEY', 4096, 2048, 2),
        ('gpt-3.5-turbo-16k', 'gpt-3.5-16k', true, 'https://api.openai.com/v1/chat/completions', 'Authorization', 'OPENAI_API_KEY', 16384, 8192, 2),
        ('claude-v1-100k', 'claude-v1-100k', false, 'https://api.anthropic.com/v1/complete', 'x-api-key', 'CLAUDE_API_KEY', 102400, 102400, 3),
        ('claude-instant-1', 'claude-instant(small,fast)', false, 'https://api.anthropic.com/v1/complete', 'x-api-key', 'CLAUDE_API_KEY',  9192, 2048, 4 ),
        ('claude-instant-v1', 'claude-instant(small,fast)', false, 'https://api.anthropic.com/v1/complete', 'x-api-key', 'CLAUDE_API_KEY',  9192, 2048, 4 ),
        ('gpt-4', 'gpt-4(chatgpt)', false, 'https://api.openai.com/v1/chat/completions', 'Authorization', 'OPENAI_API_KEY',  9192, 2048, 5),
        ('gpt-4-32k', 'gpt-4-32k(chatgpt)', false, 'https://api.openai.com/v1/chat/completions', 'Authorization', 'OPENAI_API_KEY',  9192, 2048, 6),
        ('text-davinci-003', 'text-davinci-003', false, 'https://api.openai.com/v1/completions', 'Authorization', 'OPENAI_API_KEY', 4096, 2048, 7),
        ('echo','echo',false,'https://bestqa_workerd.bestqa.workers.dev/echo','Authorization','ECHO_API_KEY', 40960, 20480, 8),
        ('debug','debug',false,'https://bestqa_workerd.bestqa.workers.dev/debug','Authorization','ECHO_API_KEY', 40960, 2048, 9)
ON CONFLICT(name) DO NOTHING;

UPDATE chat_model SET enable_per_mode_ratelimit = true WHERE name = 'gpt-4';
UPDATE chat_model SET enable_per_mode_ratelimit = true WHERE name = 'gpt-4-32k';
DELETE FROM chat_model where name = 'claude-v1';
DELETE FROM chat_model where name = 'claude-v1-100k';
DELETE FROM chat_model where name = 'claude-instant-v1';
-- create index on name
CREATE INDEX IF NOT EXISTS jwt_secrets_name_idx ON jwt_secrets (name);


CREATE TABLE IF NOT EXISTS auth_user (
  id SERIAL PRIMARY KEY,
  password VARCHAR(128) NOT NULL,
  last_login TIMESTAMP default now() NOT NULL,
  is_superuser BOOLEAN default false NOT NULL,
  username VARCHAR(150) UNIQUE NOT NULL,
  first_name VARCHAR(30) default '' NOT NULL,
  last_name VARCHAR(30) default '' NOT NULL,
  email VARCHAR(254) UNIQUE NOT NULL,
  is_staff BOOLEAN default false NOT NULL,
  is_active BOOLEAN default true NOT NULL,
  date_joined TIMESTAMP default now() NOT NULL
);

-- add index on email
CREATE INDEX IF NOT EXISTS auth_user_email_idx ON auth_user (email);

CREATE TABLE IF NOT EXISTS auth_user_management (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES auth_user(id) ON DELETE CASCADE,
    rate_limit INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL
);

-- add index on user_id
CREATE INDEX IF NOT EXISTS auth_user_management_user_id_idx ON auth_user_management (user_id);


-- control specific model ratelimit, like gpt4
-- if not find gpt4 on privilege than forbiden
-- if found, then check the acess count (session messages).
-- get rate_limit by user_id, chat_session_uuid
CREATE TABLE IF NOT EXISTS user_chat_model_privilege(
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES auth_user(id) ON DELETE CASCADE,
    chat_model_id INT NOT NULL REFERENCES chat_model(id) ON DELETE CASCADE,
    rate_limit INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL,
    created_by INTEGER NOT NULL DEFAULT 0,
    updated_by INTEGER NOT NULL DEFAULT 0, 
    CONSTRAINT chat_usage_user_model_unique UNIQUE (user_id, chat_model_id)
);

CREATE TABLE IF NOT EXISTS chat_session (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL,
    --ALTER TABLE chat_session ADD COLUMN uuid character varying(255) NOT NULL DEFAULT '';
    uuid character varying(255) UNIQUE NOT NULL,
    topic character varying(255) NOT NULL,
    created_at timestamp  DEFAULT now() NOT NULL,
    updated_at timestamp  DEFAULT now() NOT NULL,
    active boolean default true NOT NULL,
    model character varying(255) NOT NULL DEFAULT 'gpt-3.5-turbo',
    max_length integer DEFAULT 0 NOT NULL,
    temperature float DEFAULT 1.0 NOT NUll,
    top_p float DEFAULT 1.0 NOT NUll,
    max_tokens int DEFAULT 512 NOT NULL,
    n  integer DEFAULT 1 NOT NULL,
    summarize_mode boolean DEFAULT false NOT NULL
);


-- chat_session
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS temperature float DEFAULT 1.0 NOT NULL;
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS top_p float DEFAULT 1.0 NOT NULL;
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS max_tokens int DEFAULT 512 NOT NULL; 
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS debug boolean DEFAULT false NOT NULL; 
ALTER TABlE chat_session ADD COLUMN IF NOT EXISTS model character varying(255) NOT NULL DEFAULT 'gpt-3.5-turbo';
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS n INTEGER DEFAULT 1 NOT NULL;
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS summarize_mode boolean DEFAULT false NOT NULL;


-- add hash index on uuid
CREATE INDEX IF NOT EXISTS chat_session_uuid_idx ON chat_session using hash (uuid) ;

-- add index on user_id
CREATE INDEX IF NOT EXISTS chat_session_user_id_idx ON chat_session (user_id);

CREATE TABLE IF NOT EXISTS chat_message (
    id SERIAL PRIMARY KEY,
    --ALTER TABLE chat_message ADD COLUMN uuid character varying(255) NOT NULL DEFAULT '';
    uuid character varying(255) NOT NULL,
    chat_session_uuid character varying(255) NOT NUll,
    role character varying(255) NOT NULL,
    content character varying NOT NULL,
    llm_summary character varying(1024) NOT NULL DEFAULT '',
    score double precision NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp DEFAULT now() Not NULL,
    created_by integer NOT NULL,
    updated_by integer NOT NULL,
    is_deleted BOOLEAN  NOT NULL DEFAULT false,
    is_pin BOOLEAN  NOT NULL DEFAULT false,
    token_count INTEGER DEFAULT 0 NOT NULL,
    raw jsonb default '{}' NOT NULL
);

-- chat_messages
ALTER TABLE chat_message ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN  NOT NULL DEFAULT false;
ALTER TABLE chat_message ADD COLUMN IF NOT EXISTS token_count INTEGER DEFAULT 0 NOT NULL;
ALTER TABLE chat_message ADD COLUMN IF NOT EXISTS is_pin BOOLEAN  NOT NULL DEFAULT false;
ALTER TABLE chat_message ADD COLUMN IF NOT EXISTS llm_summary character varying(1024) NOT NULL DEFAULT '';

-- add hash index on uuid
CREATE INDEX IF NOT EXISTS chat_message_uuid_idx ON chat_message using hash (uuid) ;

-- add index on chat_session_uuid
CREATE INDEX IF NOT EXISTS chat_message_chat_session_uuid_idx ON chat_message (chat_session_uuid);

-- add index on user_id
CREATE INDEX IF NOT EXISTS chat_message_user_id_idx ON chat_message (user_id);

-- add brin index on created_at
CREATE INDEX IF NOT EXISTS chat_message_created_at_idx ON chat_message using brin (created_at) ;

CREATE TABLE IF NOT EXISTS chat_prompt (
    id SERIAL PRIMARY KEY,
    uuid character varying(255) NOT NULL,
    chat_session_uuid character varying(255) NOT NULL, -- store the session_uuid
    role character varying(255) NOT NULL,
    content character varying NOT NULL,
    score double precision  default 0 NOT NULL,
    user_id integer default 0 NOT NULL,
    created_at timestamp  DEFAULT now() NOT NULL ,
    updated_at timestamp  DEFAULT now() NOT NULL,
    created_by integer NOT NULL,
    updated_by integer NOT NULL,
    is_deleted BOOLEAN  NOT NULL DEFAULT false,
    token_count INTEGER DEFAULT 0 NOT NULL
    -- raw jsonb default '{}' NOT NULL
);

ALTER TABLE chat_prompt ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN  NOT NULL DEFAULT false;
ALTER TABLE chat_prompt ADD COLUMN IF NOT EXISTS token_count INTEGER DEFAULT 0 NOT NULL;

-- add hash index on uuid
CREATE INDEX IF NOT EXISTS chat_prompt_uuid_idx ON chat_prompt using hash (uuid) ;

-- add index on chat_session_uuid
CREATE INDEX IF NOT EXISTS chat_prompt_chat_session_uuid_idx ON chat_prompt (chat_session_uuid);

-- add index on user_id
CREATE INDEX IF NOT EXISTS chat_prompt_user_id_idx ON chat_prompt (user_id);

CREATE TABLE IF NOT EXISTS chat_logs (
	id SERIAL PRIMARY KEY,  -- Auto-incrementing ID as primary key
	session JSONB default '{}' NOT NULL,         -- JSONB column to store chat session info
	question JSONB default '{}' NOT NULL,        -- JSONB column to store the question
	answer JSONB default '{}' NOT NULL,          -- JSONB column to store the answer 
    created_at timestamp  DEFAULT now() NOT NULL 
);

-- add brin index on created_at
CREATE INDEX IF NOT EXISTS chat_logs_created_at_idx ON chat_logs using brin (created_at) ;


-- user_id is the user who created the session
-- uuid is the session uuid
CREATE TABLE IF NOT EXISTS user_active_chat_session (
    id SERIAL PRIMARY KEY,
    user_id integer UNIQUE default '0' NOT NULL ,
    chat_session_uuid character varying(255) NOT NULL,
    created_at timestamp  DEFAULT now() NOT NULL,
    updated_at timestamp  DEFAULT now() NOT NULL
);

-- add index on user_id
CREATE INDEX IF NOT EXISTS user_active_chat_session_user_id_idx ON user_active_chat_session using hash (user_id) ;


-- for share chat feature
CREATE TABLE IF NOT EXISTS chat_snapshot (
    id SERIAL PRIMARY KEY,
    uuid VARCHAR(255) NOT NULL default '',
    user_id INTEGER NOT NULL default 0,
    title VARCHAR(255) NOT NULL default '',
    summary TEXT NOT NULL default '',
    model VARCHAR(255) NOT NULL default '',
    tags JSONB DEFAULT '{}' NOT NULL,
    session JSONB DEFAULT '{}' NOT NULL,
    conversation JSONB DEFAULT '{}' NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    text text DEFAULT '' NOT NULL,
    search_vector tsvector generated always as (setweight(to_tsvector('simple', coalesce(title, '')), 'A') || ' ' || setweight(to_tsvector('simple', coalesce(text, '')), 'B') :: tsvector) stored
);

ALTER TABLE chat_snapshot ADD COLUMN IF NOT EXISTS model VARCHAR(255) NOT NULL default '' ;
ALTER TABLE chat_snapshot ADD COLUMN IF NOT EXISTS session JSONB DEFAULT '{}' NOT NULL;
ALTER TABLE chat_snapshot ADD COLUMN IF NOT EXISTS text text DEFAULT '' NOT NULL;
ALTER TABLE chat_snapshot ADD COLUMN IF NOT EXISTS search_vector tsvector generated always as (
	setweight(to_tsvector('simple', coalesce(title, '')), 'A') || ' ' || setweight(to_tsvector('simple', coalesce(text, '')), 'B') :: tsvector
) stored; 

CREATE INDEX IF NOT EXISTS search_vector_gin_idx on chat_snapshot using GIN(search_vector);

-- add index on user id
CREATE INDEX IF NOT EXISTS chat_snapshot_user_id_idx ON chat_snapshot (user_id);

-- add index on created_at(brin)
CREATE INDEX IF NOT EXISTS chat_snapshot_created_at_idx ON chat_snapshot using brin (created_at) ;

UPDATE chat_snapshot SET model = 'gpt-3.5-turbo' WHERE model = '';


CREATE TABLE IF NOT EXISTS chat_file (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    user_id INTEGER NOT NULL default 1,
    -- foreign key chat_session_uuid
    chat_session_uuid VARCHAR(255) NOT NULL REFERENCES chat_session(uuid) ON DELETE CASCADE,
    mime_type VARCHAR(255) NOT NULL
);
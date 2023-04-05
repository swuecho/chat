
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

CREATE TABLE IF NOT EXISTS auth_user_management (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL REFERENCES auth_user(id) ON DELETE CASCADE,
    rate_limit INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP DEFAULT NOW() NOT NULL
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
    max_tokens int DEFAULT 512 NOT NULL
);

CREATE TABLE IF NOT EXISTS chat_message (
    id SERIAL PRIMARY KEY,
    --ALTER TABLE chat_message ADD COLUMN uuid character varying(255) NOT NULL DEFAULT '';
    uuid character varying(255) NOT NULL,
    chat_session_uuid character varying(255) NOT NUll,
    role character varying(255) NOT NULL,
    content character varying NOT NULL,
    score double precision NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,
    updated_at timestamp DEFAULT now() Not NULL,
    created_by integer NOT NULL,
    updated_by integer NOT NULL,
    raw jsonb default '{}' NOT NULL

);

-- alter table chat_message add column chat_session_uuid character varying(255) NOT NULL DEFAULT '';

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
    updated_by integer NOT NULL
    -- raw jsonb default '{}' NOT NULL
);


-- user_id is the user who created the session
-- uuid is the session uuid
CREATE TABLE IF NOT EXISTS user_active_chat_session (
    id SERIAL PRIMARY KEY,
    user_id integer UNIQUE default '0' NOT NULL ,
    chat_session_uuid character varying(255) NOT NULL,
    created_at timestamp  DEFAULT now() NOT NULL,
    updated_at timestamp  DEFAULT now() NOT NULL
);


-- ALTER TABLE user_active_chat_session
-- ADD CONSTRAINT unique_user_id
-- UNIQUE (user_id);

-- ALTER TABLE chat_session
-- ADD CONSTRAINT unique_uuid
-- UNIQUE (uuid);

-- ALTER TABLE chat_prompt RENAME COLUMN topic TO session_uuid;

ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS temperature float DEFAULT 1.0 NOT NULL;
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS top_p float DEFAULT 1.0 NOT NULL;
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS max_tokens int DEFAULT 512 NOT NULL; 
ALTER TABLE chat_session ADD COLUMN IF NOT EXISTS debug boolean DEFAULT false NOT NULL; 
ALTER TABlE chat_session ADD COLUMN IF NOT EXISTS  model character varying(255) NOT NULL DEFAULT 'gpt-3.5-turbo',
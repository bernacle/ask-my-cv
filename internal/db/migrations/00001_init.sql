-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE documents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_path text NOT NULL UNIQUE,
    title text NOT NULL,
    content_hash text NOT NULL,
    embedding_model text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE chunks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id uuid NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content text NOT NULL,
    embedding vector(1536) NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    source_title text NOT NULL,
    source_path text NOT NULL,
    chunk_index integer NOT NULL,
    content_hash text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (document_id, chunk_index, content_hash)
);

CREATE TABLE conversations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content text NOT NULL,
    citations jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE question_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid REFERENCES conversations(id) ON DELETE SET NULL,
    request_id text NOT NULL,
    question text NOT NULL,
    normalized_question text NOT NULL,
    retrieved_chunks integer NOT NULL,
    used_chunks integer NOT NULL,
    max_similarity double precision NOT NULL DEFAULT 0,
    answered boolean NOT NULL,
    ip_hash text,
    user_agent text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE token_usage (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id uuid REFERENCES conversations(id) ON DELETE SET NULL,
    request_id text NOT NULL,
    provider text NOT NULL,
    model text NOT NULL,
    operation text NOT NULL,
    input_tokens integer NOT NULL DEFAULT 0,
    output_tokens integer NOT NULL DEFAULT 0,
    total_tokens integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX documents_source_path_idx ON documents (source_path);
CREATE INDEX documents_created_at_idx ON documents (created_at);
CREATE INDEX chunks_document_id_idx ON chunks (document_id);
CREATE INDEX chunks_source_path_idx ON chunks (source_path);
CREATE INDEX chunks_created_at_idx ON chunks (created_at);
CREATE INDEX chunks_embedding_hnsw_idx ON chunks USING hnsw (embedding vector_cosine_ops);
CREATE INDEX conversations_created_at_idx ON conversations (created_at);
CREATE INDEX messages_conversation_id_created_at_idx ON messages (conversation_id, created_at);
CREATE INDEX question_events_created_at_idx ON question_events (created_at);
CREATE INDEX token_usage_conversation_id_created_at_idx ON token_usage (conversation_id, created_at);

-- +goose Down
DROP TABLE IF EXISTS token_usage;
DROP TABLE IF EXISTS question_events;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS chunks;
DROP TABLE IF EXISTS documents;
DROP EXTENSION IF EXISTS pgcrypto;
DROP EXTENSION IF EXISTS vector;

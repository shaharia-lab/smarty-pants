CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS datasources (
    uuid UUID PRIMARY KEY,
    name TEXT NOT NULL,
    source_type TEXT NOT NULL,
    settings JSONB NOT NULL,
    state JSONB NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS documents (
    uuid UUID PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL,
    url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    datasource_uuid UUID REFERENCES datasources(uuid) ON DELETE CASCADE
);

-- Create table for content parts
CREATE TABLE IF NOT EXISTS content_parts (
    id SERIAL PRIMARY KEY,
    document_uuid UUID NOT NULL REFERENCES documents(uuid) ON DELETE CASCADE,
    content TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS embedding_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive')),
    configuration JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS embeddings (
    id SERIAL PRIMARY KEY,
    content_part_id INTEGER NOT NULL REFERENCES content_parts(id) ON DELETE CASCADE,
    embedding vector NOT NULL,
    embedding_provider_id UUID NOT NULL REFERENCES embedding_providers(id) ON DELETE CASCADE,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    embedding_prompt_token integer NOT NULL
);

-- Create table for metadata
CREATE TABLE IF NOT EXISTS metadata (
    id SERIAL PRIMARY KEY,
    document_uuid UUID NOT NULL REFERENCES documents(uuid) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS llm_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive')),
    configuration JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE TABLE IF NOT EXISTS ai_ops_usage (
    ops_provider_id UUID,
    document_id UUID REFERENCES documents(uuid) ON DELETE CASCADE,
    input_tokens INT,
    output_tokens INT,
    dimensions INT,
    operation_type VARCHAR,
    cost_per_thousands_token FLOAT,
    created_at TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    total_latency FLOAT
);

CREATE TABLE IF NOT EXISTS app_settings (
    id SERIAL PRIMARY KEY,
    settings JSONB NOT NULL,
    last_updated_at TIMESTAMP NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

-- Create the interactions table
CREATE TABLE IF NOT EXISTS interactions (
    uuid UUID PRIMARY KEY,
    query TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

-- Create the conversations table
CREATE TABLE IF NOT EXISTS conversations (
    uuid UUID PRIMARY KEY,
    role TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    interaction_uuid UUID NOT NULL,
    FOREIGN KEY (interaction_uuid) REFERENCES interactions(uuid)
);

-- Indexes for efficient searching
CREATE INDEX IF NOT EXISTS idx_embedding_providers_name ON embedding_providers (name);
CREATE INDEX IF NOT EXISTS idx_llm_providers_name ON llm_providers (name);
CREATE INDEX IF NOT EXISTS idx_documents_title ON documents (title);
CREATE INDEX IF NOT EXISTS idx_documents_body ON documents (body);
CREATE INDEX IF NOT EXISTS idx_metadata_key_value ON metadata (key, value);

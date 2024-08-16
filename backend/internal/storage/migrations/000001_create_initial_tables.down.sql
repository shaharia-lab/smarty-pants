DROP INDEX IF EXISTS idx_metadata_key_value;
DROP INDEX IF EXISTS idx_documents_body;
DROP INDEX IF EXISTS idx_documents_title;
DROP INDEX IF EXISTS idx_llm_providers_name;
DROP INDEX IF EXISTS idx_embedding_providers_name;

DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS interactions;
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS ai_ops_usage;
DROP TABLE IF EXISTS llm_providers;
DROP TABLE IF EXISTS metadata;
DROP TABLE IF EXISTS embeddings;
DROP TABLE IF EXISTS embedding_providers;
DROP TABLE IF EXISTS content_parts;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS datasources;

DROP EXTENSION IF EXISTS vector;
-- 1. Embedding Providers Analytics
-- Purpose: This query retrieves information about embedding providers,
-- including the total count and details of active providers.
-- It helps understand the diversity and status of embedding services in use.

SELECT
    COUNT(*) AS total_embedding_providers,
    SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS active_embedding_providers,
    STRING_AGG(
        CASE WHEN status = 'active'
        THEN name || ' (' || provider || ')'
        ELSE NULL END,
        ', '
    ) AS active_providers_list
FROM embedding_providers;

-- 2. Daily Embedding Generation
-- Purpose: This query shows the number of embeddings generated per day.
-- It helps track the daily workload and usage patterns of the embedding feature.

SELECT
    DATE(generated_at) AS date,
    COUNT(*) AS embeddings_generated
FROM embeddings
GROUP BY DATE(generated_at)
ORDER BY date DESC;

-- 3. Total Prompt Tokens
-- Purpose: This query calculates the total prompt tokens used per day and the grand total.
-- It helps monitor token usage, which is often tied to costs in AI services.

SELECT
    DATE(generated_at) AS date,
    SUM(embedding_prompt_token) AS daily_prompt_tokens,
    SUM(SUM(embedding_prompt_token)) OVER () AS total_prompt_tokens
FROM embeddings
GROUP BY DATE(generated_at)
ORDER BY date DESC;

-- 4. LLM Providers Analytics
-- Purpose: This query retrieves information about LLM providers,
-- including the total count and details of active providers.
-- It helps understand the diversity and status of LLM services in use.

SELECT
    COUNT(*) AS total_llm_providers,
    SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS active_llm_providers,
    STRING_AGG(
        CASE WHEN status = 'active'
        THEN name || ' (' || provider || ')'
        ELSE NULL END,
        ', '
    ) AS active_providers_list
FROM llm_providers;

-- 5. Document Fetching Over Time
-- Purpose: This query shows the number of documents fetched over time.
-- It helps visualize the growth of the document database and identify any patterns or spikes in document acquisition.

SELECT
    DATE(fetched_at) AS date,
    COUNT(*) AS documents_fetched
FROM documents
GROUP BY DATE(fetched_at)
ORDER BY date DESC;

-- 6. Total Documents by Datasource
-- Purpose: This query counts the number of documents associated with each datasource.
-- It helps understand which datasources are contributing the most to the document database.

SELECT
    d.name AS datasource_name,
    COUNT(doc.uuid) AS document_count
FROM datasources d
LEFT JOIN documents doc ON d.uuid = doc.datasource_uuid
GROUP BY d.uuid, d.name
ORDER BY document_count DESC;

-- 7. Interactions and Conversations Over Time
-- Purpose: This query aggregates the number of interactions and associated conversations over time.
-- It helps track user engagement and the volume of AI interactions.

SELECT
    DATE(i.created_at) AS date,
    COUNT(DISTINCT i.uuid) AS interaction_count,
    COUNT(c.uuid) AS conversation_count
FROM interactions i
LEFT JOIN conversations c ON i.uuid = c.interaction_uuid
GROUP BY DATE(i.created_at)
ORDER BY date DESC;

-- 8. Average Tokens Per Embedding
-- Purpose: This query calculates the average number of tokens used per embedding for each provider.
-- It helps optimize token usage and compare efficiency across different providers.

SELECT
    ep.name AS provider_name,
    AVG(e.embedding_prompt_token) AS avg_tokens_per_embedding
FROM embeddings e
JOIN embedding_providers ep ON e.embedding_provider_id = ep.id
GROUP BY ep.id, ep.name
ORDER BY avg_tokens_per_embedding DESC;

-- 9. Document Processing Latency
-- Purpose: This query calculates the average time taken to process documents
-- (from creation to embedding generation) for each datasource.
-- It helps identify any bottlenecks in the document processing pipeline.

SELECT
    d.name AS datasource_name,
    AVG(EXTRACT(EPOCH FROM (e.generated_at - doc.created_at))) AS avg_processing_time_seconds
FROM documents doc
JOIN content_parts cp ON doc.uuid = cp.document_uuid
JOIN embeddings e ON cp.id = e.content_part_id
JOIN datasources d ON doc.datasource_uuid = d.uuid
GROUP BY d.uuid, d.name
ORDER BY avg_processing_time_seconds DESC;

-- 10. AI Operations Cost Analysis
-- Purpose: This query calculates the total cost of AI operations per day,
-- helping to track and manage expenses related to AI usage.

SELECT
    DATE(created_at) AS date,
    SUM((input_tokens + output_tokens) * cost_per_thousands_token / 1000) AS total_cost
FROM ai_ops_usage
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- 11. Most Active Documents
-- Purpose: This query identifies the most frequently accessed documents in interactions.
-- It helps understand which documents are most valuable or relevant to users.

WITH document_mentions AS (
    SELECT
        document_uuid,
        COUNT(*) AS mention_count
    FROM metadata
    WHERE key = 'referenced_document'
    GROUP BY document_uuid
)
SELECT
    d.title,
    d.url,
    dm.mention_count
FROM document_mentions dm
JOIN documents d ON dm.document_uuid = d.uuid
ORDER BY dm.mention_count DESC
LIMIT 10;

-- 12. System Performance Over Time
-- Purpose: This query tracks the average latency of AI operations over time.
-- It helps monitor system performance and identify any degradation or improvements.

SELECT
    DATE(created_at) AS date,
    AVG(total_latency) AS avg_latency
FROM ai_ops_usage
GROUP BY DATE(created_at)
ORDER BY date DESC;
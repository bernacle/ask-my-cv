-- name: InsertChunk :one
INSERT INTO chunks (
    document_id,
    content,
    embedding,
    metadata,
    source_title,
    source_path,
    chunk_index,
    content_hash
) VALUES (
    sqlc.arg(document_id),
    sqlc.arg(content),
    sqlc.arg(embedding)::vector,
    sqlc.arg(metadata),
    sqlc.arg(source_title),
    sqlc.arg(source_path),
    sqlc.arg(chunk_index),
    sqlc.arg(content_hash)
)
RETURNING id, document_id, content, metadata, source_title, source_path, chunk_index, content_hash, created_at;

-- name: SearchChunks :many
SELECT
    c.id,
    c.document_id,
    c.content,
    c.metadata,
    c.source_title,
    c.source_path,
    c.chunk_index,
    (c.embedding <=> sqlc.arg(embedding)::vector)::double precision AS distance,
    (1 - (c.embedding <=> sqlc.arg(embedding)::vector))::double precision AS similarity
FROM chunks c
JOIN documents d ON d.id = c.document_id
WHERE d.embedding_model = sqlc.arg(embedding_model)
ORDER BY c.embedding <=> sqlc.arg(embedding)::vector
LIMIT sqlc.arg(match_count);

-- name: SearchChunksByKeywords :many
SELECT
    c.id,
    c.document_id,
    c.content,
    c.metadata,
    c.source_title,
    c.source_path,
    c.chunk_index,
    0::double precision AS distance,
    1::double precision AS similarity
FROM chunks c
JOIN documents d ON d.id = c.document_id
WHERE d.embedding_model = sqlc.arg(embedding_model)
  AND EXISTS (
    SELECT 1
    FROM unnest(sqlc.arg(keywords)::text[]) AS keyword
    WHERE c.content ILIKE ('%' || keyword || '%')
  )
ORDER BY c.source_path, c.chunk_index
LIMIT sqlc.arg(match_count);

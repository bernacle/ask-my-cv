-- name: UpsertDocument :one
INSERT INTO documents (
    source_path,
    title,
    content_hash,
    embedding_model
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (source_path) DO UPDATE SET
    title = EXCLUDED.title,
    content_hash = EXCLUDED.content_hash,
    embedding_model = EXCLUDED.embedding_model,
    updated_at = now()
RETURNING id, source_path, title, content_hash, embedding_model, created_at, updated_at;

-- name: GetDocumentByPath :one
SELECT id, source_path, title, content_hash, embedding_model, created_at, updated_at
FROM documents
WHERE source_path = $1
LIMIT 1;

-- name: DeleteChunksByDocument :exec
DELETE FROM chunks
WHERE document_id = $1;

-- name: ListSources :many
SELECT
    d.id,
    d.source_path,
    d.title,
    d.content_hash,
    d.embedding_model,
    d.created_at,
    d.updated_at,
    count(c.id)::integer AS chunk_count
FROM documents d
LEFT JOIN chunks c ON c.document_id = d.id
GROUP BY d.id
ORDER BY d.title ASC;

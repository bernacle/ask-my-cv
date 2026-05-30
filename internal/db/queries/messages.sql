-- name: CreateMessage :one
INSERT INTO messages (
    conversation_id,
    role,
    content,
    citations
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, conversation_id, role, content, citations, created_at;

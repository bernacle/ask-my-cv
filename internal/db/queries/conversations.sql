-- name: CreateConversation :one
INSERT INTO conversations DEFAULT VALUES
RETURNING id, created_at, updated_at;

-- name: GetConversation :one
SELECT id, created_at, updated_at
FROM conversations
WHERE id = $1
LIMIT 1;

-- name: TouchConversation :exec
UPDATE conversations
SET updated_at = now()
WHERE id = $1;

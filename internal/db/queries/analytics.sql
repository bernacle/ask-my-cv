-- name: CreateQuestionEvent :one
INSERT INTO question_events (
    conversation_id,
    request_id,
    question,
    normalized_question,
    retrieved_chunks,
    used_chunks,
    max_similarity,
    answered,
    ip_hash,
    user_agent
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, conversation_id, request_id, question, normalized_question, retrieved_chunks, used_chunks, max_similarity, answered, ip_hash, user_agent, created_at;

-- name: CreateTokenUsage :one
INSERT INTO token_usage (
    conversation_id,
    request_id,
    provider,
    model,
    operation,
    input_tokens,
    output_tokens,
    total_tokens
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, conversation_id, request_id, provider, model, operation, input_tokens, output_tokens, total_tokens, created_at;

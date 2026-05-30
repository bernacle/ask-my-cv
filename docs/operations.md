# Operations

## Cost Controls

- Per-IP in-memory rate limits: `RATE_LIMIT_PER_MINUTE` and `RATE_LIMIT_PER_DAY`.
- Maximum question length: `MAX_QUESTION_CHARS`.
- Maximum retrieved chunks: `MAX_RETRIEVED_CHUNKS`.
- Maximum generated output: `MAX_OUTPUT_TOKENS`.
- Repeated exact normalized questions are cached in memory for `CACHE_TTL`.
- Embedding and chat token usage are written to `token_usage`.
- Tune `SIMILARITY_THRESHOLD` from `question_events.max_similarity`; for this small CV corpus, `0.60` is a practical starting value.

## Observability

The API emits structured JSON logs with request IDs. The RAG flow records:

- question events
- retrieved and used chunk counts
- max similarity
- answered vs unsupported responses
- token usage by operation and model

## Security Notes

- The indexed knowledge base should contain only public professional content.
- Retrieved Markdown is treated as untrusted quoted context in the system prompt.
- The assistant is instructed not to invent unsupported claims.
- CORS is allowlist-based.
- API keys are read from environment variables only.

# Ask My CV

Ask My CV is a small Go RAG backend for a portfolio chat widget. It indexes curated Markdown files, stores embeddings in PostgreSQL with pgvector, retrieves relevant chunks, and answers with citations.

## Stack

- Go + Chi
- PostgreSQL + pgvector
- sqlc + pgx
- goose migrations
- OpenAI embeddings and Responses API
- Railway deployment target

## Local Development

1. Copy `.env.example` to `.env` and fill `OPENAI_API_KEY`.
2. Start Postgres:

```bash
task up -- -d
```

3. Run migrations:

```bash
task db-migrate-up
```

4. Generate SQL code after query changes:

```bash
task sqlc
```

5. Ingest curated Markdown:

```bash
task ingest
```

6. Start the API:

```bash
task serve
```

Useful shortcuts:

```bash
task test
task build
task dev
task health
task ask QUESTION="What AI/RAG experience does Bruno have?"
```

## API

- `GET /health`
- `GET /api/v1/sources`
- `POST /api/v1/ask`

```json
{
  "question": "What backend engineering experience does Bruno have?",
  "conversation_id": "optional-uuid"
}
```

Responses include an answer, citations, source metadata, a conversation ID, and token usage when available.

## Deployment

Deploy one Railway service for the Go API and one Railway PostgreSQL database with pgvector enabled. Configure the environment variables in `.env.example`, run goose migrations during deploy or as a one-off command, then run ingestion whenever curated Markdown changes.

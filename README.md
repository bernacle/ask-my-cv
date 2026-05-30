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
docker compose -f deployment/docker-compose.yml up -d
```

3. Run migrations:

```bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

4. Generate SQL code after query changes:

```bash
sqlc generate
```

5. Ingest curated Markdown:

```bash
go run main.go ingest ./data
```

6. Start the API:

```bash
go run main.go serve-api
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

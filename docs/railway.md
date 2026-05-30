# Railway Deployment

Deploy two Railway services:

- Go API service from this repository.
- Railway PostgreSQL service with pgvector support.

## Environment Variables

Set the API service variables from `.env.example`:

```bash
APP_ENV=production
HTTP_ADDR=:8080
DATABASE_URL=${{Postgres.DATABASE_URL}}
OPENAI_API_KEY=...
OPENAI_BASE_URL=https://api.openai.com
OPENAI_EMBEDDING_MODEL=text-embedding-3-small
OPENAI_CHAT_MODEL=gpt-5.4-mini
OPENAI_REASONING_EFFORT=low
OPENAI_TEXT_VERBOSITY=low
CORS_ALLOWED_ORIGINS=https://brunosimplicio.com
MAX_QUESTION_CHARS=1000
MAX_RETRIEVED_CHUNKS=6
SIMILARITY_THRESHOLD=0.72
MAX_OUTPUT_TOKENS=500
RATE_LIMIT_PER_MINUTE=10
RATE_LIMIT_PER_DAY=100
CACHE_TTL=24h
```

## Migrations

Run migrations as a one-off command:

```bash
goose -dir internal/db/migrations postgres "$DATABASE_URL" up
```

## Ingestion

After curated Markdown changes, run:

```bash
go run main.go ingest ./data
```

The API command is:

```bash
go run main.go serve-api
```

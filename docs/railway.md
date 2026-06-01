# Railway Deployment

Deploy two Railway services:

- Go API service from this repository.
- Railway PostgreSQL service with pgvector support.

## Environment Variables

Set the API service variables from `.env.example`:

```bash
APP_ENV=production
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
SIMILARITY_THRESHOLD=0.60
MAX_OUTPUT_TOKENS=500
RATE_LIMIT_PER_MINUTE=10
RATE_LIMIT_PER_DAY=100
CACHE_TTL=24h
```

Do not set `HTTP_ADDR` on Railway unless you need a fixed port. The app reads Railway's injected `PORT` variable and listens on that port automatically.

## Migrations

Migrations and ingestion run automatically before deploy through `railway.toml`:

```bash
ask-my-cv init /app/data
```

That command is built into the production binary. It applies migrations, then ingests the Markdown files copied into the Docker image at `/app/data`. The ingest step is idempotent, so unchanged documents are skipped.

## Ingestion

After curated Markdown changes, push and redeploy the API service. The pre-deploy initialization command reingests changed files automatically.

You can also run ingestion manually from a Railway container if needed:

```bash
railway ssh --service ask-my-cv -- /usr/local/bin/ask-my-cv ingest /app/data
```

The API command is:

```bash
ask-my-cv serve-api
```

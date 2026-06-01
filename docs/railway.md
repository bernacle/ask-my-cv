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

Migrations run automatically before deploy through `railway.toml`:

```bash
ask-my-cv migrate-up
```

That command is built into the production binary, so the deployed image does not need the external `goose` CLI.

## Ingestion

After curated Markdown changes, run:

```bash
railway run task ingest
```

This runs ingestion from your local checkout using Railway's production variables, so it writes embeddings into the Railway Postgres database.

The API command is:

```bash
ask-my-cv serve-api
```

DATABASE_URL ?= postgres://ask_my_cv:ask_my_cv@localhost:5432/ask_my_cv?sslmode=disable

.PHONY: test build sqlc migrate-up migrate-down ingest serve

test:
	go test ./...

build:
	go build ./...

sqlc:
	sqlc generate

migrate-up:
	goose -dir internal/db/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir internal/db/migrations postgres "$(DATABASE_URL)" down

ingest:
	go run main.go ingest ./data

serve:
	go run main.go serve-api

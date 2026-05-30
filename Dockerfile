FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /ask-my-cv .

FROM alpine:3.22
RUN adduser -D -H appuser
WORKDIR /app
COPY --from=build /ask-my-cv /usr/local/bin/ask-my-cv
COPY internal/db/migrations ./internal/db/migrations
COPY data ./data
USER appuser
EXPOSE 8080
CMD ["ask-my-cv", "serve-api"]

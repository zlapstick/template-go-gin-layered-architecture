FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -o migration ./cmd/migration && \
    CGO_ENABLED=0 GOOS=linux go build -o seed ./cmd/seed

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/main .
COPY --from=builder /app/migration .
COPY --from=builder /app/seed .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations
COPY .env.example .
COPY docs ./docs

ENV PORT=8084

EXPOSE 8084
CMD ["sh", "-c", "./migration migrate && ./seed seed && exec ./main"]
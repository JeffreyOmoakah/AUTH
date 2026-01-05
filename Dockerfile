# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

# Build from the ./cmd directory where your main.go and api.go live
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-api ./cmd

# Stage 2: Runtime
FROM alpine:3.20

RUN apk add --no-cache ca-certificates libc6-compat

RUN adduser -D appuser
WORKDIR /app

# 1. Copy the application binary
COPY --from=builder /auth-api ./auth-api

# 2. Copy the goose binary
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# 3. Copy migration files
COPY --from=builder /app/internal/adapters/postgresql/migrations ./migrations

RUN chown -R appuser:appuser /app && chmod +x ./auth-api
USER appuser

EXPOSE 8080

# Execute the binary in the current directory
CMD ["./auth-api"]
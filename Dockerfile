# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Install build dependencies 
RUN apk add --no-cache git

WORKDIR /app

# Docker cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Install goose binary in the build stage
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-api ./cmd/

# Stage 2: Runtime
FROM alpine:3.20

# Install necessary libraries for SSL/Postgres
RUN apk add --no-cache ca-certificates libc6-compat

# Add non-root user
RUN adduser -D appuser
WORKDIR /app

# 1. Copy the application binary
COPY --from=builder /auth-api .

# 2. Copy the goose binary from the Go path in Stage 1
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# 3. Copy your migration files specifically
# We put them in a flat folder at /app/migrations for simplicity
COPY --from=builder /app/internal/adapters/postgresql/migrations ./migrations

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

CMD ["./auth-api"]
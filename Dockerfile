# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Install goose into /go/bin/goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .
# Build the binary from the cmd folder (containing main.go and api.go)
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-api ./cmd/api

# Stage 2: Runtime
FROM alpine:3.20
RUN apk add --no-cache ca-certificates libc6-compat
WORKDIR /app

# 1. Copy the application binary
COPY --from=builder /auth-api ./auth-api

# 2. Copy the goose binary to a SYSTEM PATH so it works anywhere
COPY --from=builder /go/bin/goose /usr/local/bin/goose

# 3. Copy migrations to a predictable flat path
COPY --from=builder /app/internal/adapters/postgresql/migrations ./internal/adapters/postgresql/migrations

# Permissions
RUN chmod +x ./auth-api /usr/local/bin/goose

EXPOSE 8080
CMD ["./auth-api"]
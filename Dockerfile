# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Install build dependencies 
RUN apk add --no-cache git

WORKDIR /app

#  Docker cache for dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary. 
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-api ./cmd/main.go

# Stage 2: Runtime
FROM alpine:3.20

# Add non-root user for security
RUN adduser -D appuser
USER appuser

WORKDIR /app

# Copy binary and migrations (needed if running migrations at startup)
COPY --from=builder /auth-api .
COPY --from=builder /app/internal/adapters/postgresql/migrations ./migrations

# Railway provides the PORT environment variable
EXPOSE 8080

CMD ["./auth-api"]
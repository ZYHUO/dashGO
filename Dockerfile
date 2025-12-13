# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies (gcc, musl-dev for CGO/SQLite support)
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled for SQLite support
RUN CGO_ENABLED=1 GOOS=linux go build -a -o dashgo ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create necessary directories
RUN mkdir -p /app/data /app/configs /app/web/dist

# Copy binary
COPY --from=builder /app/dashgo .

# Copy web frontend if exists
COPY --from=builder /app/web/dist ./web/dist 2>/dev/null || true

# Expose port
EXPOSE 8080

# Run
CMD ["./dashgo", "-config", "/app/configs/config.yaml"]

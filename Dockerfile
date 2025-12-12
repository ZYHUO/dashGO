# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dashgo ./cmd/server

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
CMD ["./xboard", "-config", "/app/configs/config.yaml"]

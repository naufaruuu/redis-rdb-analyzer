# Multi-stage build for Redis RDB Analyzer
# Stage 1: Build the Go application
FROM golang:1.18-alpine AS builder

# Install build dependencies (gcc, musl-dev for CGO, sqlite)
RUN apk add --no-cache \
    gcc \
    musl-dev \
    sqlite-dev \
    git

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled (required for sqlite3)
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'v1.0') \
              -X main.BuildTime=$(date -u '+%Y-%m-%d_%H:%M:%S') \
              -X main.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
    -o redis-rdb-analyzer \
    .

# Stage 2: Runtime image with kubectl
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    curl \
    bash

# Install kubectl
ARG KUBECTL_VERSION=v1.35.0
RUN curl -LO "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/

# Create non-root user
RUN addgroup -g 1000 rdr && \
    adduser -D -u 1000 -G rdr rdr

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/redis-rdb-analyzer /app/redis-rdb-analyzer

# Copy views directory (HTML templates)
COPY --from=builder /build/views /app/views

# Create directories for runtime data
RUN mkdir -p /app/tmp /app/data && \
    chown -R rdr:rdr /app

# Switch to non-root user
USER rdr

# Expose port
EXPOSE 8080


# Run the application (port can be set via RDR_PORT env var)
CMD ["/app/redis-rdb-analyzer"]

# ---------- Build Stage ----------
FROM golang:1.25.1-bookworm AS builder

# Install build dependencies (needed for CGO and SQLite)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled (required for whatsmeow SQLite store)
ENV CGO_ENABLED=1

RUN go build -o main -ldflags="-s -w" .

# ---------- Run Stage ----------
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    tzdata \
    wget \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create app user for security
RUN groupadd -g 1000 appuser && \
    useradd -r -u 1000 -g appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Create necessary directories
RUN mkdir -p /app/stores /app/uploads/whatsapp /app/keys && \
    chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (configured via env)
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:${PORT:-3000}/health || exit 1

# Run the application
CMD ["./main"]

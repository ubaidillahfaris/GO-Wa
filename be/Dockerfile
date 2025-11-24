# ---------- Build Stage ----------
FROM golang:1.25.1-bookworm AS builder

# Install build dependencies (needed for CGO and SQLite)
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*


WORKDIR /app

# Copy go modules files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations
ENV CGO_ENABLED=1

RUN go build -ldflags="-s -w" -o main .

# ---------- Run Stage ----------
FROM debian:bookworm-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    tzdata \
    curl \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set timezone
ENV TZ=Asia/Jakarta

# Create app user for security
RUN groupadd -g 1000 appuser && \
    useradd -r -u 1000 -g appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Create necessary directories with proper permissions
RUN mkdir -p /app/stores /app/uploads/whatsapp /app/keys /app/logs && \
    chown -R appuser:appuser /app && \
    chmod -R 755 /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Run the application
CMD ["./main"]

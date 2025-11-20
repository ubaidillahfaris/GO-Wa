# ---------- Build Stage ----------
FROM golang:1.25.1-alpine AS builder

# Install build dependencies (needed for CGO and SQLite)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go modules files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled (required for whatsmeow SQLite store)
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -o main -ldflags="-s -w" .

# ---------- Run Stage ----------
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite-libs \
    tzdata

# Create app user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

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
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-3000}/health || exit 1

# Run the application
CMD ["./main"]

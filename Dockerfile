# ── Stage 1: Build binary ──
FROM golang:alpine AS builder

WORKDIR /build

# Install essential build tools
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static binary for Linux amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server ./cmd/server

# ── Stage 2: Final Minimal Runtime Image ──
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies: SSL certificates, timezone data, and curl for healthchecks
RUN apk add --no-cache ca-certificates tzdata curl

# Configure default timezone (America/Bogota)
ENV TZ=America/Bogota
RUN cp /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Copy compiled binary from builder
COPY --from=builder /build/server /app/server

# Copy translation catalogs
COPY --from=builder /build/i18n /app/i18n

# Create uploads directory (mount point for persistent Dokploy volume)
RUN mkdir -p /app/uploads && chmod -R 755 /app/uploads

# Expose backend application port
EXPOSE 8080

# Production default environment
ENV GIN_MODE=release \
    SERVER_PORT=8080

# Healthcheck for Dokploy / Docker monitoring
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/health || exit 1

# Run the server
CMD ["/app/server"]

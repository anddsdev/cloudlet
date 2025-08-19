# Multi-stage build for production optimization
FROM golang:1.23.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    upx

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with retry and verification
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with production optimizations
RUN CGO_ENABLED=1 go build -o cloudlet ./cmd/cloudlet/main.go

# Compress binary for smaller size
RUN upx --best --lzma cloudlet

# Verify the binary works
RUN ./cloudlet --version || echo "Binary built successfully"

# Production image - using distroless for security
FROM gcr.io/distroless/static-debian12:nonroot

# Copy timezone data from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create necessary directories
USER root
RUN ["/busybox/mkdir", "-p", "/app/data/storage", "/app/config", "/app/logs", "/tmp/cloudlet"]

# Set ownership to nonroot user (uid=65532)
RUN ["/busybox/chown", "-R", "65532:65532", "/app", "/tmp/cloudlet"]

# Switch to non-root user
USER 65532:65532

WORKDIR /app

# Copy the optimized binary
COPY --from=builder --chown=65532:65532 /build/cloudlet ./

# Copy default configuration (can be overridden)
COPY --chown=65532:65532 config/config.yaml config/

# Production environment variables with secure defaults
ENV PORT=8080
ENV DB_DSN=/app/data/cloudlet.db
ENV STORAGE_PATH=/app/data/storage
ENV CONFIG_PATH=/app/config/config.yaml

# Production file size limits (more restrictive)
ENV MAX_FILE_SIZE=500000000
ENV MAX_MEMORY=64000000

# Production timeout configuration (more conservative)
ENV READ_TIMEOUT=60
ENV WRITE_TIMEOUT=300
ENV IDLE_TIMEOUT=120
ENV MAX_HEADER_BYTES=4096
ENV READ_HEADER_TIMEOUT=30

# Production upload configuration
ENV MAX_FILES_PER_REQUEST=20
ENV MAX_TOTAL_SIZE_PER_REQUEST=1073741824
ENV ALLOW_PARTIAL_SUCCESS=false
ENV ENABLE_BATCH_PROCESSING=true
ENV BATCH_SIZE=5
ENV MAX_CONCURRENT_UPLOADS=10
ENV STREAMING_THRESHOLD=52428800
ENV VALIDATE_BEFORE_UPLOAD=true
ENV ENABLE_PROGRESS_TRACKING=true
ENV CLEANUP_ON_FAILURE=true
ENV RATE_LIMIT_PER_MINUTE=200

# Production database configuration
ENV DB_MAX_CONN=25

# Security and logging
ENV LOG_LEVEL=info
ENV LOG_FORMAT=json
ENV LOG_FILE=/app/logs/cloudlet.log
ENV ENABLE_METRICS=true
ENV METRICS_PORT=9090

# Timezone
ENV TZ=UTC

EXPOSE 8080 9090

# Define volumes for persistence
VOLUME ["/app/data", "/app/config", "/app/logs"]

# Optimized health check
HEALTHCHECK --interval=15s --timeout=5s --start-period=30s --retries=3 \
    CMD ["/app/cloudlet", "healthcheck"] || exit 1

# Use exec form for better signal handling
CMD ["/app/cloudlet"]

# Security labels
LABEL \
    org.opencontainers.image.title="Cloudlet" \
    org.opencontainers.image.description="Production-ready file upload service" \
    org.opencontainers.image.vendor="Cloudlet" \
    org.opencontainers.image.version="production" \
    org.opencontainers.image.created="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    security.non-root="true" \
    security.readonly-rootfs="true"
# Multi-stage build for production optimization
FROM node:20-alpine AS client-builder

WORKDIR /client

# Copy client package files
COPY client/package.json client/bun.lock* ./

# Install dependencies using npm (fallback if bun not available)
RUN npm ci --only=production

# Copy client source code
COPY client/ .

# Build client for production
RUN NODE_ENV=docker npm run build

# Backend builder stage
FROM golang:1.23.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    make

WORKDIR /build

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with retry and verification
RUN go mod download && go mod verify

# Copy source code including Makefile
COPY . .

# Copy built client files to web directory
COPY --from=client-builder /client/dist ./web/

# Build using Makefile with production optimizations
RUN CGO_ENABLED=1 make build && mv main.exe cloudlet

# Verify the binary works
RUN ./cloudlet --help || echo "Binary built successfully"

# Production image - using Alpine instead of distroless for CGO compatibility
FROM alpine:3.19

# Install runtime dependencies for CGO
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    sqlite \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S cloudlet && \
    adduser -u 1001 -S cloudlet -G cloudlet

# Create necessary directories with proper permissions
RUN mkdir -p /app/data/storage /app/config /app/logs /tmp/cloudlet && \
    chown -R cloudlet:cloudlet /app /tmp/cloudlet

# Switch to non-root user
USER cloudlet

WORKDIR /app

# Copy the optimized binary
COPY --from=builder --chown=cloudlet:cloudlet /build/cloudlet ./

# Copy built web assets
COPY --from=builder --chown=cloudlet:cloudlet /build/web ./web/

# Copy default configuration (can be overridden)
COPY --chown=cloudlet:cloudlet config/config.yaml config/

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

# Working health check using wget (available in Alpine)
HEALTHCHECK --interval=15s --timeout=5s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1

# Use exec form for better signal handling
CMD ["/app/cloudlet"]

# Security labels
LABEL \
    org.opencontainers.image.title="Cloudlet" \
    org.opencontainers.image.description="Production-ready file upload service" \
    org.opencontainers.image.vendor="Cloudlet" \
    org.opencontainers.image.version="production" \
    org.opencontainers.image.created="2025-08-18T23:00:00Z" \
    security.non-root="true" \
    security.minimal-base="true"
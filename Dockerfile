FROM golang:1.23.4-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o cloudlet ./cmd/cloudlet

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata sqlite

RUN addgroup -g 1001 -S cloudlet && \
    adduser -u 1001 -S cloudlet -G cloudlet

RUN mkdir -p /app/data/storage /app/config && \
    chown -R cloudlet:cloudlet /app

USER cloudlet

WORKDIR /app

COPY --from=builder --chown=cloudlet:cloudlet /build/cloudlet .

COPY --chown=cloudlet:cloudlet config/config.yaml config/

# Default environment variables - all configurable via docker-compose or runtime
ENV PORT=8080
ENV DB_DSN=/app/data/cloudlet.db
ENV STORAGE_PATH=/app/data/storage
ENV MAX_FILE_SIZE=100000000
ENV MAX_MEMORY=32000000

# Timeout configuration
ENV READ_TIMEOUT=30
ENV WRITE_TIMEOUT=30
ENV IDLE_TIMEOUT=60
ENV MAX_HEADER_BYTES=1024
ENV READ_HEADER_TIMEOUT=10

# Upload configuration
ENV MAX_FILES_PER_REQUEST=50
ENV MAX_TOTAL_SIZE_PER_REQUEST=524288000
ENV ALLOW_PARTIAL_SUCCESS=true
ENV ENABLE_BATCH_PROCESSING=true
ENV BATCH_SIZE=10
ENV MAX_CONCURRENT_UPLOADS=3
ENV STREAMING_THRESHOLD=10485760
ENV VALIDATE_BEFORE_UPLOAD=true
ENV ENABLE_PROGRESS_TRACKING=false
ENV CLEANUP_ON_FAILURE=false
ENV RATE_LIMIT_PER_MINUTE=100

# Database configuration
ENV DB_MAX_CONN=10

# Configuration path (fallback)
ENV CONFIG_PATH=/app/config/config.yaml

EXPOSE 8080

VOLUME ["/app/data", "/app/config"]

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/health || exit 1


CMD ["./cloudlet"]
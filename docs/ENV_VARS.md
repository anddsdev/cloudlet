# Environment Variables Configuration

This document describes all available environment variables for configuring Cloudlet.

## Configuration Priority

1. **Environment Variables** (Primary - highest priority)
2. **YAML Configuration File** (Fallback)
3. **Default Values** (Last resort)

## Server Configuration

### Basic Server Settings

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `PORT` | string | `"8080"` | Server port |
| `MAX_MEMORY` | int | `32000000` | Maximum memory for uploads (32MB) |
| `MAX_FILE_SIZE` | int64 | `100000000` | Maximum file size (100MB) |
| `STORAGE_PATH` | string | `"./data/storage"` | File storage directory path |

### Timeout Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `READ_TIMEOUT` | int | `30` | Read timeout in seconds |
| `WRITE_TIMEOUT` | int | `30` | Write timeout in seconds |
| `IDLE_TIMEOUT` | int | `60` | Idle timeout in seconds |
| `MAX_HEADER_BYTES` | int | `1024` | Maximum header bytes |
| `READ_HEADER_TIMEOUT` | int | `10` | Read header timeout in seconds |

## Upload Configuration

### File Upload Limits

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `MAX_FILES_PER_REQUEST` | int | `50` | Maximum files per upload request |
| `MAX_TOTAL_SIZE_PER_REQUEST` | int64 | `524288000` | Maximum total size per request (500MB) |
| `STREAMING_THRESHOLD` | int64 | `10485760` | File size threshold for streaming (10MB) |
| `MAX_CONCURRENT_UPLOADS` | int | `3` | Maximum concurrent uploads |
| `RATE_LIMIT_PER_MINUTE` | int | `100` | Rate limit per minute |

### Upload Behavior

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `ALLOW_PARTIAL_SUCCESS` | bool | `true` | Allow partial success in batch uploads |
| `ENABLE_BATCH_PROCESSING` | bool | `true` | Enable batch processing |
| `BATCH_SIZE` | int | `10` | Batch processing size |
| `VALIDATE_BEFORE_UPLOAD` | bool | `true` | Validate files before upload |
| `ENABLE_PROGRESS_TRACKING` | bool | `false` | Enable upload progress tracking |
| `CLEANUP_ON_FAILURE` | bool | `false` | Clean up files on upload failure |

## Database Configuration

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `DB_DSN` | string | `"./data/cloudlet.db"` | Database connection string |
| `DB_MAX_CONN` | int | `10` | Maximum database connections |

## Boolean Value Formats

Boolean environment variables accept multiple formats:

**True values:** `true`, `1`, `yes`, `on`, `enabled` (case insensitive)
**False values:** `false`, `0`, `no`, `off`, `disabled` (case insensitive)

## Usage Examples

### Basic Development Setup
```bash
export PORT=3000
export MAX_FILE_SIZE=50000000
export STORAGE_PATH=/tmp/cloudlet-storage
./cloudlet
```

### Production Setup
```bash
export PORT=80
export MAX_FILE_SIZE=1000000000
export MAX_CONCURRENT_UPLOADS=10
export RATE_LIMIT_PER_MINUTE=500
export VALIDATE_BEFORE_UPLOAD=true
export CLEANUP_ON_FAILURE=true
export DB_MAX_CONN=25
./cloudlet
```

### Docker Environment File (.env)
```bash
# Server Configuration
PORT=8080
MAX_MEMORY=32000000
MAX_FILE_SIZE=100000000
STORAGE_PATH=/app/data/storage

# Upload Configuration
MAX_FILES_PER_REQUEST=50
MAX_TOTAL_SIZE_PER_REQUEST=524288000
STREAMING_THRESHOLD=10485760
MAX_CONCURRENT_UPLOADS=3
RATE_LIMIT_PER_MINUTE=100

# Boolean Settings
ALLOW_PARTIAL_SUCCESS=true
ENABLE_BATCH_PROCESSING=true
VALIDATE_BEFORE_UPLOAD=true
ENABLE_PROGRESS_TRACKING=false
CLEANUP_ON_FAILURE=false

# Database
DB_DSN=/app/data/cloudlet.db
DB_MAX_CONN=10
```

### High Performance Setup
```bash
export PORT=8080
export MAX_FILE_SIZE=2147483648      # 2GB
export MAX_MEMORY=134217728          # 128MB
export MAX_FILES_PER_REQUEST=100
export MAX_TOTAL_SIZE_PER_REQUEST=5368709120  # 5GB
export ENABLE_BATCH_PROCESSING=true
export BATCH_SIZE=20
export MAX_CONCURRENT_UPLOADS=15
export STREAMING_THRESHOLD=52428800  # 50MB
export RATE_LIMIT_PER_MINUTE=1000
export DB_MAX_CONN=50
```

## Configuration Validation

The application will:
1. Load environment variables first (if any are set)
2. Fall back to YAML configuration file if no environment variables are detected
3. Use built-in defaults if both environment variables and YAML file are unavailable

## Testing Configuration

You can test your configuration by checking the startup logs:

```bash
PORT=9090 MAX_FILE_SIZE=200000000 ./cloudlet
```

Expected output:
```
Configuration loaded successfully
Server will start on port: 9090
Storage path: ./data/storage
Database DSN: ./data/cloudlet.db
starting server on port 9090
```

## Environment Variable Detection

The system detects if environment variables are being used by checking for the presence of any configuration environment variable. If at least one is found, the system will use environment variable mode exclusively.

If you want to use a mix of environment variables and YAML, you should set all required environment variables explicitly.
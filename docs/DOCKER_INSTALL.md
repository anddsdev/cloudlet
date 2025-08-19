# Cloudlet Docker Installation

## 📋 Prerequisites

- Docker Engine 20.10+
- Docker Compose v2.0+
- Minimum 512MB RAM
- Available port (default 8080)

## 🚀 Quick Installation

### 1. Clone or Download the Project

```bash
git clone <repository-url> cloudlet
cd cloudlet
```

### 2. Configure Environment Variables

```bash
# Copy configuration template
cp .env.example .env

# Edit configuration (optional)
nano .env
```

### 3. Run the Service

```bash
# Build and run in background
docker-compose up -d

# View logs
docker-compose logs -f cloudlet
```

### 4. Verify Installation

```bash
# Health check
curl http://localhost:8080/health

# Upload test
curl -X POST -F "file=@test.txt" http://localhost:8080/upload
```

## ⚙️ Custom Configuration

### Main Environment Variables

#### **Server Configuration**

```bash
CLOUDLET_PORT=8080                    # Service port
MAX_FILE_SIZE=100000000              # Maximum file size (100MB)
MAX_MEMORY=32000000                  # Maximum memory for uploads (32MB)
```

#### **Upload Configuration**

```bash
MAX_FILES_PER_REQUEST=50             # Maximum files per request
MAX_TOTAL_SIZE_PER_REQUEST=524288000 # Maximum total size per request (500MB)
STREAMING_THRESHOLD=10485760         # Streaming threshold (10MB)
MAX_CONCURRENT_UPLOADS=3             # Maximum concurrent uploads
RATE_LIMIT_PER_MINUTE=100           # Request limit per minute
```

#### **Timeout Configuration**

```bash
READ_TIMEOUT=30                      # Read timeout (seconds)
WRITE_TIMEOUT=30                     # Write timeout (seconds)
IDLE_TIMEOUT=60                      # Idle timeout (seconds)
```

#### **Container Resources**

```bash
CONTAINER_MEMORY_LIMIT=512M          # Memory limit
CONTAINER_CPU_LIMIT=1.0              # CPU limit
```

## 🛠️ Common Use Cases

### Development Server

```bash
# .env for development
CLOUDLET_PORT=3000
MAX_FILE_SIZE=50000000              # 50MB
ENABLE_PROGRESS_TRACKING=true
CLEANUP_ON_FAILURE=true
```

### Production Server

```bash
# .env for production
CLOUDLET_PORT=80
MAX_FILE_SIZE=1000000000            # 1GB
MAX_CONCURRENT_UPLOADS=10
RATE_LIMIT_PER_MINUTE=500
CONTAINER_MEMORY_LIMIT=2G
CONTAINER_CPU_LIMIT=2.0
```

### High Volume Server

```bash
# .env for high volume
MAX_FILES_PER_REQUEST=100
MAX_TOTAL_SIZE_PER_REQUEST=2147483648  # 2GB
ENABLE_BATCH_PROCESSING=true
BATCH_SIZE=20
DB_MAX_CONN=25
```

## 🔧 Management Commands

### Service Management

```bash
# Start service
docker-compose up -d

# Stop service
docker-compose down

# Restart service
docker-compose restart

# View logs in real-time
docker-compose logs -f

# Check service status
docker-compose ps
```

### Data Management

```bash
# Backup data
docker-compose exec cloudlet tar -czf /tmp/backup.tar.gz /app/data
docker cp $(docker-compose ps -q cloudlet):/tmp/backup.tar.gz ./backup.tar.gz

# Restore data
docker cp ./backup.tar.gz $(docker-compose ps -q cloudlet):/tmp/
docker-compose exec cloudlet tar -xzf /tmp/backup.tar.gz -C /
```

### Updates

```bash
# Update image
docker-compose pull
docker-compose up -d

# Rebuild from source
docker-compose build --no-cache
docker-compose up -d
```

## 🐳 Installation with Docker Run

### Basic Command

```bash
docker run -d \
  --name cloudlet \
  -p 8080:8080 \
  -v cloudlet_data:/app/data \
  -e CLOUDLET_PORT=8080 \
  -e MAX_FILE_SIZE=100000000 \
  cloudlet:latest
```

### Complete Command with Configuration

```bash
docker run -d \
  --name cloudlet \
  -p 8080:8080 \
  -v cloudlet_data:/app/data \
  -v cloudlet_config:/app/config \
  -e CLOUDLET_PORT=8080 \
  -e MAX_FILE_SIZE=100000000 \
  -e MAX_MEMORY=32000000 \
  -e MAX_FILES_PER_REQUEST=50 \
  -e RATE_LIMIT_PER_MINUTE=100 \
  --restart unless-stopped \
  cloudlet:latest
```

## 🔒 Security Configuration

### Security Environment Variables

```bash
# Limit features for enhanced security
VALIDATE_BEFORE_UPLOAD=true
CLEANUP_ON_FAILURE=true
RATE_LIMIT_PER_MINUTE=50
MAX_FILES_PER_REQUEST=10
```

## 🚨 Troubleshooting

### Common Issues

```bash
# View detailed logs
docker-compose logs --tail=100 cloudlet

# Verify configuration
docker-compose config

# Check resources
docker stats cloudlet

# Access container
docker-compose exec cloudlet sh
```

### Health Verification

```bash
# Health endpoint
curl -f http://localhost:8080/health || echo "Service unhealthy"

# Verify with timeout
timeout 5 curl http://localhost:8080/health
```

## 📞 Support

For specific issues:

1. Review container logs
2. Verify environment variable configuration
3. Check port availability
4. Validate system resources

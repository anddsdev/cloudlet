# ☁️ Cloudlet

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-AGPL-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/anddsdev/cloudlet)

A lightweight, self-hosted file storage service designed to be simple, fast, and easy to use. Cloudlet provides a minimalist alternative to complex file management systems like Nextcloud, focusing on essential file operations without unnecessary overhead.

## ✨ Features

- 🌐 **Modern Web Interface**: Beautiful, responsive React-based dashboard with drag & drop support
- 📁 **File Management**: Upload, download, and organize files through both web UI and API
- 📂 **Directory Operations**: Create, navigate, and manage folder structures
- 🔄 **File Operations**: Move, rename, and delete files and directories
  - Smart deletion with recursive directory support
  - Confirmation dialogs for destructive operations
  - Atomic transactions for data consistency
- 📊 **Statistics**: View file information, directory sizes, and item counts
- 🗂️ **Multiple File Types**: Support for various file formats with MIME type detection
- ⚡ **High Performance**: Built with Go for speed and efficiency
- 🛡️ **Advanced Security**: 
  - Path validation with directory traversal protection
  - SQL injection prevention with SafeQueryBuilder
  - Dangerous file detection and validation
  - Enhanced input validation and sanitization
- 🚀 **Multiple Upload Options**:
  - Traditional single file uploads
  - Multiple file upload with batch processing
  - Streaming uploads for large files
  - Chunked uploads with progress tracking
  - Concurrent and sequential processing strategies
- 🔒 **Atomic Operations**: Thread-safe file operations with data integrity guarantees
- 🐳 **Easy Deployment**: Simple configuration and deployment
- 💾 **SQLite Database**: Lightweight database with optimized queries
- 🌐 **REST API**: Complete RESTful API for integration
- 🧪 **Comprehensive Testing**: Extensive test coverage with benchmarks

## 🚀 Quick Start

### Prerequisites

- Go 1.23.4 or later
- SQLite3 (included)

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/anddsdev/cloudlet.git
   cd cloudlet
   ```

2. **Build the application**

   ```bash
   go mod download
   go build -o cloudlet ./cmd/cloudlet
   ```

3. **Run Cloudlet**
   ```bash
   ./cloudlet
   ```

The server will start on `http://localhost:8080` by default and will serve the web interface at the root path.

### Frontend Development

To run the React frontend in development mode:

1. **Navigate to the client directory**
   ```bash
   cd client
   ```

2. **Install dependencies**
   ```bash
   bun install
   ```

3. **Start the development server**
   ```bash
   bun run dev
   ```

The frontend will start on `http://localhost:5173` and will proxy API requests to the backend on port 8080.

4. **Build for production**

   Using Make (builds both client and backend):
   ```bash
   make build-prod
   ```

   Or build individually:
   ```bash
   make build-client  # Builds React client to web/
   make build         # Builds Go backend
   ```

   Or using npm/bun directly (client only):
   ```bash
   cd client && bun run build
   ```

The built client files are output to `web/` directory to be served by the Go backend.

### Using Docker

Build and run with Docker:
```bash
# Build Docker image (includes both client and backend)
make docker-build
# or
docker build -t cloudlet:latest .

# Run container
make docker-run
# or
docker run -d \
  --name cloudlet \
  -p 8080:8080 \
  -v ./data:/app/data \
  cloudlet:latest
```

## ⚙️ Configuration

Cloudlet uses a YAML configuration file located at `config/config.yaml`:

```yaml
server:
  port: 8080
  max_memory: 32000000 # 32MB
  max_file_size: 100000000 # 100MB
  storage:
    path: ./data/storage
  timeout:
    read_timeout: 30
    write_timeout: 30
    idle_timeout: 60
    max_header_bytes: 1024
    read_header_timeout: 10
  
  upload:
    max_files_per_request: 50
    max_total_size_per_request: 524288000 # 500MB in bytes
    allow_partial_success: true        
    enable_batch_processing: true      
    batch_size: 10          
    max_concurrent_uploads: 3
    streaming_threshold: 10485760 # 10MB in bytes
    validate_before_upload: true
    enable_progress_tracking: false
    cleanup_on_failure: false    
    rate_limit_per_minute: 100

database:
  driver: sqlite3
  dsn: ./data/cloudlet.db
  max_conn: 10
```

### Configuration Options

| Option                                      | Description                                | Default              |
| ------------------------------------------- | ------------------------------------------ | -------------------- |
| `server.port`                               | HTTP server port                           | `8080`               |
| `server.max_memory`                         | Maximum memory for file uploads            | `32MB`               |
| `server.max_file_size`                      | Maximum file size allowed                  | `100MB`              |
| `server.storage.path`                       | Physical storage directory                 | `./data/storage`     |
| `server.timeout.read_timeout`               | HTTP read timeout (seconds)                | `30`                 |
| `server.timeout.write_timeout`              | HTTP write timeout (seconds)               | `30`                 |
| `server.timeout.idle_timeout`               | HTTP idle timeout (seconds)                | `60`                 |
| `server.upload.max_files_per_request`       | Maximum files per upload request           | `50`                 |
| `server.upload.max_total_size_per_request`  | Maximum total size per request             | `500MB`              |
| `server.upload.allow_partial_success`       | Allow partial upload success               | `true`               |
| `server.upload.enable_batch_processing`     | Enable batch processing                    | `true`               |
| `server.upload.batch_size`                  | Batch size for processing                  | `10`                 |
| `server.upload.max_concurrent_uploads`      | Maximum concurrent uploads                 | `3`                  |
| `server.upload.streaming_threshold`         | File size threshold for streaming          | `10MB`               |
| `server.upload.validate_before_upload`      | Validate files before upload               | `true`               |
| `server.upload.enable_progress_tracking`    | Enable upload progress tracking            | `false`              |
| `server.upload.cleanup_on_failure`          | Clean up files on upload failure          | `false`              |
| `server.upload.rate_limit_per_minute`       | Upload rate limit per minute               | `100`                |
| `database.dsn`                              | SQLite database file path                  | `./data/cloudlet.db` |
| `database.max_conn`                         | Maximum database connections               | `10`                 |

## 📖 API Documentation

### Endpoints

#### Files

| Method   | Endpoint                        | Description                      |
| -------- | ------------------------------- | -------------------------------- |
| `GET`    | `/api/v1/files`                 | List files in root directory     |
| `GET`    | `/api/v1/files/{path}`          | List files in specific directory |
| `POST`   | `/api/v1/upload`                | Upload single file               |
| `POST`   | `/api/v1/upload/multiple`       | Upload multiple files            |
| `POST`   | `/api/v1/upload/batch`          | Batch upload with validation     |
| `POST`   | `/api/v1/upload/stream`         | Streaming upload for large files |
| `POST`   | `/api/v1/upload/chunked`        | Chunked upload with progress     |
| `POST`   | `/api/v1/upload/progress`       | Upload with progress tracking    |
| `GET`    | `/api/v1/download/{path}`       | Download a file                  |
| `DELETE` | `/api/v1/files/{path}`          | Delete a file or directory       |

#### Directories

| Method | Endpoint                     | Description             |
| ------ | ---------------------------- | ----------------------- |
| `POST` | `/api/v1/directories`        | Create a new directory  |
| `GET`  | `/api/v1/directories/{path}` | List directory contents |

#### Operations

| Method | Endpoint         | Description                 |
| ------ | ---------------- | --------------------------- |
| `POST` | `/api/v1/move`   | Move files or directories   |
| `POST` | `/api/v1/rename` | Rename files or directories |

#### Health

| Method | Endpoint  | Description           |
| ------ | --------- | --------------------- |
| `GET`  | `/health` | Health check endpoint |

### Request/Response Examples

#### Upload a single file

```bash
curl -X POST \
  -F "file=@example.txt" \
  -F "path=/" \
  http://localhost:8080/api/v1/upload
```

#### Upload multiple files

```bash
curl -X POST \
  -F "files=@file1.txt" \
  -F "files=@file2.txt" \
  -F "path=/" \
  http://localhost:8080/api/v1/upload/multiple
```

#### Streaming upload for large files

```bash
curl -X POST \
  -F "file=@largefile.zip" \
  -F "path=/" \
  http://localhost:8080/api/v1/upload/stream
```

#### Create a directory

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name": "documents", "parent_path": "/"}' \
  http://localhost:8080/api/v1/directories
```

#### List directory contents

```bash
curl http://localhost:8080/api/v1/files/documents
```

Response:

```json
{
  "path": "/documents",
  "parent_path": "/",
  "files": [...],
  "directories": [...],
  "total_files": 5,
  "total_directories": 2,
  "total_size": 1048576,
  "breadcrumbs": [...]
}
```

#### Delete files and directories

Delete a single file:
```bash
curl -X DELETE http://localhost:8080/api/v1/files/document.txt
```

Delete an empty directory:
```bash
curl -X DELETE http://localhost:8080/api/v1/files/empty-folder
```

Delete a directory and all its contents recursively:
```bash
curl -X DELETE "http://localhost:8080/api/v1/files/folder-with-content?recursive=true"
```

Response for successful deletion:
```json
{
  "success": true,
  "message": "File deleted successfully",
  "path": "/folder-with-content"
}
```

Response for non-empty directory without recursive parameter:
```json
{
  "error": true,
  "message": "Directory not empty",
  "status": 409
}
```

## 🏗️ Architecture

Cloudlet follows clean architecture principles with clear separation of concerns:

```
cloudlet/
├── cmd/cloudlet/           # Application entry point
├── config/                 # Configuration management
├── internal/
│   ├── database/          # Database utilities and SafeQueryBuilder
│   ├── handlers/          # HTTP handlers (including specialized upload handlers)
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   ├── security/          # Security components (PathValidator)
│   ├── server/            # HTTP server and routing
│   ├── services/          # Business logic
│   ├── storage/           # AtomicFileOperations
│   ├── transaction/       # TransactionManager
│   └── utils/             # Utility functions and validators
├── client/                # React frontend application
│   ├── src/
│   │   ├── components/    # React components (Dashboard, FileList, etc.)
│   │   ├── services/      # API communication layer
│   │   └── lib/           # Utilities and helpers
├── web/                   # Built frontend assets (served by Go)
├── data/                  # SQLite database
└── storage/               # File storage directory
```

### Key Components

- **Handlers**: Process HTTP requests and responses with specialized upload handlers
- **Services**: Contain business logic and orchestrate operations
  - **FileService**: Core file management with atomic operations
  - **MultipleUploadService**: Handles batch and concurrent uploads
  - **StorageService**: Physical file operations with atomic guarantees
- **Repository**: Handle database operations with SQLite and SafeQueryBuilder
- **Security**: 
  - **PathValidator**: Prevents directory traversal attacks
  - **SafeQueryBuilder**: SQL injection prevention
  - **Validator**: Enhanced input validation and sanitization
- **Storage**: **AtomicFileOperations**: Thread-safe file operations
- **Transaction**: **TransactionManager**: Ensures data consistency
- **Models**: Define data structures and API contracts

#### Frontend Architecture

- **React Components**: Modern functional components with hooks
  - **Dashboard**: Main interface with file management
  - **FileList**: Table view with file operations
  - **FileUploadZone**: Drag & drop upload with progress
  - **Modals**: Confirmation dialogs for destructive operations
- **Services Layer**: API communication with strategy pattern
  - **uploadStrategy**: Smart upload method selection
  - **fileService**: Core file operations API calls
- **UI Components**: shadcn/ui for consistent design system
- **State Management**: React hooks with real-time notifications

## 🔐 Security & Performance

### Security Features

- **Path Validation**: Comprehensive protection against directory traversal attacks
- **SQL Injection Prevention**: SafeQueryBuilder ensures all database queries are secure
- **File Validation**: Detection and prevention of dangerous file uploads
- **Input Sanitization**: Enhanced validation across all endpoints
- **Atomic Operations**: Thread-safe file operations prevent race conditions

### Performance Optimizations

- **Streaming Uploads**: Handle large files efficiently without memory overhead
- **Chunked Processing**: Break large operations into manageable chunks
- **Concurrent Operations**: Parallel processing for multiple file operations
- **Batch Processing**: Optimize multiple file uploads with intelligent batching
- **Transaction Management**: Ensure database consistency with proper rollback capabilities

### Testing & Quality

- **Comprehensive Test Suite**: Over 5,800 lines of test coverage
- **Benchmark Tests**: Performance validation for critical operations
- **Mock Services**: Robust testing infrastructure
- **Edge Case Validation**: Security testing against various attack vectors

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

1. **Fork the repository**
2. **Clone your fork**
   ```bash
   git clone https://github.com/yourusername/cloudlet.git
   ```
3. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
4. **Make your changes**
5. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
6. **Push to your branch**
   ```bash
   git push origin feature/amazing-feature
   ```
7. **Open a Pull Request**

### Code Style

**Backend (Go):**
- Follow Go conventions and use `gofmt`
- Write clear, descriptive commit messages
- Add tests for new features
- Update documentation as needed

**Frontend (React/TypeScript):**
- Use TypeScript for type safety
- Follow React hooks patterns
- Use shadcn/ui components for consistency
- Implement proper error handling with user feedback
- Write responsive and accessible interfaces

## 📋 Roadmap

- [x] **Web UI interface** - Modern React dashboard with drag & drop uploads
  - [x] File and directory management
  - [x] Smart upload strategies (single, multiple, batch, stream)
  - [x] Recursive directory deletion with confirmation
  - [x] Real-time progress tracking and notifications
- [ ] User authentication and authorization
- [ ] File sharing with expirable links
- [ ] File versioning
- [ ] Bulk operations (move, copy multiple files)
- [ ] File search functionality
- [x] Docker container support
- [ ] Cloud storage backends (S3, etc.)
- [ ] File encryption
- [ ] API rate limiting

## 🐛 Issues and Support

If you encounter any issues or have questions:

1. Check the [Issues](https://github.com/anddsdev/cloudlet/issues) page
2. Search for existing issues before creating a new one
3. Provide detailed information about your environment and the problem
4. Include steps to reproduce the issue

## 📄 License

This project is licensed under the AGPL-3.0 License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

**Backend:**
- Built with [Go](https://golang.org/)
- Database powered by [SQLite](https://sqlite.org/)
- Configuration with [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

**Frontend:**
- [React](https://reactjs.org/) - UI framework
- [TypeScript](https://www.typescriptlang.org/) - Type safety
- [Tailwind CSS](https://tailwindcss.com/) - Styling framework
- [shadcn/ui](https://ui.shadcn.com/) - Component library
- [Vite](https://vitejs.dev/) - Build tool and dev server
- [Lucide React](https://lucide.dev/) - Icon library
- [React Dropzone](https://react-dropzone.js.org/) - Drag & drop uploads
- [Sonner](https://sonner.emilkowal.ski/) - Toast notifications

## 📬 Contact

- GitHub: [@anddsdev](https://github.com/anddsdev)
- Project Link: [https://github.com/anddsdev/cloudlet](https://github.com/anddsdev/cloudlet)

---

⭐ If you find Cloudlet useful, please consider giving it a star on GitHub!

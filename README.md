# ☁️ Cloudlet

[![Go Version](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/anddsdev/cloudlet)

A lightweight, self-hosted file storage service designed to be simple, fast, and easy to use. Cloudlet provides a minimalist alternative to complex file management systems like Nextcloud, focusing on essential file operations without unnecessary overhead.

## ✨ Features

- 📁 **File Management**: Upload, download, and organize files
- 📂 **Directory Operations**: Create, navigate, and manage folder structures
- 🔄 **File Operations**: Move, rename, and delete files and directories
- 📊 **Statistics**: View file information, directory sizes, and item counts
- 🗂️ **Multiple File Types**: Support for various file formats with MIME type detection
- ⚡ **High Performance**: Built with Go for speed and efficiency
- 🛡️ **Security**: Input validation and secure file handling
- 🐳 **Easy Deployment**: Simple configuration and deployment
- 💾 **SQLite Database**: Lightweight database with optimized queries
- 🌐 **REST API**: Complete RESTful API for integration

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

The server will start on `http://localhost:8080` by default.

### Using Docker (Coming Soon)

```bash
docker run -d \
  --name cloudlet \
  -p 8080:8080 \
  -v ./storage:/app/storage \
  -v ./data:/app/data \
  cloudlet:latest
```

## ⚙️ Configuration

Cloudlet uses a YAML configuration file located at `config/config.yaml`:

```yaml
server:
  port: 8080
  max_memory: 32000000      # 32MB
  max_file_size: 100000000  # 100MB
  storage:
    path: ./storage
  timeout:
    read_timeout: 30
    write_timeout: 30
    idle_timeout: 60
    max_header_bytes: 1024
    read_header_timeout: 10

database:
  driver: sqlite3
  dsn: ./data/cloudlet.db
  max_conn: 10
```

### Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `server.port` | HTTP server port | `8080` |
| `server.max_memory` | Maximum memory for file uploads | `32MB` |
| `server.max_file_size` | Maximum file size allowed | `100MB` |
| `server.storage.path` | Physical storage directory | `./storage` |
| `database.dsn` | SQLite database file path | `./data/cloudlet.db` |
| `database.max_conn` | Maximum database connections | `10` |

## 📖 API Documentation

### Endpoints

#### Files

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/files` | List files in root directory |
| `GET` | `/api/v1/files/{path}` | List files in specific directory |
| `POST` | `/api/v1/upload` | Upload files |
| `GET` | `/api/v1/download/{path}` | Download a file |
| `DELETE` | `/api/v1/files` | Delete a file |

#### Directories

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/directories` | Create a new directory |
| `GET` | `/api/v1/directories/{path}` | List directory contents |

#### Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/move` | Move files or directories |
| `POST` | `/api/v1/rename` | Rename files or directories |

#### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check endpoint |

### Request/Response Examples

#### Upload a file
```bash
curl -X POST \
  -F "file=@example.txt" \
  -F "path=/" \
  http://localhost:8080/api/v1/upload
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

## 🏗️ Architecture

Cloudlet follows clean architecture principles with clear separation of concerns:

```
cloudlet/
├── cmd/cloudlet/           # Application entry point
├── config/                 # Configuration management
├── internal/
│   ├── handlers/          # HTTP handlers
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   ├── server/            # HTTP server and routing
│   ├── services/          # Business logic
│   └── utils/             # Utility functions
├── data/                  # SQLite database
└── storage/               # File storage directory
```

### Key Components

- **Handlers**: Process HTTP requests and responses
- **Services**: Contain business logic and orchestrate operations
- **Repository**: Handle database operations with SQLite
- **Storage Service**: Manage physical file operations
- **Models**: Define data structures and API contracts

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
5. **Run tests**
   ```bash
   go test ./...
   ```
6. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
7. **Push to your branch**
   ```bash
   git push origin feature/amazing-feature
   ```
8. **Open a Pull Request**

### Code Style

- Follow Go conventions and use `gofmt`
- Write clear, descriptive commit messages
- Add tests for new features
- Update documentation as needed

## 📋 Roadmap

- [ ] Web UI interface
- [ ] User authentication and authorization
- [ ] File sharing with expirable links
- [ ] File versioning
- [ ] Bulk operations
- [ ] File search functionality
- [ ] Docker container support
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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Database powered by [SQLite](https://sqlite.org/)
- Configuration with [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

## 📬 Contact

- GitHub: [@anddsdev](https://github.com/anddsdev)
- Project Link: [https://github.com/anddsdev/cloudlet](https://github.com/anddsdev/cloudlet)

---

⭐ If you find Cloudlet useful, please consider giving it a star on GitHub!

#!/bin/bash

# Cloudlet Docker Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/anddsdev/cloudlet/main/scripts/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
DEFAULT_PORT=8080
DEFAULT_DIR="cloudlet"
DEFAULT_BRANCH="main"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking prerequisites..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    # Check Git
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed. Please install Git first."
        exit 1
    fi
    
    log_success "All prerequisites are installed"
}

prompt_configuration() {
    echo
    log_info "Installation configuration:"
    
    # Installation directory
    read -p "Installation directory (${DEFAULT_DIR}): " INSTALL_DIR
    INSTALL_DIR=${INSTALL_DIR:-$DEFAULT_DIR}
    
    # Port configuration
    read -p "Service port (${DEFAULT_PORT}): " PORT
    PORT=${PORT:-$DEFAULT_PORT}
    
    # Environment selection
    echo
    log_info "Select installation type:"
    echo "1) Development (basic configuration)"
    echo "2) Production (optimized configuration)"
    echo "3) High volume (maximum performance)"
    echo "4) Custom (configure manually)"
    
    read -p "Option (1-4): " ENV_TYPE
    ENV_TYPE=${ENV_TYPE:-1}
}

download_project() {
    log_info "Downloading Cloudlet..."
    
    if [ -d "$INSTALL_DIR" ]; then
        log_warning "Directory $INSTALL_DIR already exists"
        read -p "Do you want to continue and overwrite? (y/N): " CONFIRM
        if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
            log_error "Installation cancelled"
            exit 1
        fi
        rm -rf "$INSTALL_DIR"
    fi
    
    # Clone repository (replace with actual repository URL)
    if [[ -n "${REPO_URL:-}" ]]; then
        git clone "$REPO_URL" "$INSTALL_DIR"
    else
        log_warning "Using local files for installation"
        mkdir -p "$INSTALL_DIR"
        # Copy current directory files (for local development)
        cp -r . "$INSTALL_DIR/" 2>/dev/null || true
    fi
    
    cd "$INSTALL_DIR"
    log_success "Project downloaded to $INSTALL_DIR"
}

configure_environment() {
    log_info "Configuring environment variables..."
    
    # Create .env file based on selection
    case $ENV_TYPE in
        1) # Development
            cat > .env << EOF
CLOUDLET_PORT=$PORT
MAX_FILE_SIZE=50000000
MAX_MEMORY=32000000
ENABLE_PROGRESS_TRACKING=true
CLEANUP_ON_FAILURE=true
RATE_LIMIT_PER_MINUTE=100
CONTAINER_MEMORY_LIMIT=512M
CONTAINER_CPU_LIMIT=1.0
EOF
            log_success "Development configuration applied"
            ;;
            
        2) # Production
            cat > .env << EOF
CLOUDLET_PORT=$PORT
MAX_FILE_SIZE=1000000000
MAX_MEMORY=64000000
MAX_CONCURRENT_UPLOADS=10
RATE_LIMIT_PER_MINUTE=500
VALIDATE_BEFORE_UPLOAD=true
CLEANUP_ON_FAILURE=true
CONTAINER_MEMORY_LIMIT=2G
CONTAINER_CPU_LIMIT=2.0
DB_MAX_CONN=25
EOF
            log_success "Production configuration applied"
            ;;
            
        3) # High volume
            cat > .env << EOF
CLOUDLET_PORT=$PORT
MAX_FILE_SIZE=2000000000
MAX_MEMORY=128000000
MAX_FILES_PER_REQUEST=100
MAX_TOTAL_SIZE_PER_REQUEST=5368709120
ENABLE_BATCH_PROCESSING=true
BATCH_SIZE=20
MAX_CONCURRENT_UPLOADS=15
RATE_LIMIT_PER_MINUTE=1000
CONTAINER_MEMORY_LIMIT=4G
CONTAINER_CPU_LIMIT=4.0
DB_MAX_CONN=50
EOF
            log_success "High volume configuration applied"
            ;;
            
        4) # Custom
            cp .env.example .env
            log_info ".env file created. Please edit it according to your needs:"
            echo "  nano .env"
            read -p "Press Enter when you have finished editing..."
            ;;
    esac
}

install_service() {
    log_info "Installing and starting the service..."
    
    # Build and start services
    docker-compose build
    docker-compose up -d
    
    # Wait for service to be ready
    log_info "Waiting for service to be ready..."
    sleep 10
    
    # Health check
    for i in {1..30}; do
        if curl -sf "http://localhost:$PORT/health" > /dev/null 2>&1; then
            log_success "Service started successfully on port $PORT"
            return 0
        fi
        sleep 2
    done
    
    log_error "Service could not start correctly"
    log_info "Checking logs..."
    docker-compose logs --tail=20 cloudlet
    return 1
}

show_completion_info() {
    echo
    log_success "🎉 Installation completed!"
    echo
    echo "📋 Service information:"
    echo "  • URL: http://localhost:$PORT"
    echo "  • Health: http://localhost:$PORT/health"
    echo "  • Directory: $(pwd)"
    echo
    echo "🛠️  Useful commands:"
    echo "  • View logs: docker-compose logs -f"
    echo "  • Stop service: docker-compose down"
    echo "  • Restart: docker-compose restart"
    echo "  • Status: docker-compose ps"
    echo
    echo "📖 Full documentation at: ./DOCKER_INSTALL.md"
    
    # Test upload
    echo
    read -p "Would you like to test a file upload? (y/N): " TEST_UPLOAD
    if [[ $TEST_UPLOAD =~ ^[Yy]$ ]]; then
        echo "Creating test file..."
        echo "Cloudlet test file" > test_upload.txt
        
        if curl -X POST -F "file=@test_upload.txt" "http://localhost:$PORT/upload" > /dev/null 2>&1; then
            log_success "✅ Upload test successful"
        else
            log_warning "❌ Upload test failed"
        fi
        
        rm -f test_upload.txt
    fi
}

cleanup_on_error() {
    log_error "Error during installation"
    if [ -d "$INSTALL_DIR" ]; then
        cd "$INSTALL_DIR"
        docker-compose down 2>/dev/null || true
    fi
    exit 1
}

# Main installation process
main() {
    echo "🐳 Cloudlet Docker Installer"
    echo "=============================="
    
    # Set trap for errors
    trap cleanup_on_error ERR
    
    check_requirements
    prompt_configuration
    download_project
    configure_environment
    install_service
    show_completion_info
}

# Run if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
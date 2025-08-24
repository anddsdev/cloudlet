# Simple Makefile for a Go project

# Build the application (backend + client)
all: build-client build test

# Build only the Go backend
build:
	@echo "Building backend..."
	@go build -o main.exe cmd/cloudlet/main.go

# Build the client
build-client:
	@echo "Building client..."
	@if [ -d "client" ]; then \
		mkdir -p web; \
		if command -v bun >/dev/null 2>&1; then \
			echo "Using bun to build client..."; \
			cd client && bun run build; \
		else \
			echo "Bun not found, using npm to build client..."; \
			cd client && npm run build; \
		fi; \
		echo "Client built successfully to web/"; \
	else \
		echo "Client directory not found, skipping client build"; \
	fi

# Build everything for production
build-prod: build-client build
	@echo "Production build complete"

# Run the application
run:
	@go run cmd/cloudlet/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@powershell -ExecutionPolicy Bypass -Command "if (Get-Command air -ErrorAction SilentlyContinue) { \
		air; \
		Write-Output 'Watching...'; \
	} else { \
		Write-Output 'Installing air...'; \
		go install github.com/air-verse/air@latest; \
		air; \
		Write-Output 'Watching...'; \
	}"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t cloudlet:latest .

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 -v $(PWD)/data:/app/data cloudlet:latest

# Development commands
dev-client:
	@echo "Starting client development server..."
	@cd client && npm run dev

dev-backend:
	@echo "Starting backend development server..."
	@go run cmd/cloudlet/main.go

# Clean up build artifacts
clean-all: clean
	@echo "Cleaning client build..."
	@rm -rf client/dist web/*

.PHONY: all build build-client build-prod run test clean clean-all watch docker-build docker-run dev-client dev-backend

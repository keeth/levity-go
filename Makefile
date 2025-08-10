# Levity - OCPP Charge Point Management System
# Makefile for build automation and development workflow

.PHONY: help build test clean deps lint run migrate docker-build docker-run

# Variables
BINARY_NAME=levity
BUILD_DIR=build
MAIN_PATH=./cmd/levity
CONFIG_FILE=config/config.yaml

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	go test -v ./...
	@echo "Tests complete"

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean
	@echo "Clean complete"

deps: ## Install/update dependencies
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

lint: ## Run code quality checks
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting complete"

format: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting complete"

run: ## Run the application
	@echo "Starting $(BINARY_NAME)..."
	go run $(MAIN_PATH)

run-build: build ## Build and run the application
	@echo "Starting $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

migrate: ## Run database migrations
	@echo "Running database migrations..."
	go run $(MAIN_PATH) migrate

dev: ## Start development mode with hot reload
	@echo "Starting development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		go run $(MAIN_PATH); \
	fi

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME) .
	@echo "Docker image built: $(BINARY_NAME)"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME)

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# Development setup
setup: deps install-tools ## Setup development environment
	@echo "Development environment setup complete"

# CI/CD targets
ci: deps lint test build ## Run CI pipeline
	@echo "CI pipeline complete"

# Release targets
release: clean test build ## Prepare release build
	@echo "Release build ready in $(BUILD_DIR)/"

# Database targets
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "WARNING: This will delete all database data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		rm -f levity.db; \
		echo "Database reset complete"; \
	else \
		echo "Database reset cancelled"; \
	fi

# Monitoring targets
monitor: ## Start monitoring (if enabled)
	@echo "Starting monitoring..."
	@if [ -f "$(CONFIG_FILE)" ]; then \
		go run $(MAIN_PATH) monitor; \
	else \
		echo "Config file not found. Run 'make config' first."; \
	fi

# Configuration
config: ## Create default configuration file
	@echo "Creating default configuration..."
	@mkdir -p config
	@if [ ! -f "$(CONFIG_FILE)" ]; then \
		echo "server:" > $(CONFIG_FILE); \
		echo "  address: \":8080\"" >> $(CONFIG_FILE); \
		echo "  read_timeout: \"30s\"" >> $(CONFIG_FILE); \
		echo "  write_timeout: \"30s\"" >> $(CONFIG_FILE); \
		echo "  max_header_bytes: 1048576" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "database:" >> $(CONFIG_FILE); \
		echo "  path: \"./levity.db\"" >> $(CONFIG_FILE); \
		echo "  max_open_conns: 25" >> $(CONFIG_FILE); \
		echo "  max_idle_conns: 5" >> $(CONFIG_FILE); \
		echo "  conn_max_lifetime: \"5m\"" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "ocpp:" >> $(CONFIG_FILE); \
		echo "  heartbeat_interval: \"60s\"" >> $(CONFIG_FILE); \
		echo "  max_message_size: 1048576" >> $(CONFIG_FILE); \
		echo "  connection_timeout: \"30s\"" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "log:" >> $(CONFIG_FILE); \
		echo "  level: \"info\"" >> $(CONFIG_FILE); \
		echo "  format: \"json\"" >> $(CONFIG_FILE); \
		echo "" >> $(CONFIG_FILE); \
		echo "monitoring:" >> $(CONFIG_FILE); \
		echo "  enabled: true" >> $(CONFIG_FILE); \
		echo "  address: \":9090\"" >> $(CONFIG_FILE); \
		echo "Configuration file created: $(CONFIG_FILE)"; \
	else \
		echo "Configuration file already exists: $(CONFIG_FILE)"; \
	fi

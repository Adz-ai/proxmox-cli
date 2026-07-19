# Makefile for Proxmox CLI

# Variables
BINARY_NAME=proxmox-cli
BUILD_DIR=build
MAIN_PATH=main.go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

# Build the binary
.PHONY: build
build: ## Build the binary to build/ directory
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Run all tests
.PHONY: test
test: ## Run all tests (unit + BDD)
	@echo "Running tests..."
	go test -v ./...
	@echo "All tests completed"

# Run BDD tests only
.PHONY: test-bdd
test-bdd: ## Run BDD/Gherkin tests
	@echo "Running BDD tests..."
	go test -v ./test/bdd
	@echo "BDD tests completed"

# Run unit tests only (excluding BDD tests)
.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -v -short ./cmd/...
	@echo "Unit tests completed"

# Run tests with coverage (unit tests only for meaningful coverage)
.PHONY: test-coverage
test-coverage: ## Run unit tests with coverage report
	@echo "Running unit tests with coverage..."
	go test -v -coverprofile=coverage.out -short ./cmd/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install binary to system
.PHONY: install
install: build ## Install binary to system PATH
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(BINARY_NAME) installed to /usr/local/bin/"

# Format code
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

.PHONY: generate
generate: ## Regenerate GoMock implementations
	go generate ./internal/interfaces

# Run go vet
.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet completed"

# Run linter (if available)
.PHONY: lint
lint: ## Run golangci-lint (install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "Linting completed"; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Download dependencies
.PHONY: deps
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies updated"

# Development setup
.PHONY: dev-setup
dev-setup: deps ## Set up development environment
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@echo "Development environment ready"

# Run the application
.PHONY: run
run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Build for multiple platforms
.PHONY: build-all
build-all: ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Multi-platform builds completed"

# Check everything
.PHONY: check
check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)
	@echo "All checks passed"

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

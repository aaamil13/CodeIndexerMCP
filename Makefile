.PHONY: all build test clean install run-index run-mcp fmt lint

# Build variables
BINARY_NAME=code-indexer
BUILD_DIR=bin
MAIN_PATH=cmd/code-indexer/main.go

# Go variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)

all: clean build

build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(GOBIN)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(GOBIN)/$(BINARY_NAME)"

build-all:
	@echo "🔨 Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(GOBIN)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(GOBIN)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(GOBIN)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(GOBIN)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Multi-platform build complete"

test:
	@echo "🧪 Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage:
	@echo "📊 Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

bench:
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@find . -name ".projectIndex" -type d -exec rm -rf {} + 2>/dev/null || true
	@echo "✅ Cleaned"

install: build
	@echo "📦 Installing $(BINARY_NAME)..."
	@cp $(GOBIN)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)"

run-index:
	@echo "🚀 Running indexer on current directory..."
	@go run $(MAIN_PATH) index .

run-mcp:
	@echo "🚀 Starting MCP server..."
	@go run $(MAIN_PATH) mcp .

fmt:
	@echo "🎨 Formatting code..."
	@go fmt ./...
	@echo "✅ Formatted"

lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, skipping..."; \
	fi

deps:
	@echo "📥 Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Dependencies downloaded"

help:
	@echo "Code Indexer MCP - Makefile Commands"
	@echo ""
	@echo "  make build        - Build the binary"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make bench        - Run benchmarks"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make install      - Install binary to GOPATH/bin"
	@echo "  make run-index    - Run indexer on current directory"
	@echo "  make run-mcp      - Start MCP server"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Lint code"
	@echo "  make deps         - Download dependencies"
	@echo "  make help         - Show this help"

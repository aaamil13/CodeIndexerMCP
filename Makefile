.PHONY: setup
setup:
	@echo "Downloading Tree-sitter Go bindings..."
	go mod tidy
	@echo "Tree-sitter setup complete!"

.PHONY: test-sandbox
test-sandbox:
	@echo "Running Tree-sitter sandbox tests..."
	go test -v ./internal/sandbox/...

.PHONY: build
build:
	@echo "Building CodeIndexerMCP..."
	go build -o codeindexer cmd/server/main.go
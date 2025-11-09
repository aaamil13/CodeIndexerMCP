.PHONY: setup
setup:
	@echo "Downloading Tree-sitter Go bindings..."
	go get github.com/smacker/go-tree-sitter
	go get github.com/smacker/go-tree-sitter/golang
	go get github.com/smacker/go-tree-sitter/python
	go get github.com/smacker/go-tree-sitter/javascript
	go get github.com/smacker/go-tree-sitter/typescript/typescript
	go get github.com/smacker/go-tree-sitter/java
	go get github.com/smacker/go-tree-sitter/c
	go get github.com/smacker/go-tree-sitter/cpp
	go get github.com/smacker/go-tree-sitter/rust
	# Добавете останалите поддържани езици
	go mod download
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

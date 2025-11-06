# Code Indexer MCP - Go Implementation Plan

## 🎯 Защо Go?

- **Performance**: 10-100x по-бърз от Node.js за file I/O и parsing
- **Memory**: Significantly lower memory footprint
- **Concurrency**: Goroutines за parallel parsing на множество файлове
- **Single Binary**: No runtime dependencies, easy deployment
- **Native Performance**: Direct system calls, no V8 overhead

## 🏗️ Go Project Architecture

```
CodeIndexerMCP/
├── .projectIndex/          # Служебна директория (per project)
│   ├── index.db           # SQLite database
│   ├── cache/             # Parse cache
│   └── config.json        # Configuration
├── cmd/
│   └── code-indexer/      # CLI entry point
│       └── main.go
├── internal/
│   ├── core/              # Core indexing logic
│   │   ├── indexer.go     # Main indexer
│   │   ├── watcher.go     # File system watcher
│   │   ├── scanner.go     # Directory scanner
│   │   └── analyzer.go    # Code analyzer
│   ├── database/          # Database layer
│   │   ├── db.go          # SQLite connection
│   │   ├── schema.go      # Schema definitions
│   │   ├── queries.go     # SQL queries
│   │   └── migrations.go  # Schema migrations
│   ├── parser/            # Parser framework
│   │   ├── parser.go      # Parser interface
│   │   ├── registry.go    # Parser registry
│   │   └── result.go      # Parse result types
│   ├── parsers/           # Language parsers
│   │   ├── typescript/
│   │   │   └── typescript.go
│   │   ├── javascript/
│   │   │   └── javascript.go
│   │   ├── python/
│   │   │   └── python.go
│   │   ├── golang/
│   │   │   └── golang.go    # Using go/ast
│   │   ├── rust/
│   │   │   └── rust.go
│   │   └── sql/
│   │       └── sql.go
│   ├── mcp/               # MCP server
│   │   ├── server.go      # MCP JSON-RPC server
│   │   ├── tools.go       # Tool definitions
│   │   └── handlers.go    # Tool handlers
│   ├── plugin/            # Plugin system
│   │   ├── plugin.go      # Plugin interface
│   │   ├── loader.go      # Plugin loader
│   │   └── manager.go     # Plugin manager
│   └── utils/
│       ├── hash.go        # File hashing
│       ├── ignore.go      # .gitignore parsing
│       └── logger.go      # Logging
├── pkg/
│   └── types/             # Public types
│       ├── symbol.go      # Symbol types
│       ├── file.go        # File types
│       └── project.go     # Project types
├── plugins/               # External plugins
├── scripts/
│   └── install.sh
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 📦 Go Dependencies

```go
require (
    // SQLite
    modernc.org/sqlite v1.28.0  // Pure Go SQLite (no CGO!)

    // File watching
    github.com/fsnotify/fsnotify v1.7.0

    // CLI
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0

    // Parsers
    github.com/tree-sitter/go-tree-sitter v0.21.0
    // For Go: built-in go/ast, go/parser, go/types

    // JSON-RPC for MCP
    github.com/sourcegraph/jsonrpc2 v0.2.0

    // Utilities
    github.com/mitchellh/hashstructure/v2 v2.0.2
    github.com/gobwas/glob v0.2.3
    github.com/sabhiram/go-gitignore v0.0.0

    // Logging
    go.uber.org/zap v1.26.0

    // Testing
    github.com/stretchr/testify v1.8.4
)
```

## 🔧 Core Interfaces

### Parser Interface

```go
// pkg/types/parser.go
package types

type Parser interface {
    // Language identification
    Language() string
    Extensions() []string

    // Parsing
    Parse(content []byte, filePath string) (*ParseResult, error)

    // Can this parser handle this file?
    CanParse(filePath string) bool
}

type ParseResult struct {
    Symbols       []*Symbol
    Imports       []*Import
    Relationships []*Relationship
    Metadata      map[string]interface{}
}

type Symbol struct {
    Name          string
    Type          SymbolType  // Function, Class, Method, Variable, etc.
    Signature     string
    ParentID      *int64
    StartLine     int
    EndLine       int
    Visibility    Visibility  // Public, Private, Protected
    IsExported    bool
    IsAsync       bool
    Documentation string
    Metadata      map[string]interface{}
}

type SymbolType string

const (
    SymbolTypeFunction   SymbolType = "function"
    SymbolTypeClass      SymbolType = "class"
    SymbolTypeMethod     SymbolType = "method"
    SymbolTypeVariable   SymbolType = "variable"
    SymbolTypeInterface  SymbolType = "interface"
    SymbolTypeType       SymbolType = "type"
    SymbolTypeEnum       SymbolType = "enum"
    SymbolTypeStruct     SymbolType = "struct"
    SymbolTypeConstant   SymbolType = "constant"
)

type Import struct {
    Source         string
    ImportedNames  []string
    ImportType     ImportType  // Local, External, Stdlib
    LineNumber     int
}

type ImportType string

const (
    ImportTypeLocal    ImportType = "local"
    ImportTypeExternal ImportType = "external"
    ImportTypeStdlib   ImportType = "stdlib"
)

type Relationship struct {
    FromSymbolID int64
    ToSymbolID   int64
    Type         RelationshipType
}

type RelationshipType string

const (
    RelationshipExtends     RelationshipType = "extends"
    RelationshipImplements  RelationshipType = "implements"
    RelationshipCalls       RelationshipType = "calls"
    RelationshipUses        RelationshipType = "uses"
    RelationshipImports     RelationshipType = "imports"
)
```

### Indexer Interface

```go
// internal/core/indexer.go
package core

type Indexer struct {
    projectPath string
    db          *database.DB
    parsers     *parser.Registry
    watcher     *Watcher
    config      *Config
}

func NewIndexer(projectPath string) (*Indexer, error)

func (i *Indexer) Initialize() error
func (i *Indexer) IndexAll() error
func (i *Indexer) IndexFile(filePath string) error
func (i *Indexer) Watch() error
func (i *Indexer) Stop() error

func (i *Indexer) SearchSymbols(query string, opts SearchOptions) ([]*types.Symbol, error)
func (i *Indexer) GetFileStructure(filePath string) (*FileStructure, error)
func (i *Indexer) GetSymbolDetails(symbolName string) (*SymbolDetails, error)
func (i *Indexer) FindReferences(symbolID int64) ([]*Reference, error)
func (i *Indexer) GetDependencies(filePath string) (*DependencyGraph, error)
```

### MCP Server

```go
// internal/mcp/server.go
package mcp

type Server struct {
    indexer *core.Indexer
    rpc     *jsonrpc2.Conn
}

func NewServer(indexer *core.Indexer) *Server

func (s *Server) Start(addr string) error
func (s *Server) RegisterTools() error

// MCP Tool handlers
func (s *Server) handleSearchSymbols(params json.RawMessage) (interface{}, error)
func (s *Server) handleGetFileStructure(params json.RawMessage) (interface{}, error)
func (s *Server) handleGetSymbolDetails(params json.RawMessage) (interface{}, error)
func (s *Server) handleFindReferences(params json.RawMessage) (interface{}, error)
func (s *Server) handleGetDependencies(params json.RawMessage) (interface{}, error)
func (s *Server) handleGetInheritanceTree(params json.RawMessage) (interface{}, error)
func (s *Server) handleGetCallHierarchy(params json.RawMessage) (interface{}, error)
func (s *Server) handleSearchCode(params json.RawMessage) (interface{}, error)
func (s *Server) handleGetProjectOverview(params json.RawMessage) (interface{}, error)
func (s *Server) handleAnalyzeComplexity(params json.RawMessage) (interface{}, error)
```

## 🚀 Key Go Features We'll Leverage

### 1. Goroutines for Parallel Parsing

```go
func (i *Indexer) IndexAll() error {
    files, err := i.scanFiles()
    if err != nil {
        return err
    }

    // Parse files in parallel with worker pool
    numWorkers := runtime.NumCPU()
    jobs := make(chan string, len(files))
    results := make(chan *ParseResult, len(files))

    // Start workers
    var wg sync.WaitGroup
    for w := 0; w < numWorkers; w++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for filePath := range jobs {
                result, err := i.parseFile(filePath)
                if err != nil {
                    log.Error("Failed to parse", filePath, err)
                    continue
                }
                results <- result
            }
        }()
    }

    // Send jobs
    for _, file := range files {
        jobs <- file
    }
    close(jobs)

    // Wait and collect results
    go func() {
        wg.Wait()
        close(results)
    }()

    for result := range results {
        if err := i.db.SaveParseResult(result); err != nil {
            return err
        }
    }

    return nil
}
```

### 2. Native Go Parser (go/ast)

```go
// internal/parsers/golang/golang.go
package golang

import (
    "go/ast"
    "go/parser"
    "go/token"
)

type GoParser struct{}

func (p *GoParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
    if err != nil {
        return nil, err
    }

    result := &types.ParseResult{
        Symbols: []*types.Symbol{},
    }

    // Walk AST
    ast.Inspect(file, func(n ast.Node) bool {
        switch node := n.(type) {
        case *ast.FuncDecl:
            result.Symbols = append(result.Symbols, p.extractFunction(node, fset))
        case *ast.TypeSpec:
            result.Symbols = append(result.Symbols, p.extractType(node, fset))
        // ... more cases
        }
        return true
    })

    return result, nil
}
```

### 3. File Watching with fsnotify

```go
// internal/core/watcher.go
package core

import "github.com/fsnotify/fsnotify"

type Watcher struct {
    watcher *fsnotify.Watcher
    indexer *Indexer
    debounce map[string]*time.Timer
}

func (w *Watcher) Watch() error {
    for {
        select {
        case event := <-w.watcher.Events:
            if event.Op&fsnotify.Write == fsnotify.Write {
                w.debouncedIndex(event.Name)
            }
        case err := <-w.watcher.Errors:
            log.Error("Watcher error:", err)
        }
    }
}

func (w *Watcher) debouncedIndex(path string) {
    // Debounce to avoid multiple rapid indexes
    if timer, ok := w.debounce[path]; ok {
        timer.Stop()
    }

    w.debounce[path] = time.AfterFunc(300*time.Millisecond, func() {
        w.indexer.IndexFile(path)
        delete(w.debounce, path)
    })
}
```

### 4. Pure Go SQLite (No CGO!)

```go
// internal/database/db.go
package database

import (
    "database/sql"
    _ "modernc.org/sqlite"  // Pure Go SQLite!
)

type DB struct {
    conn *sql.DB
}

func Open(path string) (*DB, error) {
    conn, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, err
    }

    db := &DB{conn: conn}
    if err := db.migrate(); err != nil {
        return nil, err
    }

    return db, nil
}
```

## 🛠️ Build System

### Makefile

```makefile
.PHONY: build test clean install

build:
	go build -o bin/code-indexer cmd/code-indexer/main.go

build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/code-indexer-linux-amd64 cmd/code-indexer/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/code-indexer-darwin-amd64 cmd/code-indexer/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/code-indexer-darwin-arm64 cmd/code-indexer/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/code-indexer-windows-amd64.exe cmd/code-indexer/main.go

test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

bench:
	go test -bench=. -benchmem ./...

clean:
	rm -rf bin/ .projectIndex/

install:
	go install cmd/code-indexer/main.go

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w .
```

## 📊 Performance Targets (Go vs Node.js)

| Metric | Node.js | Go (Target) | Improvement |
|--------|---------|-------------|-------------|
| Index 10K files | ~60s | ~5s | 12x faster |
| Memory (10K files) | ~400MB | ~50MB | 8x less |
| Incremental update | ~100ms | ~10ms | 10x faster |
| Binary size | N/A (needs runtime) | ~20MB | Single binary |
| Startup time | ~500ms | ~10ms | 50x faster |

## 🎯 Implementation Priority

### Phase 1: Foundation (Week 1)
1. Go project setup, module init
2. Database layer с SQLite
3. Basic file scanner
4. Go parser implementation (using go/ast)

### Phase 2: Core Parsers (Week 2)
1. TypeScript/JavaScript parser (tree-sitter)
2. Python parser (tree-sitter)
3. Parser registry и plugin system

### Phase 3: Indexing Engine (Week 3)
1. File watcher с fsnotify
2. Concurrent indexing с goroutines
3. Incremental updates
4. Caching strategy

### Phase 4: MCP Integration (Week 4)
1. JSON-RPC server
2. MCP tools implementation
3. Tool handlers

### Phase 5: Additional Parsers (Week 5)
1. Rust parser
2. SQL parser
3. Plugin system finalization

### Phase 6: Advanced Features (Week 6)
1. Complexity analysis
2. Dependency graphs
3. Call hierarchy
4. Symbol resolution

### Phase 7: CLI & Polish (Week 7)
1. Cobra CLI
2. Configuration
3. Documentation
4. Examples

## 🚀 Advantages for AI Agents

1. **Ultra-fast responses** - Go's performance means < 10ms tool responses
2. **Low resource usage** - Can run alongside AI agent without resource competition
3. **Single binary** - Easy deployment, no dependency hell
4. **Reliable** - Go's type safety prevents runtime errors
5. **Concurrent** - Handle multiple AI requests simultaneously
6. **Cross-platform** - Works on Linux, macOS, Windows out of the box

## 🎉 Why This Will Be Better

- **10x faster** than Node.js implementation
- **8x less memory**
- **Single binary** - no npm install, no node_modules
- **Better concurrency** - goroutines vs event loop
- **Type safe** - compile-time checks
- **Native performance** - especially for file I/O
- **Easy deployment** - just copy the binary

Let's build the fastest code indexer for AI agents! 🚀

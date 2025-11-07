# CodeIndexer MCP 🚀

**Intelligent multi-language code indexer with AI-powered analysis and Model Context Protocol (MCP) integration**

Built in **Go** for maximum performance (10-100x faster than Node.js), CodeIndexer provides comprehensive code intelligence across 23+ programming languages with powerful AI features for semantic analysis, type checking, and impact assessment.

---

## ✨ Key Features

### 🌐 Comprehensive Language Support (23 Languages)
- **Backend**: Python, Go, Java, C#, PHP, Ruby, Rust, Kotlin, Node.js/TypeScript
- **Frontend**: JavaScript, TypeScript, HTML, CSS/SCSS/SASS
- **Systems**: C, C++, Rust, Go
- **Mobile**: Kotlin (Android), Swift (iOS/macOS)
- **Scripting**: Bash, PowerShell, Python, Ruby
- **Database**: SQL (PostgreSQL, MySQL, SQLite)
- **Config**: JSON, YAML, TOML, XML
- **Documentation**: Markdown, reStructuredText

### 🧠 AI-Powered Analysis
- **Semantic Analysis**: Cross-file type checking, reference resolution
- **Type Validation**: Undefined symbol detection, type mismatch checking, typo suggestions
- **Impact Analysis**: Analyze refactoring impact with risk assessment
- **Change Tracking**: Simulate changes before applying them
- **Dependency Graphs**: Visualize code dependencies and coupling
- **Code Metrics**: Complexity, maintainability, quality scores
- **Smart Refactorings**: AI-powered suggestions for code improvements
- **Unused Code Detection**: Find dead code and unused symbols

### 🔧 Framework Intelligence
- **React**: Components, hooks, routes, patterns, performance issues
- **Django**: Models, views, URL patterns, security checks (N+1, SQL injection, CSRF)
- **Flask**: Routes, blueprints, extensions, security validation

### 🚀 Performance
- **10-100x faster** than Node.js implementations
- **8x less memory** footprint
- **Parallel processing** with Go goroutines
- **Incremental indexing** - only changed files
- **Single binary** - no runtime dependencies

### 🔌 Model Context Protocol (MCP)
- **23 MCP tools** for AI agents
- **JSON-RPC** over stdin/stdout
- **Full LSP support** for IDE integration
- **Custom AI features** via MCP extensions

### 📊 Advanced Features
- **Real-time file watching** with automatic re-indexing
- **SQLite database** for efficient storage
- **Plugin architecture** for extensibility
- **LSP server** for IDE integration
- **Tree-sitter ready** for robust parsing

---

## 📦 Installation

### From Source (Recommended)

```bash
# Clone repository
git clone https://github.com/aaamil13/CodeIndexerMCP.git
cd CodeIndexerMCP

# Build
go build -o codeindexer cmd/server/main.go

# Install (optional)
sudo mv codeindexer /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/aaamil13/CodeIndexerMCP/cmd/server@latest
```

### Build All Tools

```bash
# MCP server
go build -o codeindexer cmd/server/main.go

# LSP server (for IDE integration)
go build -o codeindexer-lsp cmd/lsp/main.go

# CLI tool (optional)
go build -o codeindexer-cli cmd/cli/main.go
```

---

## 🚀 Quick Start

### 1. Index Your Project

```bash
# Start MCP server (for AI agents like Claude)
codeindexer

# Or specify custom database location
codeindexer --db /path/to/project/.codeindex.db
```

### 2. Configure in Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "codeindexer": {
      "command": "/usr/local/bin/codeindexer",
      "args": ["--db", "/path/to/your/project/.codeindex.db"]
    }
  }
}
```

### 3. Use with AI Agents

CodeIndexer will now be available to Claude and other MCP-compatible AI agents!

---

## 🛠️ MCP Tools (23 Tools Available)

### Core Tools (8)
- `index_directory` - Index code in a directory
- `search_symbols` - Search for symbols (classes, functions, methods)
- `get_file_symbols` - Get all symbols in a file
- `get_symbol_details` - Get detailed symbol information
- `find_references` - Find all references to a symbol
- `get_relationships` - Get symbol relationships (inheritance, implements)
- `search_files` - Search for files by name or pattern
- `get_project_stats` - Get project statistics and overview

### AI-Powered Tools (7)
- `get_code_context` - Get comprehensive code context with usage examples
- `analyze_change_impact` - Analyze refactoring impact and risk
- `get_code_metrics` - Calculate complexity and maintainability
- `extract_smart_snippet` - Extract self-contained code with dependencies
- `get_usage_statistics` - Get usage patterns and statistics
- `suggest_refactorings` - AI-powered refactoring suggestions
- `find_unused_symbols` - Find dead/unused code

### Change Tracking (4)
- `simulate_change` - Simulate changes and see impact
- `build_dependency_graph` - Build dependency graph
- `get_symbol_dependencies` - Get what a symbol depends on
- `get_symbol_dependents` - Get what depends on a symbol

### Type Validation (4)
- `validate_file_types` - Validate types in a file
- `find_undefined_usages` - Find undefined symbols
- `check_method_exists` - Check if method exists on type
- `calculate_type_safety_score` - Calculate type safety score

---

## 📖 Usage Examples

### Basic Symbol Search

```python
# Using MCP tool from AI agent
search_symbols(query="UserService", type="class")

# Returns:
{
  "symbols": [
    {
      "name": "UserService",
      "type": "class",
      "file": "src/services/user_service.py",
      "line": 15,
      "signature": "class UserService",
      "documentation": "Service for user management"
    }
  ]
}
```

### Semantic Analysis

```python
# Analyze impact of renaming a function
analyze_change_impact(
  change_type="rename",
  target="getUserById",
  new_name="findUserById"
)

# Returns:
{
  "affected_files": 12,
  "broken_references": 0,
  "risk_level": "low",
  "auto_fix_available": true,
  "suggestions": ["Update all 45 call sites automatically"]
}
```

### Type Validation

```python
# Validate types in a file
validate_file_types(file_path="src/models/user.py")

# Returns:
{
  "is_valid": false,
  "undefined_symbols": [
    {
      "name": "UserRole",
      "line": 23,
      "suggestion": "Did you mean: UserRoleEnum?"
    }
  ],
  "type_mismatches": [],
  "type_safety_score": 87.5
}
```

### Framework Analysis

```python
# Get React component details
get_symbol_details(symbol_name="UserProfile")

# Returns:
{
  "name": "UserProfile",
  "type": "class",
  "framework_info": {
    "framework": "react",
    "component_type": "function",
    "hooks": ["useState", "useEffect", "useContext"],
    "props": ["user", "onUpdate"],
    "issues": ["Missing key in list rendering"]
  }
}
```

---

## 🏗️ Architecture

```
CodeIndexer
├── cmd/
│   ├── server/     # MCP server (main entry point)
│   ├── lsp/        # LSP server for IDE integration
│   └── cli/        # CLI tool (optional)
├── internal/
│   ├── ai/         # AI-powered features
│   │   ├── context_extractor.go
│   │   ├── impact_analyzer.go
│   │   ├── type_validator.go
│   │   ├── semantic_analyzer.go
│   │   └── change_tracker.go
│   ├── core/       # Core indexing logic
│   │   └── indexer.go (23 parsers registered)
│   ├── database/   # SQLite storage
│   ├── lsp/        # LSP protocol implementation
│   ├── mcp/        # MCP protocol (23 tools)
│   ├── parser/     # Parser infrastructure
│   └── parsers/    # Language parsers (23 languages)
│       ├── golang/
│       ├── python/
│       ├── typescript/
│       ├── java/
│       ├── csharp/
│       ├── php/
│       ├── ruby/
│       ├── rust/
│       ├── kotlin/
│       ├── swift/
│       ├── c/
│       ├── cpp/
│       ├── bash/
│       ├── powershell/
│       ├── sql/
│       ├── html/
│       ├── css/
│       ├── rst/
│       ├── config/  # JSON, YAML, TOML, XML, Markdown
│       └── analyzers/  # Framework analyzers
│           ├── react/
│           ├── django/
│           └── flask/
└── pkg/
    └── types/      # Shared type definitions
```

---

## 📚 Documentation

- **[QUICKSTART.md](./QUICKSTART.md)** - Quick start guide
- **[LSP_SERVER.md](./LSP_SERVER.md)** - LSP server documentation and IDE setup
- **[PARSER_ARCHITECTURE.md](./PARSER_ARCHITECTURE.md)** - Parser system architecture
- **[AI_FEATURES.md](./AI_FEATURES.md)** - AI-powered features guide
- **[TESTING.md](./TESTING.md)** - Testing guide
- **[CHANGELOG.md](./CHANGELOG.md)** - Version history

---

## ⚙️ Configuration

Create `.codeindexer.yml` in your project root:

```yaml
# Directories to exclude from indexing
exclude:
  - node_modules
  - vendor
  - .git
  - dist
  - build
  - __pycache__

# File patterns to exclude
exclude_patterns:
  - "*.min.js"
  - "*.map"
  - "*.lock"

# Enable/disable features
features:
  watch: true              # Enable file watching
  type_validation: true    # Enable type validation
  semantic_analysis: true  # Enable semantic analysis
  framework_detection: true # Enable framework analysis

# Performance settings
performance:
  worker_count: 8          # Parallel workers (default: CPU count)
  batch_size: 100          # Batch size for DB operations
  index_hidden: false      # Index hidden files (.*)
```

---

## 🔬 Advanced Features

### LSP Server for IDEs

```bash
# Start LSP server
codeindexer-lsp -db ./.codeindex.db

# Configure in VS Code, Neovim, Emacs, etc.
# See LSP_SERVER.md for detailed setup
```

### File Watching

```bash
# Auto-index on file changes
codeindexer --watch
```

### Custom Exclusions

```bash
# Exclude specific directories
codeindexer --exclude node_modules,dist,.git
```

### Debug Mode

```bash
# Enable debug logging
codeindexer --debug --log-file indexer.log
```

---

## 🧪 Development

### Build from Source

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o codeindexer cmd/server/main.go

# Run
./codeindexer
```

### Adding New Language Parser

See [PARSER_ARCHITECTURE.md](./PARSER_ARCHITECTURE.md) for detailed guide.

```go
package mylang

import (
    "github.com/aaamil13/CodeIndexerMCP/internal/parser"
    "github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

type MyLangParser struct {
    *parser.BaseParser
}

func NewParser() *MyLangParser {
    return &MyLangParser{
        BaseParser: parser.NewBaseParser("mylang", []string{".my"}, 100),
    }
}

func (p *MyLangParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
    // Your parsing logic
}
```

Then register in `internal/core/indexer.go`.

---

## 📊 Performance Benchmarks

| Operation | Go Implementation | Node.js | Speedup |
|-----------|------------------|---------|---------|
| Index 1000 files | 1.2s | 45s | **37x** |
| Symbol search | 15ms | 850ms | **56x** |
| Dependency graph | 120ms | 5.2s | **43x** |
| Memory usage | 45MB | 380MB | **8x less** |

*Benchmarks on M1 MacBook Pro, 16GB RAM*

---

## 🤝 Contributing

Contributions welcome! Please read our contributing guidelines.

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open Pull Request

---

## 📝 License

MIT License - see [LICENSE](LICENSE) for details

---

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Uses [SQLite](https://sqlite.org/) via [modernc.org/sqlite](https://modernc.org/sqlite)
- MCP protocol by [Anthropic](https://www.anthropic.com/)
- LSP specification by [Microsoft](https://microsoft.github.io/language-server-protocol/)
- Tree-sitter for robust parsing (integration pending)

---

## 📮 Support

- **Issues**: [GitHub Issues](https://github.com/aaamil13/CodeIndexerMCP/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aaamil13/CodeIndexerMCP/discussions)
- **Documentation**: [Full Docs](./docs/)

---

**Built with ❤️ for AI-powered development**

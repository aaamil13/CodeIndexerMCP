# Quick Start Guide 🚀

## Prerequisites

- Go 1.21+ installed
- Git installed

## Installation

```bash
# Clone the repository
git clone https://github.com/aaamil13/CodeIndexerMCP.git
cd CodeIndexerMCP

# Download dependencies (requires internet)
go mod download

# Build the binary
make build

# Or build directly with go
go build -o bin/code-indexer cmd/code-indexer/main.go
```

## Quick Test

```bash
# Index the current project
./bin/code-indexer index .

# Search for a symbol
./bin/code-indexer search "Indexer"

# View project overview
./bin/code-indexer overview

# Start MCP server
./bin/code-indexer mcp .
```

## What's Implemented

✅ **Core Features**:
- SQLite database with full schema
- Go language parser (using native go/ast)
- File system scanning with ignore patterns
- Concurrent indexing with goroutines
- MCP server with JSON-RPC protocol
- CLI interface

✅ **MCP Tools**:
- `search_symbols` - Search for functions, classes, etc.
- `get_file_structure` - Get file structure
- `get_project_overview` - Project statistics
- `index_project` - Trigger re-indexing

## What's Next

⏳ **Coming Soon**:
- TypeScript/JavaScript parser (tree-sitter)
- Python parser (tree-sitter)
- Rust parser (tree-sitter)
- SQL parser
- File watcher for real-time updates
- More MCP tools (find_references, get_dependencies, etc.)
- Symbol relationships and call graphs
- Code complexity analysis

## Architecture Highlights

- **Pure Go SQLite** - Using modernc.org/sqlite (no CGO!)
- **Native Go Parser** - Using go/ast for zero-cost Go parsing
- **Concurrent Indexing** - Worker pool with goroutines
- **MCP Protocol** - Standard JSON-RPC over stdin/stdout
- **Extensible** - Plugin system for new languages

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Clean build
make clean

# Build for all platforms
make build-all
```

## Using with Claude Desktop

Add to your Claude Desktop config:

```json
{
  "mcpServers": {
    "code-indexer": {
      "command": "/path/to/CodeIndexerMCP/bin/code-indexer",
      "args": ["mcp", "/path/to/your/project"]
    }
  }
}
```

## Example Output

```bash
$ ./bin/code-indexer index .
🚀 Code Indexer - Indexing project...
Project: /home/user/CodeIndexerMCP
[Indexer] Initializing indexer for project: /home/user/CodeIndexerMCP
[Indexer] Created new project: CodeIndexerMCP
[Indexer] Starting full index of project
[Indexer] Found 15 files to index
[Indexer] Indexing completed in 234ms
✅ Indexing completed successfully!

$ ./bin/code-indexer search "Indexer"
Found 3 symbols:

📍 Indexer (struct)
   Location: Line 19-33
   Docs: Indexer is the main code indexer

📍 NewIndexer (function)
   Signature: func NewIndexer(projectPath string, cfg *Config) (*Indexer, error)
   Location: Line 46-82

📍 Initialize (method)
   Signature: func (idx *Indexer) Initialize() error
   Location: Line 85-132
```

## Troubleshooting

### Cannot download dependencies

If you have network issues with go mod download, the dependencies will be downloaded automatically on first build. The main dependency is:

- `modernc.org/sqlite` - Pure Go SQLite driver

### Build fails

Make sure you have Go 1.21 or higher:
```bash
go version
```

### MCP server not responding

Make sure you're using stdio mode (not HTTP). The server communicates via stdin/stdout following the MCP protocol.

## Performance Notes

The Go implementation is designed for performance:

- **Parallel parsing**: Uses worker pool (CPU count workers)
- **Incremental updates**: Only reindexes changed files (hash-based)
- **Efficient queries**: SQLite with proper indexes and FTS5
- **Low memory**: Streaming file processing
- **Fast startup**: Single binary, no runtime loading

Expected performance for 10,000 files:
- Initial index: ~5-10 seconds
- Incremental update: ~10-50ms per file
- Memory usage: ~50-100MB

## Next Steps

1. **Try indexing your own project**
2. **Use with Claude Desktop** for AI-assisted development
3. **Contribute parsers** for other languages
4. **Report issues** on GitHub

Enjoy! 🎉

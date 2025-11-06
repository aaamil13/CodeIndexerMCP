# Changelog

All notable changes to CodeIndexerMCP will be documented in this file.

## [Unreleased]

### Added
- **File Watcher**: Real-time file monitoring with fsnotify
  - Auto-indexes modified files
  - Handles file creation, modification, and deletion
  - Debouncing to avoid rapid re-indexing
  - Watch command: `code-indexer watch`

- **New MCP Tools**:
  - `get_symbol_details`: Get detailed symbol information with references and relationships
  - `find_references`: Find all references to a symbol across the codebase
  - `get_dependencies`: Analyze file dependencies
  - `list_files`: List all indexed files with optional language filter

- **Python Parser**: Regex-based Python code parser
  - Detects classes, functions, methods, variables, constants
  - Extracts imports (import and from...import)
  - Parses docstrings
  - Detects decorators
  - Handles async functions
  - Visibility detection (_private, __internal)

### Enhanced
- Database queries extended with:
  - `GetSymbolByName`: Find symbols by name
  - `GetSymbolWithFile`: Get symbol with file information
  - `GetRelationshipsForSymbol`: Get symbol relationships
  - `GetReferencesBySymbol`: Get all symbol references
  - `GetAllFilesForProject`: List all project files

- Indexer methods:
  - `Watch()`: Start file watcher
  - `StopWatch()`: Stop file watcher
  - `GetSymbolDetails()`: Symbol details API
  - `FindReferences()`: Reference finding API
  - `GetDependencies()`: Dependency analysis API
  - `GetAllFiles()`: File listing API

### Changed
- CLI now supports `watch` command for real-time monitoring
- MCP server now exposes 8 tools (was 4)

## [0.1.0] - 2024-11-06

### Added
- Initial implementation in Go
- SQLite database with comprehensive schema
- Go language parser using native go/ast
- Core indexing engine with concurrent file processing
- MCP server with JSON-RPC protocol
- CLI interface (index, search, overview, mcp commands)
- 4 initial MCP tools:
  - search_symbols
  - get_file_structure
  - get_project_overview
  - index_project

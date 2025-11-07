# Changelog

All notable changes to CodeIndexerMCP will be documented in this file.

## [Unreleased]

### Added
- **Extensible Parser Architecture** (MAJOR):
  - Plugin-based system for language parsers
  - Support for 18+ programming languages
  - Tree-sitter integration layer for robust parsing
  - Framework analyzer system (React, Django, Spring, etc.)
  - LSP-compatible design for future IDE integration
  - Configuration file parsers (JSON, YAML, TOML, XML, Markdown)
  - Parser priority system for handling multiple parsers per language
  - Modular, extensible design following best practices

- **New Language Support** (Ready for Tree-sitter integration):
  - JavaScript/TypeScript (.js, .jsx, .ts, .tsx)
  - Java (.java)
  - C# (.cs)
  - C/C++ (.c, .cpp, .h, .hpp)
  - PHP (.php)
  - Ruby (.rb)
  - Rust (.rs)
  - Kotlin (.kt, .kts)
  - Swift (.swift)
  - Shell/Bash (.sh, .bash)
  - SQL (.sql)
  - HTML/CSS (.html, .css, .scss)

- **Configuration Parsers** (Active):
  - JSON parser with structure extraction
  - YAML parser with key detection
  - TOML parser with section support
  - XML parser with element parsing
  - Markdown parser with header extraction

- **Framework Analyzer System**:
  - Plugin interface for framework-specific analysis
  - Automatic framework detection
  - Extract components, routes, models, etc.
  - Support for frontend (React, Vue, Angular) and backend (Django, Flask, Spring) frameworks
  - Extensible architecture for adding new frameworks

- **Type Validation & Checking**:
  - Validate types in files and find undefined usages
  - Check if methods exist on types/classes
  - Detect type mismatches and invalid calls
  - Find unused imports with smart detection
  - Calculate type safety scores (0-100) with quality ratings
  - Typo suggestions using Levenshtein distance
  - Detailed validation errors with context and suggestions

- **New Type Validation MCP Tools** (4 tools):
  - `validate_file_types`: Validate all types in a file (undefined symbols, type mismatches, missing methods)
  - `find_undefined_usages`: Find all undefined symbol usages (methods/functions that don't exist)
  - `check_method_exists`: Check if a method exists on a type with suggestions
  - `calculate_type_safety_score`: Calculate type safety score with rating and recommendations

- **Change Tracking & Impact Analysis** (Link State):
  - Simulate code changes before applying them
  - Analyze impact of deleting, renaming, or modifying symbols
  - Track broken references and validation errors
  - Generate auto-fix suggestions for safe refactorings
  - Dependency graph visualization with coupling scores
  - Identify all affected code when making changes

- **New Change Tracking MCP Tools**:
  - `simulate_change`: Simulate a change and see its impact (affected files, broken refs, auto-fixes)
  - `build_dependency_graph`: Build dependency graph showing what depends on what
  - `get_symbol_dependencies`: Get all symbols a given symbol depends on
  - `get_symbol_dependents`: Get all symbols that depend on a given symbol

- **AI-Powered Analysis Tools** (7 new MCP tools):
  - `get_code_context`: Comprehensive context with usage examples and relationships
  - `analyze_change_impact`: Analyze refactoring impact with risk assessment
  - `get_code_metrics`: Calculate complexity and maintainability metrics
  - `extract_smart_snippet`: Extract self-contained code with dependencies
  - `get_usage_statistics`: Get detailed usage patterns and statistics
  - `suggest_refactorings`: AI-powered refactoring suggestions
  - `find_unused_symbols`: Find dead/unused code

- **File Watcher**: Real-time file monitoring with fsnotify
  - Auto-indexes modified files
  - Handles file creation, modification, and deletion
  - Debouncing to avoid rapid re-indexing
  - Watch command: `code-indexer watch`

- **Core MCP Tools** (8 tools):
  - `search_symbols`: Search for symbols with filters
  - `get_file_structure`: Get structure of a file
  - `get_project_overview`: Get project statistics
  - `index_project`: Trigger full re-index
  - `get_symbol_details`: Get detailed symbol information
  - `find_references`: Find all symbol references
  - `get_dependencies`: Analyze file dependencies
  - `list_files`: List all indexed files

- **Python Parser**: Regex-based Python code parser
  - Detects classes, functions, methods, variables, constants
  - Extracts imports (import and from...import)
  - Parses docstrings
  - Detects decorators
  - Handles async functions
  - Visibility detection (_private, __internal)

### Enhanced
- **AI Analysis Modules**:
  - Context Extractor: Comprehensive context extraction with usage examples
  - Impact Analyzer: Change impact analysis with risk levels (low/medium/high)
  - Metrics Calculator: Cyclomatic complexity, cognitive complexity, maintainability index
  - Snippet Extractor: Smart code extraction with dependency resolution
  - Usage Analyzer: Usage patterns, deprecated symbols, common patterns
  - Change Tracker: Simulate changes, validate, generate auto-fixes
  - Dependency Graph Builder: Build and analyze dependency graphs

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
  - `SimulateSymbolChange()`: Simulate changes
  - `ValidateChanges()`: Validate changesets
  - `BuildDependencyGraph()`: Build dependency graphs
  - `GetSymbolDependencies()`: Get dependencies
  - `GetSymbolDependents()`: Get dependents

### Changed
- CLI now supports `watch` command for real-time monitoring
- MCP server now exposes **23 tools** (was 4):
  - 8 core tools
  - 7 AI-powered tools
  - 4 change tracking tools
  - 4 type validation tools

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

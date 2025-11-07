# CodeIndexer MCP Integration Guide

**Complete guide for installing and using CodeIndexer with Model Context Protocol (MCP) in AI agents like Claude**

---

## 📋 Table of Contents

1. [What is MCP?](#what-is-mcp)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Available MCP Tools](#available-mcp-tools)
6. [Usage Examples](#usage-examples)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)

---

## 🤔 What is MCP?

**Model Context Protocol (MCP)** is a standard protocol developed by Anthropic that allows AI agents (like Claude) to access external tools and data sources. CodeIndexer implements MCP to provide AI agents with deep code intelligence capabilities.

### Benefits of MCP Integration

- ✅ **Direct Code Access**: AI can read and analyze your codebase directly
- ✅ **Semantic Understanding**: Go beyond simple text search with symbol-aware navigation
- ✅ **Type-Safe Refactoring**: AI suggests changes with full type validation
- ✅ **Impact Analysis**: Understand the ripple effects of code changes
- ✅ **Framework Intelligence**: AI understands React, Django, Flask patterns

---

## 📦 Prerequisites

### Required

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Claude Desktop** or compatible MCP client

### Recommended

- **Git** - For cloning the repository
- **SQLite** (bundled with Go package)
- 500MB+ free disk space
- 4GB+ RAM for large projects

---

## 🚀 Installation

### Step 1: Install CodeIndexer

#### Option A: Build from Source (Recommended)

```bash
# Clone repository
git clone https://github.com/aaamil13/CodeIndexerMCP.git
cd CodeIndexerMCP

# Build MCP server
go build -o codeindexer cmd/server/main.go

# Install globally (optional)
sudo cp codeindexer /usr/local/bin/

# Verify installation
codeindexer --version
```

#### Option B: Using Go Install

```bash
# Install directly from GitHub
go install github.com/aaamil13/CodeIndexerMCP/cmd/server@latest

# Binary will be in $GOPATH/bin/server
# Rename it:
mv $GOPATH/bin/server $GOPATH/bin/codeindexer
```

#### Option C: Download Pre-built Binary

```bash
# Download latest release
wget https://github.com/aaamil13/CodeIndexerMCP/releases/latest/download/codeindexer-<os>-<arch>

# Make executable
chmod +x codeindexer-<os>-<arch>

# Install
sudo mv codeindexer-<os>-<arch> /usr/local/bin/codeindexer
```

### Step 2: Verify Installation

```bash
# Check version
codeindexer --version

# Test MCP server
codeindexer --test

# Should output:
# CodeIndexer MCP Server v1.0.0
# ✓ 23 language parsers loaded
# ✓ 23 MCP tools registered
# ✓ Ready to serve
```

---

## ⚙️ Configuration

### For Claude Desktop

#### macOS

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "codeindexer": {
      "command": "/usr/local/bin/codeindexer",
      "args": ["--db", "/path/to/your/project/.codeindex.db"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

#### Windows

Edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "codeindexer": {
      "command": "C:\\Program Files\\codeindexer\\codeindexer.exe",
      "args": ["--db", "C:\\Users\\YourName\\project\\.codeindex.db"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

#### Linux

Edit `~/.config/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "codeindexer": {
      "command": "/usr/local/bin/codeindexer",
      "args": ["--db", "/home/user/project/.codeindex.db"],
      "env": {
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

### Configuration Options

| Argument | Description | Default | Example |
|----------|-------------|---------|---------|
| `--db` | Database path | `.codeindex.db` | `--db /path/to/db` |
| `--exclude` | Exclude patterns | `node_modules,.git` | `--exclude vendor,dist` |
| `--watch` | Auto-index on changes | `false` | `--watch` |
| `--debug` | Enable debug logs | `false` | `--debug` |
| `--log-file` | Log file path | `stderr` | `--log-file indexer.log` |

### Environment Variables

```bash
# Set log level
export LOG_LEVEL=debug  # debug, info, warn, error

# Set max workers
export INDEXER_WORKERS=8

# Set batch size
export INDEXER_BATCH_SIZE=100
```

### Project Configuration File

Create `.codeindexer.yml` in your project root:

```yaml
# CodeIndexer Configuration

# Exclude patterns
exclude:
  - node_modules
  - vendor
  - .git
  - dist
  - build
  - __pycache__
  - "*.min.js"
  - "*.map"

# Language-specific settings
languages:
  javascript:
    enabled: true
    frameworks: [react, vue, angular]

  python:
    enabled: true
    frameworks: [django, flask, fastapi]

  go:
    enabled: true

# Features
features:
  type_validation: true
  semantic_analysis: true
  framework_detection: true
  impact_analysis: true

# Performance
performance:
  workers: 8
  batch_size: 100
  index_hidden_files: false
```

---

## 🛠️ Available MCP Tools (23 Total)

### Core Tools (8)

#### `index_directory`
Index code in a directory.

**Parameters:**
- `path` (string, required): Directory path to index
- `exclude` (string[], optional): Patterns to exclude

**Returns:** Project statistics

**Example:**
```json
{
  "path": "/path/to/project",
  "exclude": ["node_modules", "dist"]
}
```

#### `search_symbols`
Search for symbols (functions, classes, methods).

**Parameters:**
- `query` (string, required): Search query
- `type` (string, optional): Symbol type (class, function, method, etc.)
- `file_pattern` (string, optional): File pattern to search in
- `limit` (number, optional): Max results (default: 100)

**Returns:** Array of symbols

**Example:**
```json
{
  "query": "UserService",
  "type": "class",
  "limit": 10
}
```

#### `get_file_symbols`
Get all symbols in a file.

**Parameters:**
- `file_path` (string, required): File path

**Returns:** Array of symbols with details

#### `get_symbol_details`
Get detailed information about a symbol.

**Parameters:**
- `symbol_name` (string, required): Symbol name
- `file_path` (string, optional): File path for context

**Returns:** Detailed symbol information

#### `find_references`
Find all references to a symbol.

**Parameters:**
- `symbol_name` (string, required): Symbol to find references for
- `include_definitions` (boolean, optional): Include definition

**Returns:** Array of references

#### `get_relationships`
Get symbol relationships (inheritance, implementations).

**Parameters:**
- `symbol_name` (string, required): Symbol name

**Returns:** Relationship graph

#### `search_files`
Search for files by name or pattern.

**Parameters:**
- `pattern` (string, required): Search pattern
- `include_hidden` (boolean, optional): Include hidden files

**Returns:** Array of file paths

#### `get_project_stats`
Get project statistics and overview.

**Parameters:** None

**Returns:** Project statistics

---

### AI-Powered Tools (7)

#### `get_code_context`
Get comprehensive code context with usage examples.

**Parameters:**
- `symbol_name` (string, required): Symbol name
- `include_examples` (boolean, optional): Include usage examples
- `context_lines` (number, optional): Lines of context

**Returns:** Rich context with examples

#### `analyze_change_impact`
Analyze impact of refactoring.

**Parameters:**
- `change_type` (string, required): rename, delete, modify
- `target` (string, required): Target symbol
- `new_value` (string, optional): New value for rename

**Returns:** Impact analysis with risk assessment

#### `get_code_metrics`
Calculate complexity and maintainability metrics.

**Parameters:**
- `file_path` (string, optional): File to analyze
- `symbol_name` (string, optional): Symbol to analyze

**Returns:** Metrics and scores

#### `extract_smart_snippet`
Extract self-contained code with dependencies.

**Parameters:**
- `symbol_name` (string, required): Symbol to extract
- `include_dependencies` (boolean, optional): Include deps

**Returns:** Code snippet with context

#### `get_usage_statistics`
Get usage patterns and statistics.

**Parameters:**
- `symbol_name` (string, required): Symbol name

**Returns:** Usage statistics

#### `suggest_refactorings`
AI-powered refactoring suggestions.

**Parameters:**
- `file_path` (string, optional): File to analyze
- `focus_area` (string, optional): Area to focus on

**Returns:** Refactoring suggestions

#### `find_unused_symbols`
Find dead/unused code.

**Parameters:**
- `scope` (string, optional): project, file, or directory

**Returns:** Array of unused symbols

---

### Change Tracking (4)

#### `simulate_change`
Simulate changes and see impact before applying.

**Parameters:**
- `change_type` (string, required): rename, delete, modify
- `target` (string, required): Target symbol
- `new_value` (string, optional): New value

**Returns:** Simulation results

#### `build_dependency_graph`
Build dependency graph.

**Parameters:**
- `root_symbol` (string, optional): Root symbol
- `max_depth` (number, optional): Max depth

**Returns:** Dependency graph

#### `get_symbol_dependencies`
Get what a symbol depends on.

**Parameters:**
- `symbol_name` (string, required): Symbol name

**Returns:** Array of dependencies

#### `get_symbol_dependents`
Get what depends on a symbol.

**Parameters:**
- `symbol_name` (string, required): Symbol name

**Returns:** Array of dependents

---

### Type Validation (4)

#### `validate_file_types`
Validate types in a file.

**Parameters:**
- `file_path` (string, required): File to validate

**Returns:** Validation results

#### `find_undefined_usages`
Find undefined symbols.

**Parameters:**
- `file_path` (string, optional): File to check
- `project_wide` (boolean, optional): Check entire project

**Returns:** Array of undefined usages

#### `check_method_exists`
Check if method exists on type.

**Parameters:**
- `type_name` (string, required): Type name
- `method_name` (string, required): Method name

**Returns:** Validation result

#### `calculate_type_safety_score`
Calculate type safety score.

**Parameters:**
- `file_path` (string, optional): File to analyze
- `project_wide` (boolean, optional): Analyze entire project

**Returns:** Type safety score and rating

---

## 💡 Usage Examples

### Example 1: Basic Code Search

**Scenario**: Find all React components

**Prompt to Claude:**
```
Search for all React components in the project
```

**Claude uses:**
```json
search_symbols({
  "query": "Component",
  "type": "class",
  "file_pattern": "*.tsx"
})
```

### Example 2: Refactoring Analysis

**Scenario**: Rename a function safely

**Prompt to Claude:**
```
I want to rename getUserById to findUserById.
What will be affected?
```

**Claude uses:**
```json
analyze_change_impact({
  "change_type": "rename",
  "target": "getUserById",
  "new_value": "findUserById"
})
```

**Response:**
```
Impact Analysis:
- 12 files will be affected
- 45 references need updating
- Risk level: LOW
- Auto-fix available: YES
- Estimated time: 30 seconds
```

### Example 3: Type Validation

**Scenario**: Check for type errors

**Prompt to Claude:**
```
Check this file for type errors: src/services/user.ts
```

**Claude uses:**
```json
validate_file_types({
  "file_path": "src/services/user.ts"
})
```

**Response:**
```
Type Validation Results:
✗ Line 23: Undefined type 'UserRole'
  Suggestion: Did you mean 'UserRoleEnum'?

✓ Type safety score: 87.5/100 (Good)
```

### Example 4: Dependency Analysis

**Scenario**: Understand dependencies

**Prompt to Claude:**
```
Show me what depends on the UserService class
```

**Claude uses:**
```json
get_symbol_dependents({
  "symbol_name": "UserService"
})
```

**Response:**
```
Dependencies on UserService:
- UserController (uses UserService)
- AuthService (uses UserService)
- AdminDashboard (uses UserService)
- 15 test files

Total: 18 dependents
```

### Example 5: Framework-Specific Analysis

**Scenario**: Analyze React component

**Prompt to Claude:**
```
Analyze the UserProfile component for best practices
```

**Claude uses:**
```json
get_symbol_details({
  "symbol_name": "UserProfile"
})
```

**Response:**
```
UserProfile Component Analysis:

Type: Function Component
Framework: React

Hooks Used:
- useState (user data)
- useEffect (fetch user)
- useContext (theme)

Props:
- userId: number
- onUpdate: (user) => void

Issues Found:
⚠ Missing key in list rendering (line 45)
⚠ Direct state mutation detected (line 67)

Suggestions:
1. Add key prop to list items
2. Use setter function for state updates
3. Consider useMemo for expensive calculations
```

---

## 🔧 Troubleshooting

### Common Issues

#### 1. MCP Server Not Starting

**Symptom**: Claude says "CodeIndexer not available"

**Solutions:**
```bash
# Check if server is installed
which codeindexer

# Test manually
codeindexer --test

# Check Claude config file
cat "~/Library/Application Support/Claude/claude_desktop_config.json"

# Restart Claude Desktop
```

#### 2. Database Errors

**Symptom**: "Failed to open database"

**Solutions:**
```bash
# Check database path exists
ls -la /path/to/.codeindex.db

# Recreate database
rm .codeindex.db
codeindexer --db .codeindex.db

# Check permissions
chmod 644 .codeindex.db
```

#### 3. Slow Indexing

**Symptom**: Indexing takes too long

**Solutions:**
```yaml
# In .codeindexer.yml
exclude:
  - node_modules
  - vendor
  - dist
  - "*.min.js"

performance:
  workers: 16  # Increase workers
  batch_size: 200
```

#### 4. Memory Issues

**Symptom**: Out of memory errors

**Solutions:**
```bash
# Reduce batch size
codeindexer --batch-size 50

# Exclude large directories
codeindexer --exclude node_modules,vendor,dist
```

### Debug Mode

Enable debug logging:

```bash
# In Claude config
{
  "mcpServers": {
    "codeindexer": {
      "command": "/usr/local/bin/codeindexer",
      "args": ["--debug", "--log-file", "/tmp/codeindexer.log"]
    }
  }
}

# View logs
tail -f /tmp/codeindexer.log
```

---

## ✅ Best Practices

### 1. Project Setup

```bash
# Navigate to project root
cd /path/to/your/project

# Create exclusion config
cat > .codeindexer.yml <<EOF
exclude:
  - node_modules
  - vendor
  - .git
  - dist
EOF

# Initial index
codeindexer --db .codeindex.db
```

### 2. Regular Maintenance

```bash
# Re-index after major changes
codeindexer --db .codeindex.db --reindex

# Clean up old database
rm .codeindex.db
codeindexer --db .codeindex.db
```

### 3. Performance Optimization

- ✅ Exclude unnecessary directories
- ✅ Use `.codeindexer.yml` for configuration
- ✅ Enable file watching for real-time updates
- ✅ Increase workers for large projects

### 4. Claude Prompts

**Good prompts:**
- ✅ "Find all classes that extend BaseService"
- ✅ "Show me the dependencies of UserController"
- ✅ "What would break if I rename this function?"
- ✅ "Analyze this file for type errors"

**Avoid:**
- ❌ "Show me everything" (too broad)
- ❌ Generic searches without context

### 5. Security

```yaml
# Exclude sensitive files
exclude:
  - "*.env"
  - "*.key"
  - "*.pem"
  - secrets/
  - credentials/
```

---

## 📊 Performance Tips

### Indexing Speed

| Project Size | Time | Memory |
|--------------|------|--------|
| Small (< 1K files) | < 5s | 30MB |
| Medium (1-10K files) | 10-30s | 50MB |
| Large (10-50K files) | 1-3min | 100MB |
| Huge (50K+ files) | 5-10min | 200MB |

### Optimization Strategies

1. **Exclude build artifacts**: `dist`, `build`, `.next`
2. **Exclude dependencies**: `node_modules`, `vendor`, `.venv`
3. **Increase workers**: Use `--workers 16` for large projects
4. **Use SSD**: SQLite performs better on SSDs

---

## 🆘 Getting Help

- **Documentation**: See [full docs](./README.md)
- **Issues**: [GitHub Issues](https://github.com/aaamil13/CodeIndexerMCP/issues)
- **Discussions**: [GitHub Discussions](https://github.com/aaamil13/CodeIndexerMCP/discussions)

---

## 📚 Additional Resources

- [MCP Specification](https://spec.modelcontextprotocol.io/)
- [Claude Desktop](https://claude.ai/desktop)
- [LSP Integration](./LSP_SERVER.md)
- [Parser Architecture](./PARSER_ARCHITECTURE.md)
- [AI Features Guide](./AI_FEATURES.md)

---

**Built with ❤️ for AI-powered development**

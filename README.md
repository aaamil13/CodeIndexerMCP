# Code Indexer MCP 🚀

Intelligent code indexer with Model Context Protocol (MCP) integration for AI agents. Built in **Go** for maximum performance and efficiency.

## ✨ Features

- 🔍 **Real-time code indexing** - Automatically scans and indexes project structure
- 🧠 **AI-optimized** - Designed specifically for AI agent consumption via MCP
- 🌐 **Multi-language** - Supports Go, TypeScript, JavaScript, Python, Rust, SQL
- 🔌 **Plugin system** - Easy to add new language parsers
- 📊 **Rich analysis** - Extracts functions, classes, dependencies, inheritance
- ⚡ **Fast incremental updates** - Only reindexes changed files
- 💾 **SQLite storage** - Efficient local database in `.projectIndex/`
- 🚀 **Native performance** - Written in Go with goroutines for parallel processing
- 📦 **Single binary** - No runtime dependencies

## 🎯 Why Go?

- **10-100x faster** than Node.js for file I/O and parsing
- **8x less memory** footprint
- **Single binary** - no npm install, no node_modules
- **Excellent concurrency** - goroutines for parallel file parsing
- **Cross-platform** - works on Linux, macOS, Windows

## MCP Tools for AI Agents

- `search_symbols` - Find functions, classes, methods
- `get_file_structure` - Get structure of a file
- `get_symbol_details` - Detailed symbol information
- `find_references` - Find all uses of a symbol
- `get_dependencies` - Analyze dependencies
- `get_inheritance_tree` - Class hierarchy
- `get_call_hierarchy` - Function call graphs
- `search_code` - Semantic code search
- `get_project_overview` - Project statistics
- `analyze_complexity` - Code complexity metrics
- `find_similar_code` - Code clone detection
- `get_api_endpoints` - API endpoint extraction

## Installation

```bash
npm install -g code-indexer-mcp
```

## Usage

### CLI

```bash
# Index current directory
code-indexer index

# Watch for changes
code-indexer watch

# Search symbols
code-indexer search "myFunction"

# Get project overview
code-indexer overview
```

### MCP Server

```bash
# Start MCP server
code-indexer mcp

# Or with custom port
code-indexer mcp --port 3000
```

### Programmatic

```typescript
import { CodeIndexer } from 'code-indexer-mcp';

const indexer = new CodeIndexer('/path/to/project');
await indexer.initialize();
await indexer.indexAll();

// Search symbols
const results = await indexer.searchSymbols('MyClass');

// Watch for changes
indexer.watch();
```

## Configuration

Create `.projectIndex/config.json`:

```json
{
  "exclude": ["node_modules", "dist", ".git"],
  "languages": ["typescript", "python", "go"],
  "indexOnSave": true,
  "plugins": []
}
```

## Plugin Development

```typescript
import { LanguageParser, IndexerPlugin } from 'code-indexer-mcp';

export class MyLanguageParser implements LanguageParser {
  language = 'mylang';
  extensions = ['.my'];

  parse(content: string, filePath: string) {
    // Your parsing logic
  }
}

export const plugin: IndexerPlugin = {
  name: 'mylang-plugin',
  version: '1.0.0',
  parser: new MyLanguageParser()
};
```

## Architecture

See [PLAN.md](./PLAN.md) for detailed architecture and implementation plan.

## License

MIT

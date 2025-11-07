# CodeIndexer LSP Server

This document describes the Language Server Protocol (LSP) implementation for CodeIndexer, which enables IDE integration for advanced code intelligence features.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [IDE Configuration](#ide-configuration)
- [Supported Features](#supported-features)
- [Custom AI Features](#custom-ai-features)
- [Architecture](#architecture)
- [Development](#development)

## Overview

The CodeIndexer LSP Server provides intelligent code analysis and navigation features for multiple programming languages through the Language Server Protocol. It leverages CodeIndexer's powerful indexing, semantic analysis, and AI capabilities to provide:

- **Cross-file intelligence**: Navigate and search across your entire codebase
- **Framework awareness**: Understands React, Django, Flask and other frameworks
- **Type checking**: Advanced type validation and inference
- **AI-powered features**: Semantic analysis, dependency graphs, impact analysis

## Features

### Standard LSP Features

✅ **Text Document Synchronization**
- Document open/close/change tracking
- Incremental and full sync modes

✅ **Code Navigation**
- Go to Definition
- Find References
- Document Symbols
- Workspace Symbol Search

✅ **Code Intelligence**
- Auto-completion with context
- Hover information
- Signature help
- Rename symbol across files

✅ **Diagnostics** (Coming Soon)
- Syntax errors
- Type errors
- Undefined references
- Code quality warnings

✅ **Code Actions** (Coming Soon)
- Quick fixes
- Refactoring suggestions
- Auto-imports

### AI-Powered Custom Features

🤖 **Semantic Analysis** (`codeindexer/analyze`)
- Project-wide semantic analysis
- Type errors and undefined references
- Circular dependencies
- Code quality scores

🤖 **Type Checking** (`codeindexer/typeCheck`)
- Advanced type validation
- Method existence checking
- Type inference with confidence
- Smart typo suggestions

🤖 **Call Graph** (`codeindexer/callGraph`)
- Function/method call relationships
- Call count analysis
- Recursive function detection

🤖 **Dependency Analysis** (`codeindexer/dependencies`)
- Module dependency graphs
- Circular dependency detection
- Coupling metrics

🤖 **Find Unused Symbols** (`codeindexer/findUnused`)
- Detect unused functions, variables, classes
- Export analysis
- Dead code detection

## Installation

### Build from Source

```bash
# Clone repository
git clone https://github.com/aaamil13/CodeIndexerMCP.git
cd CodeIndexerMCP

# Build LSP server
go build -o codeindexer-lsp cmd/lsp/main.go

# Install (optional)
sudo mv codeindexer-lsp /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/aaamil13/CodeIndexerMCP/cmd/lsp@latest
```

## IDE Configuration

### VS Code

1. Install the generic LSP client extension (or create a custom extension)
2. Configure in `.vscode/settings.json`:

```json
{
  "languageServerExample.trace.server": "verbose",
  "lsp-sample.server": {
    "command": "codeindexer-lsp",
    "args": ["-db", "./.codeindex.db"],
    "languages": [
      "go",
      "python",
      "javascript",
      "typescript",
      "java",
      "c",
      "cpp",
      "csharp",
      "rust"
    ]
  }
}
```

**Custom Extension (Recommended)**

Create a VS Code extension for CodeIndexer:

```json
// package.json
{
  "name": "codeindexer-lsp",
  "displayName": "CodeIndexer LSP",
  "version": "1.0.0",
  "publisher": "your-name",
  "engines": {
    "vscode": "^1.60.0"
  },
  "activationEvents": ["*"],
  "main": "./out/extension.js",
  "contributes": {
    "configuration": {
      "type": "object",
      "title": "CodeIndexer",
      "properties": {
        "codeindexer.serverPath": {
          "type": "string",
          "default": "codeindexer-lsp",
          "description": "Path to CodeIndexer LSP server"
        },
        "codeindexer.databasePath": {
          "type": "string",
          "default": "./.codeindex.db",
          "description": "Path to CodeIndexer database"
        }
      }
    }
  }
}
```

```typescript
// src/extension.ts
import * as path from 'path';
import { workspace, ExtensionContext } from 'vscode';
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind
} from 'vscode-languageclient/node';

let client: LanguageClient;

export function activate(context: ExtensionContext) {
  const config = workspace.getConfiguration('codeindexer');
  const serverPath = config.get<string>('serverPath') || 'codeindexer-lsp';
  const dbPath = config.get<string>('databasePath') || './.codeindex.db';

  const serverOptions: ServerOptions = {
    command: serverPath,
    args: ['-db', dbPath],
    transport: TransportKind.stdio
  };

  const clientOptions: LanguageClientOptions = {
    documentSelector: [
      { scheme: 'file', language: 'go' },
      { scheme: 'file', language: 'python' },
      { scheme: 'file', language: 'javascript' },
      { scheme: 'file', language: 'typescript' },
      { scheme: 'file', language: 'java' },
      { scheme: 'file', language: 'c' },
      { scheme: 'file', language: 'cpp' },
      { scheme: 'file', language: 'csharp' },
      { scheme: 'file', language: 'rust' }
    ],
    synchronize: {
      fileEvents: workspace.createFileSystemWatcher('**/*')
    }
  };

  client = new LanguageClient(
    'codeindexer',
    'CodeIndexer LSP',
    serverOptions,
    clientOptions
  );

  client.start();
}

export function deactivate(): Thenable<void> | undefined {
  if (!client) {
    return undefined;
  }
  return client.stop();
}
```

### Neovim

Using `nvim-lspconfig`:

```lua
-- ~/.config/nvim/lua/lsp/codeindexer.lua
local configs = require 'lspconfig.configs'
local util = require 'lspconfig.util'

if not configs.codeindexer then
  configs.codeindexer = {
    default_config = {
      cmd = { 'codeindexer-lsp', '-db', './.codeindex.db' },
      filetypes = {
        'go', 'python', 'javascript', 'typescript',
        'java', 'c', 'cpp', 'rust', 'csharp'
      },
      root_dir = util.root_pattern('.git', '.codeindex.db'),
      settings = {},
    },
  }
end

-- Enable for all supported languages
require('lspconfig').codeindexer.setup {
  on_attach = function(client, bufnr)
    -- Custom keybindings
    local opts = { noremap=true, silent=true, buffer=bufnr }
    vim.keymap.set('n', 'gd', vim.lsp.buf.definition, opts)
    vim.keymap.set('n', 'gr', vim.lsp.buf.references, opts)
    vim.keymap.set('n', 'K', vim.lsp.buf.hover, opts)
    vim.keymap.set('n', '<leader>rn', vim.lsp.buf.rename, opts)

    -- Custom AI commands
    vim.keymap.set('n', '<leader>ca', function()
      vim.lsp.buf.execute_command({
        command = 'codeindexer/analyze',
        arguments = { { uri = vim.uri_from_bufnr(0) } }
      })
    end, opts)
  end,
  capabilities = require('cmp_nvim_lsp').default_capabilities(),
}
```

### Emacs

Using `lsp-mode`:

```elisp
;; ~/.emacs.d/init.el
(use-package lsp-mode
  :hook ((go-mode python-mode typescript-mode) . lsp)
  :commands lsp
  :config
  (lsp-register-client
   (make-lsp-client :new-connection (lsp-stdio-connection
                                     '("codeindexer-lsp" "-db" "./.codeindex.db"))
                    :major-modes '(go-mode python-mode typescript-mode
                                  javascript-mode java-mode c-mode c++-mode
                                  rust-mode csharp-mode)
                    :server-id 'codeindexer)))
```

### Sublime Text

Using LSP package:

```json
// Packages/User/LSP.sublime-settings
{
  "clients": {
    "codeindexer": {
      "command": ["codeindexer-lsp", "-db", "./.codeindex.db"],
      "enabled": true,
      "languages": [
        {
          "languageId": "go",
          "scopes": ["source.go"],
          "syntaxes": ["Packages/Go/Go.sublime-syntax"]
        },
        {
          "languageId": "python",
          "scopes": ["source.python"],
          "syntaxes": ["Packages/Python/Python.sublime-syntax"]
        }
      ]
    }
  }
}
```

### Vim (with vim-lsp)

```vim
" ~/.vimrc
if executable('codeindexer-lsp')
  au User lsp_setup call lsp#register_server({
    \ 'name': 'codeindexer',
    \ 'cmd': {server_info->['codeindexer-lsp', '-db', './.codeindex.db']},
    \ 'allowlist': ['go', 'python', 'javascript', 'typescript', 'java', 'c', 'cpp', 'rust'],
    \ })
endif

" Keybindings
function! s:on_lsp_buffer_enabled() abort
    setlocal omnifunc=lsp#complete
    nmap <buffer> gd <plug>(lsp-definition)
    nmap <buffer> gr <plug>(lsp-references)
    nmap <buffer> K <plug>(lsp-hover)
    nmap <buffer> <leader>rn <plug>(lsp-rename)
endfunction

augroup lsp_install
    au!
    autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```

## Supported Features

### By Language

| Feature | Go | Python | JS/TS | Java | Rust | C/C++ | C# |
|---------|-------|--------|-------|------|------|-------|-----|
| Completion | ✅ | ✅ | ✅ | 🔄 | 🔄 | 🔄 | 🔄 |
| Hover | ✅ | ✅ | ✅ | 🔄 | 🔄 | 🔄 | 🔄 |
| Definition | ✅ | ✅ | ✅ | 🔄 | 🔄 | 🔄 | 🔄 |
| References | ✅ | ✅ | ✅ | 🔄 | 🔄 | 🔄 | 🔄 |
| Symbols | ✅ | ✅ | ✅ | 🔄 | 🔄 | 🔄 | 🔄 |
| Rename | ✅ | ✅ | ✅ | 🔄 | 🔄 | 🔄 | 🔄 |
| Diagnostics | 🔄 | 🔄 | 🔄 | 🔄 | 🔄 | 🔄 | 🔄 |

✅ = Fully implemented
🔄 = In progress
❌ = Not yet implemented

## Custom AI Features

### Semantic Analysis

Request full semantic analysis for a file/project:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "codeindexer/analyze",
  "params": {
    "uri": "file:///path/to/file.go"
  }
}
```

Response includes:
- Type errors and mismatches
- Undefined references
- Unused symbols
- Circular dependencies
- Quality score (0-100)

### Type Checking

Request type validation for a file:

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "codeindexer/typeCheck",
  "params": {
    "uri": "file:///path/to/file.py"
  }
}
```

Response includes:
- Undefined symbols with typo suggestions
- Type mismatches
- Missing methods
- Type safety score

### Call Graph

Request call graph for a project:

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "codeindexer/callGraph",
  "params": {
    "projectId": 1
  }
}
```

Response includes:
- Function/method nodes
- Call edges with counts
- Recursive functions
- Most called functions

## Architecture

### LSP Server Components

```
┌─────────────────────────────────────────┐
│         LSP Server (stdio)              │
├─────────────────────────────────────────┤
│  Message Handler                        │
│  - Initialize / Shutdown                │
│  - Document Sync                        │
│  - Language Features                    │
├─────────────────────────────────────────┤
│  Feature Handlers                       │
│  - Completion                           │
│  - Hover                                │
│  - Definition / References              │
│  - Symbols                              │
│  - Rename                               │
├─────────────────────────────────────────┤
│  CodeIndexer Integration                │
│  - Database queries                     │
│  - Indexer                              │
│  - Semantic Analyzer                    │
│  - Type Validator                       │
└─────────────────────────────────────────┘
```

### Communication Flow

```
IDE ←→ LSP Server ←→ CodeIndexer DB
                ↓
          Semantic Analyzer
                ↓
          Type Validator
```

### Database Indexing

The LSP server automatically indexes your workspace when it starts and re-indexes files as they change. For large codebases, consider:

1. **Pre-indexing**: Run `codeindexer index` before starting LSP server
2. **Incremental updates**: LSP server only re-indexes changed files
3. **Background indexing**: Indexing happens asynchronously

## Development

### Building

```bash
# Build LSP server
go build -o codeindexer-lsp cmd/lsp/main.go

# Run with debug logging
./codeindexer-lsp -debug -db ./test.db
```

### Testing

Test the LSP server manually using `nc` or `telnet`:

```bash
# Start server
./codeindexer-lsp -debug -db ./test.db

# Send initialize request (paste JSON)
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "processId": null,
    "rootUri": "file:///path/to/project",
    "capabilities": {}
  }
}
```

### Adding New Features

1. **Add method handler** in `server.go`:
```go
func (s *Server) handleYourFeature(msg *Message) (interface{}, error) {
    // Implementation
}
```

2. **Register in handleMessage**:
```go
case "textDocument/yourFeature":
    return s.handleYourFeature(msg)
```

3. **Update capabilities** in `NewServer`:
```go
capabilities: ServerCapabilities{
    YourFeatureProvider: true,
}
```

## Performance

- **Startup**: ~100ms (depends on database size)
- **Completion**: <50ms
- **Go to Definition**: <10ms
- **Find References**: <100ms (depends on reference count)
- **Semantic Analysis**: 100ms - 2s (depends on project size)

### Optimization Tips

1. **Use SSD**: Database performance is I/O bound
2. **Pre-index**: Index before starting LSP server
3. **Exclude directories**: Ignore `node_modules`, `vendor`, etc.
4. **Incremental indexing**: Only changed files are re-indexed

## Troubleshooting

### LSP Server Not Starting

```bash
# Check if server is accessible
which codeindexer-lsp

# Test manually
codeindexer-lsp -debug -db ./test.db

# Check logs (VS Code)
Output → Select "CodeIndexer LSP" from dropdown
```

### Completion Not Working

1. Ensure file is indexed: Check database
2. Check language support: See supported languages
3. Verify LSP connection: Check IDE's LSP status

### Slow Performance

1. Check database size: `ls -lh .codeindex.db`
2. Optimize database: `VACUUM;` in SQLite
3. Reduce indexed files: Add exclusions
4. Pre-index workspace: Run `codeindexer index` first

## Future Enhancements

- [ ] Real-time diagnostics (errors, warnings)
- [ ] Code actions and quick fixes
- [ ] Auto-imports
- [ ] Code formatting
- [ ] Refactoring support
- [ ] Workspace-wide search/replace
- [ ] Call hierarchy
- [ ] Type hierarchy
- [ ] Code lens (references count)
- [ ] Inlay hints (type annotations)
- [ ] Semantic tokens (better syntax highlighting)

## References

- [Language Server Protocol Specification](https://microsoft.github.io/language-server-protocol/)
- [LSP Implementations](https://langserver.org/)
- [Go LSP Libraries](https://pkg.go.dev/golang.org/x/tools/gopls)

## Contributing

To contribute to the LSP server:

1. Fork the repository
2. Create feature branch
3. Add tests for new features
4. Submit pull request

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

See [LICENSE](LICENSE) for details.

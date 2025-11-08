# Parser Architecture

This document describes the extensible parser architecture in CodeIndexerMCP, which supports **18+ programming languages** and **framework-specific analysis**.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Supported Languages](#supported-languages)
- [Framework Analyzers](#framework-analyzers)
- [Adding New Languages](#adding-new-languages)
- [Adding Framework Analyzers](#adding-framework-analyzers)
- [Tree-sitter Integration](#tree-sitter-integration)

## Overview

CodeIndexerMCP uses a **plugin-based architecture** for parsers and framework analyzers:

1. **Parser Plugins**: Handle language-specific parsing (syntax, symbols, imports)
2. **Framework Analyzers**: Add framework-specific knowledge on top of parsers
3. **Tree-sitter Integration**: Use Tree-sitter for robust, fast parsing
4. **Configuration Parsers**: Parse config files (JSON, YAML, TOML, XML, etc.)

### Key Principles

✅ **Modular**: Each parser is an independent plugin
✅ **Extensible**: Easy to add new languages via plugin interface
✅ **Framework-Aware**: Separate analyzers for React, Django, Spring, etc.
✅ **LSP-Compatible**: Design follows Language Server Protocol patterns
✅ **Performance**: Tree-sitter based parsers are fast and incremental

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────┐
│           CodeIndexerMCP Core                │
└──────────────────┬──────────────────────────┘
                   │
          ┌────────▼─────────┐
          │  Parser Registry  │
          └────────┬──────────┘
                   │
     ┌─────────────┼─────────────┐
     │             │             │
┌────▼────┐  ┌────▼────┐  ┌────▼────┐
│ Parser  │  │Framework│  │ Tree-   │
│ Plugins │  │Analyzers│  │sitter   │
└─────────┘  └─────────┘  └─────────┘
```

### Parser Plugin Interface

```go
type ParserPlugin interface {
    Language() string
    Extensions() []string
    Parse(content []byte, filePath string) (*ParseResult, error)
    SupportsFramework(framework string) bool
    Priority() int
}
```

### Framework Analyzer Interface

```go
type FrameworkAnalyzer interface {
    Framework() string
    Language() string
    Analyze(result *ParseResult, content []byte) (*FrameworkInfo, error)
    DetectFramework(content []byte, filePath string) bool
}
```

## Supported Languages

### Programming Languages (18+)

| Language | Status | Parser Type | Extensions |
|----------|--------|-------------|------------|
| **Python** | ✅ Active | Regex + Tree-sitter | `.py`, `.pyi` |
| **Go** | ✅ Active | Native go/ast | `.go` |
| **JavaScript** | 🔄 Tree-sitter | Tree-sitter | `.js`, `.jsx`, `.mjs` |
| **TypeScript** | 🔄 Tree-sitter | Tree-sitter | `.ts`, `.tsx` |
| **Java** | 🔄 Tree-sitter | Tree-sitter | `.java` |
| **C#** | 🔄 Tree-sitter | Tree-sitter | `.cs` |
| **C** | 🔄 Tree-sitter | Tree-sitter | `.c`, `.h` |
| **C++** | 🔄 Tree-sitter | Tree-sitter | `.cpp`, `.hpp`, `.cc` |
| **PHP** | 🔄 Tree-sitter | Tree-sitter | `.php` |
| **Ruby** | 🔄 Tree-sitter | Tree-sitter | `.rb` |
| **Rust** | 🔄 Tree-sitter | Tree-sitter | `.rs` |
| **Kotlin** | 🔄 Tree-sitter | Tree-sitter | `.kt`, `.kts` |
| **Swift** | 🔄 Tree-sitter | Tree-sitter | `.swift` |
| **Shell/Bash** | 🔄 Tree-sitter | Tree-sitter | `.sh`, `.bash` |
| **PowerShell** | 🔄 Planned | Tree-sitter | `.ps1` |
| **SQL** | 🔄 Tree-sitter | Tree-sitter | `.sql` |
| **HTML** | 🔄 Tree-sitter | Tree-sitter | `.html`, `.htm` |
| **CSS** | 🔄 Tree-sitter | Tree-sitter | `.css`, `.scss` |

### Configuration & Markup Languages

| Format | Status | Parser Type | Extensions |
|--------|--------|-------------|------------|
| **JSON** | ✅ Active | encoding/json | `.json` |
| **YAML** | ✅ Active | Custom | `.yaml`, `.yml` |
| **TOML** | ✅ Active | Custom | `.toml` |
| **XML** | ✅ Active | encoding/xml | `.xml` |
| **Markdown** | ✅ Active | Custom | `.md`, `.markdown` |

## Framework Analyzers

Framework analyzers add framework-specific knowledge on top of language parsers.

### Supported Frameworks (Planned)

#### Frontend

- **React** (JavaScript/TypeScript)
  - Components, Hooks, Props, State
  - JSX/TSX syntax
  - React Router routes

- **Vue** (JavaScript/TypeScript)
  - Components, Directives, Computed
  - Single File Components (.vue)

- **Angular** (TypeScript)
  - Components, Services, Modules
  - Dependency Injection
  - Decorators

#### Backend

- **Django** (Python)
  - Models, Views, URLs
  - ORM relationships
  - Admin configuration

- **Flask** (Python)
  - Routes, Blueprints
  - Request handlers

- **Express** (JavaScript/TypeScript)
  - Routes, Middleware
  - REST endpoints

- **Spring** (Java)
  - Controllers, Services, Repositories
  - Dependency Injection
  - Annotations

- **ASP.NET** (C#)
  - Controllers, Models, Views
  - Entity Framework

- **Rails** (Ruby)
  - Models, Controllers, Routes
  - ActiveRecord

#### Mobile

- **Android** (Java/Kotlin)
  - Activities, Fragments, Services
  - Lifecycle methods

- **iOS** (Swift)
  - ViewControllers, Views
  - SwiftUI components

### Framework Detection

Frameworks are detected automatically by:
1. Import/require statements
2. File naming conventions
3. Project structure
4. Configuration files

Example:
```python
# Detects Django
from django.db import models
from django.views import View

class User(models.Model):  # <- Detected as Django Model
    name = models.CharField(max_length=100)
```

## Adding New Languages

### Step 1: Create Parser Plugin

```go
// internal/parsers/mylang/mylang.go
package mylang

import (
    "github.com/aaamil13/CodeIndexerMCP/internal/parser"
)

type MyLangParser struct {
    *parser.BaseParser
}

func NewMyLangParser() *MyLangParser {
    return &MyLangParser{
        BaseParser: parser.NewBaseParser(
            "mylang",
            []string{".ml"},
            100, // priority
        ),
    }
}

func (p *MyLangParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
    // Your parsing logic here
    result := &types.ParseResult{
        Symbols: make([]*types.Symbol, 0),
        Imports: make([]*types.Import, 0),
    }

    // Extract symbols, imports, relationships

    return result, nil
}
```

### Step 2: Register Parser

```go
// In parser registry initialization
registry.RegisterParser(mylang.NewMyLangParser())
```

### Step 3: Add Tests

```go
// internal/parsers/mylang/mylang_test.go
func TestMyLangParser(t *testing.T) {
    parser := NewMyLangParser()
    code := []byte(`your test code`)

    result, err := parser.Parse(code, "test.ml")
    if err != nil {
        t.Fatalf("Parse failed: %v", err)
    }

    // Assert symbols, imports, etc.
}
```

## Adding Framework Analyzers

### Step 1: Create Framework Analyzer

```go
// internal/parsers/analyzers/myframework/myframework.go
package myframework

type MyFrameworkAnalyzer struct{}

func NewMyFrameworkAnalyzer() *MyFrameworkAnalyzer {
    return &MyFrameworkAnalyzer{}
}

func (a *MyFrameworkAnalyzer) Framework() string {
    return "myframework"
}

func (a *MyFrameworkAnalyzer) Language() string {
    return "python" // or whatever language
}

func (a *MyFrameworkAnalyzer) DetectFramework(content []byte, filePath string) bool {
    // Check for framework imports/patterns
    return strings.Contains(string(content), "import myframework")
}

func (a *MyFrameworkAnalyzer) Analyze(result *types.ParseResult, content []byte) (*types.FrameworkInfo, error) {
    info := &types.FrameworkInfo{
        Name:       "myframework",
        Type:       "backend",
        Components: make([]*types.FrameworkComponent, 0),
        Routes:     make([]*types.Route, 0),
    }

    // Analyze framework-specific patterns
    // Extract components, routes, models, etc.

    return info, nil
}
```

### Step 2: Register Analyzer

```go
registry.RegisterFrameworkAnalyzer(myframework.NewMyFrameworkAnalyzer())
```

## Tree-sitter Integration

Tree-sitter provides fast, incremental parsing with error recovery.

### Using Tree-sitter

1. **Install Tree-sitter grammar** for your language
2. **Create queries** for symbol extraction
3. **Wrap in parser plugin**

Example query file (`queries/typescript/symbols.scm`):

```scheme
; Functions
(function_declaration
  name: (identifier) @name) @function

; Classes
(class_declaration
  name: (type_identifier) @name) @class

; Methods
(method_definition
  name: (property_identifier) @name) @method
```

### Tree-sitter Parser Template

```go
import (
    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/typescript"
)

type TreeSitterTSParser struct {
    parser *sitter.Parser
    query  *sitter.Query
}

func NewTreeSitterTSParser() *TreeSitterTSParser {
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Load queries
    query, _ := sitter.NewQuery([]byte(symbolQuery), typescript.GetLanguage())

    return &TreeSitterTSParser{
        parser: parser,
        query:  query,
    }
}

func (p *TreeSitterTSParser) Parse(content []byte, filePath string) (*types.ParseResult, error) {
    tree := p.parser.Parse(nil, content)
    defer tree.Close()

    // Execute queries and extract symbols
    cursor := sitter.NewQueryCursor()
    defer cursor.Close()

    matches := cursor.Matches(p.query, tree.RootNode(), content)

    // Process matches and build ParseResult

    return result, nil
}
```

## Parser Priority System

When multiple parsers can handle a file, priority determines which is used:

| Priority | Parser Type | Example |
|----------|-------------|---------|
| 150 | Native language tools | go/ast for Go |
| 100 | Tree-sitter parsers | TypeScript, Java, Rust |
| 50 | Custom/regex parsers | Python (current) |
| 10 | Fallback parsers | Generic text parser |

## Performance Considerations

### Tree-sitter Benefits

✅ **Fast**: ~1ms for typical files
✅ **Incremental**: Re-parse only changed parts
✅ **Error Recovery**: Parses even with syntax errors
✅ **Memory Efficient**: Streaming parse

### Optimization Tips

1. **Cache parse results**: Don't re-parse unchanged files
2. **Parallel parsing**: Use worker pools for multiple files
3. **Lazy analysis**: Run framework analyzers only when needed
4. **Query optimization**: Write efficient Tree-sitter queries

## LSP Compatibility

The architecture follows LSP patterns for future integration:

```go
// LSP-style capabilities
type ParserCapabilities struct {
    SymbolProvider       bool
    CompletionProvider   bool
    DefinitionProvider   bool
    ReferenceProvider    bool
    HoverProvider        bool
    CodeActionProvider   bool
}
```

## Configuration

Parser behavior can be configured via `.code-indexer.yaml`:

```yaml
parsers:
  go:
    enabled: true
    priority: 150

  typescript:
    enabled: true
    tree_sitter: true
    frameworks:
      - react
      - vue

  python:
    enabled: true
    frameworks:
      - django
      - flask
```

## Future Enhancements

- [ ] Complete Tree-sitter integration for all languages
- [ ] Framework analyzer for React, Django, Spring
- [ ] Semantic analysis (type checking, symbol resolution)
- [ ] Cross-language analysis (e.g., Python calling Go via CGo)
- [ ] IDE protocol support (LSP server mode)
- [ ] Incremental parsing
- [ ] AST caching
- [ ] Query-based symbol search
- [ ] Code generation from templates

## References

- [Tree-sitter](https://tree-sitter.github.io/tree-sitter/)
- [Language Server Protocol](https://microsoft.github.io/language-server-protocol/)
- [Go AST Package](https://pkg.go.dev/go/ast)
- [Parser Plugin Pattern](https://en.wikipedia.org/wiki/Plugin_(computing))

## Contributing

To add a new language parser:

1. Create parser in `internal/parsers/yourlang/`
2. Implement `ParserPlugin` interface
3. Add tests in `*_test.go`
4. Register in parser registry
5. Add documentation
6. Submit PR with examples

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

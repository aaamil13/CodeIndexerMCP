# CodeIndexerMCP Architecture Analysis

This document provides a detailed overview of the CodeIndexerMCP project's architecture, based on a thorough code review.

## 1. High-Level Overview

CodeIndexerMCP is a sophisticated command-line tool designed to parse and index source code from various programming languages. It creates a local database of symbols (classes, functions, etc.), imports, and their relationships. This index powers several features, including code searching, project analysis, and a suite of advanced AI-driven insights.

The application is structured around a central `Indexer` component that orchestrates the entire process, from file discovery and parsing to database management and AI analysis.

## 2. Core Components

The system is built on a few key components that work together:

### a. The `Indexer` (`internal/core/indexer.go`)

This is the heart of the application. It is responsible for:
- **Initialization**: Setting up the database, parser registry, and AI modules.
- **File Scanning**: Walking the project directory to find relevant source files, respecting `.gitignore` and `.indexerignore` rules.
- **Concurrent Indexing**: Managing a pool of worker goroutines to parse files in parallel for efficiency.
- **Database Interaction**: Saving parsed data (files, symbols, imports) into the SQLite database.
- **Providing an API**: Exposing methods for searching, watching files, and accessing the various AI-powered analysis tools.

### b. The `Database` (`internal/database/db.go`)

The project uses a SQLite database (`.projectIndex/index.db`) to store all indexed information. The schema is designed to capture:
- Project metadata
- File information (path, language, hash)
- Symbols (functions, classes, methods, etc.)
- Imports
- Relationships between symbols (e.g., inheritance, function calls)

### c. The `Parser Registry` (`internal/parser/plugin.go`)

This component is responsible for managing and selecting the correct language parser for each file.
- **Plugin-Based**: It uses a `ParserPlugin` interface, allowing different parsing strategies to be used interchangeably.
- **Extension-Based Routing**: It maps file extensions (e.g., `.py`, `.go`) to a list of capable parsers.
- **Priority System**: It includes a priority mechanism to resolve conflicts when multiple parsers can handle the same file type. This is key to the project's future migration path.

## 3. Parsing Architecture: A Tale of Two Parsers

The most interesting aspect of the architecture is its dual-parser strategy:

- **Current Reality: Regex-Based Parsers**: As seen in `internal/parsers/python/python.go`, the currently active parsers are built using regular expressions. These parsers scan files line-by-line to identify symbols and imports. While functional, this approach is less robust and harder to maintain than a full syntax tree analysis.

- **Future Goal: Tree-sitter Integration**: The file `internal/parser/treesitter/treesitter.go` lays the complete groundwork for a migration to Tree-sitter, a powerful parser generator. It contains configurations for over 20 languages and placeholder structs. However, the core parsing logic is **not yet implemented**. The high priority assigned to the placeholder Tree-sitter parsers ensures that once they are implemented, they will automatically be used instead of the regex-based ones.

This hybrid state indicates a project in transition, with a clear and well-designed path toward a more robust parsing engine.

## 4. AI & Analysis Features

A significant portion of the codebase (`internal/ai/`) is dedicated to advanced code analysis features built on top of the indexed data. These include:
- **Context Extraction**: Understanding the code surrounding a specific symbol.
- **Impact Analysis**: Predicting the effects of a code change.
- **Dependency Graphing**: Visualizing the relationships between different parts of the code.
- **Type Validation**: Finding potential type-related errors in dynamically typed languages.

## 5. Architectural Diagram

The following Mermaid diagram illustrates the flow of data and control within the system.

```mermaid
graph TD
    subgraph "CLI (cmd/code-indexer/main.go)"
        A[index]
        B[watch]
        C[mcp]
        D[search]
    end

    subgraph "Core Engine (internal/core/indexer.go)"
        E[Indexer]
    end

    subgraph "Parsing Subsystem (internal/parser/)"
        F[Parser Registry]
        G[Regex Parsers]
        H[Tree-sitter Facade (Future)]
    end

    subgraph "Data Layer"
        I[SQLite Database]
    end

    subgraph "AI Subsystem (internal/ai/)"
        J[AI Helpers]
    end

    A --> E
    B --> E
    C --> E
    D --> E

    E -- Scans Files --> F
    F -- Selects Parser --> G
    F -- Selects Parser --> H

    G -- Parses --> E
    H -- Parses --> E

    E -- Saves Data --> I
    E -- Reads Data --> I

    J -- Reads Data --> I
    E -- Uses --> J
```

This concludes my analysis of the project's architecture.
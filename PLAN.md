# Code Indexer MCP - Подробен План

## 🎯 Цел на Проекта

Създаване на интелигентен код индексер, който предоставя на AI агенти пълна и структурирана информация за софтуерни проекти чрез Model Context Protocol (MCP).

## 🏗️ Архитектура

### 1. Ядрени Компоненти

```
CodeIndexerMCP/
├── .projectIndex/          # Служебна директория за всеки проект
│   ├── index.db           # SQLite база данни
│   ├── cache/             # Кеш за парсинг резултати
│   └── config.json        # Конфигурация на индексера
├── src/
│   ├── core/              # Основна логика
│   │   ├── indexer.ts     # Главен индексер
│   │   ├── watcher.ts     # File system watcher
│   │   ├── database.ts    # Database операции
│   │   └── analyzer.ts    # Код анализ
│   ├── parsers/           # Language парсери
│   │   ├── base.ts        # Абстрактен базов парсер
│   │   ├── typescript.ts  # TypeScript/JavaScript парсер
│   │   ├── python.ts      # Python парсер
│   │   ├── go.ts          # Go парсер
│   │   ├── rust.ts        # Rust парсер
│   │   └── sql.ts         # SQL парсер
│   ├── plugins/           # Плъгин система
│   │   ├── manager.ts     # Plugin manager
│   │   └── loader.ts      # Dynamic plugin loading
│   ├── mcp/               # MCP сървър
│   │   ├── server.ts      # MCP server
│   │   └── tools.ts       # MCP tools definition
│   └── cli/               # CLI interface
│       └── index.ts
├── plugins/               # Външни плъгини
└── tests/
```

### 2. Database Schema (SQLite)

```sql
-- Проекти
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    language_stats JSON,
    last_indexed TIMESTAMP,
    created_at TIMESTAMP
);

-- Файлове
CREATE TABLE files (
    id INTEGER PRIMARY KEY,
    project_id INTEGER,
    path TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    language TEXT,
    size INTEGER,
    lines_of_code INTEGER,
    hash TEXT,
    last_modified TIMESTAMP,
    last_indexed TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

-- Символи (функции, класове, променливи)
CREATE TABLE symbols (
    id INTEGER PRIMARY KEY,
    file_id INTEGER,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'function', 'class', 'method', 'variable', 'interface', 'type', 'enum'
    signature TEXT,
    parent_id INTEGER, -- За методи в класове
    start_line INTEGER,
    end_line INTEGER,
    visibility TEXT, -- 'public', 'private', 'protected'
    is_exported BOOLEAN,
    is_async BOOLEAN,
    docstring TEXT,
    metadata JSON, -- Допълнителна информация
    FOREIGN KEY (file_id) REFERENCES files(id),
    FOREIGN KEY (parent_id) REFERENCES symbols(id)
);

-- Зависимости/Импорти
CREATE TABLE imports (
    id INTEGER PRIMARY KEY,
    file_id INTEGER,
    source TEXT NOT NULL, -- Импортиран модул/пакет
    imported_symbols JSON, -- Конкретни символи
    import_type TEXT, -- 'local', 'external', 'stdlib'
    line_number INTEGER,
    FOREIGN KEY (file_id) REFERENCES files(id)
);

-- Наследяване и имплементации
CREATE TABLE relationships (
    id INTEGER PRIMARY KEY,
    from_symbol_id INTEGER,
    to_symbol_id INTEGER,
    relationship_type TEXT, -- 'extends', 'implements', 'calls', 'uses'
    FOREIGN KEY (from_symbol_id) REFERENCES symbols(id),
    FOREIGN KEY (to_symbol_id) REFERENCES symbols(id)
);

-- Референции (къде се използва символ)
CREATE TABLE references (
    id INTEGER PRIMARY KEY,
    symbol_id INTEGER,
    file_id INTEGER,
    line_number INTEGER,
    column_number INTEGER,
    reference_type TEXT, -- 'call', 'assignment', 'type_reference'
    FOREIGN KEY (symbol_id) REFERENCES symbols(id),
    FOREIGN KEY (file_id) REFERENCES files(id)
);

-- Индекси за бързо търсене
CREATE INDEX idx_symbols_name ON symbols(name);
CREATE INDEX idx_symbols_type ON symbols(type);
CREATE INDEX idx_files_path ON files(relative_path);
CREATE INDEX idx_imports_source ON imports(source);
```

### 3. MCP Tools за AI Агенти

#### Основни Tools:

1. **`search_symbols`**
   - Търси функции, класове, методи, променливи
   - Параметри: `query`, `type`, `language`, `file_pattern`
   - Връща: Списък със символи, локация, сигнатура

2. **`get_file_structure`**
   - Връща структурата на конкретен файл
   - Параметри: `file_path`
   - Връща: Дървовидна структура с всички символи

3. **`get_symbol_details`**
   - Подробна информация за символ
   - Параметри: `symbol_name`, `file_path`
   - Връща: Сигнатура, документация, използвания

4. **`find_references`**
   - Намира всички референции към символ
   - Параметри: `symbol_name`, `symbol_type`
   - Връща: Списък с файлове и локации

5. **`get_dependencies`**
   - Връща зависимостите на файл или проект
   - Параметри: `file_path`, `include_transitive`
   - Връща: Граф на зависимости

6. **`get_inheritance_tree`**
   - Връща йерархия на наследяване
   - Параметри: `class_name`
   - Връща: Дървовидна структура

7. **`get_call_hierarchy`**
   - Показва кои функции извикват дадена функция
   - Параметри: `function_name`, `direction` (callers/callees)
   - Връща: Граф на извиквания

8. **`search_code`**
   - Семантично търсене в кода
   - Параметри: `query`, `language`, `context`
   - Връща: Релевантни код сегменти

9. **`get_project_overview`**
   - Обща информация за проекта
   - Параметри: няма
   - Връща: Статистики, структура, главни компоненти

10. **`analyze_complexity`**
    - Анализ на сложност на код
    - Параметри: `file_path`, `function_name`
    - Връща: Cyclomatic complexity, cognitive complexity

11. **`find_similar_code`**
    - Намира подобен код (code clones)
    - Параметри: `code_snippet`, `threshold`
    - Връща: Подобни сегменти

12. **`get_api_endpoints`**
    - Извлича API endpoints (REST, GraphQL)
    - Параметри: няма
    - Връща: Списък с endpoints, методи, параметри

### 4. Language Parsers

#### Базов Interface:
```typescript
interface LanguageParser {
    language: string;
    extensions: string[];

    parse(content: string, filePath: string): ParseResult;
    extractSymbols(ast: any): Symbol[];
    extractImports(ast: any): Import[];
    extractRelationships(ast: any): Relationship[];
    getDocumentation(node: any): string | null;
}
```

#### Имплементации:

1. **TypeScript/JavaScript Parser**
   - Използва: `@typescript-eslint/parser` или `@babel/parser`
   - Разпознава: classes, functions, interfaces, types, enums
   - Извлича: JSDoc коментари

2. **Python Parser**
   - Използва: `tree-sitter-python` или `py-ast-parser`
   - Разпознава: classes, functions, decorators, type hints
   - Извлича: docstrings

3. **Go Parser**
   - Използва: `tree-sitter-go`
   - Разпознава: packages, functions, structs, interfaces, methods
   - Извлича: Go doc коментари

4. **Rust Parser**
   - Използва: `tree-sitter-rust`
   - Разпознава: modules, structs, enums, traits, functions, impls
   - Извлича: doc коментари

5. **SQL Parser**
   - Използва: `node-sql-parser`
   - Разпознава: tables, views, procedures, functions, triggers
   - Извлича: DDL/DML структури

### 5. File System Watcher

- Използва `chokidar` за мониторинг
- Incremental indexing - само променени файлове
- Debouncing за batch операции
- Игнорира: `node_modules`, `.git`, `dist`, `.projectIndex`

### 6. Plugin System

```typescript
interface IndexerPlugin {
    name: string;
    version: string;

    // Language parser plugin
    parser?: LanguageParser;

    // Custom analyzer
    analyzer?: (file: File) => AnalysisResult;

    // Custom MCP tools
    tools?: MCPTool[];

    // Lifecycle hooks
    onInit?: () => void;
    onFileIndexed?: (file: File) => void;
}
```

Плъгините се зареждат от:
- `plugins/` директория в проекта
- npm пакети с префикс `code-indexer-plugin-`

## 🚀 Допълнителни Features

### 1. **Code Quality Metrics**
   - Lines of code, complexity
   - Maintainability index
   - Test coverage mapping
   - TODO/FIXME коментари

### 2. **Smart Code Navigation**
   - Go to definition
   - Find all implementations
   - Show type hierarchy
   - Workspace symbols

### 3. **Documentation Generation**
   - Auto-extract API documentation
   - Generate markdown docs
   - OpenAPI/Swagger specs от код

### 4. **Change Impact Analysis**
   - Какви файлове ще се засегнат от промяна
   - Breaking changes detection
   - Dependency update impact

### 5. **Code Patterns Recognition**
   - Design patterns detection
   - Anti-patterns detection
   - Best practices suggestions

### 6. **Multi-Project Support**
   - Workspace с множество проекти
   - Cross-project references
   - Monorepo support

### 7. **Performance Optimization**
   - Parallel parsing
   - Lazy loading
   - Cache strategy
   - Incremental updates

### 8. **Security Analysis**
   - Hardcoded secrets detection
   - Vulnerable dependencies
   - SQL injection patterns

### 9. **AI-Specific Features**
   - Code embeddings за semantic search
   - Summary generation за файлове/класове
   - Intent detection (какво прави този код)
   - Example usage extraction

### 10. **Export/Import**
   - Export index to JSON
   - Share index between machines
   - Integration с други tools

## 📋 Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
- [ ] Project setup (TypeScript, dependencies)
- [ ] Database schema и SQLite integration
- [ ] File system watcher
- [ ] Basic indexer engine

### Phase 2: Core Parsers (Week 3-4)
- [ ] TypeScript/JavaScript parser
- [ ] Python parser
- [ ] Base plugin system

### Phase 3: MCP Integration (Week 5)
- [ ] MCP server setup
- [ ] Core tools implementation
- [ ] Testing с AI agents

### Phase 4: Advanced Parsers (Week 6-7)
- [ ] Go parser
- [ ] Rust parser
- [ ] SQL parser

### Phase 5: Advanced Features (Week 8-9)
- [ ] Code quality metrics
- [ ] Symbol resolution
- [ ] Relationship analysis
- [ ] Performance optimization

### Phase 6: Polish (Week 10)
- [ ] CLI interface
- [ ] Documentation
- [ ] Examples
- [ ] Testing

## 🛠️ Technology Stack

- **Runtime**: Node.js 18+
- **Language**: TypeScript
- **Database**: SQLite (better-sqlite3)
- **Parsers**: tree-sitter, @typescript-eslint/parser
- **File Watching**: chokidar
- **MCP**: @modelcontextprotocol/sdk
- **CLI**: commander, chalk
- **Testing**: Jest, vitest

## 📊 Success Metrics

1. Индексира 10K+ файлове за < 1 минута
2. Incremental updates за < 100ms на файл
3. Точност на symbol resolution > 95%
4. Memory usage < 200MB за average project
5. MCP tool response time < 500ms

## 🎯 Ключови Предимства за AI Агенти

1. **Бързо контекстно разбиране** - агентът получава структурирана информация вместо да чете цели файлове
2. **Точна навигация** - директни референции към файл:ред
3. **Семантично търсене** - намира код по намерение, не само по текст
4. **Relationship awareness** - разбира връзките между компоненти
5. **Real-time updates** - винаги актуална информация
6. **Multi-language support** - работи с цели проекти с mixed languages
7. **Extensible** - лесно добавяне на нови езици и анализи

## 🔄 Next Steps

1. Започваме с TypeScript setup
2. Имплементираме database layer
3. Създаваме TypeScript parser като пилот
4. Интегрираме MCP server
5. Итеративно добавяме features

---

**Този план обхваща всички аспекти за създаване на production-ready код индексер, специално оптимизиран за работа с AI агенти!**

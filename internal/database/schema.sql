CREATE TABLE IF NOT EXISTS symbols (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    file_path TEXT NOT NULL,
    language TEXT NOT NULL,
    type TEXT,
    start_line INTEGER NOT NULL,
    start_column INTEGER NOT NULL,
    start_byte INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    end_column INTEGER NOT NULL,
    end_byte INTEGER NOT NULL,
    content_hash TEXT NOT NULL
);

-- File table
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    path TEXT NOT NULL UNIQUE,
    relative_path TEXT NOT NULL,
    language TEXT NOT NULL,
    size INTEGER NOT NULL,
    lines_of_code INTEGER NOT NULL,
    hash TEXT NOT NULL,
    last_modified TEXT NOT NULL,
    last_indexed TEXT NOT NULL
);

-- Project table
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    language_stats TEXT,
    last_indexed TEXT,
    created_at TEXT NOT NULL
);

-- Таблица за импорти
CREATE TABLE IF NOT EXISTS imports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    import_path TEXT NOT NULL,
    alias TEXT,
    is_wildcard BOOLEAN DEFAULT 0,
    start_line INTEGER
);

-- Таблица за референции между символи
CREATE TABLE IF NOT EXISTS code_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_symbol_id TEXT NOT NULL,
    target_symbol_name TEXT NOT NULL,
    reference_type TEXT, -- 'calls', 'uses', 'instantiates'
    file_path TEXT NOT NULL,
    line INTEGER NOT NULL,
    column INTEGER NOT NULL,
    
    FOREIGN KEY (source_symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);
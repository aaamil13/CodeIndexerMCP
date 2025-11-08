-- Основна таблица за символи
CREATE TABLE IF NOT EXISTS symbols (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,
    file_path TEXT NOT NULL,
    language TEXT NOT NULL,
    signature TEXT,
    documentation TEXT,
    visibility TEXT,
    
    -- Позиция в кода
    start_line INTEGER NOT NULL,
    start_column INTEGER NOT NULL,
    start_byte INTEGER NOT NULL,
    end_line INTEGER NOT NULL,
    end_column INTEGER NOT NULL,
    end_byte INTEGER NOT NULL,
    
    -- 💡 ПОДОБРЕНИЕ #5: Content Hash за Incremental Indexing
    content_hash TEXT NOT NULL,
    
    -- AI Development metadata
    status TEXT DEFAULT 'completed',
    priority INTEGER DEFAULT 0,
    assigned_agent TEXT,
    
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
    
    -- JSON метаданни
    metadata TEXT
);

-- Таблица за функции/методи
CREATE TABLE IF NOT EXISTS functions (
    symbol_id TEXT PRIMARY KEY,
    return_type TEXT,
    is_async BOOLEAN DEFAULT 0,
    is_generator BOOLEAN DEFAULT 0,
    body TEXT,
    receiver_type TEXT, -- За методи
    is_static BOOLEAN DEFAULT 0,
    
    FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);

-- Таблица за параметри
CREATE TABLE IF NOT EXISTS parameters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    function_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT,
    default_value TEXT,
    position INTEGER NOT NULL,
    is_optional BOOLEAN DEFAULT 0,
    is_variadic BOOLEAN DEFAULT 0,
    
    FOREIGN KEY (function_id) REFERENCES functions(symbol_id) ON DELETE CASCADE
);

-- Таблица за класове
CREATE TABLE IF NOT EXISTS classes (
    symbol_id TEXT PRIMARY KEY,
    is_abstract BOOLEAN DEFAULT 0,
    is_interface BOOLEAN DEFAULT 0,
    
    FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);

-- Таблица за полета на клас
CREATE TABLE IF NOT EXISTS fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    class_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT,
    default_value TEXT,
    visibility TEXT,
    is_static BOOLEAN DEFAULT 0,
    is_constant BOOLEAN DEFAULT 0,
    
    FOREIGN KEY (class_id) REFERENCES classes(symbol_id) ON DELETE CASCADE
);

-- Таблица за наследяване
CREATE TABLE IF NOT EXISTS inheritance (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    child_id TEXT NOT NULL,
    parent_name TEXT NOT NULL,
    kind TEXT, -- 'extends', 'implements'
    
    FOREIGN KEY (child_id) REFERENCES symbols(id) ON DELETE CASCADE
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

-- Таблица за build tasks (AI-driven)
CREATE TABLE IF NOT EXISTS build_tasks (
    id TEXT PRIMARY KEY,
    task_type TEXT NOT NULL,
    target_symbol TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'planned',
    priority INTEGER DEFAULT 0,
    assigned_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME
);

-- Таблица за task dependencies
CREATE TABLE IF NOT EXISTS task_dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id TEXT NOT NULL,
    depends_on_task_id TEXT NOT NULL,
    
    FOREIGN KEY (task_id) REFERENCES build_tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (depends_on_task_id) REFERENCES build_tasks(id) ON DELETE CASCADE
);

-- Таблица за test definitions
CREATE TABLE IF NOT EXISTS test_definitions (
    id TEXT PRIMARY KEY,
    target_symbol_id TEXT NOT NULL,
    test_name TEXT NOT NULL,
    description TEXT,
    expected_behavior TEXT,
    status TEXT NOT NULL DEFAULT 'planned',
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (target_symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);

-- Таблица за test assertions
CREATE TABLE IF NOT EXISTS test_assertions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    test_id TEXT NOT NULL,
    assertion_text TEXT NOT NULL,
    position INTEGER,
    
    FOREIGN KEY (test_id) REFERENCES test_definitions(id) ON DELETE CASCADE
);

-- Full-text search за символи
CREATE VIRTUAL TABLE IF NOT EXISTS symbols_fts USING fts5(
    name,
    signature,
    documentation,
    content='symbols',
    content_rowid='id'
);

-- Triggers за sync на FTS
CREATE TRIGGER IF NOT EXISTS symbols_ai AFTER INSERT ON symbols BEGIN
    INSERT INTO symbols_fts(rowid, name, signature, documentation)
    VALUES (new.id, new.name, new.signature, new.documentation);
END;

CREATE TRIGGER IF NOT EXISTS symbols_ad AFTER DELETE ON symbols BEGIN
    DELETE FROM symbols_fts WHERE rowid = old.id;
END;

CREATE TRIGGER IF NOT EXISTS symbols_au AFTER UPDATE ON symbols BEGIN
    UPDATE symbols_fts 
    SET name = new.name,
        signature = new.signature,
        documentation = new.documentation
    WHERE rowid = new.id;
END;

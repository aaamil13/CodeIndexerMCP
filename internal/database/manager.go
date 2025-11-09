package database

import (
	"database/sql"
	"encoding/json" // Added for JSON marshalling
	"fmt"
	"log"
	"strconv" // Added for type conversion
	"sync"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const (
	DriverName = "sqlite3"
	Schema     = `
	PRAGMA journal_mode = WAL;
	PRAGMA foreign_keys = ON;
	CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		path TEXT NOT NULL UNIQUE,
		language_stats JSON, -- New field for LanguageStats
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		path TEXT NOT NULL,
		relative_path TEXT NOT NULL,
		language TEXT NOT NULL,
		size INTEGER NOT NULL,
		lines_of_code INTEGER NOT NULL,
		hash TEXT NOT NULL,
		last_modified DATETIME NOT NULL,
		last_indexed DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id),
		UNIQUE(project_id, relative_path)
	);
	CREATE TABLE IF NOT EXISTS file_symbols (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL UNIQUE, -- Added UNIQUE constraint
		symbols_json TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS symbols (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		kind TEXT NOT NULL,
		file_path TEXT NOT NULL,
		language TEXT NOT NULL,
		line_number INTEGER NOT NULL,
		column_number INTEGER NOT NULL,
		end_line_number INTEGER NOT NULL,
		end_column_number INTEGER NOT NULL,
		parent TEXT, -- e.g., for methods, the parent class/struct
		signature TEXT,
		documentation TEXT,
		visibility TEXT,
		status TEXT DEFAULT 'unassigned',
		priority TEXT DEFAULT 'medium',
		assigned_agent TEXT,
		content_hash TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metadata JSON,
		FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
	);
	CREATE VIRTUAL TABLE IF NOT EXISTS symbols_fts USING fts5(name, signature, documentation, content='symbols', content_rowid='id');
	CREATE TABLE IF NOT EXISTS "references" (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_symbol_id INTEGER NOT NULL,
		target_symbol_name TEXT NOT NULL,
		reference_type TEXT NOT NULL,
		file_path TEXT NOT NULL,
		line INTEGER NOT NULL,
		column INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE IF NOT EXISTS "relationships" (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		source_symbol INTEGER NOT NULL, -- Changed to INTEGER
		target_symbol TEXT NOT NULL, -- Changed to TEXT
		file_path TEXT NOT NULL,
		line INTEGER NOT NULL,
		metadata JSON,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
)

type Manager struct {
	db *sql.DB
	mu sync.RWMutex
}

func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open(DriverName, dbPath+"?_journal=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	db.SetMaxOpenConns(1) // Ensure only one connection to prevent "database is locked" errors

	// Ping the database to ensure the connection is established
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Execute schema
	if _, err = db.Exec(Schema); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	// Manually create triggers for FTS5
	_, err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS symbols_ai AFTER INSERT ON symbols BEGIN
			INSERT INTO symbols_fts(rowid, name, signature, documentation) VALUES (new.id, new.name, new.signature, new.documentation);
		END;
		CREATE TRIGGER IF NOT EXISTS symbols_ad AFTER DELETE ON symbols BEGIN
			INSERT INTO symbols_fts(symbols_fts, rowid, name, signature, documentation) VALUES('delete', old.id, old.name, old.signature, old.documentation);
		END;
		CREATE TRIGGER IF NOT EXISTS symbols_au AFTER UPDATE ON symbols BEGIN
			INSERT INTO symbols_fts(symbols_fts, rowid, name, signature, documentation) VALUES('delete', old.id, old.name, old.signature, old.documentation);
			INSERT INTO symbols_fts(rowid, name, signature, documentation) VALUES (new.id, new.name, new.signature, new.documentation);
		END;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create FTS5 triggers: %w", err)
	}

	return &Manager{db: db}, nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}

func (m *Manager) Transaction(txFunc func(*sql.Tx) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	err = txFunc(tx)
	return err
}

// Project Operations
func (m *Manager) SaveProject(project *model.Project) error {
	return m.Transaction(func(tx *sql.Tx) error {
		// Marshal LanguageStats to JSON
		languageStatsJSON, err := json.Marshal(project.LanguageStats)
		if err != nil {
			return fmt.Errorf("failed to marshal language stats: %w", err)
		}

		stmt, err := tx.Prepare(`
			INSERT INTO projects (name, path, language_stats) VALUES (?, ?, ?)
			ON CONFLICT(path) DO UPDATE SET name = EXCLUDED.name, language_stats = EXCLUDED.language_stats, updated_at = CURRENT_TIMESTAMP
			RETURNING id;
		`)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		err = stmt.QueryRow(project.Name, project.Path, languageStatsJSON).Scan(&project.ID)
		if err != nil {
			return fmt.Errorf("failed to save project: %w", err)
		}
		return nil
	})
}

func (m *Manager) GetProjectByID(id int) (*model.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, name, path, language_stats, created_at, updated_at FROM projects WHERE id = ?", id)
	project := &model.Project{}
	var languageStatsJSON sql.NullString
	err := row.Scan(&project.ID, &project.Name, &project.Path, &languageStatsJSON, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Project not found
		}
		return nil, fmt.Errorf("failed to get project by ID: %w", err)
	}

	if languageStatsJSON.Valid && languageStatsJSON.String != "null" {
		if err := json.Unmarshal([]byte(languageStatsJSON.String), &project.LanguageStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal language stats for project %d: %w", project.ID, err)
		}
	} else {
		project.LanguageStats = make(map[string]int) // Ensure it's not nil if no data
	}

	return project, nil
}

func (m *Manager) GetProjectByPath(path string) (*model.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, name, path, language_stats, created_at, updated_at FROM projects WHERE path = ?", path)
	project := &model.Project{}
	var languageStatsJSON sql.NullString
	err := row.Scan(&project.ID, &project.Name, &project.Path, &languageStatsJSON, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Project not found
		}
		return nil, fmt.Errorf("failed to get project by path: %w", err)
	}

	if languageStatsJSON.Valid && languageStatsJSON.String != "null" {
		if err := json.Unmarshal([]byte(languageStatsJSON.String), &project.LanguageStats); err != nil {
			return nil, fmt.Errorf("failed to unmarshal language stats for project %d: %w", project.ID, err)
		}
	} else {
		project.LanguageStats = make(map[string]int) // Ensure it's not nil if no data
	}

	return project, nil
}

// File Operations
func (m *Manager) SaveFileTx(tx *sql.Tx, file *model.File) error {
	stmt, err := tx.Prepare(`
		INSERT INTO files (project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, relative_path) DO UPDATE SET
			path = EXCLUDED.path,
			language = EXCLUDED.language,
			size = EXCLUDED.size,
			lines_of_code = EXCLUDED.lines_of_code,
			hash = EXCLUDED.hash,
			last_modified = EXCLUDED.last_modified,
			last_indexed = EXCLUDED.last_indexed,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for SaveFileTx: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(
		file.ProjectID,
		file.Path,
		file.RelativePath,
		file.Language,
		file.Size,
		file.LinesOfCode,
		file.Hash,
		file.LastModified,
		file.LastIndexed,
	).Scan(&file.ID)
	if err != nil {
		return fmt.Errorf("failed to save file in transaction: %w", err)
	}
	return nil
}

// SaveFile saves a file without an external transaction (for standalone use)
func (m *Manager) SaveFile(file *model.File) error {
	return m.Transaction(func(tx *sql.Tx) error {
		return m.SaveFileTx(tx, file)
	})
}

func (m *Manager) GetFileByID(id int) (*model.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed, created_at, updated_at FROM files WHERE id = ?", id)
	file := &model.File{}
	err := row.Scan(&file.ID, &file.ProjectID, &file.Path, &file.RelativePath, &file.Language, &file.Size, &file.LinesOfCode, &file.Hash, &file.LastModified, &file.LastIndexed, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file by ID: %w", err)
	}
	return file, nil
}

func (m *Manager) GetFileByPath(projectID int, relativePath string) (*model.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed, created_at, updated_at FROM files WHERE project_id = ? AND relative_path = ?", projectID, relativePath)
	file := &model.File{}
	err := row.Scan(&file.ID, &file.ProjectID, &file.Path, &file.RelativePath, &file.Language, &file.Size, &file.LinesOfCode, &file.Hash, &file.LastModified, &file.LastIndexed, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file by path: %w", err)
	}
	return file, nil
}

func (m *Manager) GetAllFilesForProject(projectID int) ([]*model.File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := m.db.Query("SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed, created_at, updated_at FROM files WHERE project_id = ?", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query files for project: %w", err)
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		file := &model.File{}
		err := rows.Scan(&file.ID, &file.ProjectID, &file.Path, &file.RelativePath, &file.Language, &file.Size, &file.LinesOfCode, &file.Hash, &file.LastModified, &file.LastIndexed, &file.CreatedAt, &file.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file row: %w", err)
		}
		files = append(files, file)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return files, nil
}

// FileSymbols Operations
func (m *Manager) SaveFileSymbolsTx(tx *sql.Tx, fileSymbols *model.FileSymbols) error {
	stmt, err := tx.Prepare(`
		INSERT INTO file_symbols (file_id, symbols_json) VALUES (?, ?)
		ON CONFLICT(file_id) DO UPDATE SET symbols_json = EXCLUDED.symbols_json, updated_at = CURRENT_TIMESTAMP
		RETURNING id;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for SaveFileSymbolsTx: %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(fileSymbols.FileID, fileSymbols.SymbolsJSON).Scan(&fileSymbols.ID)
	if err != nil {
		return fmt.Errorf("failed to save file symbols in transaction: %w", err)
	}
	return nil
}

func (m *Manager) SaveFileSymbols(fileSymbols *model.FileSymbols) error {
	return m.Transaction(func(tx *sql.Tx) error {
		return m.SaveFileSymbolsTx(tx, fileSymbols)
	})
}

func (m *Manager) GetFileSymbolsByFileID(fileID int) (*model.FileSymbols, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, file_id, symbols_json, created_at, updated_at FROM file_symbols WHERE file_id = ?", fileID)
	fs := &model.FileSymbols{}
	err := row.Scan(&fs.ID, &fs.FileID, &fs.SymbolsJSON, &fs.CreatedAt, &fs.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file symbols by file ID: %w", err)
	}
	return fs, nil
}

// Symbol Operations
func (m *Manager) SaveSymbolTx(tx *sql.Tx, symbol *model.Symbol) error {
	stmt, err := tx.Prepare(`
		INSERT INTO symbols (file_id, name, kind, file_path, language, line_number, column_number, end_line_number, end_column_number, parent, signature, documentation, visibility, status, priority, assigned_agent, content_hash, metadata)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT DO UPDATE SET
			name = EXCLUDED.name,
			kind = EXCLUDED.kind,
			file_path = EXCLUDED.file_path,
			language = EXCLUDED.language,
			line_number = EXCLUDED.line_number,
			column_number = EXCLUDED.column_number,
			end_line_number = EXCLUDED.end_line_number,
			end_column_number = EXCLUDED.end_column_number,
			parent = EXCLUDED.parent,
			signature = EXCLUDED.signature,
			documentation = EXCLUDED.documentation,
			visibility = EXCLUDED.visibility,
			status = EXCLUDED.status,
			priority = EXCLUDED.priority,
			assigned_agent = EXCLUDED.assigned_agent,
			content_hash = EXCLUDED.content_hash,
			metadata = EXCLUDED.metadata,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id;
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for SaveSymbolTx: %w", err)
	}
	defer stmt.Close()

	metadataJSONBytes, err := json.Marshal(symbol.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	metadataJSON := string(metadataJSONBytes)

	err = stmt.QueryRow(
		symbol.FileID,
		symbol.Name,
		symbol.Kind,
		symbol.FilePath,
		symbol.Language,
		symbol.LineNumber,
		symbol.ColumnNumber,
		symbol.EndLineNumber,
		symbol.EndColumnNumber,
		symbol.Parent,
		symbol.Signature,
		symbol.Documentation,
		symbol.Visibility,
		symbol.Status,
		symbol.Priority,
		symbol.AssignedAgent,
		symbol.ContentHash, // Added content_hash
		metadataJSON,
	).Scan(&symbol.ID)
	if err != nil {
		return fmt.Errorf("failed to save symbol in transaction: %w", err)
	}
	return nil
}

func (m *Manager) SaveSymbolsTx(tx *sql.Tx, symbols []*model.Symbol) error {
	for _, symbol := range symbols {
		if err := m.SaveSymbolTx(tx, symbol); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) GetSymbol(fileID int, name, kind string) (*model.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, file_id, name, kind, file_path, language, line_number, column_number, end_line_number, end_column_number, parent, signature, documentation, visibility, status, priority, assigned_agent, content_hash, created_at, updated_at, metadata FROM symbols WHERE file_id = ? AND name = ? AND kind = ?", fileID, name, kind)
	return scanSymbol(row)
}

func (m *Manager) GetSymbolByID(id int) (*model.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	row := m.db.QueryRow("SELECT id, file_id, name, kind, file_path, language, line_number, column_number, end_line_number, end_column_number, parent, signature, documentation, visibility, status, priority, assigned_agent, content_hash, created_at, updated_at, metadata FROM symbols WHERE id = ?", id)
	return scanSymbol(row)
}

func (m *Manager) GetSymbolsByFile(fileID int) ([]*model.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := m.db.Query("SELECT id, file_id, name, kind, file_path, language, line_number, column_number, end_line_number, end_column_number, parent, signature, documentation, visibility, status, priority, assigned_agent, content_hash, created_at, updated_at, metadata FROM symbols WHERE file_id = ?", fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query symbols by file ID: %w", err)
	}
	defer rows.Close()

	var symbols []*model.Symbol
	for rows.Next() {
		symbol, err := scanSymbol(rows)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return symbols, nil
}

func (m *Manager) GetImportsByFile(fileID int) ([]*model.Import, error) {
	// This method is not implemented yet. Returning empty slice for now.
	return []*model.Import{}, nil
}

func (m *Manager) GetSymbolByName(name string) (*model.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// This is a simplified implementation. In a real scenario, you might need
	// more context (e.g., project ID, file ID, kind) to uniquely identify a symbol by name.
	row := m.db.QueryRow("SELECT id, file_id, name, kind, file_path, language, line_number, column_number, end_line_number, end_column_number, parent, signature, documentation, visibility, status, priority, assigned_agent, content_hash, created_at, updated_at, metadata FROM symbols WHERE name = ?", name)
	return scanSymbol(row)
}

func (m *Manager) GetReferencesBySymbol(symbolID int) ([]*model.Reference, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := m.db.Query("SELECT id, source_symbol_id, target_symbol_name, reference_type, file_path, line, column, created_at, updated_at FROM references WHERE source_symbol_id = ?", symbolID)
	if err != nil {
		return nil, fmt.Errorf("failed to query references by symbol ID: %w", err)
	}
	defer rows.Close()

	var references []*model.Reference
	for rows.Next() {
		ref := &model.Reference{}
		var referenceType string
		err := rows.Scan(
			&ref.ID,
			&ref.SourceSymbolID,
			&ref.TargetSymbolName,
			&referenceType,
			&ref.FilePath,
			&ref.Line,
			&ref.Column,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reference row: %w", err)
		}
		ref.ReferenceType = model.ReferenceType(referenceType) // Convert string to model.ReferenceType
		references = append(references, ref)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return references, nil
}

func (m *Manager) SearchSymbols(query string, projectID int) ([]*model.Symbol, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var rows *sql.Rows
	var err error

	if projectID > 0 {
		rows, err = m.db.Query(`
			SELECT s.id, s.file_id, s.name, s.kind, s.file_path, s.language, s.line_number, s.column_number, s.end_line_number, s.end_column_number, s.parent, s.signature, s.documentation, s.visibility, s.status, s.priority, s.assigned_agent, s.content_hash, s.created_at, s.updated_at, s.metadata
			FROM symbols_fts AS sf
			JOIN symbols AS s ON sf.rowid = s.id
			JOIN files AS f ON s.file_id = f.id
			WHERE sf.name MATCH ? AND f.project_id = ?
		`, query+"*", projectID)
	} else {
		rows, err = m.db.Query(`
			SELECT s.id, s.file_id, s.name, s.kind, s.file_path, s.language, s.line_number, s.column_number, s.end_line_number, s.end_column_number, s.parent, s.signature, s.documentation, s.visibility, s.status, s.priority, s.assigned_agent, s.content_hash, s.created_at, s.updated_at, s.metadata
			FROM symbols_fts AS sf
			JOIN symbols AS s ON sf.rowid = s.id
			WHERE sf.name MATCH ?
		`, query+"*")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search symbols: %w", err)
	}
	defer rows.Close()

	var symbols []*model.Symbol
	for rows.Next() {
		symbol, err := scanSymbol(rows)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return symbols, nil
}

func (m *Manager) DeleteFile(projectID int, relativePath string) error {
	return m.Transaction(func(tx *sql.Tx) error {
		// First, get the file ID to delete associated symbols and file_symbols
		var fileID int
		err := tx.QueryRow("SELECT id FROM files WHERE project_id = ? AND relative_path = ?", projectID, relativePath).Scan(&fileID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil // File not found, nothing to delete
			}
			return fmt.Errorf("failed to get file ID for deletion: %w", err)
		}

		// The ON DELETE CASCADE should handle symbols and file_symbols, so we just delete the file
		_, err = tx.Exec("DELETE FROM files WHERE id = ?", fileID)
		if err != nil {
			return fmt.Errorf("failed to delete file: %w", err)
		}
		return nil
	})
}

// Reference Operations
func (m *Manager) SaveReferenceTx(tx *sql.Tx, ref *model.Reference) error {
	stmt, err := tx.Prepare(`
		INSERT INTO "references" (source_symbol_id, target_symbol_name, reference_type, file_path, line, column)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_symbol_id, target_symbol_name, file_path, line, column) DO UPDATE SET
			reference_type = EXCLUDED.reference_type, updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for SaveReferenceTx: %w", err)
	}
	defer stmt.Close()

	_, err = tx.Exec(strconv.Itoa(ref.SourceSymbolID), ref.TargetSymbolName, string(ref.ReferenceType), ref.FilePath, ref.Line, ref.Column)
	if err != nil {
		return fmt.Errorf("failed to save reference in transaction: %w", err)
	}
	return nil
}

func (m *Manager) SaveReference(ref *model.Reference) error {
	return m.Transaction(func(tx *sql.Tx) error {
		return m.SaveReferenceTx(tx, ref)
	})
}

// Relationship Operations
func (m *Manager) SaveRelationshipTx(tx *sql.Tx, rel *model.Relationship) error {
	stmt, err := tx.Prepare(`
		INSERT INTO "relationships" (type, source_symbol, target_symbol, file_path, line, metadata)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(type, source_symbol, target_symbol, file_path, line) DO UPDATE SET
			metadata = EXCLUDED.metadata, updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for SaveRelationshipTx: %w", err)
	}
	defer stmt.Close()

	metadataJSON, err := json.Marshal(rel.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = tx.Exec(string(rel.Type), rel.SourceSymbol, rel.TargetSymbol, rel.FilePath, rel.Line, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to save relationship in transaction: %w", err)
	}
	return nil
}

func (m *Manager) SaveRelationship(rel *model.Relationship) error {
	return m.Transaction(func(tx *sql.Tx) error {
		return m.SaveRelationshipTx(tx, rel)
	})
}

type RowScanner interface {
	Scan(dest ...interface{}) error
}

func scanSymbol(row RowScanner) (*model.Symbol, error) {
	symbol := &model.Symbol{}
	var parent, signature, documentation, visibility, assignedAgent sql.NullString
	var status sql.NullString // Use sql.NullString for the status field
	var contentHash sql.NullString // Added content_hash
	var metadataJSON sql.NullString
	err := row.Scan(
		&symbol.ID,
		&symbol.FileID,
		&symbol.Name,
		&symbol.Kind,
		&symbol.FilePath, // Added FilePath
		&symbol.Language, // Added Language
		&symbol.LineNumber,
		&symbol.ColumnNumber,
		&symbol.EndLineNumber,
		&symbol.EndColumnNumber,
		&parent,
		&signature,
		&documentation,
		&visibility,
		&status,
		&symbol.Priority,
		&assignedAgent,
		&contentHash, // Added content_hash
		&symbol.CreatedAt,
		&symbol.UpdatedAt,
		&metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan symbol row: %w", err)
	}

	symbol.Parent = parent.String
	symbol.Signature = signature.String
	symbol.Documentation = documentation.String
	symbol.Visibility = model.Visibility(visibility.String)
	symbol.Status = model.DevelopmentStatus(status.String)
	symbol.AssignedAgent = assignedAgent.String
	symbol.ContentHash = contentHash.String // Added content_hash
	if metadataJSON.Valid && metadataJSON.String != "null" {
		var metadata map[string]string
		if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		symbol.Metadata = metadata
	} else {
		symbol.Metadata = nil
	}

	return symbol, nil
}

func (m *Manager) HasSymbolChanged(symbolID int, newHash string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var currentHash string
	err := m.db.QueryRow("SELECT content_hash FROM symbols WHERE id = ?", symbolID).Scan(&currentHash) // Changed to content_hash
	if err != nil {
		if err == sql.ErrNoRows {
			return true, nil // New symbol, so it's changed
		}
		return false, fmt.Errorf("failed to get current symbol hash: %w", err)
	}
	return currentHash != newHash, nil
}

func (m *Manager) SaveSymbolIfChanged(symbol *model.Symbol) (bool, error) {
	changed, err := m.HasSymbolChanged(symbol.ID, symbol.ContentHash)
	if err != nil {
		return false, err
	}
	if changed {
		err := m.Transaction(func(tx *sql.Tx) error {
			return m.SaveSymbolTx(tx, symbol)
		})
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (m *Manager) DebugPrintAllFiles() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := m.db.Query("SELECT id, project_id, path, relative_path, language FROM files")
	if err != nil {
		log.Printf("DebugPrintAllFiles: failed to query files: %v", err)
		return
	}
	defer rows.Close()

	log.Println("--- DEBUG: All Files in DB ---")
	for rows.Next() {
		var id, projectID int
		var path, relativePath, language string
		if err := rows.Scan(&id, &projectID, &path, &relativePath, &language); err != nil {
			log.Printf("DebugPrintAllFiles: failed to scan file row: %v", err)
			continue
		}
		log.Printf("ID: %d, ProjectID: %d, Path: %s, RelativePath: %s, Language: %s", id, projectID, path, relativePath, language)
	}
	log.Println("------------------------------")
}

func (m *Manager) DebugPrintAllSymbols() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rows, err := m.db.Query("SELECT id, file_id, name, kind FROM symbols")
	if err != nil {
		log.Printf("DebugPrintAllSymbols: failed to query symbols: %v", err)
		return
	}
	defer rows.Close()

	log.Println("--- DEBUG: All Symbols in DB ---")
	for rows.Next() {
		var id, fileID int
		var name, kind string
		if err := rows.Scan(&id, &fileID, &name, &kind); err != nil {
			log.Printf("DebugPrintAllSymbols: failed to scan symbol row: %v", err)
			continue
		}
		log.Printf("ID: %d, FileID: %d, Name: %s, Kind: %s", id, fileID, name, kind)
	}
	log.Println("--------------------------------")
}

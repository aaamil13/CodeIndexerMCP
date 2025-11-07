package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// Project operations

// CreateProject creates a new project
func (db *DB) CreateProject(project *types.Project) error {
	langStatsJSON, err := toJSON(project.LanguageStats)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO projects (path, name, language_stats, last_indexed, created_at)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := db.conn.Exec(query,
		project.Path,
		project.Name,
		langStatsJSON,
		project.LastIndexed,
		project.CreatedAt,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	project.ID = id
	return nil
}

// GetProject retrieves a project by path
func (db *DB) GetProject(path string) (*types.Project, error) {
	query := `SELECT id, path, name, language_stats, last_indexed, created_at FROM projects WHERE path = ?`

	var project types.Project
	var langStatsJSON string

	err := db.conn.QueryRow(query, path).Scan(
		&project.ID,
		&project.Path,
		&project.Name,
		&langStatsJSON,
		&project.LastIndexed,
		&project.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := fromJSON(langStatsJSON, &project.LanguageStats); err != nil {
		return nil, err
	}

	return &project, nil
}

// UpdateProject updates a project
func (db *DB) UpdateProject(project *types.Project) error {
	langStatsJSON, err := toJSON(project.LanguageStats)
	if err != nil {
		return err
	}

	query := `
		UPDATE projects
		SET name = ?, language_stats = ?, last_indexed = ?
		WHERE id = ?
	`

	_, err = db.conn.Exec(query, project.Name, langStatsJSON, project.LastIndexed, project.ID)
	return err
}

// File operations

// SaveFile creates or updates a file
func (db *DB) SaveFile(file *types.File) error {
	query := `
		INSERT INTO files (project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(project_id, relative_path) DO UPDATE SET
			path = excluded.path,
			language = excluded.language,
			size = excluded.size,
			lines_of_code = excluded.lines_of_code,
			hash = excluded.hash,
			last_modified = excluded.last_modified,
			last_indexed = excluded.last_indexed
		RETURNING id
	`

	err := db.conn.QueryRow(query,
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

	return err
}

// GetFile retrieves a file by ID
func (db *DB) GetFile(id int64) (*types.File, error) {
	query := `SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed FROM files WHERE id = ?`

	var file types.File
	err := db.conn.QueryRow(query, id).Scan(
		&file.ID,
		&file.ProjectID,
		&file.Path,
		&file.RelativePath,
		&file.Language,
		&file.Size,
		&file.LinesOfCode,
		&file.Hash,
		&file.LastModified,
		&file.LastIndexed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// GetFileByPath retrieves a file by its relative path
func (db *DB) GetFileByPath(projectID int64, relativePath string) (*types.File, error) {
	query := `SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed FROM files WHERE project_id = ? AND relative_path = ?`

	var file types.File
	err := db.conn.QueryRow(query, projectID, relativePath).Scan(
		&file.ID,
		&file.ProjectID,
		&file.Path,
		&file.RelativePath,
		&file.Language,
		&file.Size,
		&file.LinesOfCode,
		&file.Hash,
		&file.LastModified,
		&file.LastIndexed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// DeleteFile deletes a file and all its symbols
func (db *DB) DeleteFile(id int64) error {
	_, err := db.conn.Exec("DELETE FROM files WHERE id = ?", id)
	return err
}

// Symbol operations

// SaveSymbol creates a new symbol
func (db *DB) SaveSymbol(symbol *types.Symbol) error {
	metadataJSON, err := toJSON(symbol.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO symbols (
			file_id, name, type, signature, parent_id,
			start_line, end_line, start_column, end_column,
			visibility, is_exported, is_async, is_static, is_abstract,
			documentation, metadata
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	err = db.conn.QueryRow(query,
		symbol.FileID,
		symbol.Name,
		symbol.Type,
		nullString(symbol.Signature),
		nullInt64(symbol.ParentID),
		symbol.StartLine,
		symbol.EndLine,
		symbol.StartColumn,
		symbol.EndColumn,
		symbol.Visibility,
		symbol.IsExported,
		symbol.IsAsync,
		symbol.IsStatic,
		symbol.IsAbstract,
		nullString(symbol.Documentation),
		metadataJSON,
	).Scan(&symbol.ID)

	return err
}

// DeleteSymbolsByFile deletes all symbols for a file
func (db *DB) DeleteSymbolsByFile(fileID int64) error {
	_, err := db.conn.Exec("DELETE FROM symbols WHERE file_id = ?", fileID)
	return err
}

// SearchSymbols searches for symbols by name
func (db *DB) SearchSymbols(opts types.SearchOptions) ([]*types.Symbol, error) {
	query := `
		SELECT id, file_id, name, type, signature, parent_id,
			start_line, end_line, start_column, end_column,
			visibility, is_exported, is_async, is_static, is_abstract,
			documentation, metadata
		FROM symbols
		WHERE name LIKE ?
	`
	args := []interface{}{"%" + opts.Query + "%"}

	if opts.Type != nil {
		query += " AND type = ?"
		args = append(args, *opts.Type)
	}

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	} else {
		query += " LIMIT 100" // Default limit
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []*types.Symbol
	for rows.Next() {
		symbol, err := scanSymbol(rows)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	return symbols, rows.Err()
}

// GetSymbolsByFile retrieves all symbols for a file
func (db *DB) GetSymbolsByFile(fileID int64) ([]*types.Symbol, error) {
	query := `
		SELECT id, file_id, name, type, signature, parent_id,
			start_line, end_line, start_column, end_column,
			visibility, is_exported, is_async, is_static, is_abstract,
			documentation, metadata
		FROM symbols
		WHERE file_id = ?
		ORDER BY start_line
	`

	rows, err := db.conn.Query(query, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []*types.Symbol
	for rows.Next() {
		symbol, err := scanSymbol(rows)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	return symbols, rows.Err()
}

// Import operations

// SaveImport creates a new import
func (db *DB) SaveImport(imp *types.Import) error {
	namesJSON, err := toJSON(imp.ImportedNames)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO imports (file_id, source, imported_names, import_type, line_number, imported_symbol)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id
	`

	err = db.conn.QueryRow(query,
		imp.FileID,
		imp.Source,
		namesJSON,
		imp.ImportType,
		imp.LineNumber,
		nullString(imp.ImportedSymbol),
	).Scan(&imp.ID)

	return err
}

// DeleteImportsByFile deletes all imports for a file
func (db *DB) DeleteImportsByFile(fileID int64) error {
	_, err := db.conn.Exec("DELETE FROM imports WHERE file_id = ?", fileID)
	return err
}

// GetImportsByFile retrieves all imports for a file
func (db *DB) GetImportsByFile(fileID int64) ([]*types.Import, error) {
	query := `SELECT id, file_id, source, imported_names, import_type, line_number, imported_symbol FROM imports WHERE file_id = ?`

	rows, err := db.conn.Query(query, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imports []*types.Import
	for rows.Next() {
		imp, err := scanImport(rows)
		if err != nil {
			return nil, err
		}
		imports = append(imports, imp)
	}

	return imports, rows.Err()
}

// Relationship operations

// SaveRelationship creates a relationship
func (db *DB) SaveRelationship(rel *types.Relationship) error {
	query := `
		INSERT INTO relationships (from_symbol_id, to_symbol_id, relationship_type)
		VALUES (?, ?, ?)
		ON CONFLICT DO NOTHING
		RETURNING id
	`

	err := db.conn.QueryRow(query, rel.FromSymbolID, rel.ToSymbolID, rel.Type).Scan(&rel.ID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}

// Helper scanning functions

func scanSymbol(scanner interface {
	Scan(dest ...interface{}) error
}) (*types.Symbol, error) {
	var symbol types.Symbol
	var signature, documentation, metadataJSON sql.NullString
	var parentID sql.NullInt64

	err := scanner.Scan(
		&symbol.ID,
		&symbol.FileID,
		&symbol.Name,
		&symbol.Type,
		&signature,
		&parentID,
		&symbol.StartLine,
		&symbol.EndLine,
		&symbol.StartColumn,
		&symbol.EndColumn,
		&symbol.Visibility,
		&symbol.IsExported,
		&symbol.IsAsync,
		&symbol.IsStatic,
		&symbol.IsAbstract,
		&documentation,
		&metadataJSON,
	)

	if err != nil {
		return nil, err
	}

	if signature.Valid {
		symbol.Signature = signature.String
	}
	if parentID.Valid {
		pid := parentID.Int64
		symbol.ParentID = &pid
	}
	if documentation.Valid {
		symbol.Documentation = documentation.String
	}
	if metadataJSON.Valid {
		fromJSON(metadataJSON.String, &symbol.Metadata)
	}

	return &symbol, nil
}

func scanImport(scanner interface {
	Scan(dest ...interface{}) error
}) (*types.Import, error) {
	var imp types.Import
	var namesJSON string
	var importedSymbol sql.NullString

	err := scanner.Scan(
		&imp.ID,
		&imp.FileID,
		&imp.Source,
		&namesJSON,
		&imp.ImportType,
		&imp.LineNumber,
		&importedSymbol,
	)

	if err != nil {
		return nil, err
	}

	if err := fromJSON(namesJSON, &imp.ImportedNames); err != nil {
		return nil, err
	}

	if importedSymbol.Valid {
		imp.ImportedSymbol = importedSymbol.String
	}

	return &imp, nil
}

// GetSymbolByName retrieves a symbol by name
func (db *DB) GetSymbolByName(name string) (*types.Symbol, error) {
	query := `
		SELECT id, file_id, name, type, signature, parent_id,
			start_line, end_line, start_column, end_column,
			visibility, is_exported, is_async, is_static, is_abstract,
			documentation, metadata
		FROM symbols
		WHERE name = ?
		LIMIT 1
	`

	symbol, err := scanSymbol(db.conn.QueryRow(query, name))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return symbol, err
}

// GetSymbolWithFile retrieves a symbol with its file information
func (db *DB) GetSymbolWithFile(symbolID int64) (*types.Symbol, *types.File, error) {
	query := `
		SELECT s.id, s.file_id, s.name, s.type, s.signature, s.parent_id,
			s.start_line, s.end_line, s.start_column, s.end_column,
			s.visibility, s.is_exported, s.is_async, s.is_static, s.is_abstract,
			s.documentation, s.metadata,
			f.id, f.project_id, f.path, f.relative_path, f.language, f.size,
			f.lines_of_code, f.hash, f.last_modified, f.last_indexed
		FROM symbols s
		JOIN files f ON s.file_id = f.id
		WHERE s.id = ?
	`

	row := db.conn.QueryRow(query, symbolID)

	var symbol types.Symbol
	var file types.File
	var signature, documentation, metadataJSON sql.NullString
	var parentID sql.NullInt64

	err := row.Scan(
		&symbol.ID, &symbol.FileID, &symbol.Name, &symbol.Type, &signature, &parentID,
		&symbol.StartLine, &symbol.EndLine, &symbol.StartColumn, &symbol.EndColumn,
		&symbol.Visibility, &symbol.IsExported, &symbol.IsAsync, &symbol.IsStatic, &symbol.IsAbstract,
		&documentation, &metadataJSON,
		&file.ID, &file.ProjectID, &file.Path, &file.RelativePath, &file.Language, &file.Size,
		&file.LinesOfCode, &file.Hash, &file.LastModified, &file.LastIndexed,
	)

	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	if signature.Valid {
		symbol.Signature = signature.String
	}
	if parentID.Valid {
		pid := parentID.Int64
		symbol.ParentID = &pid
	}
	if documentation.Valid {
		symbol.Documentation = documentation.String
	}
	if metadataJSON.Valid {
		fromJSON(metadataJSON.String, &symbol.Metadata)
	}

	return &symbol, &file, nil
}

// GetRelationshipsForSymbol retrieves relationships for a symbol
func (db *DB) GetRelationshipsForSymbol(symbolID int64) ([]*types.Relationship, error) {
	query := `
		SELECT id, from_symbol_id, to_symbol_id, relationship_type
		FROM relationships
		WHERE from_symbol_id = ? OR to_symbol_id = ?
	`

	rows, err := db.conn.Query(query, symbolID, symbolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relationships []*types.Relationship
	for rows.Next() {
		var rel types.Relationship
		if err := rows.Scan(&rel.ID, &rel.FromSymbolID, &rel.ToSymbolID, &rel.Type); err != nil {
			return nil, err
		}
		relationships = append(relationships, &rel)
	}

	return relationships, rows.Err()
}

// SaveReference creates a reference
func (db *DB) SaveReference(ref *types.Reference) error {
	query := `
		INSERT INTO references (symbol_id, file_id, line_number, column_number, reference_type)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id
	`

	err := db.conn.QueryRow(query,
		ref.SymbolID,
		ref.FileID,
		ref.LineNumber,
		ref.ColumnNumber,
		ref.ReferenceType,
	).Scan(&ref.ID)

	return err
}

// GetReferencesBySymbol retrieves all references to a symbol
func (db *DB) GetReferencesBySymbol(symbolID int64) ([]*types.Reference, error) {
	query := `
		SELECT id, symbol_id, file_id, line_number, column_number, reference_type
		FROM references
		WHERE symbol_id = ?
	`

	rows, err := db.conn.Query(query, symbolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*types.Reference
	for rows.Next() {
		var ref types.Reference
		if err := rows.Scan(&ref.ID, &ref.SymbolID, &ref.FileID, &ref.LineNumber, &ref.ColumnNumber, &ref.ReferenceType); err != nil {
			return nil, err
		}
		references = append(references, &ref)
	}

	return references, rows.Err()
}

// GetAllFilesForProject retrieves all files in a project
func (db *DB) GetAllFilesForProject(projectID int64) ([]*types.File, error) {
	query := `
		SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed
		FROM files
		WHERE project_id = ?
		ORDER BY relative_path
	`

	rows, err := db.conn.Query(query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*types.File
	for rows.Next() {
		var file types.File
		if err := rows.Scan(
			&file.ID, &file.ProjectID, &file.Path, &file.RelativePath,
			&file.Language, &file.Size, &file.LinesOfCode, &file.Hash,
			&file.LastModified, &file.LastIndexed,
		); err != nil {
			return nil, err
		}
		files = append(files, &file)
	}

	return files, rows.Err()
}

// GetReferencesByFile retrieves all references in a file
func (db *DB) GetReferencesByFile(fileID int64) ([]*types.Reference, error) {
	query := `
		SELECT id, symbol_id, file_id, line_number, column_number, reference_type
		FROM references
		WHERE file_id = ?
		ORDER BY line_number
	`

	rows, err := db.conn.Query(query, fileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*types.Reference
	for rows.Next() {
		var ref types.Reference
		if err := rows.Scan(&ref.ID, &ref.SymbolID, &ref.FileID, &ref.LineNumber, &ref.ColumnNumber, &ref.ReferenceType); err != nil {
			return nil, err
		}
		references = append(references, &ref)
	}

	return references, rows.Err()
}

// GetMethodsForType retrieves all methods for a given type (struct/class)
func (db *DB) GetMethodsForType(typeSymbolID int64) ([]*types.Symbol, error) {
	query := `
		SELECT id, file_id, name, type, signature, parent_id,
			start_line, end_line, start_column, end_column,
			visibility, is_exported, is_async, is_static, is_abstract,
			documentation, metadata
		FROM symbols
		WHERE parent_id = ? AND type = ?
		ORDER BY name
	`

	rows, err := db.conn.Query(query, typeSymbolID, types.SymbolTypeMethod)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []*types.Symbol
	for rows.Next() {
		method, err := scanSymbol(rows)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, rows.Err()
}

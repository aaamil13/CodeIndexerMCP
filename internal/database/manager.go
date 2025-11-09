package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	_ "embed"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
	_ "github.com/mattn/go-sqlite3" // Use mattn/go-sqlite3 driver
)

//go:embed schema.sql
var schemaSQL string

func applySchema(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	return err
}

type Manager struct {
	db *sql.DB
	logger *utils.Logger // Add logger field
}

func NewManager(dbPath string, logger *utils.Logger) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath) // Change driver name to "sqlite3"
	if err != nil {
		return nil, err
	}

	// Прилагане на schema
	if err := applySchema(db); err != nil {
		return nil, err
	}

	return &Manager{db: db, logger: logger}, nil
}

func (m *Manager) Close() error {
	return m.db.Close()
}

func (m *Manager) Stats() (map[string]int, error) {
	// TODO: Implement this properly
	return nil, nil
}

func (m *Manager) SearchSymbols(opts model.SearchOptions) ([]*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language,
               start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash
        FROM symbols
    `
	m.logger.Debug("SearchSymbols query (simplified):", "query", query)

	rows, err := m.db.Query(query)
	if err != nil {
		m.logger.Error("SearchSymbols query failed:", "error", err)
		return nil, err
	}
	defer rows.Close()

	var symbols []*model.Symbol
	for rows.Next() {
		s := &model.Symbol{}
		err := rows.Scan(
			&s.ID, &s.Name, &s.Kind, &s.File, &s.Language,
			&s.Range.Start.Line, &s.Range.Start.Column, &s.Range.Start.Byte,
			&s.Range.End.Line, &s.Range.End.Column, &s.Range.End.Byte, &s.ContentHash,
		)
		if err != nil {
			m.logger.Error("SearchSymbols scan failed:", "error", err)
			return nil, err
		}
		m.logger.Debug("Found symbol in search:", "ID", s.ID, "Name", s.Name)
		symbols = append(symbols, s)
	}
	m.logger.Debug("SearchSymbols found:", "count", len(symbols))

	return symbols, nil
}

func (m *Manager) SaveRelationship(rel *model.Relationship) error {
	return nil
}

func (m *Manager) SaveImport(imp *model.Import) error {
	query := `
        INSERT INTO imports (file_path, import_path, alias, is_wildcard, start_line)
        VALUES (?, ?, ?, ?, ?)
    `
	_, err := m.db.Exec(query,
		imp.FilePath, imp.Path, imp.Alias, imp.IsWildcard, imp.Range.Start.Line,
	)
	return err
}

func (m *Manager) SaveReference(ref *model.Reference) error {
	query := `
        INSERT INTO code_references (source_symbol_id, target_symbol_name, reference_type, file_path, line, column)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err := m.db.Exec(query,
		ref.SourceSymbolID, ref.TargetSymbolName, ref.ReferenceType, ref.FilePath, ref.Line, ref.Column,
	)
	return err
}

func (m *Manager) DeleteFile(fileID int) error {
	return nil
}

func (m *Manager) DeleteImportsByFile(fileID int) error {
	_, err := m.db.Exec("DELETE FROM imports WHERE file_path = (SELECT path FROM files WHERE id = ?)", fileID)
	return err
}

func (m *Manager) DeleteSymbolsByFile(fileID int) error {
	_, err := m.db.Exec("DELETE FROM symbols WHERE file_path = (SELECT path FROM files WHERE id = ?)", fileID)
	return err
}

func (m *Manager) SaveFile(file *model.File) error {
	if file.ID == 0 { // Insert new file
		query := `
            INSERT INTO files (
                project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `
		result, err := m.db.Exec(query,
			file.ProjectID, file.Path, file.RelativePath, file.Language, file.Size,
			file.LinesOfCode, file.Hash, file.LastModified.Format(time.RFC3339), file.LastIndexed.Format(time.RFC3339),
		)
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		file.ID = int(id)
	} else { // Update existing file
		query := `
            UPDATE files
            SET project_id = ?, path = ?, relative_path = ?, language = ?, size = ?,
                lines_of_code = ?, hash = ?, last_modified = ?, last_indexed = ?
            WHERE id = ?
        `
		_, err := m.db.Exec(query,
			file.ProjectID, file.Path, file.RelativePath, file.Language, file.Size,
			file.LinesOfCode, file.Hash, file.LastModified.Format(time.RFC3339), file.LastIndexed.Format(time.RFC3339),
			file.ID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Transaction(f func(tx *sql.Tx) error) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (m *Manager) GetFileByPath(projectID int, relPath string) (*model.File, error) {
	query := `
        SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed
        FROM files
        WHERE project_id = ? AND relative_path = ?
    `
	row := m.db.QueryRow(query, projectID, relPath)

	f := &model.File{}
	var lastModifiedStr, lastIndexedStr string
	err := row.Scan(
		&f.ID, &f.ProjectID, &f.Path, &f.RelativePath, &f.Language, &f.Size, &f.LinesOfCode, &f.Hash, &lastModifiedStr, &lastIndexedStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	f.LastModified, err = time.Parse(time.RFC3339, lastModifiedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_modified time: %w", err)
	}
	f.LastIndexed, err = time.Parse(time.RFC3339, lastIndexedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_indexed time: %w", err)
	}

	return f, nil
}

func (m *Manager) GetFile(id int64) (*model.File, error) {
	query := `
        SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed
        FROM files
        WHERE id = ?
    `
	row := m.db.QueryRow(query, id)

	f := &model.File{}
	var lastModifiedStr, lastIndexedStr string
	err := row.Scan(
		&f.ID, &f.ProjectID, &f.Path, &f.RelativePath, &f.Language, &f.Size, &f.LinesOfCode, &f.Hash, &lastModifiedStr, &lastIndexedStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	f.LastModified, err = time.Parse(time.RFC3339, lastModifiedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_modified time: %w", err)
	}
	f.LastIndexed, err = time.Parse(time.RFC3339, lastIndexedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_indexed time: %w", err)
	}
	return f, nil
}

func (m *Manager) UpdateProject(project *model.Project) error {
	query := `
        UPDATE projects
        SET name = ?, language_stats = ?, last_indexed = ?
        WHERE id = ?
    `
	languageStatsJSON, err := json.Marshal(project.LanguageStats)
	if err != nil {
		return fmt.Errorf("failed to marshal language stats: %w", err)
	}

	_, err = m.db.Exec(query,
		project.Name, string(languageStatsJSON), project.LastIndexed.Format(time.RFC3339), project.ID,
	)
	return err
}

func (m *Manager) CreateProject(project *model.Project) error {
	query := `
        INSERT INTO projects (path, name, language_stats, last_indexed, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	languageStatsJSON, err := json.Marshal(project.LanguageStats)
	if err != nil {
		return fmt.Errorf("failed to marshal language stats: %w", err)
	}

	result, err := m.db.Exec(query,
		project.Path, project.Name, string(languageStatsJSON),
		project.LastIndexed.Format(time.RFC3339), project.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	project.ID = int(id)
	return nil
}

func (m *Manager) GetProject(projectPath string) (*model.Project, error) {
	query := `
        SELECT id, path, name, language_stats, last_indexed, created_at
        FROM projects
        WHERE path = ?
    `
	row := m.db.QueryRow(query, projectPath)

	p := &model.Project{}
	var languageStatsJSON, lastIndexedStr, createdAtStr sql.NullString
	err := row.Scan(
		&p.ID, &p.Path, &p.Name, &languageStatsJSON, &lastIndexedStr, &createdAtStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if languageStatsJSON.Valid {
		err = json.Unmarshal([]byte(languageStatsJSON.String), &p.LanguageStats)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal language stats: %w", err)
		}
	} else {
		p.LanguageStats = make(map[string]int)
	}

	if lastIndexedStr.Valid {
		p.LastIndexed, err = time.Parse(time.RFC3339, lastIndexedStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_indexed time: %w", err)
		}
	}
	if createdAtStr.Valid {
		p.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at time: %w", err)
		}
	}

	return p, nil
}
func (m *Manager) SaveSymbol(symbol *model.Symbol) error {
    metadata, _ := json.Marshal(symbol.Metadata)
    
    query := `
        INSERT OR REPLACE INTO symbols (
            id, name, kind, file_path, language, signature, documentation,
            visibility, start_line, start_column, start_byte,
            end_line, end_column, end_byte, content_hash, status, priority,
            assigned_agent, created_at, updated_at, metadata
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err := m.db.Exec(query,
        symbol.ID, symbol.Name, symbol.Kind, symbol.File, symbol.Language,
        symbol.Signature, symbol.Documentation, symbol.Visibility,
        symbol.Range.Start.Line, symbol.Range.Start.Column, symbol.Range.Start.Byte,
        symbol.Range.End.Line, symbol.Range.End.Column, symbol.Range.End.Byte,
        symbol.ContentHash, symbol.Status, symbol.Priority, symbol.AssignedAgent,
        symbol.CreatedAt, symbol.UpdatedAt, string(metadata),
    )
    
    return err
}

// 💡 ПОДОБРЕНИЕ #5: Проверка за промени чрез content hash
func (m *Manager) HasSymbolChanged(symbolID, newContentHash string) (bool, error) {
	var oldHash string
	query := `SELECT content_hash FROM symbols WHERE id = ?`

	err := m.db.QueryRow(query, symbolID).Scan(&oldHash)
	if err == sql.ErrNoRows {
		// Символът не съществува - счита се за промяна
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return oldHash != newContentHash, nil
}

// Оптимизирано запазване - пропуска непроменени символи
func (m *Manager) SaveSymbolIfChanged(symbol *model.Symbol) (bool, error) {
	changed, err := m.HasSymbolChanged(symbol.ID, symbol.ContentHash)
	if err != nil {
		return false, err
	}

	if !changed {
		// Символът не е променен, пропускаме записа
		return false, nil
	}

	// Символът е променен или нов, записваме го
	return true, m.SaveSymbol(symbol)
}

func (m *Manager) SaveFunction(function *model.Function) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Запазване на символа
	if err := m.SaveSymbol(&function.Symbol); err != nil {
		return err
	}

	// Запазване на функция детайли
	funcQuery := `
        INSERT OR REPLACE INTO functions (
            symbol_id, return_type, is_async, is_generator, body, receiver_type, is_static
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	_, err = tx.Exec(funcQuery,
		function.ID, function.ReturnType, function.IsAsync,
		function.IsGenerator, function.Body, "", false,
	)
	if err != nil {
		return err
	}

	// Запазване на параметри
	for i, param := range function.Parameters {
		paramQuery := `
            INSERT INTO parameters (
                function_id, name, type, default_value, position,
                is_optional, is_variadic
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
        `
		_, err = tx.Exec(paramQuery,
			function.ID, param.Name, param.Type, param.DefaultValue,
			i, param.IsOptional, param.IsVariadic,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (m *Manager) SaveMethod(method *model.Method) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Запазване на символа
    if err := m.SaveSymbol(&method.Symbol); err != nil {
        return err
    }
    
    // Запазване на функция детайли
    funcQuery := `
        INSERT OR REPLACE INTO functions (
            symbol_id, return_type, is_async, is_generator, body, receiver_type, is_static
        ) VALUES (?, ?, ?, ?, ?, ?, ?)
    `
    
    _, err = tx.Exec(funcQuery,
        method.ID, method.ReturnType, method.IsAsync,
        method.IsGenerator, method.Body, method.ReceiverType, method.IsStatic,
    )
    if err != nil {
        return err
    }
    
    // Запазване на параметри
    for i, param := range method.Parameters {
        paramQuery := `
            INSERT INTO parameters (
                function_id, name, type, default_value, position,
                is_optional, is_variadic
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
        `
        _, err = tx.Exec(paramQuery,
            method.ID, param.Name, param.Type, param.DefaultValue,
            i, param.IsOptional, param.IsVariadic,
        )
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

func (m *Manager) SaveClass(class *model.Class) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Запазване на символа
    if err := m.SaveSymbol(&class.Symbol); err != nil {
        return err
    }
    
    // Запазване на клас детайли
    classQuery := `
        INSERT OR REPLACE INTO classes (
            symbol_id, is_abstract, is_interface
        ) VALUES (?, ?, ?)
    `
    
    _, err = tx.Exec(classQuery,
        class.ID, class.IsAbstract, class.IsInterface,
    )
    if err != nil {
        return err
    }
    
    // Запазване на полета
    for _, field := range class.Fields {
        fieldQuery := `
            INSERT INTO fields (
                class_id, name, type, default_value, visibility, is_static, is_constant
            ) VALUES (?, ?, ?, ?, ?, ?, ?)
        `
        _, err = tx.Exec(fieldQuery,
            class.ID, field.Name, field.Type, field.DefaultValue,
            field.Visibility, field.IsStatic, field.IsConstant,
        )
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

func (m *Manager) SaveFileSymbols(fileSymbols *model.FileSymbols) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Изтриване на стари символи от файла
	_, err = tx.Exec("DELETE FROM symbols WHERE file_path = ?", fileSymbols.FilePath)
	if err != nil {
		return err
	}

	// Запазване на функции
	for _, fn := range fileSymbols.Functions {
		if err := m.SaveFunction(fn); err != nil {
			return err
		}
	}

	// Запазване на методи
	for _, method := range fileSymbols.Methods {
		if err := m.SaveMethod(method); err != nil {
			return err
		}
	}

	// Запазване на класове
	for _, class := range fileSymbols.Classes {
		if err := m.SaveClass(class); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// AI-driven methods

func (m *Manager) CreateBuildTask(task *model.BuildTask) error {
	query := `
        INSERT INTO build_tasks (
            id, task_type, target_symbol, description, status,
            priority, assigned_agent, created_at, updated_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := m.db.Exec(query,
		task.ID, task.Type, task.TargetSymbol, task.Description,
		task.Status, task.Priority, task.AssignedAgent,
		task.CreatedAt, task.UpdatedAt,
	)

	return err
}

func (m *Manager) GetTasksByStatus(status model.DevelopmentStatus) ([]*model.BuildTask, error) {
	query := `
        SELECT id, task_type, target_symbol, description, status,
               priority, assigned_agent, created_at, updated_at
        FROM build_tasks
        WHERE status = ?
        ORDER BY priority DESC, created_at ASC
    `

	rows, err := m.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*model.BuildTask
	for rows.Next() {
		task := &model.BuildTask{}
		err := rows.Scan(
			&task.ID, &task.Type, &task.TargetSymbol, &task.Description,
			&task.Status, &task.Priority, &task.AssignedAgent,
			&task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (m *Manager) UpdateSymbolStatus(symbolID string, status model.DevelopmentStatus) error {
	query := `UPDATE symbols SET status = ?, updated_at = ? WHERE id = ?`
	_, err := m.db.Exec(query, status, time.Now(), symbolID)
	return err
}

func (m *Manager) GetSymbolsByStatus(status model.DevelopmentStatus) ([]*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language, signature,
               status, priority, assigned_agent
        FROM symbols
        WHERE status = ?
        ORDER BY priority DESC
    `

	rows, err := m.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []*model.Symbol
	for rows.Next() {
		s := &model.Symbol{}
		err := rows.Scan(
			&s.ID, &s.Name, &s.Kind, &s.File, &s.Language, &s.Signature,
			&s.Status, &s.Priority, &s.AssignedAgent,
		)
		if err != nil {
			return nil, err
		}
		symbols = append(symbols, s)
	}

	return symbols, nil
}

func (m *Manager) GetSymbolByName(name string) (*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language, type,
               start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash
        FROM symbols
        WHERE name = ?
    `
	row := m.db.QueryRow(query, name)
	return m.scanSymbol(row)
}

func (m *Manager) GetSymbol(id string) (*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language, type,
               start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash
        FROM symbols
        WHERE id = ?
    `
	row := m.db.QueryRow(query, id)
	return m.scanSymbol(row)
}

func (m *Manager) scanSymbol(row *sql.Row) (*model.Symbol, error) {
	s := &model.Symbol{}
	var symbolType sql.NullString // Use sql.NullString for nullable 'type' column

	err := row.Scan(
		&s.ID, &s.Name, &s.Kind, &s.File, &s.Language, &symbolType,
		&s.Range.Start.Line, &s.Range.Start.Column, &s.Range.Start.Byte,
		&s.Range.End.Line, &s.Range.End.Column, &s.Range.End.Byte, &s.ContentHash,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if symbolType.Valid {
		s.Type = symbolType.String
	}

	return s, nil
}

func (m *Manager) GetReferencesBySymbol(symbolID string) ([]*model.Reference, error) {
	query := `
        SELECT source_symbol_id, target_symbol_name, reference_type, file_path, line, column
        FROM code_references
        WHERE source_symbol_id = ? OR target_symbol_name = (SELECT name FROM symbols WHERE id = ?)
    `
	rows, err := m.db.Query(query, symbolID, symbolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*model.Reference
	for rows.Next() {
		r := &model.Reference{}
		err := rows.Scan(
			&r.SourceSymbolID, &r.TargetSymbolName, &r.ReferenceType, &r.FilePath, &r.Line, &r.Column,
		)
		if err != nil {
			return nil, err
		}
		references = append(references, r)
	}

	return references, nil
}

func (m *Manager) GetImportsByFile(filePath string) ([]*model.Import, error) {
	query := `
        SELECT file_path, import_path, alias, is_wildcard, start_line
        FROM imports
        WHERE file_path = ?
    `
	rows, err := m.db.Query(query, filePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imports []*model.Import
	for rows.Next() {
		imp := &model.Import{}
		var alias sql.NullString
		var isWildcard bool // Use bool for BOOLEAN
		var startLine sql.NullInt64
		err := rows.Scan(
			&imp.FilePath, &imp.Path, &alias, &isWildcard, &startLine,
		)
		if err != nil {
			return nil, err
		}
		if alias.Valid {
			imp.Alias = alias.String
		}
		imp.IsWildcard = isWildcard // Assign the scanned bool value
		if startLine.Valid {
			imp.Range.Start.Line = int(startLine.Int64)
		}
		imports = append(imports, imp)
	}

	return imports, nil
}

func (m *Manager) GetSymbolsByFile(filePath string) ([]*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language, type,
               start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash
        FROM symbols
        WHERE file_path = ?
    `
	rows, err := m.db.Query(query, filePath)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var symbols []*model.Symbol
	for rows.Next() {
		s := &model.Symbol{}
		var symbolType sql.NullString
		err := rows.Scan(
			&s.ID, &s.Name, &s.Kind, &s.File, &s.Language, &symbolType,
			&s.Range.Start.Line, &s.Range.Start.Column, &s.Range.Start.Byte,
			&s.Range.End.Line, &s.Range.End.Column, &s.Range.End.Byte, &s.ContentHash,
		)
		if err != nil {
			return nil, err
		}
		if symbolType.Valid {
			s.Type = symbolType.String
		}
		symbols = append(symbols, s)
	}

	return symbols, nil
}

func (m *Manager) GetAllFilesForProject(projectID int) ([]*model.File, error) {
	query := `
        SELECT id, project_id, path, relative_path, language, size, lines_of_code, hash, last_modified, last_indexed
        FROM files
        WHERE project_id = ?
    `
	m.logger.Debug("GetAllFilesForProject query:", "query", query, "projectID", projectID)

	rows, err := m.db.Query(query, projectID)
	if err != nil {
		m.logger.Error("GetAllFilesForProject query failed:", "error", err)
		return nil, err
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		f := &model.File{}
		var lastModifiedStr, lastIndexedStr string
		err := rows.Scan(
			&f.ID, &f.ProjectID, &f.Path, &f.RelativePath, &f.Language, &f.Size, &f.LinesOfCode, &f.Hash, &lastModifiedStr, &lastIndexedStr,
		)
		if err != nil {
			m.logger.Error("GetAllFilesForProject scan failed:", "error", err)
			return nil, err
		}

		f.LastModified, err = time.Parse(time.RFC3339, lastModifiedStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_modified time: %w", err)
		}
		f.LastIndexed, err = time.Parse(time.RFC3339, lastIndexedStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last_indexed time: %w", err)
		}
		files = append(files, f)
		m.logger.Debug("Found file in GetAllFilesForProject:", "ID", f.ID, "Path", f.Path)
	}
	m.logger.Debug("GetAllFilesForProject found:", "count", len(files))

	return files, nil
}

func (m *Manager) GetMethodsForType(typeSymbolID string) ([]*model.Method, error) {
	// This assumes that methods are stored with a receiver type that matches the typeSymbolID
	// Or that there's a relationship between the type and its methods.
	// For now, we'll query functions table and filter by receiver_type
	query := `
        SELECT s.id, s.name, s.kind, s.file_path, s.language, s.signature, s.documentation,
               s.visibility, s.start_line, s.start_column, s.start_byte,
               s.end_line, s.end_column, s.end_byte, s.content_hash, s.status, s.priority,
               s.assigned_agent, s.created_at, s.updated_at, s.metadata,
               f.return_type, f.is_async, f.is_generator, f.body, f.receiver_type, f.is_static
        FROM symbols s
        JOIN functions f ON s.id = f.symbol_id
        WHERE f.receiver_type = (SELECT name FROM symbols WHERE id = ?) AND s.kind = 'method'
    `
	rows, err := m.db.Query(query, typeSymbolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []*model.Method
	for rows.Next() {
		mth := &model.Method{}
		var metadataStr string
		var createdAt, updatedAt time.Time
		var assignedAgent sql.NullString

		err := rows.Scan(
			&mth.ID, &mth.Name, &mth.Kind, &mth.File, &mth.Language, &mth.Signature, &mth.Documentation,
			&mth.Visibility, &mth.Range.Start.Line, &mth.Range.Start.Column, &mth.Range.Start.Byte,
			&mth.Range.End.Line, &mth.Range.End.Column, &mth.Range.End.Byte, &mth.ContentHash, &mth.Status, &mth.Priority,
			&assignedAgent, &createdAt, &updatedAt, &metadataStr,
			&mth.ReturnType, &mth.IsAsync, &mth.IsGenerator, &mth.Body, &mth.ReceiverType, &mth.IsStatic,
		)
		if err != nil {
			return nil, err
		}

		mth.CreatedAt = createdAt
		mth.UpdatedAt = updatedAt
		if assignedAgent.Valid {
			mth.AssignedAgent = assignedAgent.String
		}

		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &mth.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		methods = append(methods, mth)
	}

	return methods, nil
}

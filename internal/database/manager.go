package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	_ "embed"

	"github.com/aaamil13/CodeIndexerMCP/internal/model"
	"github.com/aaamil13/CodeIndexerMCP/internal/utils"
	// _ "github.com/mattn/go-sqlite3" // Use mattn/go-sqlite3 driver
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
	// TODO: Implement this properly
	return nil, nil
}

func (m *Manager) SaveRelationship(rel *model.Relationship) error {
	return nil
}

func (m *Manager) SaveImport(imp *model.Import) error {
	return nil
}

func (m *Manager) SaveReference(ref *model.Reference) error {
	return nil
}

func (m *Manager) DeleteFile(fileID int) error {
	return nil
}

func (m *Manager) DeleteImportsByFile(fileID int) error {
	// TODO: Implement this properly
	return nil
}

func (m *Manager) DeleteSymbolsByFile(fileID int) error {
	// TODO: Implement this properly
	return nil
}

func (m *Manager) SaveFile(file *model.File) error {
	// TODO: Implement this properly
	return nil
}

func (m *Manager) Transaction(f func(tx *sql.Tx) error) error {
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := f(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Manager) GetFileByPath(projectID int, relPath string) (*model.File, error) {
	// TODO: Implement this properly
	return nil, nil
}

func (m *Manager) UpdateProject(project *model.Project) error {
	// TODO: Implement this properly
	return nil
}

func (m *Manager) CreateProject(project *model.Project) error {
	// TODO: Implement this properly
	return nil
}

func (m *Manager) GetProject(projectPath string) (*model.Project, error) {
	// TODO: Implement this properly
	return &model.Project{
		ID:   1,
		Path: projectPath,
		Name: "dummy",
	}, nil
}

func (m *Manager) SaveSymbol(symbol *model.Symbol) error {
	query := `
        INSERT OR REPLACE INTO symbols (
            id, name, kind, file_path, language,
            signature, documentation, visibility,
            start_line, start_column, start_byte,
            end_line, end_column, end_byte, content_hash, status, priority,
            assigned_agent, created_at, updated_at, metadata
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	
	toNullString := func(s string) sql.NullString {
		if s == "" {
			return sql.NullString{Valid: false}
		}
		return sql.NullString{String: s, Valid: true}
	}

	m.logger.Debug("Saving symbol to DB:",
		"ID", symbol.ID,
		"Name", symbol.Name,
		"Kind", string(symbol.Kind),
		"File", symbol.File,
		"Language", symbol.Language,
		"Signature", symbol.Signature,
		"Documentation", symbol.Documentation,
		"Visibility", string(symbol.Visibility),
		"StartLine", symbol.Range.Start.Line,
		"StartColumn", symbol.Range.Start.Column,
		"StartByte", symbol.Range.Start.Byte,
		"EndLine", symbol.Range.End.Line,
		"EndColumn", symbol.Range.End.Column,
		"EndByte", symbol.Range.End.Byte,
		"ContentHash", symbol.ContentHash,
		"Status", string(symbol.Status),
		"Priority", symbol.Priority,
		"AssignedAgent", symbol.AssignedAgent,
		"CreatedAt", symbol.CreatedAt.Format(time.RFC3339),
		"UpdatedAt", symbol.UpdatedAt.Format(time.RFC3339),
		"Metadata", symbol.Metadata,
	)

	var metadataNullString sql.NullString
	if symbol.Metadata == nil || len(symbol.Metadata) == 0 {
		metadataNullString = sql.NullString{Valid: false}
	} else {
		var marshalErr error
		metadataJSON, marshalErr := json.Marshal(symbol.Metadata)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal metadata: %w", marshalErr)
		}
		metadataNullString = toNullString(string(metadataJSON))
	}

	_, err := m.db.Exec(query,
		symbol.ID, symbol.Name, string(symbol.Kind), symbol.File, symbol.Language,
		toNullString(symbol.Signature),
		toNullString(symbol.Documentation),
		toNullString(string(symbol.Visibility)),
		int64(symbol.Range.Start.Line), int64(symbol.Range.Start.Column), int64(symbol.Range.Start.Byte),
		int64(symbol.Range.End.Line), int64(symbol.Range.End.Column), int64(symbol.Range.End.Byte),
		symbol.ContentHash, string(symbol.Status), int64(symbol.Priority),
		toNullString(symbol.AssignedAgent),
		symbol.CreatedAt.Format(time.RFC3339),
		symbol.UpdatedAt.Format(time.RFC3339),
		metadataNullString,
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
	// TODO: Implement
	return nil
}

func (m *Manager) SaveClass(class *model.Class) error {
	// TODO: Implement
	return nil
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
        SELECT id, name, kind, file_path, language, signature, documentation,
               visibility, start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash, status, priority,
               assigned_agent, created_at, updated_at, metadata
        FROM symbols
        WHERE name = ?
    `
	row := m.db.QueryRow(query, name)
	return m.scanSymbol(row)
}

func (m *Manager) GetSymbol(id string) (*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language, signature, documentation,
               visibility, start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash, status, priority,
               assigned_agent, created_at, updated_at, metadata
        FROM symbols
        WHERE id = ?
    `
	row := m.db.QueryRow(query, id)
	return m.scanSymbol(row)
}

func (m *Manager) scanSymbol(row *sql.Row) (*model.Symbol, error) {
	s := &model.Symbol{}
	var metadataStr string
	var createdAt, updatedAt time.Time
	var assignedAgent sql.NullString // Use sql.NullString for nullable columns

	err := row.Scan(
		&s.ID, &s.Name, &s.Kind, &s.File, &s.Language, &s.Signature, &s.Documentation,
		&s.Visibility, &s.Range.Start.Line, &s.Range.Start.Column, &s.Range.Start.Byte,
		&s.Range.End.Line, &s.Range.End.Column, &s.Range.End.Byte, &s.ContentHash, &s.Status, &s.Priority,
		&assignedAgent, &createdAt, &updatedAt, &metadataStr,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	s.CreatedAt = createdAt
	s.UpdatedAt = updatedAt
	if assignedAgent.Valid {
		s.AssignedAgent = assignedAgent.String
	}

	if metadataStr != "" {
		err = json.Unmarshal([]byte(metadataStr), &s.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
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
		var startLine sql.NullInt64
		err := rows.Scan(
			&imp.FilePath, &imp.Path, &alias, &imp.IsWildcard, &startLine,
		)
		if err != nil {
			return nil, err
		}
		if alias.Valid {
			imp.Alias = alias.String
		}
		if startLine.Valid {
			imp.Range.Start.Line = int(startLine.Int64)
		}
		imports = append(imports, imp)
	}

	return imports, nil
}

func (m *Manager) GetSymbolsByFile(filePath string) ([]*model.Symbol, error) {
	query := `
        SELECT id, name, kind, file_path, language, signature, documentation,
               visibility, start_line, start_column, start_byte,
               end_line, end_column, end_byte, content_hash, status, priority,
               assigned_agent, created_at, updated_at, metadata
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
		var metadataStr string
		var createdAt, updatedAt time.Time
		var assignedAgent sql.NullString

		err := rows.Scan(
			&s.ID, &s.Name, &s.Kind, &s.File, &s.Language, &s.Signature, &s.Documentation,
			&s.Visibility, &s.Range.Start.Line, &s.Range.Start.Column, &s.Range.Start.Byte,
			&s.Range.End.Line, &s.Range.End.Column, &s.Range.End.Byte, &s.ContentHash, &s.Status, &s.Priority,
			&assignedAgent, &createdAt, &updatedAt, &metadataStr,
		)
		if err != nil {
			return nil, err
		}

		s.CreatedAt = createdAt
		s.UpdatedAt = updatedAt
		if assignedAgent.Valid {
			s.AssignedAgent = assignedAgent.String
		}

		if metadataStr != "" {
			err = json.Unmarshal([]byte(metadataStr), &s.Metadata)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}
		symbols = append(symbols, s)
	}

	return symbols, nil
}

func (m *Manager) GetAllFilesForProject(projectID int) ([]*model.File, error) {
	// TODO: Implement this properly
	return nil, nil
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
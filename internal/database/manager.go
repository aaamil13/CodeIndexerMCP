package database

import (
    "database/sql"
    "encoding/json"
    "time"
    _ "embed"
    
    "github.com/aaamil13/CodeIndexerMCP/internal/model"
    _ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

func applySchema(db *sql.DB) error {
    _, err := db.Exec(schemaSQL)
    return err
}

type Manager struct {
    db *sql.DB
}

func NewManager(dbPath string) (*Manager, error) {
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, err
    }
    
    // Прилагане на schema
    if err := applySchema(db); err != nil {
        return nil, err
    }
    
    return &Manager{db: db}, nil
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

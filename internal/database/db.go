package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver

	"github.com/aaamil13/CodeIndexerMCP/pkg/types"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
	path string
}

// Open opens or creates a database at the given path
func Open(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := ensureDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open database with pragmas for performance
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	// Run migrations
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	_, err := db.conn.Exec(Schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// Transaction executes a function within a transaction
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// ensureDir creates a directory if it doesn't exist
func ensureDir(dir string) error {
	// This will be implemented in utils
	return nil
}

// Helper functions to convert between types and JSON

func toJSON(v interface{}) (string, error) {
	if v == nil {
		return "{}", nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fromJSON(data string, v interface{}) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), v)
}

func nullInt64(n *int64) sql.NullInt64 {
	if n == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *n, Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// Ping checks database connectivity
func (db *DB) Ping() error {
	return db.conn.Ping()
}

// Stats returns database statistics
func (db *DB) Stats() (map[string]int, error) {
	stats := make(map[string]int)

	queries := map[string]string{
		"projects":      "SELECT COUNT(*) FROM projects",
		"files":         "SELECT COUNT(*) FROM files",
		"symbols":       "SELECT COUNT(*) FROM symbols",
		"imports":       "SELECT COUNT(*) FROM imports",
		"relationships": "SELECT COUNT(*) FROM relationships",
		"references":    "SELECT COUNT(*) FROM references",
	}

	for name, query := range queries {
		var count int
		if err := db.conn.QueryRow(query).Scan(&count); err != nil {
			return nil, err
		}
		stats[name] = count
	}

	return stats, nil
}

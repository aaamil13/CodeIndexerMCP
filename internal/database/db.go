package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver

	"github.com/aaamil13/CodeIndexerMCP/internal/utils" // Import utils for logging
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
	path string
	logger *utils.Logger // Added logger field
	writeQueue chan *dbOperation // Channel for write operations
	writerOnce sync.Once // Ensures writer goroutine starts once
	writerWg sync.WaitGroup // Waits for writer goroutine to finish
}

// dbOperation represents a database operation to be executed by the single writer
type dbOperation struct {
	query string
	args []interface{}
	resultChan chan dbResult
	isQueryRow bool // True if it's a QueryRow operation, false for Exec
}

// dbResult holds the result of a database operation
type dbResult struct {
	res sql.Result
	row *sql.Row
	err error
}

// Open opens or creates a database at the given path
func Open(dbPath string) (*DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := ensureDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// Open database with pragmas for performance and busy timeout
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", dbPath)
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
		logger: utils.NewLogger("[Database]"), // Initialize logger
		writeQueue: make(chan *dbOperation, 1000), // Buffered channel for operations
	}

	db.logger.Infof("Opened database: %s", dbPath)

	// Start the single writer goroutine
	db.startWriter()

	// Run migrations
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// startWriter starts the single goroutine responsible for executing all write operations
func (db *DB) startWriter() {
	db.writerOnce.Do(func() {
		db.writerWg.Add(1) // Increment WaitGroup
		go func() {
			defer db.writerWg.Done() // Decrement WaitGroup when goroutine exits
			for op := range db.writeQueue {
				if op.isQueryRow {
					op.resultChan <- dbResult{row: db.conn.QueryRow(op.query, op.args...)}
				} else {
					res, err := db.conn.Exec(op.query, op.args...)
					op.resultChan <- dbResult{res: res, err: err}
				}
			}
		}()
	})
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		db.logger.Info("Closing database connection")
		close(db.writeQueue) // Close the queue to signal writer to exit
		db.writerWg.Wait() // Wait for the writer goroutine to finish
		return db.conn.Close()
	}
	return nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	db.logger.Info("Running database migrations")
	// Migrations should not go through the queue as they are part of initialization
	_, err := db.conn.Exec(Schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// Transaction executes a function within a transaction
func (db *DB) Transaction(fn func(*sql.Tx) error) error {
	db.logger.Debug("Beginning database transaction")
	tx, err := db.conn.Begin()
	if err != nil {
		db.logger.Errorf("Failed to begin transaction: %v", err)
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			db.logger.Error("Transaction panicked, rolling back")
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		db.logger.Errorf("Transaction failed, rolling back: %v", err)
		return err
	}

	db.logger.Debug("Committing database transaction")
	return tx.Commit()
}

// enqueueWrite enqueues a write operation for the single writer goroutine
func (db *DB) enqueueWrite(query string, args ...interface{}) (sql.Result, error) {
	resultChan := make(chan dbResult, 1)
	op := &dbOperation{
		query: query,
		args: args,
		resultChan: resultChan,
		isQueryRow: false,
	}
	select {
	case db.writeQueue <- op:
		res := <-resultChan
		return res.res, res.err
	default:
		return nil, fmt.Errorf("database write queue is closed")
	}
}

// enqueueQueryRow enqueues a QueryRow operation for the single writer goroutine
func (db *DB) enqueueQueryRow(query string, args ...interface{}) *sql.Row {
	resultChan := make(chan dbResult, 1)
	op := &dbOperation{
		query: query,
		args: args,
		resultChan: resultChan,
		isQueryRow: true,
	}
	select {
	case db.writeQueue <- op:
		res := <-resultChan
		return res.row
	default:
		// Return a row with an error, similar to how QueryRow would behave on error
		// This is a bit tricky as sql.Row doesn't expose a way to set an error directly.
		// The caller will eventually get an error when calling Scan().
		// For now, we'll log and return an empty row.
		db.logger.Errorf("Attempted to enqueue query row to a closed database write queue")
		return db.conn.QueryRow("SELECT 0 WHERE 1=0") // Return an empty result set
	}
}


// ensureDir creates a directory if it doesn't exist
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
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
		"code_references":    "SELECT COUNT(*) FROM code_references",
	}

	for name, query := range queries {
		var count int
		// Stats queries can be run directly, they are reads
		if err := db.conn.QueryRow(query).Scan(&count); err != nil {
			return nil, err
		}
		stats[name] = count
	}

	return stats, nil
}

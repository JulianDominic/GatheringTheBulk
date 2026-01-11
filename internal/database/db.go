package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

var DB *sql.DB

// InitDB initializes the SQLite connection and ensures the schema is applied.
func InitDB(dsn string) error {
	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := DB.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}
	
	// Set busy timeout to avoid "database is locked" errors
	if _, err := DB.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		return fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	// Apply Schema
	if _, err := DB.Exec(schemaSQL); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	return nil
}

// Close closes the database connection.
func Close() {
	if DB != nil {
		DB.Close()
	}
}

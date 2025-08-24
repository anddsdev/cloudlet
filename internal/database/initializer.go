package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseInitializer struct {
	dsn string
}

func NewDatabaseInitializer(dsn string) *DatabaseInitializer {
	return &DatabaseInitializer{
		dsn: dsn,
	}
}

func (di *DatabaseInitializer) InitializeDatabase() error {
	dbExists, err := di.databaseExists()
	if err != nil {
		return fmt.Errorf("error checking database existence: %w", err)
	}

	if !dbExists {
		if err := di.createDatabaseDirectory(); err != nil {
			return fmt.Errorf("error creating database directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", di.dsn)
	if err != nil {
		return fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	if !dbExists {
		if err := di.createTables(db); err != nil {
			return fmt.Errorf("error creating tables: %w", err)
		}
	}

	if err := di.runMigrations(db); err != nil {
		return fmt.Errorf("error running migrations: %w", err)
	}

	return nil
}

func (di *DatabaseInitializer) databaseExists() (bool, error) {
	if di.dsn == ":memory:" {
		return false, nil
	}

	_, err := os.Stat(di.dsn)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (di *DatabaseInitializer) createDatabaseDirectory() error {
	if di.dsn == ":memory:" {
		return nil
	}

	dbDir := filepath.Dir(di.dsn)

	// Check if the directory already exists
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		// Create the directory with permissions 0755
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dbDir, err)
		}
		fmt.Printf("Created database directory: %s\n", dbDir)
	}

	return nil
}

func (di *DatabaseInitializer) createTables(db *sql.DB) error {
	fullSQL := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		path TEXT NOT NULL UNIQUE,
		size INTEGER NOT NULL DEFAULT 0,
		mime_type TEXT NOT NULL DEFAULT '',
		checksum TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		is_directory BOOLEAN NOT NULL DEFAULT FALSE,
		parent_path TEXT NOT NULL DEFAULT '/'
	);

	CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
	CREATE INDEX IF NOT EXISTS idx_files_parent_path ON files(parent_path);
	CREATE INDEX IF NOT EXISTS idx_files_name ON files(name);
	CREATE INDEX IF NOT EXISTS idx_files_is_directory ON files(is_directory);
	`

	_, err := db.Exec(fullSQL)
	if err != nil {
		return fmt.Errorf("error creating tables and indexes: %w", err)
	}

	fmt.Println("Database tables and indexes created successfully")
	return nil
}

func (di *DatabaseInitializer) runMigrations(db *sql.DB) error {
	createVersionsTable := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(createVersionsTable)
	if err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	migrations := []Migration{
		{
			Version: 1,
			SQL:     `-- Migration 1: Initial schema (already handled in createTables)`,
		},
		{
			Version: 2,
			SQL:     `ALTER TABLE files ADD COLUMN modified_at DATETIME; UPDATE files SET modified_at = updated_at WHERE modified_at IS NULL;`,
		},
		{
			Version: 3,
			SQL:     `ALTER TABLE files DROP COLUMN modified_at;`,
		},
	}

	// Run pending migrations
	for _, migration := range migrations {
		applied, err := di.isMigrationApplied(db, migration.Version)
		if err != nil {
			return fmt.Errorf("error checking migration %d: %w", migration.Version, err)
		}

		if !applied {
			if err := di.applyMigration(db, migration); err != nil {
				return fmt.Errorf("error applying migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

type Migration struct {
	Version int
	SQL     string
}

// Verify if a migration has been applied
func (di *DatabaseInitializer) isMigrationApplied(db *sql.DB, version int) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Apply a specific migration
func (di *DatabaseInitializer) applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if migration.SQL != "" && migration.SQL != `-- Migration 1: Initial schema (already handled in createTables)` {
		_, err = tx.Exec(migration.SQL)
		if err != nil {
			return err
		}
	}

	// Mark the migration as applied
	_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("Applied migration version %d\n", migration.Version)
	return nil
}

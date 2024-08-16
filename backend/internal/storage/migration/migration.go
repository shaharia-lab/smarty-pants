package migration

import (
	"database/sql"
	"fmt"
)

// MockMigrator interface defines methods for database migration
type Migrator interface {
	Migrate(migrations []Migration) error
	Rollback(migrations []Migration) error
	GetCurrentVersion() (string, error)
	EnsureMigrationTableExists() error
	AcquireMigrationLock() (bool, error)
	ReleaseMigrationLock() error
}

// Func MigrationFunc represents a function that performs a migration
type Func func(*sql.Tx) error

// Migration represents a single migration
type Migration struct {
	Version string
	Up      Func
	Down    Func
}

// DirtyMigrationError represents an error when a dirty migration is found
type DirtyMigrationError struct {
	Version string
	Message string
}

func (e *DirtyMigrationError) Error() string {
	return e.Message
}

// MigrationError represents an error that occurred during migration
type MigrationError struct {
	Version string
	Err     error
	Message string
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

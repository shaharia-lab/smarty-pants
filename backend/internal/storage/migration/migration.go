package migration

import "database/sql"

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

// Package migration handles the migration process
package migration

// Manager handles the migration process
type Manager struct {
	migrator   Migrator
	migrations []Migration
}

// NewMigrationManager creates a new MigrationManager
func NewMigrationManager(migrator Migrator, migrations []Migration) *Manager {
	return &Manager{migrator: migrator, migrations: migrations}
}

// RunMigrations runs all pending migrations
func (mm *Manager) RunMigrations() error {
	// Ensure the schema_migrations table exists
	err := mm.migrator.EnsureMigrationTableExists()
	if err != nil {
		return err
	}

	return mm.migrator.Migrate(mm.migrations)
}

// RollbackMigration rolls back the last migration
func (mm *Manager) RollbackMigration() error {
	return mm.migrator.Rollback(mm.migrations)
}

// GetCurrentVersion returns the current migration version
func (mm *Manager) GetCurrentVersion() (string, error) {
	return mm.migrator.GetCurrentVersion()
}

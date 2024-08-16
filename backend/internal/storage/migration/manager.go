// Package migration handles the migration process
package migration

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Manager handles the migration process
type Manager struct {
	migrator   Migrator
	migrations []Migration
	logger     *logrus.Logger
}

// NewMigrationManager creates a new MigrationManager
func NewMigrationManager(migrator Migrator, migrations []Migration, logger *logrus.Logger) *Manager {
	return &Manager{migrator: migrator, migrations: migrations, logger: logger}
}

// RunMigrations runs all pending migrations
// RunMigrations runs all pending migrations
func (mm *Manager) RunMigrations() error {
	mm.logger.Info("Running migrations. Ensuring migration table exists...")
	err := mm.migrator.EnsureMigrationTableExists()
	if err != nil {
		errMsg := "Failed to ensure migration table exists"
		mm.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	mm.logger.Info("Migration table ready. Attempting to acquire migration lock...")
	acquired, err := mm.migrator.AcquireMigrationLock()
	if err != nil {
		mm.logger.WithError(err).Error("Failed to acquire migration lock")
		return fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	if !acquired {
		mm.logger.Info("Another instance is running migrations, so we can exit")
		return nil
	}

	// Ensure the lock is released when we're done
	defer func() {
		if err := mm.migrator.ReleaseMigrationLock(); err != nil {
			mm.logger.WithError(err).Error("Failed to release migration lock")
		}
	}()

	mm.logger.Info("Migration lock acquired. Running migrations...")
	if err := mm.migrator.Migrate(mm.migrations); err != nil {
		mm.logger.WithError(err).Error("Failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	mm.logger.Info("Migrations completed successfully")
	return nil
}

// RollbackMigration rolls back the last migration
func (mm *Manager) RollbackMigration() error {
	return mm.migrator.Rollback(mm.migrations)
}

// GetCurrentVersion returns the current migration version
func (mm *Manager) GetCurrentVersion() (string, error) {
	return mm.migrator.GetCurrentVersion()
}

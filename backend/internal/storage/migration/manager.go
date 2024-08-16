package migration

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Manager handles the migration process
type Manager struct {
	migrator Migrator
	logger   *logrus.Logger
}

// NewMigrationManager creates a new MigrationManager
func NewMigrationManager(migrator Migrator, logger *logrus.Logger) *Manager {
	return &Manager{migrator: migrator, logger: logger}
}

// Run executes all pending migrations
func (mm *Manager) Run() error {
	mm.logger.Info("Starting migration process")

	if err := mm.ensureMigrationTable(); err != nil {
		return err
	}

	acquired, err := mm.acquireLock()
	if err != nil {
		return err
	}
	if !acquired {
		mm.logger.Info("Another instance is running migrations, exiting")
		return nil
	}

	defer mm.releaseLock()

	if err := mm.runMigrations(); err != nil {
		return err
	}

	mm.logger.Info("Migration process completed successfully")
	return nil
}

func (mm *Manager) ensureMigrationTable() error {
	mm.logger.Info("Ensuring migration table exists")
	err := mm.migrator.EnsureMigrationTableExists()
	if err != nil {
		mm.logger.WithError(err).Error("Failed to ensure migration table exists")
		return fmt.Errorf("failed to ensure migration table exists: %w", err)
	}
	mm.logger.Info("Migration table ready")
	return nil
}

func (mm *Manager) acquireLock() (bool, error) {
	mm.logger.Info("Attempting to acquire migration lock")
	acquired, err := mm.migrator.AcquireMigrationLock()
	if err != nil {
		mm.logger.WithError(err).Error("Failed to acquire migration lock")
		return false, fmt.Errorf("failed to acquire migration lock: %w", err)
	}
	if acquired {
		mm.logger.Info("Migration lock acquired")
	}
	return acquired, nil
}

func (mm *Manager) releaseLock() {
	mm.logger.Info("Releasing migration lock")
	if err := mm.migrator.ReleaseMigrationLock(); err != nil {
		mm.logger.WithError(err).Error("Failed to release migration lock")
	} else {
		mm.logger.Info("Migration lock released successfully")
	}
}

func (mm *Manager) runMigrations() error {
	mm.logger.Info("Running migrations")
	err := mm.migrator.Migrate(postgreSQLMigrations)
	if err != nil {
		return mm.handleMigrationError(err)
	}
	mm.logger.Info("All migrations applied successfully")
	return nil
}

func (mm *Manager) handleMigrationError(err error) error {
	switch specificErr := err.(type) {
	case *DirtyMigrationError:
		mm.logger.WithFields(logrus.Fields{
			"version": specificErr.Version,
			"message": specificErr.Message,
		}).Error("Dirty migration detected")
		return fmt.Errorf("dirty migration detected: %w", err)
	case *MigrationError:
		mm.logger.WithFields(logrus.Fields{
			"version": specificErr.Version,
			"message": specificErr.Message,
		}).Error("Migration failed")
		return fmt.Errorf("migration failed: %w", err)
	default:
		mm.logger.WithError(err).Error("Failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}
}

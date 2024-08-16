package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/sirupsen/logrus"
)

type Migration struct {
	db     *sql.DB
	logger *logrus.Logger
	m      *migrate.Migrate
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewMigration(db *sql.DB, logger *logrus.Logger) *Migration {
	return &Migration{
		db:     db,
		logger: logger,
	}
}

// Run executes all pending migrations
func (m *Migration) Run() error {
	if m.m == nil {
		migrator, err := m.newMigrator()
		if err != nil {
			return err
		}
		m.m = migrator
	}

	err := m.m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		errMsg := "failed to run migrations"
		m.logger.WithError(err).Error(errMsg)
		return fmt.Errorf("%s: %w", errMsg, err)
	}

	m.logger.Info("Migrations completed successfully")
	return nil
}

func (m *Migration) newMigrator() (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		errMsg := "failed to create driver instance"
		m.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		errMsg := "failed to create source instance"
		m.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	migrator, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		errMsg := "failed to create migrator instance"
		m.logger.WithError(err).Error(errMsg)
		return nil, fmt.Errorf("%s: %w", errMsg, err)
	}

	return migrator, nil
}

// ShutdownFn returns a function that can be registered with the shutdown manager
func (m *Migration) ShutdownFn(ctx context.Context) error {
	if m.m != nil {
		m.logger.Info("Gracefully stopping migration process")
		m.m.GracefulStop <- true
		select {
		case <-ctx.Done():
			m.logger.Warn("Migration shutdown timed out")
			return ctx.Err()
		case <-m.m.GracefulStop:
			m.logger.Info("Migration process stopped gracefully")
		}
	}
	return nil
}
